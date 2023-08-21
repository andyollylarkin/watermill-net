package connection_test

import (
	"bufio"
	"net"
	"testing"

	"github.com/andyollylarkin/watermill-net/pkg/connection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTCP4Connection(t *testing.T) {
	l, err := net.Listen("tcp", ":0")
	require.NoError(t, err)

	pConn := connection.NewTCPConnection(net.Dialer{})

	err = pConn.Connect(l.Addr())
	require.NoError(t, err)

	sListener := connection.NewTCP4Listener(l)
	sConn, err := sListener.Accept()
	require.NoError(t, err)

	_, err = pConn.Write([]byte("Hello\n"))
	require.NoError(t, err)

	var result string

	r := bufio.NewReader(sConn)
	s, _ := r.ReadString('\n')
	result = s

	assert.Equal(t, "Hello\n", result)
}
