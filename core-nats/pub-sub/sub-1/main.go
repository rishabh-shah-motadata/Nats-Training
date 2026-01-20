package main

import (
	"log"

	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := nats.Connect(
		"nats://localhost:4222,nats://localhost:4223,nats://localhost:4224",
		nats.Name("NATS Sample Subscriber-1"),
		nats.DisconnectHandler(func(nc *nats.Conn) {
			log.Println("Subscriber-1 disconnected from NATS server")
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Println("Subscriber-1 reconnected to NATS server")
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Println("Subscriber-1 connection to NATS server closed")
		}),
	)
	if err != nil {
		log.Println("Error connecting to NATS server:", err)
		return
	}
	defer nc.Close()

	log.Println("Connected to:", nc.ConnectedUrl())

	// Subscribe to the subject
	_, err = nc.Subscribe("events.user.created", func(msg *nats.Msg) {
		log.Println("Received message:", string(msg.Data))
	})
	if err != nil {
		log.Println("Error subscribing to subject:", err)
		return
	}

	log.Println("Subscribed to 'events.user.created'")

	select {}
}