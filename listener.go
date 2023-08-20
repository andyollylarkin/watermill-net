package watermillnet

import "net"

type Listener interface {
	// Accept waits for and returns the next connection to the listener.
	Accept() (Connection, error)
	// Close closes the listener.
	// Any blocked Accept operations will be unblocked and return errors.
	Close() error

	// Addr returns the listener's network address.
	Addr() net.Addr
}
