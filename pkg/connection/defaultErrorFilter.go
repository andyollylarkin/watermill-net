package connection

import (
	"errors"
	"net"
)

func DefaultErrorFilter(err error) error {
	var netError *net.OpError

	if errors.As(err, &netError) {
		return err
	}

	return nil
}
