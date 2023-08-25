package watermillnet_test

import (
	"context"
	"sync"
	"testing"

	watermillnet "github.com/andyollylarkin/watermill-net"
	"github.com/andyollylarkin/watermill-net/pkg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscriberCreateOK(t *testing.T) {
	sc := watermillnet.SubscriberConfig{
		Marshaler:   pkg.MessagePackMarshaler{},
		Unmarshaler: pkg.MessagePackUnmarshaler{},
	}
	_, err := watermillnet.NewSubscriber(sc)
	require.NoError(t, err)
}

func TestSubscriberError(t *testing.T) {
	tc := []struct {
		name     string
		config   watermillnet.SubscriberConfig
		expected string
	}{
		{
			name: "Err Conn nil",
			config: watermillnet.SubscriberConfig{
				Marshaler:   nil,
				Unmarshaler: pkg.MessagePackUnmarshaler{},
			},
			expected: "invalid field: Marshaler. reason: cant be nil",
		},
		{
			name: "Err Addr nil",
			config: watermillnet.SubscriberConfig{
				Marshaler:   pkg.MessagePackMarshaler{},
				Unmarshaler: nil,
			},
			expected: "invalid field: Unmarshaler. reason: cant be nil",
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			_, err := watermillnet.NewSubscriber(c.config)
			assert.Error(t, err)
			assert.ErrorContains(t, err, c.expected)
		})
	}
}

func TestPublisherConnectOK(t *testing.T) {
	pipeConn := NewPipeConnection()
	s, err := watermillnet.NewSubscriber(watermillnet.SubscriberConfig{
		Marshaler:   pkg.MessagePackMarshaler{},
		Unmarshaler: pkg.MessagePackUnmarshaler{},
	})
	require.NoError(t, err)
	err = s.Connect(NewListenerStub(pipeConn.RemoteSideConn()))
	require.NoError(t, err)
}

func TestPublisherConnectErrorClosed(t *testing.T) {
	pipeConn := NewPipeConnection()
	s, err := watermillnet.NewSubscriber(watermillnet.SubscriberConfig{
		Marshaler:   pkg.MessagePackMarshaler{},
		Unmarshaler: pkg.MessagePackUnmarshaler{},
	})
	s.SetConnection(pipeConn)
	require.NoError(t, err)
	err = s.Close()
	require.NoError(t, err)
	err = s.Connect(NewListenerStub(pipeConn.RemoteSideConn()))
	require.ErrorIs(t, err, watermillnet.ErrSubscriberClosed)
}

func TestPublisherConnectError(t *testing.T) {
	s, err := watermillnet.NewSubscriber(watermillnet.SubscriberConfig{
		Marshaler:   pkg.MessagePackMarshaler{},
		Unmarshaler: pkg.MessagePackUnmarshaler{},
	})
	require.NoError(t, err)
	tc := []struct {
		name       string
		subscriber *watermillnet.Subscriber
		argOne     watermillnet.Listener
		expected   string
	}{
		{
			name:       "Nil listener",
			subscriber: s,
			argOne:     nil,
			expected:   "listener cant be nil",
		},
	}
	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			err = c.subscriber.Connect(c.argOne)
			assert.Errorf(t, err, c.expected)
		})
	}
}

func TestCloseWaitAllConns(t *testing.T) {
	pipeConn := NewPipeConnection()
	p, err := watermillnet.NewPublisher(watermillnet.PublisherConfig{
		Marshaler:   pkg.MessagePackMarshaler{},
		Unmarshaler: pkg.MessagePackUnmarshaler{},
	}, false)
	p.SetConnection(pipeConn)
	require.NoError(t, err)
	err = p.Connect(pipeAddr{})
	require.NoError(t, err)

	s, err := watermillnet.NewSubscriber(watermillnet.SubscriberConfig{
		Marshaler:   pkg.MessagePackMarshaler{},
		Unmarshaler: pkg.MessagePackUnmarshaler{},
	})
	s.SetConnection(pipeConn)
	require.NoError(t, err)
	err = s.Connect(NewListenerStub(pipeConn.RemoteSideConn()))
	require.NoError(t, err)

	subChan1, err := s.Subscribe(context.Background(), "test_topic1")
	require.NoError(t, err)
	subChan2, err := s.Subscribe(context.Background(), "test_topic1")
	require.NoError(t, err)

	var wg sync.WaitGroup

	var m sync.Mutex

	wg.Add(2)

	var doneCounter int = 0
	go func() {
		defer wg.Done()

		for range subChan1 {
		}

		m.Lock()
		doneCounter++
		m.Unlock()
	}()

	go func() {
		defer wg.Done()

		for range subChan2 {
		}

		m.Lock()
		doneCounter++
		m.Unlock()
	}()

	err = s.Close()
	require.NoError(t, err)
	wg.Wait()
	assert.Equal(t, 2, doneCounter)
}

