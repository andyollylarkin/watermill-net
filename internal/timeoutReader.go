package internal

import (
	"bytes"
	"net"
	"strings"
	"time"
)

type TimeoutReader struct {
	r                  ReadDeadliner
	extendReadDeadline time.Duration
}

// Return reader that reads from r. If i/o timeout exceeded reader continue read.
// If another error occurs reader return that error.
func NewTimeoutReader(r ReadDeadliner, extendReadDeadline time.Duration) *TimeoutReader {
	tr := new(TimeoutReader)
	tr.r = r
	tr.extendReadDeadline = extendReadDeadline

	return tr
}

func (tr *TimeoutReader) ReadBytes(delim byte) ([]byte, error) {
	var out bytes.Buffer

	var bufSize int = 1

	buf := make([]byte, bufSize)

	for {
		n, err := tr.r.Read(buf)

		// if has error and no read bytes
		if err != nil && n == 0 {
			return out.Bytes(), err
		}

		if err != nil && n > 0 {
			e, ok := err.(*net.OpError)
			if ok && strings.Contains(e.Error(), "i/o timeout") {
				tr.r.SetReadDeadline(time.Now().Add(tr.extendReadDeadline))
				out.Write(buf)
				cleanSlice(buf)

				continue
			} else {
				out.Write(buf)

				return out.Bytes(), e
			}
		}

		pos := bytes.IndexByte(buf, delim)

		if pos == -1 {
			out.Write(buf)
			cleanSlice(buf)

			continue
		} else {
			out.Write(buf[0 : pos+1])

			break
		}
	}

	return out.Bytes(), nil
}

func (tr *TimeoutReader) Read(p []byte) (int, error) {
	return tr.r.Read(p)
}

func cleanSlice(src []byte) {
	tmp := make([]byte, len(src))
	copy(src, tmp)
}
