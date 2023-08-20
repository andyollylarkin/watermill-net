package internal

type AckMessage struct {
	UUID  string
	Acked bool // if true message is Acked. Otherwise message is Nacked.
}
