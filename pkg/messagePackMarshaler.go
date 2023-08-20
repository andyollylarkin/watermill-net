package pkg

import (
	"github.com/vmihailenco/msgpack/v5"
)

type MessagePackMarshaler struct{}

func (mpm MessagePackMarshaler) MarshalMessage(msg any) ([]byte, error) {
	return msgpack.Marshal(msg)
}

type MessagePackUnmarshaler struct{}

func (mpu MessagePackUnmarshaler) UnmarshalMessage(in []byte, out any) error {
	return msgpack.Unmarshal(in, out)
}
