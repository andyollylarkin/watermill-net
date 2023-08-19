package watermillnet

type Marshaler interface {
	MarshalMessage(msg any) ([]byte, error)
}

type Unmarshaler interface {
	UnmarshalMessage(in []byte, out any) error
}
