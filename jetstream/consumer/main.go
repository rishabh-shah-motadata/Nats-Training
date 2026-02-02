package main

import (
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := nats.Connect(
		"nats://localhost:4222, nats://localhost:4223, nats://localhost:4224",
		nats.Name("Jetstream-Limit-Publisher"),
	)
	if err != nil {
		log.Println("failed to connect with nats server")
		return
	}
	defer nc.Drain()

	js, err := nc.JetStream()
	if err != nil {
		log.Println("failed to create jetstream context")
		return
	}

	msgs, err := js.PullSubscribe(
		"orders.created",
		"ORDER_CONSUMER", // By just providing a name, we make it a durable consumer, if we put "", it will be an ephemeral consumer
		nats.BindStream("LIMIT_ORDERS"),
		nats.AckExplicit(),
		nats.MaxAckPending(5),
		nats.AckWait(1*time.Second),
		nats.MaxDeliver(2),
		nats.ReplayInstant(),
		nats.DeliverAll(),
		// nats.DeliverSubject("delivery.orders.created"), // This will make the consumer a push based consumer
	)
	if err != nil {
		log.Println("failed to subscribe to subject:", err)
		return
	}

	fetchedMsgs, err := msgs.Fetch(5, nats.MaxWait(5*time.Second))
	if err != nil {
		log.Fatal(err)
	}

	for _, m := range fetchedMsgs {
		metadata, _ := m.Metadata()
		log.Println("Received message:", string(m.Data), "Subject:", m.Subject, "Sequence:", metadata.Sequence.Stream, "Delivery:", metadata.NumDelivered)
	}

	time.Sleep(2 * time.Second)

	log.Println("Fetching messages after AckWait duration...")
	fetchedMsgs, err = msgs.Fetch(5, nats.MaxWait(5*time.Second))
	if err != nil {
		log.Println("failed to fetch messages:", err)
		return
	}

	for _, m := range fetchedMsgs {
		metadata, _ := m.Metadata()
		log.Println("Received message:", string(m.Data), "Subject:", m.Subject, "Sequence:", metadata.Sequence.Stream, "Delivery:", metadata.NumDelivered)
	}

	time.Sleep(2 * time.Second)

	log.Println("Final fetch to check for any remaining messages...")
	fetchedMsgs, err = msgs.Fetch(5, nats.MaxWait(5*time.Second))
	if err != nil {
		log.Println("failed to fetch messages:", err)
		return
	}

	if len(fetchedMsgs) == 0 {
		log.Println("No more messages to fetch.")
	} else {
		for _, m := range fetchedMsgs {
			metadata, _ := m.Metadata()
			log.Println("Received message:", string(m.Data), "Subject:", m.Subject, "Sequence:", metadata.Sequence.Stream, "Delivery:", metadata.NumDelivered)

			// Ack the message
			if err := m.Ack(); err != nil {
				log.Println("failed to ack message:", err)
			}
			log.Println("Acknowledged message:", metadata.Sequence.Stream)
		}
	}
}
