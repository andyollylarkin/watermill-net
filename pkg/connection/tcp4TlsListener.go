package connection

import (
	"crypto/tls"
	"net"
	"sync"

	watermillnet "github.com/andyollylarkin/watermill-net"
)

type TCP4TlsListener struct {
	l    net.Listener
	conn watermillnet.Connection
	mu   sync.Mutex
}

func NewTCP4TlsListener(addr string, config *tls.Config) (*TCP4Listener, error) {
	tl := new(TCP4Listener)

	l, err := tls.Listen("tcp4", addr, config)
	if err != nil {
		return nil, err
	}

	tl.l = l

	return tl, nil
}

// Accept waits for and returns the next connection to the listener.
func (tl *TCP4TlsListener) Accept() (watermillnet.Connection, error) {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	conn, err := tl.l.Accept()

	if err != nil {
		return nil, err
	}

	t4c := &TCP4TlsConnection{
		underlyingConn: conn,
	}

	tl.conn = t4c

	return t4c, nil
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (tl *TCP4TlsListener) Close() error {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	return tl.l.Close()
}

// Addr returns the listener's network address.
func (tl *TCP4TlsListener) Addr() net.Addr {
	tl.mu.Lock()
	defer tl.mu.Unlock()

	return tl.l.Addr()
}
