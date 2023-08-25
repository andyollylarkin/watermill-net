package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	watermillnet "github.com/andyollylarkin/watermill-net"
	"github.com/andyollylarkin/watermill-net/pkg"
	"github.com/andyollylarkin/watermill-net/pkg/connection"
	"github.com/sethvargo/go-retry"
)

func main() {

	pConn := connection.NewTCPConnection(net.Dialer{}, time.Second*30)

	addr := &net.TCPAddr{IP: []byte{127, 0, 0, 1}, Port: 9090}

	wpConn := connection.NewReconnectWrapper(context.Background(), pConn, retry.NewConstant(time.Second*3),
		watermill.NewStdLogger(true, true), addr, connection.DefaultErrorFilter, time.Second*4)

	p, err := watermillnet.NewPublisher(watermillnet.PublisherConfig{
		Marshaler:   pkg.MessagePackMarshaler{},
		Unmarshaler: pkg.MessagePackUnmarshaler{},
		Logger:      watermill.NewStdLogger(true, true),
	}, false)
	if err != nil {
		log.Fatal(err)
	}

	p.SetConnection(wpConn)

	err = p.Connect(addr)

	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			m := message.NewMessage("111-222", []byte("Hello"))
			err = p.Publish("test1", m)

			if err != nil {
				log.Fatal(err)
			}

			time.Sleep(time.Second * 2)
		}
	}()

	for {
		m := message.NewMessage("111-333", []byte("Hello2"))
		err = p.Publish("test2", m)

		if err != nil {
			log.Fatal(err)
		}

		time.Sleep(time.Second * 2)
	}
}
