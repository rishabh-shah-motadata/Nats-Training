package main

import (
	"context"
	"encoding/json"
	"log"
	"nats-project/internal/db"
	ns "nats-project/internal/nats"
	"os"
	"os/signal"
	"sync"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nats-io/nats.go"
)

func main() {
	var wg sync.WaitGroup
	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	// Channel for workerpool
	msgChan := make(chan *nats.Msg, 100)

	// Initialize NATS Connection
	nc, err := ns.InitNATS("Consumer-1")
	if err != nil {
		log.Fatal("error initializing NATS connection:", err)
		return
	}
	defer nc.Drain()
	log.Println("connected to NATS server:", nc.ConnectedUrl())

	// Initialize PostgreSQL Connection
	pgPool, err := db.InitPostgresDB(context.Background())
	if err != nil {
		log.Fatal("error initializing postgres database:", err)
		return
	}
	defer pgPool.Close()

	// Setup workerpool
	numWorkers := 5
	for i := 0; i < numWorkers; i++ {
		wg.Go(func() {
			processMessages(msgChan, pgPool)
		})
	}

	// Create JetStream Context
	js, err := nc.JetStream()
	if err != nil {
		log.Fatal("error creating JetStream context:", err)
		return
	}

	subs, err := js.Subscribe(
		"orders.created",
		func(msg *nats.Msg) {
			msgChan <- msg
		},
		nats.Bind("ORDERS", "ORDER_CONSUMER"),
		nats.AckExplicit(),
		nats.MaxAckPending(5),
		nats.MaxDeliver(2),
		nats.ReplayInstant(),
	)
	if err != nil {
		log.Fatal("error subscribing to subject:", err)
		return
	}

	<-quit
	log.Println("shutting down consumer...")
	err = subs.Drain()
	if err != nil {
		log.Println("error draining subscription:", err)
	}
}

func processMessages(msgChan chan *nats.Msg, pgxPool *pgxpool.Pool) {
	for msg := range msgChan {
		var payload struct {
			ID string `json:"id"`
		}

		err := json.Unmarshal(msg.Data, &payload)
		if err != nil {
			log.Printf("error unmarshalling message: %v", err)
			msg.Nak()
			continue
		}
		log.Printf("Processing order ID: %s", payload.ID)

		_, err = pgxPool.Exec(context.Background(), "UPDATE orders SET status=$1 WHERE id=$2", "PROCESSING", payload.ID)
		if err != nil {
			log.Printf("error updating order status: %v", err)
			msg.Nak()
			continue
		}
		log.Printf("Order ID %s marked as PROCESSING", payload.ID)

		err = msg.Ack()
		if err != nil {
			log.Printf("error acknowledging message: %v", err)
			continue
		}
		log.Printf("Acknowledged message for order ID: %s", payload.ID)
	}
}
