package watermillnet_test

import (
	"testing"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	watermillnet "github.com/andyollylarkin/watermill-net"
	"github.com/andyollylarkin/watermill-net/internal"
	"github.com/andyollylarkin/watermill-net/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func CreateAckMessage(t *testing.T, ack bool, uuid string) []byte {
	m := pkg.MessagePackMarshaler{}

	var ackMsg internal.AckMessage

	if ack {
		ackMsg = internal.AckMessage{
			UUID:  uuid,
			Acked: true,
		}
	} else {
		ackMsg = internal.AckMessage{
			UUID:  uuid,
			Acked: false,
		}
	}

	b, err := m.MarshalMessage(ackMsg)
	if err != nil {
		t.Fatalf("Fail marshal ack message %s", err.Error())
	}

	return internal.PrepareMessageForSend(b)
}

func TestPublishMessageRemoteSideReceiveOK(t *testing.T) {
	pipeConn := NewPipeConnection()
	uuid := watermill.NewUUID()
	config := watermillnet.PublisherConfig{
		Conn:        pipeConn,
		Addr:        pipeAddr{},
		Marshaler:   pkg.MessagePackMarshaler{},
		Unmarshaler: pkg.MessagePackUnmarshaler{},
	}

	go func() {
		out := make([]byte, 4096)
		rs := pipeConn.RemoteSideConn()
		rs.Read(out)
		ackMsg := CreateAckMessage(t, true, uuid)
		rs.Write(ackMsg)
	}()

	p, err := watermillnet.NewPublisher(config)
	require.NoError(t, err)
	err = p.Publish("test_topic", message.NewMessage("", []byte("Hello world")))
	require.NoError(t, err)
}

func TestPublishMessageRemoteSideReceiveNackResponse(t *testing.T) {
	pipeConn := NewPipeConnection()
	uuid := watermill.NewUUID()
	config := watermillnet.PublisherConfig{
		Conn:        pipeConn,
		Addr:        pipeAddr{},
		Marshaler:   pkg.MessagePackMarshaler{},
		Unmarshaler: pkg.MessagePackUnmarshaler{},
	}

	go func() {
		out := make([]byte, 512)
		rs := pipeConn.RemoteSideConn()
		rs.Read(out)
		ackMsg := CreateAckMessage(t, false, uuid)
		rs.Write(ackMsg)
	}()

	p, err := watermillnet.NewPublisher(config)
	require.NoError(t, err)
	err = p.Publish("test_topic", message.NewMessage("", []byte("Hello world")))
	require.Error(t, err)
	t.Log(err)
	assert.ErrorIs(t, err, watermillnet.ErrNacked)
}

func TestPublisherCreateError(t *testing.T) {
	// pc := NewPipeConnection()
	tc := []struct {
		name     string
		config   watermillnet.PublisherConfig
		expected string
	}{
		{
			name: "Err Conn nil",
			config: watermillnet.PublisherConfig{
				Conn:        nil,
				Addr:        pipeAddr{},
				Marshaler:   pkg.MessagePackMarshaler{},
				Unmarshaler: pkg.MessagePackUnmarshaler{},
			},
			expected: "invalid field: Conn. reason: cant be nil",
		},
		{
			name: "Err Addr nil",
			config: watermillnet.PublisherConfig{
				Conn:        &PipeConnection{},
				Addr:        nil,
				Marshaler:   pkg.MessagePackMarshaler{},
				Unmarshaler: pkg.MessagePackUnmarshaler{},
			},
			expected: "invalid field: Addr. reason: cant be nil",
		},
		{
			name: "Err Marshaler nil",
			config: watermillnet.PublisherConfig{
				Conn:        &PipeConnection{},
				Addr:        pipeAddr{},
				Marshaler:   nil,
				Unmarshaler: pkg.MessagePackUnmarshaler{},
			},
			expected: "invalid field: Marshaler. reason: cant be nil",
		},
		{
			name: "Err Unmarshaler nil",
			config: watermillnet.PublisherConfig{
				Conn:        &PipeConnection{},
				Addr:        pipeAddr{},
				Marshaler:   pkg.MessagePackMarshaler{},
				Unmarshaler: nil,
			},
			expected: "invalid field: Unmarshaler. reason: cant be nil",
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			_, err := watermillnet.NewPublisher(c.config)
			assert.Error(t, err)
			assert.ErrorContains(t, err, c.expected)
		})
	}
}

func TestPublisherCreateOK(t *testing.T) {
	// pc := NewPipeConnection()
	tc := []struct {
		name   string
		config watermillnet.PublisherConfig
	}{
		{
			name: "Publisher no error",
			config: watermillnet.PublisherConfig{
				Conn:        &PipeConnection{},
				Addr:        pipeAddr{},
				Marshaler:   pkg.MessagePackMarshaler{},
				Unmarshaler: pkg.MessagePackUnmarshaler{},
			},
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			_, err := watermillnet.NewPublisher(c.config)
			assert.NoError(t, err)
		})
	}
}

func TestPublishToClosedPublisher(t *testing.T) {
	pipeConn := NewPipeConnection()
	config := watermillnet.PublisherConfig{
		Conn:        pipeConn,
		Addr:        pipeAddr{},
		Marshaler:   pkg.MessagePackMarshaler{},
		Unmarshaler: pkg.MessagePackUnmarshaler{},
	}

	p, err := watermillnet.NewPublisher(config)
	require.NoError(t, err)
	err = p.Close()
	require.NoError(t, err)
	err = p.Publish("", message.NewMessage("", []byte("Hello world")))
	require.ErrorIs(t, err, watermillnet.ErrPublisherClosed)
}

func TestPublishMultiMessage(t *testing.T) {
	pipeConn := NewPipeConnection()
	uuid := watermill.NewUUID()
	config := watermillnet.PublisherConfig{
		Conn:        pipeConn,
		Addr:        pipeAddr{},
		Marshaler:   pkg.MessagePackMarshaler{},
		Unmarshaler: pkg.MessagePackUnmarshaler{},
	}

	go func() {
		out := make([]byte, 4096)
		rs := pipeConn.RemoteSideConn()

		// ack 2 times
		for i := 0; i < 2; i++ {
			rs.Read(out)
			ackMsg := CreateAckMessage(t, true, uuid)
			rs.Write(ackMsg)
		}
	}()

	p, err := watermillnet.NewPublisher(config)
	require.NoError(t, err)
	err = p.Publish("test_topic", message.NewMessage("", []byte("Hello world")), //send 2 messages
		message.NewMessage("", []byte("Hello world2")))
	require.NoError(t, err)
}
