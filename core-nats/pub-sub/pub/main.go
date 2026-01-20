package main

import (
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := nats.Connect(
		"nats://localhost:4222,nats://localhost:4223,nats://localhost:4224",
		nats.Name("NATS Sample Publisher"),
		nats.DisconnectHandler(func(nc *nats.Conn) {
			log.Println("Publisher disconnected from NATS server")
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			log.Println("Publisher reconnected to NATS server")
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			log.Println("Publisher connection to NATS server closed")
		}),
	)
	if err != nil {
		log.Println("Error connecting to NATS server:", err)
		return
	}
	defer nc.Close()

	log.Println("Connected to:", nc.ConnectedUrl())

	for i := 1; i <= 5; i++ {
		msg := []byte("user created event #" + time.Now().Format(time.RFC3339))
		err := nc.Publish("events.user.created", msg)
		if err != nil {
			log.Println("publish failed:", err)
		} else {
			log.Println("published:", string(msg))
		}
		time.Sleep(1 * time.Second)
	}
}
