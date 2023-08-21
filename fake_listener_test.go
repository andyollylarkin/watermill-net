package watermillnet_test

import (
	"net"
	"time"

	watermillnet "github.com/andyollylarkin/watermill-net"
)

type listenerStub struct{}

func (ls *listenerStub) Accept() (watermillnet.Connection, error) { return nil, nil }
func (ls *listenerStub) Close() error                             { return nil }
func (ls *listenerStub) Addr() net.Addr                           { return nil }

type listenerStub2 struct {
	conn net.Conn
}

type connWrapper struct {
	c net.Conn
}

func NewConnWrapper(conn net.Conn) *connWrapper {
	cw := new(connWrapper)
	cw.c = conn

	return cw
}

func (pc *connWrapper) Read(b []byte) (n int, err error) {
	return pc.c.Read(b)
}

func (pc *connWrapper) Write(b []byte) (n int, err error) {
	return pc.c.Write(b)
}

func (pc *connWrapper) Close() error {
	return pc.c.Close()
}

func (pc *connWrapper) LocalAddr() net.Addr {
	return pc.c.LocalAddr()
}

func (pc *connWrapper) RemoteAddr() net.Addr {
	return pc.c.RemoteAddr()
}

func (pc *connWrapper) SetDeadline(t time.Time) error {
	return pc.c.SetDeadline(t)
}

func (pc *connWrapper) SetReadDeadline(t time.Time) error {
	return pc.c.SetReadDeadline(t)
}

func (pc *connWrapper) SetWriteDeadline(t time.Time) error {
	return pc.c.SetWriteDeadline(t)
}

func (pc *connWrapper) Connect(addr net.Addr) error {
	return nil
}
