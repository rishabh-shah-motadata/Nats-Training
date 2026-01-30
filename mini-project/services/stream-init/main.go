package main

import (
	"context"
	"log"
	"nats-project/internal/nats"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

func main() {
	nc, err := nats.InitNATS("Jetstream-Initialization")
	if err != nil {
		log.Println("error initializing NATS connection:", err)
		return
	}
	defer nc.Drain()
	log.Println("connected to NATs server", nc.ConnectedUrl())

	js, err := jetstream.New(nc)
	if err != nil {
		log.Println("error creating JetStream context:", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	streamConfig := jetstream.StreamConfig{
		Name:         "ORDERS",
		Subjects:     []string{"orders.*"},
		Storage:      jetstream.FileStorage,
		Retention:    jetstream.WorkQueuePolicy,
		MaxMsgs:      1000,
		MaxAge:       time.Minute * 30,
		Discard:      jetstream.DiscardOld,
		MaxConsumers: 3,
	}

	stream, err := js.CreateStream(ctx, streamConfig)
	if err != nil {
		log.Println("error creating stream:", err)
		return
	}

	streamInfo, err := stream.Info(ctx)
	if err != nil {
		log.Println("error fetching stream info:", err)
		return
	}
	log.Println("stream created successfully:", streamInfo.Config.Name)

	consumer, err := js.CreateConsumer(ctx, streamConfig.Name, jetstream.ConsumerConfig{
		Durable:       "ORDER_CONSUMER",
		AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject: "orders.created",
		MaxAckPending: 5,
		AckWait:       time.Second * 1,
		MaxDeliver:    2,
		ReplayPolicy:  jetstream.ReplayInstantPolicy,
	})
	if err != nil {
		log.Println("error creating consumer:", err)
		return
	}

	consumerInfo, err := consumer.Info(ctx)
	if err != nil {
		log.Println("error fetching consumer info:", err)
		return
	}
	log.Println("consumer created successfully:", consumerInfo.Name)
}
