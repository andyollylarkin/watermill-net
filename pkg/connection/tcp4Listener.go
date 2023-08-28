package connection

import (
	"net"
	"sync"

	watermillnet "github.com/andyollylarkin/watermill-net"
)

type TCP4Listener struct {
	l    net.Listener
	conn watermillnet.Connection
	mu   sync.Mutex
}

func NewTCP4Listener(addr string) (*TCP4Listener, error) {
	tl := new(TCP4Listener)

	l, err := net.Listen("tcp4", addr)
	if err != nil {
		return nil, err
	}

	tl.l = l

	return tl, nil
}

// Accept waits for and returns the next connection to the listener.
func (tl *TCP4Listener) Accept() (watermillnet.Connection, error) {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	conn, err := tl.l.Accept()

	if err != nil {
		return nil, err
	}

	t4c := &TCP4Connection{
		underlyingConn: conn,
	}

	tl.conn = t4c

	return t4c, nil
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (tl *TCP4Listener) Close() error {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	return tl.l.Close()
}

// Addr returns the listener's network address.
func (tl *TCP4Listener) Addr() net.Addr {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	return tl.l.Addr()
}
