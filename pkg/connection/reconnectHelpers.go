package connection

import "github.com/sethvargo/go-retry"

const (
	unlocked int64 = iota
	locked
)

// Backoff alias for retry.Backoff.
type Backoff retry.Backoff

// ErrorFilter match the errors at which you want to retry the request. If err does not match the error you want,
// return nil.
type ErrorFilter func(err error) error

func retryableErrorWrap(ef ErrorFilter, err error) error {
	if e := ef(err); e != nil {
		return retry.RetryableError(e)
	}

	return nil
}
