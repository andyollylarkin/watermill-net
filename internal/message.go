package internal

import (
	"encoding/binary"

	"github.com/ThreeDotsLabs/watermill/message"
)

const (
	LenDelimiter  byte = 0xFF
	LenPayloadLen      = 9
)

var byteOrder = binary.BigEndian

type Message struct {
	Topic   string
	Message *message.Message
}

// PrepareMessageForSend add msg len to result message and delimiter byte for len end detection.
func PrepareMessageForSend(baseMsg []byte) []byte {
	b := make([]byte, len(baseMsg)+LenPayloadLen)
	byteOrder.PutUint64(b, uint64(len(baseMsg)))
	b[8] = LenDelimiter
	copy(b[9:], baseMsg)

	return b
}

func ReadLen(baseMsg []byte) uint64 {
	if len(baseMsg) > 0 {
		return byteOrder.Uint64(baseMsg)
	}

	return 0
}
