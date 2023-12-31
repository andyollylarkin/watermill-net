package watermillnet

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/andyollylarkin/watermill-net/internal"
)

const maxBodyLen = 20 << 30

type sub struct {
	Topic   string
	MsgChan chan *message.Message // client message chan
	ReadCh  chan []byte
	Closed  bool
}

type SubscriberConfig struct {
	Marshaler   Marshaler
	Unmarshaler Unmarshaler
	Logger      watermill.LoggerAdapter
}

type Subscriber struct {
	conn        Connection
	closed      bool
	mu          sync.RWMutex
	subscribers []*sub
	processWg   sync.WaitGroup
	marshaler   Marshaler
	unmarshaler Unmarshaler
	logger      watermill.LoggerAdapter
	done        chan struct{}
	started     bool
}

// NewSubscriber create new subscriber.
// ATTENTION! Set connection immediately after creation.
func NewSubscriber(config SubscriberConfig) (*Subscriber, error) {
	if err := validateSubscriberConfig(config); err != nil {
		return nil, err
	}

	s := new(Subscriber)
	s.closed = false
	s.subscribers = make([]*sub, 0)
	s.marshaler = config.Marshaler
	s.unmarshaler = config.Unmarshaler
	s.logger = config.Logger
	s.done = make(chan struct{})
	s.started = false

	return s, nil
}

func validateSubscriberConfig(c SubscriberConfig) error {
	if c.Marshaler == nil {
		return &InvalidConfigError{InvalidField: "Marshaler", InvalidReason: "cant be nil"}
	}

	if c.Unmarshaler == nil {
		return &InvalidConfigError{InvalidField: "Unmarshaler", InvalidReason: "cant be nil"}
	}

	return nil
}

func (s *Subscriber) Addr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.conn.LocalAddr().String()
}

// Subscribe returns output channel with messages from provided topic.
// Channel is closed, when Close() was called on the subscriber.
//
// To receive the next message, `Ack()` must be called on the received message.
// If message processing failed and message should be redelivered `Nack()` should be called.
//
// When provided ctx is cancelled, subscriber will close subscribe and close output channel.
// Provided ctx is set to all produced messages.
// When Nack or Ack is called on the message, context of the message is canceled.
// will wait for reconnects and will not exit the read loop when the connection is lost.
// Since it is impossible to understand whether the remote side will reconnect, this is a mandatory mechanism.
func (s *Subscriber) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn == nil {
		return nil, ErrConnectionNotSet
	}

	if s.closed {
		return nil, ErrSubscriberClosed
	}

	if !s.started {
		return nil, ErrSubscriberNotStarted
	}

	outCh := make(chan *message.Message, 0)
	readCh := make(chan []byte)

	sub := &sub{
		Topic:   topic,
		MsgChan: outCh,
		ReadCh:  readCh,
		Closed:  false,
	}
	s.subscribers = append(s.subscribers, sub)

	s.processWg.Add(1)
	go s.handle(ctx, readCh, sub)

	return outCh, nil
}

func (s *Subscriber) SetConnection(c Connection) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.started = true

	s.conn = c
}

// GetConnection get publisher connection.
func (s *Subscriber) GetConnection() (Connection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn == nil {
		return nil, ErrConnectionNotSet
	}

	if s.closed {
		return nil, ErrPublisherClosed
	}

	return s.conn, nil
}

