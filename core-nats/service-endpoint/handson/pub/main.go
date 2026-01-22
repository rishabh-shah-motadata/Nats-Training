package main

import (
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := nats.Connect(
		"nats://localhost:4222,nats://localhost:4223,nats://localhost:4224",
		nats.Name("NATS Simple Service Endpoint"),
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

	go func() {
		for range 500 {
			resp, err := nc.Request("tenant-1.get", []byte(""), nats.DefaultTimeout)
			if err != nil {
				log.Println("Error requesting get service endpoint:", err)
				return
			}
			log.Println("Response from service endpoint:", string(resp.Data))
		}
	}()

	go func() {
		for range 500 {
			resp, err := nc.Request("tenant-1.post", []byte(""), nats.DefaultTimeout)
			if err != nil {
				log.Println("Error requesting post service endpoint:", err)
				return
			}
			log.Println("Response from service endpoint:", string(resp.Data))
		}
	}()

	go func() {
		for range 500 {
			resp, err := nc.Request("tenant-2.get", []byte("Sample POST data"), nats.DefaultTimeout)
			if err != nil {
				log.Println("Error requesting get service endpoint:", err)
				return
			}
			log.Println("Response from service endpoint:", string(resp.Data))
		}
	}()

	go func() {
		for range 500 {
			resp, err := nc.Request("tenant-2.post", []byte("Sample POST data"), nats.DefaultTimeout)
			if err != nil {
				log.Println("Error requesting post service endpoint:", err)
				return
			}
			log.Println("Response from service endpoint:", string(resp.Data))
		}
	}()

	time.Sleep(10 * time.Second)
}
