package watermillnet_test

import (
	"net"
	"time"

	watermillnet "github.com/andyollylarkin/watermill-net"
)

type listenerStub struct {
	Conn watermillnet.Connection
}

func NewListenerStub(conn net.Conn) *listenerStub {
	s := new(listenerStub)
	s.Conn = NewConnWrapper(conn)

	return s
}

func NewListenerStubWithFakeConn(conn watermillnet.Connection) *listenerStub {
	s := new(listenerStub)
	s.Conn = conn

	return s
}

func (ls *listenerStub) Accept() (watermillnet.Connection, error) {
	return ls.Conn, nil
}
func (ls *listenerStub) Close() error   { return ls.Conn.Close() }
func (ls *listenerStub) Addr() net.Addr { return ls.Conn.LocalAddr() }

type connWrapper struct {
	C      net.Conn
	Closed bool
}

func NewConnWrapper(conn net.Conn) *connWrapper {
	cw := new(connWrapper)
	cw.C = conn

	return cw
}

func (pc *connWrapper) Read(b []byte) (n int, err error) {
	return pc.C.Read(b)
}

func (pc *connWrapper) Write(b []byte) (n int, err error) {
	return pc.C.Write(b)
}

func (pc *connWrapper) Close() error {
	pc.Closed = true
	return pc.C.Close()
}

func (pc *connWrapper) LocalAddr() net.Addr {
	return pc.C.LocalAddr()
}

func (pc *connWrapper) RemoteAddr() net.Addr {
	return pc.C.RemoteAddr()
}

func (pc *connWrapper) SetDeadline(t time.Time) error {
	return pc.C.SetDeadline(t)
}

func (pc *connWrapper) SetReadDeadline(t time.Time) error {
	return pc.C.SetReadDeadline(t)
}

func (pc *connWrapper) SetWriteDeadline(t time.Time) error {
	return pc.C.SetWriteDeadline(t)
}

func (pc *connWrapper) Connect(addr net.Addr) error {
	return nil
}
