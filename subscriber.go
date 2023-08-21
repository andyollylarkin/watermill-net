package watermillnet

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/andyollylarkin/watermill-net/internal"
)

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

// Connect establish connection.
func (s *Subscriber) Connect(l Listener) error {
	if l == nil {
		return fmt.Errorf("listener cant be nil")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrSubscriberClosed
	}

	conn, err := l.Accept()
	if err != nil {
		return err
	}

	go s.reconnect(l)

	s.conn = conn
	s.started = true

	go s.readContent()

	return nil
}

// reconnect wait new connection and replace existed connection.
func (s *Subscriber) reconnect(l Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			return
		}

		s.mu.Lock()
		s.conn = conn
		s.mu.Unlock()
	}
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
	for {
		select {
		case <-s.done:
			return
		default:
			// TODO: non blocking read
			r := bufio.NewReader(s.conn)

			lenRaw, err := r.ReadBytes(internal.LenDelimiter)

			s.mu.RLock()

			if err != nil {
				if s.logger != nil {
					s.logger.Error("Error read message", err, nil)
				}

				s.mu.RUnlock()

				continue
			}

			readLen := internal.ReadLen(lenRaw[:len(lenRaw)-1]) // trim len delimiter
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

	if s.conn != nil {
		return s.conn.Close()
	}

	return nil
}
