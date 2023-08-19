package watermillnet

import (
	"github.com/ThreeDotsLabs/watermill/message"
)

type Publisher struct {
}

// Publish publishes provided messages to given topic.
// Publish can be synchronous or asynchronous - it depends on the implementation.
//
// Most publishers implementations don't support atomic publishing of messages.
// This means that if publishing one of the messages fails, the next messages will not be published.
//
// Publish must be thread safe.
func (p *Publisher) Publish(topic string, messages ...*message.Message) error {
	panic("not implemented") // TODO: Implement
}

// Close should flush unsent messages, if publisher is async.
func (p *Publisher) Close() error {
	panic("not implemented") // TODO: Implement
}
