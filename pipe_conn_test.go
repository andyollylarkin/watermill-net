package watermillnet_test

import (
	"net"
	"time"
)

type pipeAddr struct{}

func (pa pipeAddr) Network() string { return "pipe" }
func (pa pipeAddr) String() string  { return "pipe" }

type PipeConnection struct {
	LeftSide  net.Conn
	RightSide net.Conn
}

func NewPipeConnection() *PipeConnection {
	pc := new(PipeConnection)
	pc.LeftSide, pc.RightSide = net.Pipe()

	return pc
}

func (pc *PipeConnection) RemoteSideConn() net.Conn {
	return pc.RightSide
}

func (pc *PipeConnection) Read(b []byte) (n int, err error) {
	return pc.LeftSide.Read(b)
}

func (pc *PipeConnection) Write(b []byte) (n int, err error) {
	return pc.LeftSide.Write(b)
}

func (pc *PipeConnection) Close() error {
	return pc.LeftSide.Close()
}

func (pc *PipeConnection) LocalAddr() net.Addr {
	return pc.LeftSide.LocalAddr()
}

func (pc *PipeConnection) RemoteAddr() net.Addr {
	return pc.LeftSide.RemoteAddr()
}

func (pc *PipeConnection) SetDeadline(t time.Time) error {
	return pc.LeftSide.SetDeadline(t)
}

func (pc *PipeConnection) SetReadDeadline(t time.Time) error {
	return pc.LeftSide.SetReadDeadline(t)
}

func (pc *PipeConnection) SetWriteDeadline(t time.Time) error {
	return pc.LeftSide.SetWriteDeadline(t)
}

func (pc *PipeConnection) Connect(addr net.Addr) error {
	return nil
}
