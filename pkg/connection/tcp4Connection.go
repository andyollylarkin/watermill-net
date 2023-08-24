package connection

import (
	"net"
	"sync"
	"time"
)

type TCP4Connection struct {
	dialer         net.Dialer
	underlyingConn net.Conn
	mu             sync.RWMutex
	Closed         bool
}

func NewTCPConnection(dialer net.Dialer) *TCP4Connection {
	c := new(TCP4Connection)
	c.dialer = dialer

	return c
}

// Read reads data from the connection.
// Read can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetReadDeadline.
func (tc *TCP4Connection) Read(b []byte) (n int, err error) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	if tc.underlyingConn == nil {
		return 0, ErrNotConnected
	}

	return tc.underlyingConn.Read(b)
}

// Write writes data to the connection.
// Write can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetWriteDeadline.
func (tc *TCP4Connection) Write(b []byte) (n int, err error) {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	if tc.underlyingConn == nil {
		return 0, ErrNotConnected
	}

	return tc.underlyingConn.Write(b)
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (tc *TCP4Connection) Close() error {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	if tc.underlyingConn == nil {
		return ErrNotConnected
	}

	tc.Closed = true

	return tc.underlyingConn.Close()
}

// LocalAddr returns the local network address, if known.
func (tc *TCP4Connection) LocalAddr() net.Addr {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	if tc.underlyingConn == nil {
		return nil
	}

	return tc.underlyingConn.LocalAddr()
}

// RemoteAddr returns the remote network address, if known.
func (tc *TCP4Connection) RemoteAddr() net.Addr {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	if tc.underlyingConn == nil {
		return nil
	}

	return tc.underlyingConn.RemoteAddr()
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail instead of blocking. The deadline applies to all future
// and pending I/O, not just the immediately following call to
// Read or Write. After a deadline has been exceeded, the
// connection can be refreshed by setting a deadline in the future.
//
// If the deadline is exceeded a call to Read or Write or to other
// I/O methods will return an error that wraps os.ErrDeadlineExceeded.
// This can be tested using errors.Is(err, os.ErrDeadlineExceeded).
// The error's Timeout method will return true, but note that there
// are other possible errors for which the Timeout method will
// return true even if the deadline has not been exceeded.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (tc *TCP4Connection) SetDeadline(t time.Time) error {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	if tc.underlyingConn == nil {
		return ErrNotConnected
	}

	return tc.underlyingConn.SetDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (tc *TCP4Connection) SetReadDeadline(t time.Time) error {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	if tc.underlyingConn == nil {
		return ErrNotConnected
	}

	return tc.underlyingConn.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (tc *TCP4Connection) SetWriteDeadline(t time.Time) error {
	tc.mu.RLock()
	defer tc.mu.RUnlock()

	if tc.underlyingConn == nil {
		return ErrNotConnected
	}

	return tc.underlyingConn.SetWriteDeadline(t)
}

// Establish connection with remote side.
// LocalAddr get local addr.
func (tc *TCP4Connection) Connect(addr net.Addr) error {
	tc.mu.Lock()
	defer tc.mu.Unlock()

	conn, err := tc.dialer.Dial("tcp4", addr.String())
	if err != nil {
		return err
	}

	tc.underlyingConn = conn

	return nil
}
