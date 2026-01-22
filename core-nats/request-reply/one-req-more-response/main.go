package main

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	nc, _ := nats.Connect(nats.DefaultURL)
	defer nc.Close()

	// Instance 1
	nc.Subscribe("health.check", func(msg *nats.Msg) {
		msg.Respond([]byte(`{"instance": "server-1", "status": "healthy"}`))
	})

	// Instance 2
	nc.Subscribe("health.check", func(msg *nats.Msg) {
		msg.Respond([]byte(`{"instance": "server-2", "status": "healthy"}`))
	})

	// Instance 3
	nc.Subscribe("health.check", func(msg *nats.Msg) {
		msg.Respond([]byte(`{"instance": "server-3", "status": "healthy"}`))
	})

	time.Sleep(100 * time.Millisecond)

	// Collect ALL responses
	inbox := nc.NewInbox()
	sub, _ := nc.SubscribeSync(inbox)
	defer sub.Unsubscribe()

	// Send request
	nc.PublishMsg(&nats.Msg{
		Subject: "health.check",
		Reply:   inbox,
	})

	// Collect responses for 500ms
	responses := []string{}
	deadline := time.Now().Add(500 * time.Millisecond)

	for time.Now().Before(deadline) {
		msg, err := sub.NextMsg(100 * time.Millisecond)
		if err == nats.ErrTimeout {
			break // No more responses
		}
		if err == nil {
			responses = append(responses, string(msg.Data))
		}
	}

	fmt.Printf("Received %d responses:\n", len(responses))
	for _, resp := range responses {
		fmt.Printf("  - %s\n", resp)
	}
}
