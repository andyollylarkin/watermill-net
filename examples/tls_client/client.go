package main

import (
	"crypto/tls"
	"log"
	"net"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	watermillnet "github.com/andyollylarkin/watermill-net"
	"github.com/andyollylarkin/watermill-net/pkg"
	"github.com/andyollylarkin/watermill-net/pkg/connection"
	connectionhelpers "github.com/andyollylarkin/watermill-net/pkg/helpers/connectionHelpers"
)

func main() {
	certs, err := connectionhelpers.LoadCerts("/home/denis/Desktop/local.test.ru.crt",
		"/home/denis/Desktop/local.test.ru.key")
	if err != nil {
		log.Fatal(err)
	}
	rootCert, err := connectionhelpers.LoadCertPool("/home/denis/Documents/rootca.crt")
	if err != nil {
		log.Fatal(err)
	}

	pConn := connection.NewTCPTlsConnection(time.Minute*3, &tls.Config{
		Certificates: certs,
		RootCAs:      rootCert,
	})

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
