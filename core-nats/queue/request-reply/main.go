package main

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	nc, _ := nats.Connect(nats.DefaultURL)
	defer nc.Drain()

	for i := 1; i <= 3; i++ {
		serverID := i
		sub, err := nc.QueueSubscribe("api.getUser", "api", func(msg *nats.Msg) {
			response := fmt.Sprintf(`{"server": %d, "user": "Alice"}`, serverID)
			msg.Respond([]byte(response))
		})
		if err != nil {
			fmt.Printf("Error subscribing to queue: %v\n", err)
			return
		}
		defer func(sub *nats.Subscription) {
			fmt.Println("serverID", serverID, "unsubscribing")
			sub.Unsubscribe()
		}(sub)
	}

	time.Sleep(100 * time.Millisecond)

	// Make 6 requests - load balanced
	for i := 1; i <= 6; i++ {
		response, err := nc.Request("api.getUser", []byte("{}"), 1*time.Second)
		if err != nil {
			fmt.Printf("Error making request %d: %v\n", i, err)
			continue
		}
		fmt.Printf("Request %d response: %s\n", i, string(response.Data))
	}
}
