package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	watermillnet "github.com/andyollylarkin/watermill-net"
	"github.com/andyollylarkin/watermill-net/pkg"
	"github.com/andyollylarkin/watermill-net/pkg/connection"
)

func main() {
	logger := watermill.NewStdLogger(true, true)

	lNet, err := connection.NewTCP4Listener(":9090")
	if err != nil {
		log.Fatal(err)
	}

	s, err := watermillnet.NewSubscriber(watermillnet.SubscriberConfig{
		Marshaler:   pkg.MessagePackMarshaler{},
		Unmarshaler: pkg.MessagePackUnmarshaler{},
		Logger:      logger,
	})

	conn, err := lNet.Accept()
	if err != nil {
		log.Fatal(err)
	}

	sRetConn := connection.NewReconnectListenerWrapper(context.Background(), conn, logger, time.Second*5, lNet)

	s.SetConnection(sRetConn)
	err = s.Connect(nil)

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
