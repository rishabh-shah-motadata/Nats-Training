package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

func main() {
	nc, err := nats.Connect(
		"nats://localhost:4222,nats://localhost:4223,nats://localhost:4224",
		nats.Name("NATS Tenant-2 Service Endpoint"),
		nats.DisconnectHandler(func(nc *nats.Conn) {
			log.Println("disconnected from NATS server")
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Println("reconnected to NATS server")
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Println("connection to NATS server closed")
		}),
	)
	if err != nil {
		log.Println("Error connecting to NATS server:", err)
		return
	}
	defer nc.Drain()

	log.Println("Connected to:", nc.ConnectedUrl())

	service, err := micro.AddService(nc, micro.Config{
		Name:        "tenant-2",
		Version:     "1.0.0",
		Description: "This the simple service endpoint example",
	})
	if err != nil {
		log.Println("Error creating service endpoint:", err)
		return
	}
	log.Println("Service created", service.Info().Name, service.Info().ID)

	group1 := service.AddGroup("tenant-2")

	group1.AddEndpoint("get", micro.HandlerFunc(getHandler))
	group1.AddEndpoint("post", micro.HandlerFunc(postHandler))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down service...")
}

func getHandler(req micro.Request) {
	var response []byte
	log.Println("Received GET request for tenant-2")
	response = []byte("Response from tenant-2 GET endpoint")
	req.Respond(response)
}

func postHandler(req micro.Request) {
	var response []byte
	log.Println("Received POST request for tenant-2 with data:", string(req.Data()))
	response = []byte("Response from tenant-2 POST endpoint")
	req.Respond(response)
}