// Connect establish connection.
// If connection was set via SetConnection, we should pass nil to l.
// If listener was set, subscriber will be waiting until the client reconnects when the connection is lost.
func (s *Subscriber) Connect(l Listener) error {
	if l == nil && s.conn == nil {
		return fmt.Errorf("listener cant be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrSubscriberClosed
	}

	// if connection was set, dont accept new connection, use existed instead.
	if s.conn == nil {
		conn, err := l.Accept()
		if err != nil {
			return err
		}

		s.conn = conn
	}

	s.started = true

	go s.readContent()

	return nil
}

func (s *Subscriber) handle(ctx context.Context, readCh <-chan []byte, sub *sub) { //nolint: gocognit,funlen
	defer s.processWg.Done()

	for {
		select {
		case <-ctx.Done():
			if s.logger != nil {
				s.logger.Error("Done consume.", ctx.Err(), nil)
			}

			s.mu.Lock()
			sub.Closed = true
			close(sub.MsgChan)
			s.mu.Unlock()

			return
		case <-s.done:
			return
		case msgBody, ok := <-readCh:
			if !ok {
				if s.logger != nil {
					s.logger.Debug("Read channel closed. Done consume.", nil)
				}

				return
			}

			s.mu.RLock()

			var msg internal.Message

			err := s.unmarshaler.UnmarshalMessage(msgBody, &msg)
			if err != nil {
				if s.logger != nil {
					s.logger.Error("Error unmarshal incoming message", err, nil)
				}
				s.mu.RUnlock()

				continue
			}

			if msg.Message == nil {
				s.mu.RUnlock()

				continue
			}

			// create new watermill message bacause after marshal/unmarshal ack/nack channels is nil
			watermillMsg := message.NewMessage(msg.Message.UUID, msg.Message.Payload)

			if msg.Topic != sub.Topic {
				s.mu.RUnlock()

				continue
			}

			if !sub.Closed {
				// send message to sub chan and wait ack or nack
				sub.MsgChan <- watermillMsg
				select {
				case <-watermillMsg.Acked():
					err = s.sendAck(true, watermillMsg.UUID)

					if err != nil {
						s.mu.RUnlock()

						continue
					}
				case <-watermillMsg.Nacked():
					err = s.sendAck(false, watermillMsg.UUID)
					if err != nil {
						s.mu.RUnlock()

						continue
					}
				}
			} else {
				return
			}

			s.mu.RUnlock()
		}
	}
}

// sendAck send acknowledge message to remote side.
// if ack = true -> ack, if ack = false -> nack.
func (s *Subscriber) sendAck(ack bool, uuid string) error {
	ackMsg := internal.AckMessage{
		UUID:  uuid,
		Acked: ack,
	}

	marshaledAck, err := s.marshaler.MarshalMessage(ackMsg)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("Error marshal ack message", err, nil)
		}

		return err
	}

	marshaledAck = internal.PrepareMessageForSend(marshaledAck)

	_, err = s.conn.Write(marshaledAck)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("Error marshal ack message", err, nil)
		}

		return err
	}

	return nil
}

func (s *Subscriber) readContent() {
	var readTimeout = time.Second * 5

	for {
		select {
		case <-s.done:
			return
		default:
			s.mu.RLock()
			s.conn.SetReadDeadline(time.Now().Add(readTimeout))
			r := internal.NewTimeoutReader(s.conn, readTimeout)

			lenRaw, err := r.ReadBytes(internal.LenDelimiter)
			if err != nil {
				if s.logger != nil && !errors.Is(err, io.EOF) {
					s.logger.Error("Error read message", err, nil)
				}

				s.mu.RUnlock()

				continue
			}

			readLen := internal.ReadLen(lenRaw[:len(lenRaw)-1]) // trim len delimiter
			if readLen <= 0 || readLen > maxBodyLen {
				if s.logger != nil {
					s.logger.Info("Message body out of range", watermill.LogFields{"len": readLen})
				}

				continue
			}

			lr := io.LimitReader(r, int64(readLen))

			respBody := make([]byte, readLen)

			_, err = lr.Read(respBody)
			if err != nil {
				if s.logger != nil {
					s.logger.Error("Error read message", err, nil)
				}

				s.mu.RUnlock()

				continue
			}

			for _, sub := range s.subscribers {
				sub.ReadCh <- respBody
			}
			s.mu.RUnlock()
		}
	}
}

// Close closes all subscriptions with their output channels and flush offsets etc. when needed.
func (s *Subscriber) Close() error {
	s.mu.Lock()

	if s.conn == nil {
		s.mu.Unlock()

		return ErrConnectionNotSet
	}

	for _, v := range s.subscribers {
		if !v.Closed {
			v.Closed = true
			close(v.MsgChan)
		}
	}

	s.closed = true
	close(s.done)

	s.mu.Unlock()

	s.processWg.Wait()

	s.mu.Lock()
	err := s.conn.Close()
	s.mu.Unlock()

	return err
}
