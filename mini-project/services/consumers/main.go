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
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nats-io/nats.go/jetstream"
)

func main() {
	var wg sync.WaitGroup
	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	// Channel for workerpool
	msgChan := make(chan jetstream.Msg, 100)

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
	for range numWorkers {
		wg.Go(func() {
			processMessages(msgChan, pgPool)
		})
	}

	// Create JetStream Context
	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatal("error creating JetStream context:", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	consumer, err := js.Consumer(ctx, "ORDERS", "ORDER_CONSUMER")
	if err != nil {
		log.Fatal("error subscribing to subject:", err)
		return
	}

	cctx, err := consumer.Consume(func(msg jetstream.Msg) {
		msgChan <- msg
	})
	if err != nil {
		log.Fatal("error creating consumer context:", err)
		return
	}
	log.Println("consumer started, waiting for messages...")

	<-quit
	log.Println("shutting down consumer...")
	cctx.Drain()
	close(msgChan)
	wg.Wait()
	log.Println("consumer shut down gracefully")
}

func processMessages(msgChan chan jetstream.Msg, pgxPool *pgxpool.Pool) {
	for msg := range msgChan {
		var payload struct {
			ID string `json:"id"`
		}

		err := json.Unmarshal(msg.Data(), &payload)
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
