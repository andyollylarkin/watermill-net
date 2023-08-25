package internal

import (
	"bufio"
	"bytes"
	"errors"
	"os"
	"time"
)

type TimeoutReader struct {
	bufioReader        bufio.Reader
	r                  ReadDeadliner
	extendReadDeadline time.Duration
}

// Return reader that reads from r. If i/o timeout exceeded reader continue read.
// If another error occurs reader return that error.
func NewTimeoutReader(r ReadDeadliner, extendReadDeadline time.Duration) *TimeoutReader {
	tr := new(TimeoutReader)
	tr.bufioReader = *bufio.NewReader(r)
	tr.r = r

	return tr
}

func (tr *TimeoutReader) ReadBytes(delim byte) ([]byte, error) {
	var outBuf bytes.Buffer

	for {
		readed, err := tr.bufioReader.ReadBytes(LenDelimiter)
		outBuf.Write(readed)

		if err != nil {
			if isTimeoutErr(err) && len(readed) > 0 {
				tr.r.SetReadDeadline(time.Now().Add(tr.extendReadDeadline))

				continue
			} else {
				return outBuf.Bytes(), err
			}
		} else {
			break
		}
	}

	return outBuf.Bytes(), nil
}

func isTimeoutErr(err error) bool {
	return errors.Is(err, os.ErrDeadlineExceeded)
}

func (tr *TimeoutReader) Read(p []byte) (int, error) {
	var outBuf bytes.Buffer

	for {
		tmpBuf := make([]byte, len(p))
		n, err := tr.bufioReader.Read(tmpBuf)
		outBuf.Write(tmpBuf)

		if err != nil {
			if isTimeoutErr(err) && n > 0 {
				tr.r.SetReadDeadline(time.Now().Add(tr.extendReadDeadline))

				continue
			} else {
				copy(p, outBuf.Bytes())

				return n, err
			}
		} else {
			break
		}
	}
	copy(p, outBuf.Bytes())

	return outBuf.Len(), nil
}

func cleanSlice(src []byte) {
	tmp := make([]byte, len(src))
	copy(src, tmp)
}
