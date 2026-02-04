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
	// Connect to LEAFNODE server (not cluster)
	// This is the critical difference - connecting to leaf-1
	nc, err := nats.Connect(
		"nats://localhost:4221", // Leafnode port
		nats.Name("leaf-consumer"),
		nats.UserInfo("app", "app"),
		nats.ReconnectWait(2*time.Second),
		nats.MaxReconnects(-1),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Drain()

	fmt.Println("Connected to NATS leafnode at localhost:4225")

	// Create JetStream context
	// Even though connected to leafnode, this JS context
	// will transparently route operations to the cluster
	js, err := jetstream.NewWithDomain(nc, "hub")
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	streamInfo, err := js.Stream(ctx, "ORDERS")
	if err != nil {
		log.Fatalf("Stream ORDERS not found via leafnode: %v", err)
	}
	log.Printf("Stream ORDERS found via leafnode: %+v", streamInfo.CachedInfo())

	// Create or get consumer
	// Consumer is created ON THE CLUSTER, not the leafnode
	// The leafnode just proxies this request
	consumer, err := js.CreateOrUpdateConsumer(ctx, "ORDERS", jetstream.ConsumerConfig{
		Name:          "leaf-consumer-1",
		Durable:       "leaf-consumer-1", // Survives consumer restarts
		Description:   "Consumer connected via leafnode",
		FilterSubject: "orders.created",
		AckPolicy:     jetstream.AckExplicitPolicy, // Manual acks
		MaxDeliver:    3,                           // Retry up to 3 times
		AckWait:       30 * time.Second,            // Wait 30s for ack
	})
	if err != nil {
		log.Fatalf("Failed to create consumer: %v", err)
	}

	fmt.Printf("Consumer created: %s\n", consumer.CachedInfo().Name)

	// Start consuming messages
	// This creates a long-lived subscription that receives messages
	// Messages flow: Cluster → Leafnode → This client
	consumerCtx, err := consumer.Consume(func(msg jetstream.Msg) {
		var order Order
		if err := json.Unmarshal(msg.Data(), &order); err != nil {
			log.Printf("Failed to unmarshal: %v", err)
			msg.Nak() // Negative acknowledge - redeliver
			return
		}

		fmt.Printf("📦 Received order: %s | Customer: %s | Amount: $%.2f | Lag: %v\n",
			order.OrderID,
			order.Customer,
			order.Amount,
			time.Since(order.Timestamp),
		)

		// Metadata available via msg
		meta, _ := msg.Metadata()
		fmt.Printf("   Stream: %s | Seq: %d | Pending: %d\n",
			meta.Stream,
			meta.Sequence.Stream,
			meta.NumPending,
		)

		// Acknowledge message
		// This ack flows: Client → Leafnode → Cluster
		if err := msg.Ack(); err != nil {
			log.Printf("Failed to ack: %v", err)
		}
	})
	if err != nil {
		log.Fatalf("Failed to start consuming: %v", err)
	}
	defer consumerCtx.Stop()

	fmt.Println("Listening for orders... (Ctrl+C to stop)")

	// Wait for interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n👋 Shutting down consumer")
}

/*
```

**Why This Works Across Leafnodes:**
```
Flow Diagram:

1. Message Published to Cluster:
   cluster-publisher → nats-1 (cluster) → ORDERS stream persisted

2. Consumer Request via Leafnode:
   leaf-consumer → leaf-1 → nats-1/2/3 (cluster) → "Create consumer on ORDERS"

3. Message Delivery:
   ORDERS stream (cluster) → nats-1 → leaf-1 → leaf-consumer

4. Acknowledgment:
   leaf-consumer → leaf-1 → cluster → consumer state updated

*/
