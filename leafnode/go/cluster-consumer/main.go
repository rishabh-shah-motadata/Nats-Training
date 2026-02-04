package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
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
	// Connect DIRECTLY to cluster (not via leafnode)
	nc, err := nats.Connect(
		"nats://localhost:4222", // Cluster node
		nats.Name("cluster-consumer"),
		nats.ReconnectWait(2*time.Second),
		nats.MaxReconnects(-1),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	fmt.Println("Connected to NATS cluster at localhost:4222")

	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create consumer on cluster
	consumer, err := js.CreateOrUpdateConsumer(ctx, "ORDERS", jetstream.ConsumerConfig{
		Name:          "cluster-consumer-1",
		Durable:       "cluster-consumer-1",
		Description:   "Consumer connected directly to cluster",
		FilterSubject: "orders.created",
		AckPolicy:     jetstream.AckExplicitPolicy,
		DeliverPolicy: jetstream.DeliverAllPolicy,
		MaxDeliver:    3,
		AckWait:       30 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}

	fmt.Printf("Consumer created: %s\n", consumer.CachedInfo().Name)

	// Consume messages
	consumerCtx, err := consumer.Consume(func(msg jetstream.Msg) {
		var order Order
		if err := json.Unmarshal(msg.Data(), &order); err != nil {
			log.Printf("Failed to unmarshal: %v", err)
			msg.Nak()
			return
		}

		// Identify source by OrderID prefix
		source := "cluster"
		if len(order.OrderID) > 4 && order.OrderID[:4] == "LEAF" {
			source = "leafnode"
		}

		fmt.Printf("Received order from %s: %s | Customer: %s | Amount: $%.2f\n",
			source,
			order.OrderID,
			order.Customer,
			order.Amount,
		)

		meta, _ := msg.Metadata()
		fmt.Printf("   Stream: %s | Seq: %d | Consumer Seq: %d\n",
			meta.Stream,
			meta.Sequence.Stream,
			meta.Sequence.Consumer,
		)

		if err := msg.Ack(); err != nil {
			log.Printf("Failed to ack: %v", err)
		}
	})
	if err != nil {
		log.Fatalf("Failed to start consuming: %v", err)
	}
	defer consumerCtx.Stop()

	fmt.Println("Listening for orders (including from leafnode)... (Ctrl+C to stop)")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down consumer")
}

/*
```

**How Cluster Receives Data from Leafnode:**
```
Unified Message Flow:

PUBLISHERS (either location):
  cluster-publisher → nats-1
  leaf-publisher → leaf-1 → nats-1

BOTH write to same stream:
  ORDERS stream (on nats-1/2/3)

CONSUMERS (either location):
  cluster-consumer → nats-1 → reads from ORDERS
  leaf-consumer → leaf-1 → nats-1 → reads from ORDERS
*/
