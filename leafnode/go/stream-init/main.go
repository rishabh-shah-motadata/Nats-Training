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
    nc, err := nats.Connect("nats://localhost:4222")
    if err != nil {
        log.Fatal(err)
    }
    defer nc.Close()

    // Create JetStream context
    js, err := jetstream.New(nc)
    if err != nil {
        log.Fatal(err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Create stream configuration
    streamCfg := jetstream.StreamConfig{
        Name:        "ORDERS",
        Description: "Stream for order events",
        Subjects:    []string{"orders.*"},
        Storage:     jetstream.FileStorage,
        Retention:   jetstream.LimitsPolicy,
        MaxAge:      24 * time.Hour,
        Discard:     jetstream.DiscardOld,
        Replicas:    3, // Replicate across all 3 cluster nodes
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
