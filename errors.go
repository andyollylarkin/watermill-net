package watermillnet

import (
	"errors"
	"fmt"
)

type InvalidConfigError struct {
	InvalidField  string
	InvalidReason string
}

func (ic *InvalidConfigError) Error() string {
	return fmt.Sprintf("invalid field: %s. reason: %s", ic.InvalidField, ic.InvalidReason)
}

var (
	ErrPublisherClosed      = errors.New("publisher closed")
	ErrSubscriberClosed     = errors.New("subscriber closed")
	ErrSubscriberNotStarted = errors.New("subscriber not started")
	ErrNacked               = errors.New("remote side sent nack for message")
	ErrIOTimeout            = errors.New("i/o timeout")
)
