package watermillnet

import (
	"context"
	"net"

	"github.com/ThreeDotsLabs/watermill/message"
)

type Subscriber struct {
	listener net.Listener
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
	panic("not implemented") // TODO: Implement
}

// Close closes all subscriptions with their output channels and flush offsets etc. when needed.
func (s *Subscriber) Close() error {
	panic("not implemented") // TODO: Implement
}
