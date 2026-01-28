package main

import (
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := nats.Connect(
		"nats://localhost:4222, nats://localhost:4223, nats://localhost:4224",
		nats.Name("Jetstream-Limit-Publisher"),
	)
	if err != nil {
		log.Println("failed to connect with nats server")
	}
	defer nc.Drain()

	js, err := nc.JetStream()
	if err != nil {
		log.Println("failed to create jetstream context")
	}

	for i := 1; i <= 5; i++ {
		data := fmt.Appendf(nil, "order-id-%d", i)

		ack, err := js.Publish("orders.created", data)
		if err != nil {
			log.Println("failed to publish message:", err)
		}

		log.Println("published message to subject:", ack.Stream, ack.Sequence, ack.Domain, ack.Duplicate)
	}
}
