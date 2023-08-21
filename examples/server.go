package main

import (
	"context"
	"fmt"
	"log"
	"net"

	watermillnet "github.com/andyollylarkin/watermill-net"
	"github.com/andyollylarkin/watermill-net/pkg"
	"github.com/andyollylarkin/watermill-net/pkg/connection"
)

func main() {
	l, _ := net.Listen("tcp4", "127.0.0.1:9090")
	s, err := watermillnet.NewSubscriber(watermillnet.SubscriberConfig{
		Marshaler:   pkg.MessagePackMarshaler{},
		Unmarshaler: pkg.MessagePackUnmarshaler{},
	})

	if err != nil {
		log.Fatal(err)
	}

	err = s.Connect(connection.NewTCP4Listener(l))

	if err != nil {
		log.Fatal(err)
	}

	sch, err := s.Subscribe(context.Background(), "test1")
	if err != nil {
		log.Fatal(err)
	}

	sch2, err := s.Subscribe(context.Background(), "test2")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for m := range sch {
			fmt.Println(1, string(m.Payload))
			m.Ack()
		}
	}()
	for m := range sch2 {
		fmt.Println(2, string(m.Payload))
		m.Ack()
	}

}
