package connection

import (
	"net"

	watermillnet "github.com/andyollylarkin/watermill-net"
)

type TCP4Listener struct {
	l    net.Listener
	conn watermillnet.Connection
}

func NewTCP4Listener(l net.Listener) *TCP4Listener {
	tl := new(TCP4Listener)
	tl.l = l

	return tl
}

// Accept waits for and returns the next connection to the listener.
func (tl *TCP4Listener) Accept() (watermillnet.Connection, error) {
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
	return tl.conn.Close()
}

// Addr returns the listener's network address.
func (tl *TCP4Listener) Addr() net.Addr {
	return tl.conn.LocalAddr()
}
