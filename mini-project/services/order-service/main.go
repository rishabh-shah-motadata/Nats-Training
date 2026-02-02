package main

import (
	"context"
	"log"
	"nats-project/internal/db"
	"nats-project/internal/nats"
	"nats-project/internal/router"
	"nats-project/services/order-service/api"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	// Initialize PostgreSQL Connection
	pgPool, err := db.InitPostgresDB(ctx)
	if err != nil {
		log.Fatal("error initializing postgres database:", err)
		return
	}
	defer pgPool.Close()

	// Initialize NATS Connection
	nc, err := nats.InitNATS("Order-Service")
	if err != nil {
		log.Fatal("error initializing NATS connection:", err)
		return
	}
	defer nc.Drain()
	log.Println("connected to NATS server:", nc.ConnectedUrl())

	// Create JetStream Context
	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatal("error creating JetStream context:", err)
		return
	}

	// Initialize Gin Router
	router := router.NewGinRouter()
	api.RegisterRoutes(router, pgPool, js)

	// Initialize Gin Server
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Start the server
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Println("error starting server:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	<-quit
	if err := server.Shutdown(context.Background()); err != nil {
		log.Println("error shutting down server:", err)
		return
	}
}
