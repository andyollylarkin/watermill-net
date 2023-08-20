package internal_test

import (
	"testing"

	"github.com/andyollylarkin/watermill-net/internal"
	"github.com/stretchr/testify/require"
)

func TestPrepareMessageForSend(t *testing.T) {
	baseMsg := []byte{0x1, 0x2, 0x3, 0x4, 0x5}
	actual := internal.PrepareMessageForSend(baseMsg)
	expected := []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x05, 0xff, 0x01, 0x02, 0x03, 0x04, 0x05}

	msgLen := byte(0x05)
	actualLen := actual[7]

	require.Equal(t, expected, actual)
	require.Equal(t, msgLen, actualLen)
	require.Equal(t, internal.LenDelimiter, actual[8])
}

func TestReadLen(t *testing.T) {
	baseMsg := []byte{0x1, 0x2, 0x3, 0x4, 0x5}
	msg := internal.PrepareMessageForSend(baseMsg)
	expected := uint64(5)

	n := internal.ReadLen(msg)

	require.Equal(t, expected, n)
}
