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

	pConn := connection.NewTCPConnection(time.Second * 30)

	p, err := watermillnet.NewPublisher(watermillnet.PublisherConfig{
		Marshaler:   pkg.MessagePackMarshaler{},
		Unmarshaler: pkg.MessagePackUnmarshaler{},
	}, false)
	p.SetConnection(pConn)

	if err != nil {
		log.Fatal(err)
	}

	err = p.Connect(&net.TCPAddr{IP: []byte{127, 0, 0, 1}, Port: 9090})

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
