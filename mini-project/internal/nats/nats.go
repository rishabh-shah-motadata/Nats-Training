package nats

import (
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

func InitNATS(name string) (*nats.Conn, error) {
	nc, err := nats.Connect(
		"nats://localhost:4222, nats://localhost:4223, nats://localhost:4224",
		nats.Name(name),
		nats.DisconnectHandler(func(nc *nats.Conn) {
			log.Println("disconnected from NATS server")
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Println("reconnected to NATS server")
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Println("connection to NATS server closed")
		}),
		nats.MaxReconnects(3),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		log.Println("error connecting to NATS server:", err)
		return nil, err
	}

	return nc, nil
}
