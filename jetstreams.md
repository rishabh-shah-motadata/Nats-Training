# NATS JetStream – Core Concepts (Beginner to Intermediate)

This document explains the **core JetStream concepts** you must understand before writing code. It is written assuming you are using the **latest NATS + JetStream versions** (no legacy APIs, no deprecated patterns).

---

## 1. Types of Streams

A **Stream** is a durable storage layer that captures messages published on one or more subjects.

### 1.1 Stream Based on Retention Policy

JetStream does not define streams as “types” explicitly, but **retention policy** effectively defines how the stream behaves.

| Stream Style              | Description                                             | Typical Use Case                           |
| ------------------------- | ------------------------------------------------------- | ------------------------------------------ |
| **Limits-based stream**   | Stores messages until size/age/count limits are reached | Event sourcing, audit logs, history replay |
| **WorkQueue stream**      | Messages are removed once consumed successfully         | Task queues, job processing                |
| **Interest-based stream** | Messages kept only while consumers exist                | Temporary subscriptions, fan-out systems   |

---

## 2. Retention Policies

Retention determines **when messages are deleted from a stream**.

### 2.1 Limits Policy (`LimitsPolicy`)

* Default retention policy
* Messages remain until one of the configured limits is hit

Limits can be:

* Max number of messages
* Max total bytes
* Max age (time-based)

✅ Best for: logs, event streams, replayable history

---

### 2.2 WorkQueue Policy (`WorkQueuePolicy`)

* Each message is delivered to **only one consumer**
* Message is deleted after it is **acknowledged**

Rules:

* Only **one consumer per subject filter**
* Requires **explicit acknowledgements**

✅ Best for: background jobs, async processing

---

### 2.3 Interest Policy (`InterestPolicy`)

* Messages are stored **only while consumers exist**
* Once all consumers have acknowledged a message, it is removed

✅ Best for: transient processing, live fan-out with durability

---

## 3. Discard Policies

Discard policy defines **what happens when a stream hits its limits**.

### 3.1 Discard Old (`DiscardOld`)

* Oldest messages are removed first
* New messages are always accepted

This is the **default behavior**.

---

### 3.2 Discard New (`DiscardNew`)

* New messages are rejected once limits are reached
* Publisher receives an error

Used when **losing old data is unacceptable**.

---

## 4. Types of Consumers

A **Consumer** reads messages from a stream and tracks delivery state.

### 4.1 Durable Consumers

* Have a **stable name**
* Delivery state is preserved across restarts
* Can resume from where they left off

✅ Use when: reliability and restart recovery matter

---

### 4.2 Ephemeral Consumers

* Created automatically
* Deleted when the client disconnects
* No long-term state

✅ Use when: temporary or short-lived processing

---

### 4.3 Push vs Pull Consumers

| Type              | Description                        | Best Use                              |
| ----------------- | ---------------------------------- | ------------------------------------- |
| **Push Consumer** | Server pushes messages to client   | Simple consumers, low volume          |
| **Pull Consumer** | Client explicitly fetches messages | High throughput, backpressure control |

⚠️ Pull consumers are **recommended** for scalable systems.

---

## 5. Persistence

JetStream supports two storage backends:

### 5.1 Memory Storage (`MemoryStorage`)

* Fastest
* Data lost on server restart

✅ Use when: speed > durability

---

### 5.2 File Storage (`FileStorage`)

* Messages stored on disk
* Survive restarts
* Can be replicated

✅ Use when: durability and recovery are required

---

## 6. Replay Policies

Replay policy controls **how stored messages are delivered** to consumers.

### 6.1 Instant Replay (`ReplayInstant`)

* Messages are delivered as fast as possible
* Ignores original publish timing

---

### 6.2 Original Replay (`ReplayOriginal`)

* Messages are replayed respecting original timestamps
* Preserves real-time spacing

---

## 7. Delivery Guarantees

JetStream provides **at-least-once delivery by default**.

### 7.1 At-Least-Once Delivery

* Message is resent if acknowledgment is not received
* Duplicates are possible

Client **must be idempotent**.

---

### 7.2 Exactly-Once Semantics (Practical)

JetStream supports effectively-once processing by:

* Using **message de-duplication** (`Msg-Id` header)
* Explicit acknowledgements
* Consumer state tracking

⚠️ Exactly-once requires **application-level discipline**.

---

## 8. Mental Model Summary

* **Stream** = durable message log
* **Consumer** = stateful reader
* **Retention** = when data is deleted
* **Discard** = behavior at capacity
* **Persistence** = where data lives
* **Replay** = how history is delivered
* **Guarantees** = reliability semantics

---
