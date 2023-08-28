package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ThreeDotsLabs/watermill"
	watermillnet "github.com/andyollylarkin/watermill-net"
	"github.com/andyollylarkin/watermill-net/pkg"
	"github.com/andyollylarkin/watermill-net/pkg/connection"
)

func main() {
	logger := watermill.NewStdLogger(true, true)
	s, err := watermillnet.NewSubscriber(watermillnet.SubscriberConfig{
		Marshaler:   pkg.MessagePackMarshaler{},
		Unmarshaler: pkg.MessagePackUnmarshaler{},
		Logger:      logger,
	})

	if err != nil {
		log.Fatal(err)
	}

	t4L, err := connection.NewTCP4Listener(":9090")
	if err != nil {
		log.Fatal(err)
	}

	err = s.Connect(t4L)

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
