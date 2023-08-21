package internal

import "time"

type ReadDeadliner interface {
	Read(p []byte) (n int, err error)
	SetReadDeadline(t time.Time) error
}
