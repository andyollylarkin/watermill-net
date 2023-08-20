package watermillnet

import (
	"net"
)

type Connection interface {
	net.Conn
	// Establish connection with remote side.
	Connect(addr net.Addr) error
	// LocalAddr get local addr.
}