func TestCloseSocketConn(t *testing.T) {
	pipeConn := NewPipeConnection()
	p, err := watermillnet.NewPublisher(watermillnet.PublisherConfig{
		Marshaler:   pkg.MessagePackMarshaler{},
		Unmarshaler: pkg.MessagePackUnmarshaler{},
	}, false)
	p.SetConnection(pipeConn)
	require.NoError(t, err)
	err = p.Connect(pipeAddr{})
	require.NoError(t, err)

	s, err := watermillnet.NewSubscriber(watermillnet.SubscriberConfig{
		Marshaler:   pkg.MessagePackMarshaler{},
		Unmarshaler: pkg.MessagePackUnmarshaler{},
	})

	wrapper := NewConnWrapper(pipeConn.RemoteSideConn())

	require.NoError(t, err)

	lStub := NewListenerStubWithFakeConn(wrapper)
	err = s.Connect(lStub)
	require.NoError(t, err)

	err = s.Close()
	require.NoError(t, err)
	assert.Equal(t, true, wrapper.Closed)
}

// Subscribe cancelled by context
func TestSubscriberCancelByContext(t *testing.T) {
	pipeConn := NewPipeConnection()
	p, err := watermillnet.NewPublisher(watermillnet.PublisherConfig{
		Marshaler:   pkg.MessagePackMarshaler{},
		Unmarshaler: pkg.MessagePackUnmarshaler{},
	}, false)
	p.SetConnection(pipeConn)
	require.NoError(t, err)
	err = p.Connect(pipeAddr{})
	require.NoError(t, err)

	s, err := watermillnet.NewSubscriber(watermillnet.SubscriberConfig{
		Marshaler:   pkg.MessagePackMarshaler{},
		Unmarshaler: pkg.MessagePackUnmarshaler{},
	})

	wrapper := NewConnWrapper(pipeConn.RemoteSideConn())

	require.NoError(t, err)

	lStub := NewListenerStubWithFakeConn(wrapper)
	err = s.Connect(lStub)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	subCh, err := s.Subscribe(ctx, "test")
	require.NoError(t, err)

	var wg sync.WaitGroup

	var mu sync.Mutex

	var complete bool

	wg.Add(1)

	go func() {
		defer wg.Done()

		for range subCh {
		}

		mu.Lock()
		complete = true
		mu.Unlock()
	}()

	cancel()
	wg.Wait()
	assert.Equal(t, true, complete)
}

func TestSubscriberConnectionNotSetError(t *testing.T) {
	retPub := func(t *testing.T) *watermillnet.Subscriber {
		s, err := watermillnet.NewSubscriber(watermillnet.SubscriberConfig{
			Marshaler:   pkg.MessagePackMarshaler{},
			Unmarshaler: pkg.MessagePackUnmarshaler{},
		})
		require.NoError(t, err)

		return s
	}

	tc := []struct {
		name   string
		exec   func(t *testing.T) error
		expect error
	}{
		{name: "Get connection", exec: func(t *testing.T) error {
			p := retPub(t)
			_, err := p.GetConnection()

			return err
		}, expect: watermillnet.ErrConnectionNotSet},
		{name: "Subscribe", exec: func(t *testing.T) error {
			p := retPub(t)
			_, err := p.Subscribe(context.Background(), "")

			return err
		}, expect: watermillnet.ErrConnectionNotSet},
		{name: "Close", exec: func(t *testing.T) error {
			p := retPub(t)
			err := p.Close()

			return err
		}, expect: watermillnet.ErrConnectionNotSet},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			res := c.exec(t)
			assert.ErrorIs(t, res, c.expect)
		})
	}
}

func TestAllowNilListenerWhenConnectionSet(t *testing.T) {
	pc := NewPipeConnection()
	s, err := watermillnet.NewSubscriber(watermillnet.SubscriberConfig{
		Marshaler:   pkg.MessagePackMarshaler{},
		Unmarshaler: pkg.MessagePackUnmarshaler{},
	})
	require.NoError(t, err)
	s.SetConnection(pc)
	err = s.Connect(nil)
	require.NoError(t, err)
}

// func TestCancelConsumeWhenReadEOF(t *testing.T) {
// 	pc := connection.NewTCPConnection(net.Dialer{}, time.Hour)
// 	nl, err := net.Listen("tcp4", ":0")
// 	require.NoError(t, err)

// 	l := connection.NewTCP4Listener(nl)

// 	var wg sync.WaitGroup

// 	wg.Add(1)

// 	go func() {
// 		_, err := l.Accept()
// 		require.NoError(t, err)
// 		// sconn.Close()
// 		wg.Done()
// 	}()

// 	err = pc.Connect(nl.Addr())
// 	require.NoError(t, err)
// 	wg.Wait()

// 	b := make([]byte, 10)
// 	_, err = pc.Read(b)
// 	require.ErrorIs(t, err, io.EOF)

// 	s, err := watermillnet.NewSubscriber(watermillnet.SubscriberConfig{
// 		Marshaler:   pkg.MessagePackMarshaler{},
// 		Unmarshaler: pkg.MessagePackUnmarshaler{},
// 	})
// 	require.NoError(t, err)
// 	s.SetConnection(pc)
// 	err = s.Connect(nil)
// 	require.NoError(t, err)

// 	sub, err := s.Subscribe(context.Background(), "")

// 	var mu sync.Mutex
// 	var done bool = false

// 	wg.Add(1)
// 	go func() {
// 		for range sub {
// 		}

// 		mu.Lock()
// 		done = true
// 		mu.Unlock()
// 		wg.Done()
// 	}()

// 	wg.Wait()
// 	require.Equal(t, true, done)
// }
