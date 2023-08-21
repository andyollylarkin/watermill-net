package main

import (
	"log"
	"net"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	watermillnet "github.com/andyollylarkin/watermill-net"
	"github.com/andyollylarkin/watermill-net/pkg"
	"github.com/andyollylarkin/watermill-net/pkg/connection"
)

func main() {

	pConn := connection.NewTCPConnection(net.Dialer{})

	p, err := watermillnet.NewPublisher(watermillnet.PublisherConfig{
		Conn:        pConn,
		RemoteAddr:  &net.TCPAddr{IP: []byte{127, 0, 0, 1}, Port: 9090},
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

	for {
		m := message.NewMessage("111-222", []byte("Hello"))
		err = p.Publish("test1", m)

		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(time.Second * 2)
	}
}
