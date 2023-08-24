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

	pConn := connection.NewTCPConnection(net.Dialer{})

	addr := &net.TCPAddr{IP: []byte{127, 0, 0, 1}, Port: 9090}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	wpConn := connection.NewReconnectWrapper(ctx, pConn, retry.NewConstant(time.Second*3),
		watermill.NewStdLogger(true, true), addr, connection.DefaultErrorFilter)

	p, err := watermillnet.NewPublisher(watermillnet.PublisherConfig{
		Conn:        wpConn,
		RemoteAddr:  addr,
		Marshaler:   pkg.MessagePackMarshaler{},
		Unmarshaler: pkg.MessagePackUnmarshaler{},
	})
	if err != nil {
		log.Fatal(err)
	}

	err = p.Connect()

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
