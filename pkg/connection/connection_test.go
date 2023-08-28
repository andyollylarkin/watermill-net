package connection_test

import (
	"bufio"
	"testing"
	"time"

	"github.com/andyollylarkin/watermill-net/pkg/connection"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTCP4Connection(t *testing.T) {
	pConn := connection.NewTCPConnection(time.Second * 30)

	sListener, err := connection.NewTCP4Listener(":0")
	require.NoError(t, err)
	err = pConn.Connect(sListener.Addr())
	require.NoError(t, err)
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
