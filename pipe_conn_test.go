package watermillnet_test

import "net"

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

func (pc *PipeConnection) LocalAddr() string {
	return pc.LeftSide.LocalAddr().String()
}

func (pc *PipeConnection) RemoteAddr() string {
	return pc.LeftSide.RemoteAddr().String()
}

func (pc *PipeConnection) RemoteSideConn() net.Conn {
	return pc.RightSide
}

func (pc *PipeConnection) Read(p []byte) (n int, err error) {
	return pc.LeftSide.Read(p)
}

func (pc *PipeConnection) Write(p []byte) (n int, err error) {
	return pc.LeftSide.Write(p)
}

func (pc *PipeConnection) Close() error {
	return pc.LeftSide.Close()
}

// Establish connection with remote side.
func (pc *PipeConnection) Connect(addr net.Addr) error {
	return nil
}
