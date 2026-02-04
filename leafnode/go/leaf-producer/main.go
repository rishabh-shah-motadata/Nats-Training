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
	// Connect to LEAFNODE (not cluster)
	nc, err := nats.Connect(
		"nats://localhost:4225",
		nats.Name("leaf-publisher"),
		nats.ReconnectWait(2*time.Second),
		nats.MaxReconnects(-1),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	fmt.Println("Connected to NATS leafnode at localhost:4225")

	// Create JetStream context
	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatal(err)
	}

	// Publish messages
	for i := 1; i <= 10; i++ {
		order := Order{
			OrderID:   fmt.Sprintf("LEAF-ORD-%d", 2000+i),
			Customer:  fmt.Sprintf("leaf-customer-%d", i),
			Amount:    149.99 * float64(i),
			Timestamp: time.Now(),
		}

		data, err := json.Marshal(order)
		if err != nil {
			log.Printf("Failed to marshal order: %v", err)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		// Publish via leafnode
		// Message flow: leaf-publisher → leaf-1 → cluster → ORDERS stream
		ack, err := js.Publish(ctx, "orders.created", data)
		if err != nil {
			log.Printf("Failed to publish: %v", err)
			cancel()
			continue
		}

		fmt.Printf("Published order %s (via leaf) | Stream: %s | Seq: %d\n",
			order.OrderID,
			ack.Stream,
			ack.Sequence,
		)

		cancel()
		time.Sleep(2 * time.Second)
	}

	fmt.Println("Publishing complete from leafnode")
}

/*
```

**Data Flow Explanation:**
```
leaf-publisher (Go app)
         ↓
    js.Publish("orders.created", data)
         ↓
leaf-1 (leafnode server) ← Connected to cluster
         ↓
nats-1/2/3 (cluster) ← Receives publish request
         ↓
ORDERS stream ← Persists message, replicates across nodes
         ↓
Returns ACK with sequence number
         ↓
leaf-1 ← Routes ACK back
         ↓
leaf-publisher ← Receives confirmation

*/
