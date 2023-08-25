package internal_test

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/andyollylarkin/watermill-net/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const delimiter byte = 0xff

type BytesBufferDeadliner struct {
	bb bytes.Buffer
}

func (bbd *BytesBufferDeadliner) Write(p []byte) (n int, err error) {
	return bbd.bb.Write(p)
}

func (bbd *BytesBufferDeadliner) Read(p []byte) (n int, err error) {
	return bbd.bb.Read(p)
}

func (bbd *BytesBufferDeadliner) SetReadDeadline(t time.Time) error {
	return nil
}

func TestReaderDelimiterOK(t *testing.T) {
	bb := BytesBufferDeadliner{bb: *bytes.NewBuffer([]byte{})}

	bb.Write(append([]byte("Hello world"), delimiter))

	r := internal.NewTimeoutReader(&bb, time.Duration(0))

	content, err := r.ReadBytes(delimiter)
	require.NoError(t, err)

	assert.Equal(t, "Hello world", string(content[:len(content)-1]))
}

func TestReaderLargeBuffer(t *testing.T) {
	bb := BytesBufferDeadliner{bb: *bytes.NewBuffer([]byte{})}

	bb.Write(append([]byte("HelloHelloHelloHelloHelloHelloHelloHello"), delimiter)) // reader read buf = 32 bytes

	r := internal.NewTimeoutReader(&bb, time.Duration(0))

	content, err := r.ReadBytes(delimiter)
	require.NoError(t, err)

	assert.Equal(t, "HelloHelloHelloHelloHelloHelloHelloHello", string(content[:len(content)-1]))
}

func TestReaderEOF(t *testing.T) {
	bb := BytesBufferDeadliner{bb: *bytes.NewBuffer([]byte{})}

	bb.Write([]byte("HelloHelloHelloHelloHelloHelloHelloHello")) // reader read buf = 32 bytes

	r := internal.NewTimeoutReader(&bb, time.Duration(0))

	_, err := r.ReadBytes(delimiter)
	require.ErrorIs(t, err, io.EOF)
}

func TestReaderErrorTimeoutZeroRead(t *testing.T) {
	tr := TestTimeoutReader{
		buf:      *bytes.NewBuffer([]byte{}),
		readZero: true,
	}

	r := internal.NewTimeoutReader(&tr, time.Duration(0))

	_, err := r.ReadBytes(delimiter)
	require.Errorf(t, err, "i/o timeout")
}

func TestReaderErrorTimeoutButFullRead(t *testing.T) {
	var bb bytes.Buffer

	bb.Write(append([]byte("Hello"), delimiter))

	tr := TestTimeoutReader{
		buf:      bb,
		readZero: false,
	}

	r := internal.NewTimeoutReader(&tr, time.Duration(0))

	content, err := r.ReadBytes(delimiter)
	require.NoError(t, err)
	assert.Equal(t, "Hello", string(content[:len(content)-1]))
}
