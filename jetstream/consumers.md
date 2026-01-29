# Consumer Patterns: Lifecycle and Delivery Models

Understanding the combination of consumer lifecycles (Durable vs. Ephemeral) and delivery models (Pull vs. Push) is critical for building resilient messaging systems.

---

## 1. Lifecycle: Durable vs. Ephemeral

### Durable Consumer
**The "Saved Game"**
Used when the server must remember the state of the consumer even if the client disconnects.

* **Named Identity:** Requires a specific name so the server can track it.
* **Stateful:** The server stores the Last Acknowledged Sequence and Redelivery Counts.
* **Resiliency:** If your app crashes or restarts, it picks up exactly where it left off.
* **Lifecycle:** Must be explicitly deleted; it survives client disconnects.

### Ephemeral Consumer
**The "Live Stream"**
A temporary window into a stream that is created on the fly and destroyed automatically.

* **Anonymous:** The server typically auto-generates a name.
* **Stateless:** No memory of past sessions. Every connection starts fresh.
* **Automatic Cleanup:** The server deletes the consumer once the client disconnects.
* **Short-lived:** Exists only for the duration of the active connection.

---

## 2. Delivery Models: Pull vs. Push

### Pull-Based Consumer
**The "On-Demand" Model**
In this model, the client is in control. The client explicitly asks the server for a batch of messages.

* **Flow Control:** The client determines the rate of processing. If the client is slow, it simply does not ask for more messages.
* **Batching:** Clients can request multiple messages at once (e.g., "Fetch 10 messages").
* **Scaling:** Ideal for horizontally scaling workers. Multiple instances can pull from the same consumer without being overwhelmed.
* **Best For:** High-volume processing, large-scale worker clusters, and resource-intensive tasks.

### Push-Based Consumer
**The "Radio Broadcast" Model**
The server is in control. As soon as a message arrives in the stream, the server pushes it to the client.

* **Latency:** Offers the lowest possible latency because the server does not wait for a request.
* **Simplicity:** Easier to implement for simple applications or real-time notifications.
* **Risk of Overload:** If the server pushes faster than the client can process, the client can become overwhelmed unless flow control is specifically configured.
* **Best For:** Real-time dashboards, low-latency alerts, and simple monitoring tools.

---

## 3. Combined Comparison Matrix

| Feature | Durable | Ephemeral | Pull-Based | Push-Based |
| :--- | :--- | :--- | :--- | :--- |
| **Identification** | Named | Anonymous | Usually Durable | Often Ephemeral |
| **Control** | Server (State) | N/A | Client (Flow) | Server (Rate) |
| **Persistence** | Yes | No | Managed by ACK | Managed by ACK |
| **Primary Benefit** | Reliability | Convenience | Scalability | Low Latency |

---

## 4. Selection Logic

### Use Durable + Pull when:
* You are building a production-grade microservice.
* You need to process large volumes of data reliably.
* You want to scale workers up and down based on demand.

### Use Durable + Push when:
* You have a single dedicated worker that must handle events in near real-time.
* Message volume is predictable and unlikely to overwhelm the consumer.

### Use Ephemeral + Pull when:
* You need to sample a specific batch of data for analysis.
* You are running a temporary job that needs to process data at its own pace.

### Use Ephemeral + Push when:
* You are tailing logs for debugging purposes.
* You are building a real-time UI notification system where missing old messages during downtime is acceptable.

---

## Summary
* **Durable/Ephemeral** defines **how long** the server remembers the consumer state.
* **Pull/Push** defines **how** the messages are delivered to the application.