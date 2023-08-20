package watermillnet

import (
	"io"
	"net"
)

type Connection interface {
	io.ReadWriteCloser
	// Establish connection with remote side.
	Connect(addr net.Addr) error
	// LocalAddr get local addr.
	LocalAddr() string
	// RemoteAddr get remote side addr.
	RemoteAddr() string
}
