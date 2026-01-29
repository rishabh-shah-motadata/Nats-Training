# JetStream Microservices Example

This repository demonstrates a **production-style setup** of NATS JetStream with microservices. It includes **stream initialization**, **publishing**, and **consuming messages** with proper durability and replication.

---

## Folder Structure

```
jetstream/
 ├── deploy/
 │     └── docker-compose.yml       # NATS cluster with JetStream enabled
 ├── stream-init/
 │     └── main.go                  # One-time stream setup
 ├── publisher/
 │     └── main.go                  # Microservice publishing messages
 └── consumer/
       └── consumer.go              # Microservice consuming messages
```

---

## How It Works

1. **Stream Initialization (`stream-init/main.go`)**

   * Run **once** per environment (dev/test/prod).
   * Defines the **stream configuration**: name, subjects, storage, retention, discard policy, replicas, and limits.
   * It also defines the consumer configuration which will be bind by consumer workers.
   * Example:

     ```text
     Stream: ORDERS
     Subjects: orders.*
     Storage: FileStorage
     Replicas: 3
     Retention: LimitsPolicy
     MaxAge: 24h
     MaxMsgs: 1_000_000
     Discard: DiscardOld
     ```
   * Ensures **durable and replicated storage** before any microservice publishes messages.

2. **Publisher (`publisher/main.go`)**

   * Publishes messages to **subjects** (e.g., `orders.created`).
   * JetStream automatically **matches the subject to the stream** and persists the message according to the stream configuration.
   * Publishers **do not create streams**; they only send messages.
   * Always check **publish ACKs** to ensure durability.

3. **Consumer (`consumer/consumer.go`)**

   * Reads messages from the stream via **JetStream consumers**.
   * Can be **durable or ephemeral**.
   * Responsible for **acknowledging processed messages**, controlling redelivery, and optionally replaying historical messages.

---

## Usage

### 1. Start NATS Cluster

```bash
docker-compose -f deploy/docker-compose.yml up -d
```

### 2. Run Stream Initialization (once)

```bash
go run stream-init/main.go
```

### 3. Start Publisher

```bash
go run publisher/main.go
```

### 4. Start Consumer

```bash
go run consumer/consumer.go
```

---

## Production Best Practices

* **Streams are created once** and managed separately from microservices.
* Publishers are **stateless** and only know subjects.
* **FileStorage + Replicas** ensures data durability and high availability.
* **Retention and Discard policies** prevent unbounded storage growth.
* **Separate consumers** handle message processing, ACKs, and replay logic.
* CI/CD should manage environment-specific stream configurations.

---

## Notes

* Do not create streams inside business microservices.
* Subject names are **versioned** for backward compatibility.
* Use monitoring (`:8222` HTTP port) to track stream health, replication, and storage usage.
