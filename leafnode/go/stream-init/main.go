package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func main() {
	// Connect to NATS cluster
	nc, err := nats.Connect("nats://localhost:4222", nats.UserInfo("app", "app"))
	if err != nil {
		log.Fatal("Failed to connect with Nats: ", err)
	}
	defer nc.Drain()

	// Create JetStream context
	js, err := jetstream.NewWithDomain(nc, "hub")
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create stream configuration
	streamCfg := jetstream.StreamConfig{
		Name:        "ORDERS",
		Description: "Stream for order events",
		Subjects:    []string{"orders.created"},
		Storage:     jetstream.FileStorage,
		Replicas:    3,
		Retention:   jetstream.LimitsPolicy,
		Discard:     jetstream.DiscardOld,
		MaxMsgs:     -1, // No limit on number of messages
		MaxBytes:    -1, // No limit on total bytes
		MaxAge:      1 * time.Hour,
		MaxMsgSize:  1024 * 1024, // 1MB
		Duplicates:  2 * time.Minute,
		AllowRollup: false,
		DenyDelete:  false,
		DenyPurge:   false,
	}

	// Create or update stream
	stream, err := js.CreateOrUpdateStream(ctx, streamCfg)
	if err != nil {
		log.Fatalf("Failed to create stream: %v", err)
	}

	fmt.Printf("✅ Stream created: %s\n", stream.CachedInfo().Config.Name)
	fmt.Printf("   Subjects: %v\n", stream.CachedInfo().Config.Subjects)
	fmt.Printf("   Replicas: %d\n", stream.CachedInfo().Config.Replicas)
	fmt.Printf("   Storage: %s\n", stream.CachedInfo().Config.Storage)
}

/*
```

**Stream Configuration Explanation:**

- **`Name: "ORDERS"`**: Unique stream identifier
- **`Subjects: ["orders.*"]`**: Wildcard - captures `orders.created`, `orders.updated`, etc.
- **`Storage: FileStorage`**: Persist to disk (survives restarts)
- **`Retention: LimitsPolicy`**: Keep messages until limits reached (age, size, count)
- **`Replicas: 3`**: Store copies on all 3 cluster nodes for HA
- **`MaxAge: 24h`**: Auto-delete messages older than 24 hours

**Where Replicas Live:**
```
nats-1: [ORDERS Replica 1] ──┐
                               │
nats-2: [ORDERS Replica 2] ──┤ ├─> Raft consensus ensures consistency
                               │
nats-3: [ORDERS Replica 3] ──┘

*/
