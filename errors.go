package watermillnet

import (
	"errors"
	"fmt"
	"os"
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
	ErrIOTimeout            = os.ErrDeadlineExceeded
	ErrConnectionNotSet     = errors.New("connection not set")
)
