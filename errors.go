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
	ErrPublisherClosed = errors.New("publisher closed")
	ErrNacked          = errors.New("remote side sent nack for message")
)
