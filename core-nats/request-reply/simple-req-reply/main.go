package main

import (
	"log"

	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := nats.Connect(
		"nats://localhost:4222,nats://localhost:4223,nats://localhost:4224",
		nats.Name("NATS Request Reply Tutorial"),
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

	sub, err := nc.Subscribe("test", func(msg *nats.Msg) {
		log.Println("Received request:", string(msg.Data))
		response := []byte(`{"name": "Alice", "age": 30}`)
		msg.Respond(response)
		/*
			You should not use msg.Respond if you don't want to send a response back to the requester.
			Using msg.Respond when no response is expected can lead to unnecessary resource usage
			and potential confusion in the communication flow.

			Also nc.Publish(msg.Reply, response) can be used instead of msg.Respond(response)
			But msg.Respond is preferred as it is more efficient and directly tied to the request message.
			
		*/
	})
	if err != nil {
		log.Println("Error subscribing to subject:", err)
		return
	}
	defer sub.Unsubscribe()

	log.Println("Subscribed to 'test' subject, waiting for requests...")

	resp, err := nc.Request("test", []byte(`{id: 42}`), nats.DefaultTimeout)
	if err != nil {
		log.Println("Error sending request:", err)
		return
	}
	log.Println("Received response:", string(resp.Data))
}
