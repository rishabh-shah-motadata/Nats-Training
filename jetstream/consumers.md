# Consumer Patterns: Durable vs. Ephemeral

Understanding the lifecycle of a consumer is critical for building resilient messaging systems. Use this guide to choose the right strategy for your workload.

---

## Durable Consumer
**The "Saved Game"**

A Durable consumer is used when the server must remember the state of the consumer even if the client disconnects.

* **Named Identity:** Requires a specific name (Durable Name) so the server can track it.
* **Stateful:** The server stores the **Last Acknowledged Sequence** and **Redelivery Counts**.
* **Resiliency:** If your app crashes, restarts, or loses network, it picks up exactly where it left off.
* **Lifecycle:** Must be explicitly deleted by the user; it survives client disconnects.

### Best For:
* **Order Processing:** Every order must be processed.
* **Payments:** Financial integrity is a priority.
* **Critical Workflows:** Any logic that cannot afford "gaps" in data.

---

## Ephemeral Consumer
**The "Live Stream"**

An Ephemeral consumer is a temporary "window" into a stream. It is created on the fly and destroyed automatically.

* **Anonymous:** The server typically auto-generates a name.
* **Stateless:** No memory of past sessions. Every connection starts fresh.
* **Automatic Cleanup:** The moment the client disconnects or the heartbeat stops, the server deletes the consumer.
* **Short-lived:** Exists only for the duration of the active connection.

### Best For:
* **Debugging:** Tailing logs or monitoring traffic in real-time.
* **Ad-hoc Analytics:** Running a quick query on the current stream.
* **Sidecars:** Temporary workers that handle non-critical, transient data.

---

## Comparison Matrix

| Feature | Durable Consumer | Ephemeral Consumer |
| :--- | :--- | :--- |
| **Durable Name** | Required | Auto-generated |
| **Server Persistence** | Yes | No |
| **Survives Restart** | Yes | No |
| **Auto-deletion** | Manual/Admin | Yes (on disconnect) |
| **Pull Support** | Yes | Yes |

---

## Summary
* **Durable = Reliability.** Use this when the business depends on the message.
* **Ephemeral = Convenience.** Use this when you are just "watching" or performing temporary tasks.