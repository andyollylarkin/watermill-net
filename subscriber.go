package watermillnet

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/andyollylarkin/watermill-net/internal"
)

type sub struct {
	Topic   string
	MsgChan chan *message.Message // client message chan
}

type Subscriber struct {
	conn                 Connection
	closed               bool
	addr                 net.Addr
	mu                   sync.RWMutex
	subscribers          []*sub
	subscribersReadChans []chan []byte
	processWg            sync.WaitGroup
	marshaler            Marshaler
	unmarshaler          Unmarshaler
	logger               watermill.LoggerAdapter
	done                 chan struct{}
	started              bool
}

func (s *Subscriber) findSubscribersByTopic(topic string) []*sub {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]*sub, 0)

	for _, s := range s.subscribers {
		if s.Topic == topic {
			out = append(out, s)
		}
	}

	return out
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
	}
	s.subscribers = append(s.subscribers, sub)
	s.subscribersReadChans = append(s.subscribersReadChans, readCh)

	go s.handle(ctx, readCh)

	return outCh, nil
}

// Connect establish connection.
// Pass listener for accept new connection or pass existed connection. Not both!
func (s *Subscriber) Connect(ctx context.Context, l Listener, conn Connection) error {
	if l != nil && conn != nil {
		return fmt.Errorf("pass listener or conn, not both")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return ErrSubscriberClosed
	}

	if l != nil {
		c, err := l.Accept()
		if err != nil {
			return err
		}

		s.conn = c
		s.started = true
	} else {
		s.conn = conn
		s.started = true
	}

	go s.readContent(ctx)

	return nil
}

func (s *Subscriber) handle(ctx context.Context, readCh <-chan []byte) { //nolint: gocognit
	s.processWg.Add(1)

	go func() {
		defer s.processWg.Done()

		for {
			select {
			case <-ctx.Done():
				if s.logger != nil {
					s.logger.Error("Read channel closed. Done consume.", ctx.Err(), nil)
				}

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

				var msg internal.Message

				err := s.unmarshaler.UnmarshalMessage(msgBody, &msg)
				if err != nil {
					if s.logger != nil {
						s.logger.Error("Error unmarshal incoming message", err, nil)
					}

					continue
				}

				subs := s.findSubscribersByTopic(msg.Topic)

				// send message to sub chan and wait ack or nack
				for _, sub := range subs {
					sub.MsgChan <- msg.Message
					select {
					case <-msg.Message.Acked():
						err = s.sendAck(true, msg.Message.UUID)
						if err != nil {
							continue
						}
					case <-msg.Message.Nacked():
						err = s.sendAck(false, msg.Message.UUID)
						if err != nil {
							continue
						}
					}
				}
			}
		}
	}()
}

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

	_, err = s.conn.Write(marshaledAck)
	if err != nil {
		if s.logger != nil {
			s.logger.Error("Error marshal ack message", err, nil)
		}

		return err
	}

	return nil
}

func (s *Subscriber) readContent(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.done:
				return
			default:
				// TODO: non blocking read
				s.mu.RLock()
				r := bufio.NewReader(s.conn)

				lenRaw, err := r.ReadBytes(internal.LenDelimiter)
				if err != nil {
					s.mu.RUnlock()

					continue
				}

				readLen := internal.ReadLen(lenRaw[:len(lenRaw)-1]) // trim len delimiter
				lr := io.LimitReader(r, int64(readLen))

				respBody := make([]byte, readLen)

				_, err = lr.Read(respBody)
				if err != nil {
					s.mu.RUnlock()

					continue
				}

				for _, rCh := range s.subscribersReadChans {
					rCh <- respBody
				}
				s.mu.RUnlock()
			}
		}
	}()
}

// Close closes all subscriptions with their output channels and flush offsets etc. when needed.
func (s *Subscriber) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, v := range s.subscribers {
		close(v.MsgChan)
	}

	s.closed = true
	close(s.done)

	s.processWg.Wait()

	return s.conn.Close()
}
