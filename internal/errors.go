package internal

import (
	"errors"
	"os"
)

func IsTimeoutError(err error) bool {
	return errors.Is(err, os.ErrDeadlineExceeded)
}
