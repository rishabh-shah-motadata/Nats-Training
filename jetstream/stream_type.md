# Stream Retention Policies: Technical Guide

This document explains the three primary retention policies for JetStream and the technical considerations for each.

---

## 1. Limits-Based Retention (LimitsPolicy)
The stream acts as a data buffer or message history. Retention is governed by physical or temporal boundaries rather than consumer state.

### Core Logic
Messages are removed when they violate a configured limit (Age, Bytes, or Count), regardless of whether they have been read or acknowledged.

### Critical Considerations
* **Discard Policy:** When a limit is reached, you must choose between `DiscardOld` (deletes the oldest message to make room) or `DiscardNew` (rejects incoming messages).
* **Storage Growth:** Without at least one limit (MaxAge, MaxBytes, or MaxMsgs), the stream will grow until it exhausts the disk or memory.
* **Replay Capability:** This is the only policy intended for "Time Machine" style replaying of historical data.

---

## 2. Interest-Based Retention (InterestPolicy)
The stream acts as a coordination point for a specific set of interested parties.

### Core Logic
A message is retained until it has been acknowledged by all active Durable Consumers. If an Ephemeral Consumer is connected, it is considered interested only while the connection is active.

### Critical Considerations
* **Consumer Management:** If a Durable Consumer is created and then abandoned (never deleted), the stream will stop deleting messages, leading to eventual disk exhaustion.
* **Limit Flags:** While MaxAge and MaxBytes can be added, they function as "hard overrides." If a limit is hit before a consumer ACKs, the message is deleted and the consumer misses the data.
* **Ephemeral Impact:** Ephemeral consumers do not "hold" messages in the stream once they disconnect. Only Durable consumers provide the "save point" guarantee.

---

## 3. Work Queue Retention (WorkQueuePolicy)
The stream acts as a distribution center where messages are treated as tasks.

### Core Logic
Once a message is acknowledged by any single consumer, it is immediately deleted from the stream.

### Critical Considerations
* **Exclusive Delivery:** The system ensures a message is processed by only one worker. If the worker fails to ACK, the message is redelivered.
* **Pull vs Push:** Pull consumers are highly recommended for Work Queues to prevent overwhelming workers and to manage flow control.
* **No History:** You cannot "replay" a Work Queue. Once the task is done, the data is gone.
* **Configuration Restriction:** You cannot have multiple consumers that receive the same message. Each consumer effectively "competes" for the next available task.

---

## Summary of Configuration Best Practices

| Feature | Limits-Based | Interest-Based | Work Queue |
| :--- | :--- | :--- | :--- |
| **Primary Goal** | History / Event Sourcing | Multi-user Coordination | Task Distribution |
| **Mandatory Flags** | At least one Limit | Durable Consumers | ACK explicit |
| **Redundancy (R)** | R1 to R5 | R3 recommended | R3 recommended |
| **MaxAge Usage** | Required for cleanup | Safety net only | Avoid (use ACK) |
| **Discard Policy** | Usually Old | Usually New | Usually New |

---

## Operational Warnings

1. **The Ghost Consumer Problem:** In Interest-based streams, always monitor your consumer list. A single offline Durable Consumer can cause the stream to grow infinitely.
2. **Limit Overrides:** Adding a 24-hour MaxAge to an Interest-based stream effectively turns off the "Guaranteed Delivery" if your systems are down for more than 24 hours.
3. **Storage Medium:** For Work Queues and Interest streams, use File storage. Memory storage will lose all "pending" work if the server node restarts.