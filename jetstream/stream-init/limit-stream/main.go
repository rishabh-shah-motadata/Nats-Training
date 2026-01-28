package main

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := nats.Connect(
		"nats://localhost:4222, nats://localhost:4223, nats://localhost:4224",
		nats.Name("Jetstream-1"),
	)
	if err != nil {
		log.Println("Error connecting to NATS server:", err)
		return
	}
	defer nc.Close()

	// Create JetStream context
	// Its through this context that we can manage streams and consumers
	// nc is the core NATS connection, while js is the JetStream specific context
	// that provides the higher level JetStream API functionality
	js, err := nc.JetStream()
	if err != nil {
		log.Println("Error creating JetStream context:", err)
		return
	}

	streamConfig := &nats.StreamConfig{
		Name:      "ORDERS",
		Subjects:  []string{"orders.*"},
		Storage:   nats.FileStorage,
		Retention: nats.LimitsPolicy,
		MaxMsgs:   1000,
		MaxAge:    time.Second,
		Discard:   nats.DiscardOld,
	}

	stream, err := js.AddStream(streamConfig)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Println("Stream created:")
	fmt.Printf("Name: %s\n", stream.Config.Name)
	fmt.Printf("Subjects: %v\n", stream.Config.Subjects)
}
