package connection

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	watermillnet "github.com/andyollylarkin/watermill-net"
	"github.com/andyollylarkin/watermill-net/internal"
	"github.com/sethvargo/go-retry"
)

func (rw *ReconnectWrapper) connectContextAdapter(ctx context.Context) error {
	connectCh := make(chan error, 0)
	go func() {
		err := rw.underlyingConnection.Connect(rw.remoteAddr)
		connectCh <- retryableErrorWrap(rw.errorFilter, err)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("Abort reconnection: %w", ctx.Err())
	case e := <-connectCh:
		if e != nil && rw.logger != nil {
			rw.logger.Info("Error when reconnect. Retry connect.", watermill.LogFields{"error": e.Error()})
		}

		return e
	}
}

// ReconnectWrapper enrich the underlying connection with a reconnect mechanism.
type ReconnectWrapper struct {
	underlyingConnection watermillnet.Connection
	backoffPolicy        Backoff
	logger               watermill.LoggerAdapter
	errorFilter          ErrorFilter
	remoteAddr           net.Addr
	readWriteTimeout     time.Duration
	lock                 int64 // lock conn wrapper for reconnect wait. 0 -> locked; 1 -> unlocked
	ctx                  context.Context
	mu                   sync.RWMutex
}

// NewReconnectWrapper enrich the underlying connection with a reconnect mechanism.
func NewReconnectWrapper(ctx context.Context, baseConn watermillnet.Connection, backoff Backoff,
	log watermill.LoggerAdapter, remoteAddr net.Addr, efilter ErrorFilter,
	readWriteTimeout time.Duration) *ReconnectWrapper { //nolint: gofumpt
	w := new(ReconnectWrapper)
	if backoff != nil {
		w.backoffPolicy = backoff
	} else {
		w.backoffPolicy = retry.NewExponential(time.Second * 1)
	}

	w.remoteAddr = remoteAddr
	w.underlyingConnection = baseConn
	w.logger = log
	w.readWriteTimeout = readWriteTimeout

	if ctx == nil {
		w.ctx = context.Background()
	} else {
		w.ctx = ctx
	}

	w.errorFilter = efilter

	return w
}

func (rw *ReconnectWrapper) reconnectLockWait() {
	for {
		val := atomic.LoadInt64(&rw.lock)
		if val == unlocked {
			break
		}
	}
}

func (rw *ReconnectWrapper) reconnectLock() {
	for {
		if atomic.CompareAndSwapInt64(&rw.lock, unlocked, locked) {
			break
		} else {
			continue
		}
	}
}

func (rw *ReconnectWrapper) reconnectUnlock() {
	for {
		if atomic.CompareAndSwapInt64(&rw.lock, locked, unlocked) {
			break
		} else {
			continue
		}
	}
}

func (rw *ReconnectWrapper) reconnect() error {
	rw.reconnectLock()
	defer rw.reconnectUnlock()

	err := retry.Do(rw.ctx, rw.backoffPolicy, rw.connectContextAdapter)
	if err != nil {
		return err
	}

	if rw.logger != nil {
		rw.logger.Info("Reconnect success.", nil)
	}

	return nil
}

// Read reads data from the connection.
// Read can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetReadDeadline.
func (rw *ReconnectWrapper) Read(b []byte) (n int, err error) { //nolint:dupl
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	rw.reconnectLockWait()

	// if a write/read error occurs, we will reconnect but lose the message we were trying to send/receive.
	// Therefore, after reconnection, the cycle will continue and there will be another attempt to send / read the message.
	// If this attempt is successful, then we will exit the loop.
	for {
		rw.underlyingConnection.SetReadDeadline(time.Now().Add(rw.readWriteTimeout))
		n, err = rw.underlyingConnection.Read(b)

		if err != nil { //nolint: nestif
			// dont reconnect when timeout happens
			if internal.IsTimeoutError(err) {
				if rw.logger != nil {
					rw.logger.Info("Timeout", watermill.LogFields{"op": "read"})
				}

				return n, watermillnet.ErrIOTimeout
			}

			if rw.logger != nil {
				rw.logger.Error("Unable to communicate with the remote side. attempt to reconnect", err,
					watermill.LogFields{"op": "read"})
			}

			err = rw.reconnect()

			if err != nil {
				return n, err
			}

			if rw.logger != nil {
				rw.logger.Debug("Reread message",
					watermill.LogFields{"op": "read"})
			}

			continue
		} else {
			break
		}
	}

	return n, err
}

// Write writes data to the connection.
// Write can be made to time out and return an error after a fixed
// time limit; see SetDeadline and SetWriteDeadline.
func (rw *ReconnectWrapper) Write(b []byte) (n int, err error) { //nolint:dupl
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	rw.reconnectLockWait()

	// if a write/read error occurs, we will reconnect but lose the message we were trying to send/receive.
	// Therefore, after reconnection, the cycle will continue and there will be another attempt to send / read the message.
	// If this attempt is successful, then we will exit the loop.
	for {
		rw.underlyingConnection.SetWriteDeadline(time.Now().Add(rw.readWriteTimeout))
		n, err = rw.underlyingConnection.Write(b)

		if err != nil { //nolint: nestif
			// dont reconnect when timeout happens
			if internal.IsTimeoutError(err) {
				if rw.logger != nil {
					rw.logger.Info("Timeout", watermill.LogFields{"op": "read"})
				}

				return n, watermillnet.ErrIOTimeout
			}

			if rw.logger != nil {
				rw.logger.Error("Unable to communicate with the remote side. attempt to reconnect", err,
					watermill.LogFields{"op": "write"})
			}

			err = rw.reconnect()

			if err != nil {
				return n, err
			}

			if rw.logger != nil {
				rw.logger.Debug("Resend message",
					watermill.LogFields{"op": "write"})
			}

			continue
		} else {
			break
		}
	}

	return n, err
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (rw *ReconnectWrapper) Close() error {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	return rw.underlyingConnection.Close()
}

// LocalAddr returns the local network address, if known.
func (rw *ReconnectWrapper) LocalAddr() net.Addr {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	return rw.underlyingConnection.LocalAddr()
}

// RemoteAddr returns the remote network address, if known.
func (rw *ReconnectWrapper) RemoteAddr() net.Addr {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	return rw.underlyingConnection.RemoteAddr()
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
func (rw *ReconnectWrapper) SetDeadline(t time.Time) error {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	return rw.underlyingConnection.SetDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (rw *ReconnectWrapper) SetReadDeadline(t time.Time) error {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	return rw.underlyingConnection.SetReadDeadline(t)
}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some of the data was successfully written.
// A zero value for t means Write will not time out.
func (rw *ReconnectWrapper) SetWriteDeadline(t time.Time) error {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	return rw.underlyingConnection.SetWriteDeadline(t)
}

// Establish connection with remote side.
// LocalAddr get local addr.
func (rw *ReconnectWrapper) Connect(addr net.Addr) error {
	rw.mu.RLock()
	defer rw.mu.RUnlock()

	rw.reconnectLockWait()
	err := rw.underlyingConnection.Connect(rw.remoteAddr)

	if err != nil {
		if rw.logger != nil {
			rw.logger.Error("Unable to communicate with the remote side. attempt to reconnect", err,
				watermill.LogFields{"op": "connect"})
		}

		err = rw.reconnect()

		if err != nil {
			return err
		}
	}

	return err
}
