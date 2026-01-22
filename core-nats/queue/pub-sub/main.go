package main

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		fmt.Printf("Error connecting to NATS server: %v\n", err)
		return
	}
	defer nc.Close()

	var count1, count2, count3 int64

	sub, err := nc.QueueSubscribe("jobs", "workers", func(msg *nats.Msg) {
		atomic.AddInt64(&count1, 1)
	})
	if err != nil {
		fmt.Printf("Error subscribing to queue: %v\n", err)
		return
	}
	defer sub.Unsubscribe()

	sub, err = nc.QueueSubscribe("jobs", "workers", func(msg *nats.Msg) {
		atomic.AddInt64(&count2, 1)
	})
	if err != nil {
		fmt.Printf("Error subscribing to queue: %v\n", err)
		return
	}
	defer sub.Unsubscribe()

	sub, err = nc.QueueSubscribe("jobs", "workers", func(msg *nats.Msg) {
		atomic.AddInt64(&count3, 1)
	})
	if err != nil {
		fmt.Printf("Error subscribing to queue: %v\n", err)
		return
	}
	defer sub.Unsubscribe()

	time.Sleep(100 * time.Millisecond)

	for range 1000 {
		err := nc.Publish("jobs", []byte("job"))
		if err != nil {
			fmt.Printf("Error publishing message: %v\n", err)
		}
	}
	err = nc.Flush()
	if err != nil {
		fmt.Printf("Error flushing messages: %v\n", err)
	}

	time.Sleep(1 * time.Second)

	fmt.Printf("Worker 1: %d messages (%.1f%%)\n", count1, float64(count1)/10)
	fmt.Printf("Worker 2: %d messages (%.1f%%)\n", count2, float64(count2)/10)
	fmt.Printf("Worker 3: %d messages (%.1f%%)\n", count3, float64(count3)/10)
}
