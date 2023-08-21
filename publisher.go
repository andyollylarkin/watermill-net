package watermillnet

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/andyollylarkin/watermill-net/internal"
)

type PublisherConfig struct {
	Conn        Connection
	RemoteAddr  net.Addr
	Marshaler   Marshaler
	Unmarshaler Unmarshaler
	Logger      watermill.LoggerAdapter
}

type Publisher struct {
	conn        Connection
	marshaler   Marshaler
	unmarshaler Unmarshaler
	addr        net.Addr
	logger      watermill.LoggerAdapter
	closed      bool
	mu          sync.Mutex
}

func NewPublisher(config PublisherConfig) (*Publisher, error) {
	if err := validatePublisherConfig(config); err != nil {
		return nil, err
	}

	p := new(Publisher)
	p.conn = config.Conn
	p.addr = config.RemoteAddr
	p.marshaler = config.Marshaler
	p.unmarshaler = config.Unmarshaler
	p.logger = config.Logger

	return p, nil
}

func validatePublisherConfig(c PublisherConfig) error {
	if c.RemoteAddr == nil {
		return &InvalidConfigError{InvalidField: "Addr", InvalidReason: "cant be nil"}
	}

	if c.Conn == nil {
		return &InvalidConfigError{InvalidField: "Conn", InvalidReason: "cant be nil"}
	}

	if c.Marshaler == nil {
		return &InvalidConfigError{InvalidField: "Marshaler", InvalidReason: "cant be nil"}
	}

	if c.Unmarshaler == nil {
		return &InvalidConfigError{InvalidField: "Unmarshaler", InvalidReason: "cant be nil"}
	}

	return nil
}

// Connect to remote side.
func (p *Publisher) Connect() error {
	if p.closed {
		return ErrPublisherClosed
	}

	return p.conn.Connect(p.addr)
}

// Publish publishes provided messages to given topic.
// Publish can be synchronous or asynchronous - it depends on the implementation.
//
// Most publishers implementations don't support atomic publishing of messages.
// This means that if publishing one of the messages fails, the next messages will not be published.
//
// Publish must be thread safe.
func (p *Publisher) Publish(topic string, messages ...*message.Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return ErrPublisherClosed
	}

	for _, msg := range messages {
		m := internal.Message{
			Topic:   topic,
			Message: msg,
		}

		b, err := p.marshaler.MarshalMessage(m)
		if err != nil {
			return err
		}

		b = internal.PrepareMessageForSend(b)
		_, err = p.conn.Write(b)

		if err != nil {
			return err
		}

		if err = p.handleResponse(); err != nil { // wait ack or nack
			return err
		}

		if p.logger != nil {
			fields := watermill.LogFields{
				"uuid":  msg.UUID,
				"topic": topic,
			}
			p.logger.Trace("Message published", fields)
		}
	}

	return nil
}

func (p *Publisher) handleResponse() error {
	r := bufio.NewReader(p.conn)
	lenRaw, err := r.ReadBytes(internal.LenDelimiter)

	if err != nil {
		return err
	}

	readLen := internal.ReadLen(lenRaw[:len(lenRaw)-1]) // trim len delimiter
	lr := io.LimitReader(r, int64(readLen))

	respBody := make([]byte, readLen)

	_, err = lr.Read(respBody)
	if err != nil {
		return fmt.Errorf("error read ack message %w", err)
	}

	var ackMsg internal.AckMessage

	err = p.unmarshaler.UnmarshalMessage(respBody, &ackMsg)
	if err != nil {
		return err
	}

	if !ackMsg.Acked {
		// TODO: handle nack for resend?
		return fmt.Errorf("%w: %s", ErrNacked, ackMsg.UUID)
	}

	return nil
}

// Close should flush unsent messages, if publisher is async.
func (p *Publisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.closed = true

	return p.conn.Close()
}
