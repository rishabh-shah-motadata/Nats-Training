package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Order struct {
	OrderID   string    `json:"order_id"`
	Customer  string    `json:"customer"`
	Amount    float64   `json:"amount"`
	Timestamp time.Time `json:"timestamp"`
}

func main() {
	// Connect to NATS cluster node
	// In production, use multiple URLs for failover:
	// nats.Connect("nats://localhost:4222,nats://localhost:4223,nats://localhost:4224")
	nc, err := nats.Connect(
		"nats://localhost:4222",
		nats.UserInfo("app", "app"),
		nats.Name("cluster-publisher"),
		nats.ReconnectWait(2*time.Second),
		nats.MaxReconnects(-1), // Infinite reconnects
	)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Drain()

	fmt.Println("Connected to NATS cluster at localhost:4222")

	// Create JetStream context
	js, err := jetstream.NewWithDomain(nc, "hub")
	if err != nil {
		log.Fatal("Failed to create context: ", err)
	}

	// Publish messages in a loop
	for i := 1; i <= 10; i++ {
		order := Order{
			OrderID:   fmt.Sprintf("ORD-%d", 1000+i),
			Customer:  fmt.Sprintf("customer-%d", i),
			Amount:    99.99 * float64(i),
			Timestamp: time.Now(),
		}

		data, err := json.Marshal(order)
		if err != nil {
			log.Printf("Failed to marshal the orders: %v", err)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		// Publish to JetStream
		// This writes to the ORDERS stream on the cluster
		ack, err := js.Publish(ctx, "orders.created", data)
		if err != nil {
			log.Printf("Failed to publish: %v", err)
			cancel()
			continue
		}

		fmt.Printf("Published order %s | Stream: %s | Seq: %d\n",
			order.OrderID,
			ack.Stream,
			ack.Sequence,
		)

		cancel()
		time.Sleep(2 * time.Second)
	}

	fmt.Println("Publishing complete")
}
