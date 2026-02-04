package main

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func main() {
	fmt.Printf("=== NATS Connectivity Diagnostic ===\n\n")

	// Test 1: Connect to hub cluster
	fmt.Println("1️⃣  Testing HUB cluster connection...")
	testHub()

	time.Sleep(1 * time.Second)

	// Test 2: Connect to leaf node
	fmt.Println("\n2️⃣  Testing LEAF node connection...")
	testLeaf()

	time.Sleep(1 * time.Second)

	// Test 3: Test leaf → hub domain access
	fmt.Println("\n3️⃣  Testing LEAF → HUB domain access...")
	testLeafToHub()
}

func testHub() {
	nc, err := nats.Connect(
		"nats://localhost:4222",
		nats.UserInfo("app", "app"),
		nats.Timeout(5*time.Second),
	)
	if err != nil {
		fmt.Printf("   ❌ Failed to connect to hub: %v\n", err)
		return
	}
	defer nc.Close()

	fmt.Printf("   ✅ Connected to hub: %s\n", nc.ConnectedUrl())
	fmt.Printf("   Server ID: %s\n", nc.ConnectedServerId())

	// Test JetStream
	js, err := jetstream.NewWithDomain(nc, "hub")
	if err != nil {
		fmt.Printf("   ❌ JetStream error: %v\n", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// List streams
	fmt.Println("   Streams in hub domain:")
	count := 0
	for stream := range js.ListStreams(ctx).Info() {
		count++
		fmt.Printf("     • %s (%d messages)\n",
			stream.Config.Name,
			stream.State.Msgs)
	}

	if count == 0 {
		fmt.Println("     (no streams found - create ORDERS stream first)")
	}
}

func testLeaf() {
	nc, err := nats.Connect(
		"nats://localhost:4221",
		nats.UserInfo("app", "app"),
		nats.Timeout(5*time.Second),
	)
	if err != nil {
		fmt.Printf("   ❌ Failed to connect to leaf: %v\n", err)
		fmt.Println("   Is the leaf node running? Check: docker ps | grep leaf")
		return
	}
	defer nc.Close()

	fmt.Printf("   ✅ Connected to leaf: %s\n", nc.ConnectedUrl())
	fmt.Printf("   Server ID: %s\n", nc.ConnectedServerId())

	// Test local leaf JetStream
	js, err := jetstream.NewWithDomain(nc, "leaf")
	if err != nil {
		fmt.Printf("   ❌ Leaf JetStream error: %v\n", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("   Streams in leaf domain:")
	count := 0
	for stream := range js.ListStreams(ctx).Info() {
		count++
		fmt.Printf("     • %s (%d messages)\n",
			stream.Config.Name,
			stream.State.Msgs)
	}

	if count == 0 {
		fmt.Println("     (no local streams)")
	}
}

func testLeafToHub() {
	nc, err := nats.Connect(
		"nats://localhost:4221",
		nats.UserInfo("app", "app"),
		nats.Timeout(5*time.Second),
	)
	if err != nil {
		fmt.Printf("   ❌ Failed to connect to leaf: %v\n", err)
		return
	}
	defer nc.Close()

	fmt.Printf("   ✅ Connected to leaf: %s\n", nc.ConnectedUrl())

	// Try to access hub domain through leaf connection
	js, err := jetstream.NewWithDomain(nc, "hub")
	if err != nil {
		fmt.Printf("   ❌ Failed to create JetStream context for hub domain: %v\n", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("   Attempting to access hub domain streams via leaf...")

	streams := js.ListStreams(ctx)
	count := 0
	hasError := false

	for result := range streams.Info() {
		count++
		fmt.Printf("     ✅ Found: %s (%d messages)\n",
			result.Config.Name,
			result.State.Msgs)
	}

	if err := streams.Err(); err != nil {
		hasError = true
		fmt.Printf("   ❌ Error accessing hub domain: %v\n", err)
		fmt.Println("\n   Possible causes:")
		fmt.Println("   • Leaf node not connected to hub cluster")
		fmt.Println("   • Hub cluster not running")
		fmt.Println("   • Domain mismatch in configuration")
		fmt.Println("\n   Check leaf node logs: docker logs leafnode")
	}

	if count == 0 && !hasError {
		fmt.Println("     (no streams in hub - create ORDERS stream first)")
	}

	// Try to get ORDERS stream specifically
	fmt.Println("\n   Testing ORDERS stream access:")
	stream, err := js.Stream(ctx, "ORDERS")
	if err != nil {
		fmt.Printf("     ❌ Cannot access ORDERS stream: %v\n", err)
		fmt.Println("     Run this to create it:")
		fmt.Println("     go run main.go  (the stream creator)")
	} else {
		fmt.Printf("     ✅ ORDERS stream accessible!\n")
		fmt.Printf("     Messages: %d\n", stream.CachedInfo().State.Msgs)
		fmt.Printf("     Subjects: %v\n", stream.CachedInfo().Config.Subjects)
	}
}
