package internal_test

import (
	"bytes"
	"errors"
	"net"
	"time"
)

type TestTimeoutReader struct {
	buf      bytes.Buffer
	readZero bool
}

func (tr *TestTimeoutReader) Read(p []byte) (n int, err error) {
	if tr.readZero {
		return tr.readZeroBytesTimeout()
	} else {
		return tr.readBytesPortion(p)
	}
}

func (tr *TestTimeoutReader) readZeroBytesTimeout() (n int, err error) {
	return 0, &net.OpError{Err: errors.New("i/o timeout")}
}

func (tr *TestTimeoutReader) readBytesPortion(p []byte) (n int, err error) {
	n, err = tr.buf.Read(p)
	if err != nil {
		return n, err
	}

	return n, &net.OpError{Err: errors.New("i/o timeout")}
}

func (tr *TestTimeoutReader) Write(p []byte) (n int, err error) {
	return tr.buf.Write(p)
}

func (bbd *TestTimeoutReader) SetReadDeadline(t time.Time) error {
	return nil
}
