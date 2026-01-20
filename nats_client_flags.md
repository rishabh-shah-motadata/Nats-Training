# NATS Go Client Options - Complete Reference

> A comprehensive guide to all `nats.Option` functions for configuring NATS connections in Go.

---

## 📋 Table of Contents

- [Connection Options](#connection-options)
- [TLS/Security Options](#tlssecurity-options)
- [Authentication Options](#authentication-options)
- [Reconnection Options](#reconnection-options)
- [Event Handlers](#event-handlers)
- [Performance Tuning](#performance-tuning)
- [Advanced Options](#advanced-options)
- [Deprecated/Legacy Options](#deprecatedlegacy-options)
- [Production Configuration Template](#production-configuration-template)
- [Quick Reference](#quick-reference)

---

## Connection Options

### `nats.Name(string)`

| Attribute | Value |
|-----------|-------|
| **Default** | `""` (empty string) |
| **Purpose** | Sets a logical name for the connection, visible in server monitoring |
| **Production** | ✅ **Always set** - Use format: `{service}-{version}-{instance}` |

```go
nats.Name("payment-service-v1.2.3-pod-42")
```

---

### `nats.Timeout(time.Duration)`

| Attribute | Value |
|-----------|-------|
| **Default** | `2 * time.Second` |
| **Purpose** | Initial connection timeout (how long to wait for server handshake) |
| **Production** | Use default unless specific network requirements |

```go
nats.Timeout(5 * time.Second) // Slower networks
nats.Timeout(500 * time.Millisecond) // Local dev, fail-fast
```

---

### `nats.NoEcho()`

| Attribute | Value |
|-----------|-------|
| **Default** | Echo enabled (receives own messages) |
| **Purpose** | Prevents receiving messages you published yourself |
| **Production** | ⚠️ Use only if you never need to receive your own messages |

```go
nats.NoEcho() // Useful for avoiding message loops
```

**⚠️ Warning:** You won't receive messages on subjects you publish to, even if subscribed.

---

### `nats.DontRandomize()`

| Attribute | Value |
|-----------|-------|
| **Default** | Server list is randomized |
| **Purpose** | Connect to servers in order (don't shuffle) |
| **Production** | ❌ Usually don't use - randomization helps load balancing |

```go
nats.DontRandomize() // Connect in specified order
```

---

## TLS/Security Options

### `nats.Secure(tlsConfig *tls.Config)`

| Attribute | Value |
|-----------|-------|
| **Default** | None (no TLS) |
| **Purpose** | Enable TLS encryption (custom config) |
| **Production** | ✅ **Always use TLS in production** |

```go
tlsConfig := &tls.Config{MinVersion: tls.VersionTLS13}
nats.Secure(tlsConfig)
```

---

### `nats.TLSHandshakeFirst()`

| Attribute | Value |
|-----------|-------|
| **Default** | Disabled |
| **Purpose** | Perform TLS handshake before NATS protocol handshake |
| **Production** | ⚠️ Use only if required by server configuration |

```go
nats.TLSHandshakeFirst()
```

---

### `nats.ClientCert(certFile, keyFile string)`

| Attribute | Value |
|-----------|-------|
| **Default** | None |
| **Purpose** | Mutual TLS (mTLS) - client certificate authentication |
| **Production** | ✅ Use for high-security environments |

```go
nats.ClientCert("/path/to/client.crt", "/path/to/client.key")
```

---

### `nats.RootCAs(caCerts ...string)`

| Attribute | Value |
|-----------|-------|
| **Default** | System CA pool |
| **Purpose** | Add custom CA certificates for server verification |
| **Production** | ✅ Use when server uses self-signed or internal CA |

```go
nats.RootCAs("/path/to/ca.crt")
```

---

### `nats.ClientTLSConfig(tlsConfig *tls.Config)`

| Attribute | Value |
|-----------|-------|
| **Default** | None |
| **Purpose** | Full control over TLS configuration |
| **Production** | ✅ Use for complete TLS customization |

```go
nats.ClientTLSConfig(&tls.Config{
    MinVersion: tls.VersionTLS13,
    CipherSuites: []uint16{tls.TLS_AES_128_GCM_SHA256},
})
```

---

## Authentication Options

### `nats.UserInfo(user, password string)`

| Attribute | Value |
|-----------|-------|
| **Default** | None |
| **Purpose** | Username/password authentication |
| **Production** | ⚠️ Use only over TLS, prefer NKey for production |

```go
nats.UserInfo("admin", "password")
```

---

### `nats.Token(string)`

| Attribute | Value |
|-----------|-------|
| **Default** | None |
| **Purpose** | Simple token-based authentication |
| **Production** | ⚠️ Less secure than NKey, use only if server requires it |

```go
nats.Token("my-secret-token")
```

---

### `nats.Nkey(string, nkeys.KeyPair)`

| Attribute | Value |
|-----------|-------|
| **Default** | None |
| **Purpose** | Authenticate using NATS NKey (Ed25519 cryptographic authentication) |
| **Production** | ✅ **Use for production** - More secure than username/password |

```go
kp, _ := nkeys.FromSeed([]byte("SUAM..."))
nats.Nkey("UABCD...", kp)
```

---

### `nats.UserCredentials(credsFile string)`

| Attribute | Value |
|-----------|-------|
| **Default** | None |
| **Purpose** | Load NKey credentials from `.creds` file |
| **Production** | ✅ **Recommended** - Easy and secure |

```go
nats.UserCredentials("/path/to/user.creds")
```

---

### `nats.UserCredentialsBytes(credsData []byte)`

| Attribute | Value |
|-----------|-------|
| **Default** | None |
| **Purpose** | Load credentials from memory (instead of file) |
| **Production** | ✅ Use when credentials stored in secrets manager |

```go
credsData := fetchFromVault()
nats.UserCredentialsBytes(credsData)
```

---

### `nats.UserJWT(jwtCB, sigCB func)`

| Attribute | Value |
|-----------|-------|
| **Default** | None |
| **Purpose** | Advanced: Provide JWT and signature via callbacks |
| **Production** | ⚠️ Advanced use case, most should use `UserCredentials` |

```go
nats.UserJWT(
    func() (string, error) { return jwt, nil },
    func(nonce []byte) ([]byte, error) { return sign(nonce), nil },
)
```

---

## Reconnection Options

### `nats.MaxReconnects(int)`

| Attribute | Value |
|-----------|-------|
| **Default** | `60` |
| **Purpose** | Maximum reconnection attempts before giving up |
| **Production** | 10-20 typical, `-1` for infinite (with caution) |

```go
nats.MaxReconnects(10)    // Give up after 10 attempts
nats.MaxReconnects(-1)    // Never give up (use with monitoring)
```

---

### `nats.ReconnectWait(time.Duration)`

| Attribute | Value |
|-----------|-------|
| **Default** | `2 * time.Second` |
| **Purpose** | Delay between reconnection attempts |
| **Production** | 1-5s depending on load |

```go
nats.ReconnectWait(2 * time.Second)
```

---

### `nats.NoReconnect()`

| Attribute | Value |
|-----------|-------|
| **Default** | Reconnect enabled |
| **Purpose** | Disable automatic reconnection |
| **Production** | ❌ **Never use** - defeats NATS resilience |

```go
nats.NoReconnect() // Connection loss = permanent failure
```

---

### `nats.ReconnectJitter(time.Duration, time.Duration)`

| Attribute | Value |
|-----------|-------|
| **Default** | None |
| **Purpose** | Add random jitter to reconnect wait (prevent thundering herd) |
| **Production** | ✅ **Use in production** with many clients |

```go
// Wait 2-4 seconds randomly between reconnects
nats.ReconnectJitter(2*time.Second, 4*time.Second)
```

---

### `nats.CustomReconnectDelay(func(attempts int) time.Duration)`

| Attribute | Value |
|-----------|-------|
| **Default** | None |
| **Purpose** | Custom backoff strategy |
| **Production** | ⚠️ Use for advanced exponential backoff |

```go
nats.CustomReconnectDelay(func(attempts int) time.Duration {
    // Exponential backoff: 1s, 2s, 4s, 8s, max 30s
    delay := time.Duration(math.Pow(2, float64(attempts))) * time.Second
    if delay > 30*time.Second {
        return 30 * time.Second
    }
    return delay
})
```

---

### `nats.ReconnectBufSize(int)`

| Attribute | Value |
|-----------|-------|
| **Default** | `8388608` (8MB) |
| **Purpose** | Buffer size for outgoing messages during reconnection |
| **Production** | Increase for high-throughput systems |

```go
nats.ReconnectBufSize(16 * 1024 * 1024) // 16MB buffer
```

---

### `nats.RetryOnFailedConnect(bool)`

| Attribute | Value |
|-----------|-------|
| **Default** | `false` |
| **Purpose** | Retry connection if initial connect fails |
| **Production** | ✅ Use in production to handle startup race conditions |

```go
nats.RetryOnFailedConnect(true)
```

---

## Event Handlers

### `nats.DisconnectErrHandler(func(*Conn, error))`

| Attribute | Value |
|-----------|-------|
| **Default** | None |
| **Purpose** | Called when connection lost (before reconnection) |
| **Production** | ✅ **Always set** - Log and update health status |

```go
nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
    log.Warn("NATS disconnected:", err)
    healthCheck.SetUnhealthy()
})
```

---

### `nats.ReconnectHandler(func(*Conn))`

| Attribute | Value |
|-----------|-------|
| **Default** | None |
| **Purpose** | Called when reconnection successful |
| **Production** | ✅ **Always set** - Log and restore health status |

```go
nats.ReconnectHandler(func(nc *nats.Conn) {
    log.Info("NATS reconnected")
    healthCheck.SetHealthy()
})
```

---

### `nats.ClosedHandler(func(*Conn))`

| Attribute | Value |
|-----------|-------|
| **Default** | None |
| **Purpose** | Called when connection permanently closed |
| **Production** | ✅ **Always set** - Trigger shutdown or alert |

```go
nats.ClosedHandler(func(nc *nats.Conn) {
    log.Error("NATS connection closed permanently")
    os.Exit(1) // Force restart
})
```

---

### `nats.ErrorHandler(func(*Conn, *Subscription, error))`

| Attribute | Value |
|-----------|-------|
| **Default** | None |
| **Purpose** | Called on async errors (e.g., slow consumer) |
| **Production** | ✅ **Always set** - Critical for catching subscription issues |

```go
nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
    log.Error("NATS async error:", err, "subject:", sub.Subject)
})
```

---

### `nats.ConnectHandler(func(*Conn))`

| Attribute | Value |
|-----------|-------|
| **Default** | None |
| **Purpose** | Called when initial connection established |
| **Production** | ⚠️ Optional - Use for logging/metrics |

```go
nats.ConnectHandler(func(nc *nats.Conn) {
    log.Info("Connected to NATS:", nc.ConnectedUrl())
})
```

---

### `nats.DiscoveredServersHandler(func(*Conn))`

| Attribute | Value |
|-----------|-------|
| **Default** | None |
| **Purpose** | Called when client discovers new servers in cluster |
| **Production** | ⚠️ Optional - Use for debugging cluster topology |

```go
nats.DiscoveredServersHandler(func(nc *nats.Conn) {
    log.Info("Discovered servers:", nc.Servers())
})
```

---

### `nats.LameDuckModeHandler(func(*Conn))`

| Attribute | Value |
|-----------|-------|
| **Default** | None |
| **Purpose** | Server entering graceful shutdown (lame duck mode) |
| **Production** | ✅ Use to gracefully reconnect to other servers |

```go
nats.LameDuckModeHandler(func(nc *nats.Conn) {
    log.Warn("Server entering lame duck mode, will reconnect")
})
```

---

### `nats.DisconnectHandler(func(*Conn))`

| Attribute | Value |
|-----------|-------|
| **Default** | None |
| **Purpose** | Called on disconnect (no error info, deprecated) |
| **Production** | ❌ Use `DisconnectErrHandler` instead |

---

## Performance Tuning

### `nats.FlusherTimeout(time.Duration)`

| Attribute | Value |
|-----------|-------|
| **Default** | `1 * time.Minute` |
| **Purpose** | Max time to wait for flush operations |
| **Production** | Use default unless experiencing flush timeouts |

```go
nats.FlusherTimeout(30 * time.Second)
```

---

### `nats.PingInterval(time.Duration)`

| Attribute | Value |
|-----------|-------|
| **Default** | `2 * time.Minute` |
| **Purpose** | Interval between PING messages to server |
| **Production** | Use default (server detects dead connections faster) |

```go
nats.PingInterval(1 * time.Minute) // More frequent health checks
```

---

### `nats.MaxPingsOutstanding(int)`

| Attribute | Value |
|-----------|-------|
| **Default** | `2` |
| **Purpose** | Number of PINGs without PONG before declaring connection dead |
| **Production** | Use default |

```go
nats.MaxPingsOutstanding(3) // More tolerant of latency
```

---

### `nats.SubChanLen(int)`

| Attribute | Value |
|-----------|-------|
| **Default** | `65536` (64K messages) |
| **Purpose** | Default channel buffer size for async subscriptions |
| **Production** | Increase for high-throughput subscriptions |

```go
nats.SubChanLen(128 * 1024) // 128K buffer
```

---

### `nats.DrainTimeout(time.Duration)`

| Attribute | Value |
|-----------|-------|
| **Default** | `30 * time.Second` |
| **Purpose** | Timeout for graceful shutdown (Drain) |
| **Production** | ✅ Use longer timeout for critical services |

```go
nats.DrainTimeout(60 * time.Second)
```

---

## Advanced Options

### `nats.Dialer(customDialer)`

| Attribute | Value |
|-----------|-------|
| **Default** | `net.Dialer` |
| **Purpose** | Custom TCP dialer (e.g., for proxies, custom DNS) |
| **Production** | ⚠️ Use for SOCKS proxy or custom networking |

```go
nats.Dialer(&net.Dialer{
    Timeout:   10 * time.Second,
    KeepAlive: 30 * time.Second,
})
```

---

### `nats.SetCustomDialer(customDialer)`

| Attribute | Value |
|-----------|-------|
| **Default** | None |
| **Purpose** | Alternative to `Dialer` |
| **Production** | ⚠️ Use `Dialer` instead |

---

### `nats.Compression(bool)`

| Attribute | Value |
|-----------|-------|
| **Default** | `false` |
| **Purpose** | Enable message compression (requires server support) |
| **Production** | ⚠️ Use only if bandwidth is limited |

```go
nats.Compression(true) // Trade CPU for bandwidth
```

---

### `nats.ProxyPath(string)`

| Attribute | Value |
|-----------|-------|
| **Default** | None |
| **Purpose** | Connect through HTTP proxy |
| **Production** | ⚠️ Use when corporate proxy required |

```go
nats.ProxyPath("http://proxy.company.com:8080")
```

---

### `nats.SkipHostLookup()`

| Attribute | Value |
|-----------|-------|
| **Default** | Perform DNS lookups |
| **Purpose** | Skip DNS resolution (use IPs directly) |
| **Production** | ⚠️ Use only if DNS is unreliable |

---

### `nats.SkipSubjectValidation()`

| Attribute | Value |
|-----------|-------|
| **Default** | Validate subjects |
| **Purpose** | Skip client-side subject validation |
| **Production** | ❌ **Never use** - subjects must be valid |

---

### `nats.CustomInboxPrefix(string)`

| Attribute | Value |
|-----------|-------|
| **Default** | `"_INBOX"` |
| **Purpose** | Custom prefix for inbox subjects (request-reply) |
| **Production** | ⚠️ Rarely needed, use for multi-tenancy |

```go
nats.CustomInboxPrefix("_INBOX.tenant1")
```

---

### `nats.InProcessServer(server)`

| Attribute | Value |
|-----------|-------|
| **Default** | None |
| **Purpose** | Connect to in-process NATS server (testing) |
| **Production** | ❌ **Testing only** |

---

### `nats.SyncQueueLen(int)`

| Attribute | Value |
|-----------|-------|
| **Default** | `1024` |
| **Purpose** | Internal sync queue size |
| **Production** | Use default |

---

### `nats.NoCallbacksAfterClientClose()`

| Attribute | Value |
|-----------|-------|
| **Default** | Callbacks fire after close |
| **Purpose** | Prevent handlers from firing after `Close()` |
| **Production** | ⚠️ Use if handlers cause issues during shutdown |

---

### `nats.PermissionErrOnSubscribe()`

| Attribute | Value |
|-----------|-------|
| **Default** | Ignore permission errors on subscribe |
| **Purpose** | Return error immediately if subscribe not permitted |
| **Production** | ⚠️ Use for strict permission enforcement |

---

### `nats.WebsocketConnectionHeaders(headers)`

| Attribute | Value |
|-----------|-------|
| **Default** | None |
| **Purpose** | Custom headers for WebSocket connections |
| **Production** | ⚠️ Use only for WebSocket transport |

```go
nats.WebsocketConnectionHeaders(map[string][]string{
    "Authorization": {"Bearer token"},
})
```

---

### `nats.WebsocketConnectionHeadersHandler(handler)`

| Attribute | Value |
|-----------|-------|
| **Default** | None |
| **Purpose** | Dynamic WebSocket headers via callback |
| **Production** | ⚠️ Use for token refresh in WebSocket clients |

---

## Deprecated/Legacy Options

| Option | Status | Alternative |
|--------|--------|-------------|
| `nats.DisconnectHandler()` | ❌ Deprecated | Use `DisconnectErrHandler` (provides error info) |
| `nats.UseOldRequestStyle()` | ❌ Deprecated | Never use - legacy request-reply only |

---

## Production Configuration Template

```go
package main

import (
    "crypto/tls"
    "log"
    "os"
    "time"

    "github.com/nats-io/nats.go"
)

func main() {
    nc, err := nats.Connect(
        "nats://nats1.prod.company.com:4222,nats2.prod.company.com:4222", // or nats.DefaultURL
        
        // ========== Identity ==========
        nats.Name("payment-service-v1.2.3-pod-42"),
        
        // ========== Security ==========
        nats.UserCredentials("/secrets/nats.creds"),
        nats.RootCAs("/certs/ca.crt"),
        nats.Secure(&tls.Config{
            MinVersion: tls.VersionTLS13,
        }),
        
        // ========== Reconnection ==========
        nats.MaxReconnects(10),
        nats.ReconnectWait(2 * time.Second),
        nats.ReconnectJitter(1*time.Second, 3*time.Second),
        nats.RetryOnFailedConnect(true),
        
        // ========== Event Handlers ==========
        nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
            log.Printf("⚠️  NATS disconnected: %v", err)
            // Update metrics: connection_status = 0
        }),
        
        nats.ReconnectHandler(func(nc *nats.Conn) {
            log.Printf("✅ NATS reconnected to %s", nc.ConnectedUrl())
            // Update metrics: connection_status = 1
        }),
        
        nats.ClosedHandler(func(nc *nats.Conn) {
            log.Printf("❌ NATS permanently closed")
            // Update metrics: connection_status = -1
            os.Exit(1) // Force container restart
        }),
        
        nats.ErrorHandler(func(nc *nats.Conn, sub *nats.Subscription, err error) {
            log.Printf("🔥 NATS async error on %s: %v", sub.Subject, err)
            // Update metrics: errors_total++
        }),
        
        nats.LameDuckModeHandler(func(nc *nats.Conn) {
            log.Printf("🦆 Server entering lame duck mode")
            // Graceful reconnect will happen automatically
        }),
        
        // ========== Performance ==========
        nats.DrainTimeout(60 * time.Second),
        nats.PingInterval(90 * time.Second),
        nats.MaxPingsOutstanding(2),
    )
    
    if err != nil {
        log.Fatalf("Failed to connect to NATS: %v", err)
    }
    defer nc.Close()

    log.Printf("✅ Connected to NATS at %s", nc.ConnectedUrl())
    
    // Your application logic here
    select {}
}
```

---

## Quick Reference

### Must-Have Options for Production

| Priority | Option | Reason |
|----------|--------|--------|
| 🔴 **Critical** | `Name()` | Debugging & monitoring |
| 🔴 **Critical** | `DisconnectErrHandler()` | Know when connection lost |
| 🔴 **Critical** | `ReconnectHandler()` | Know when recovered |
| 🔴 **Critical** | `ClosedHandler()` | Handle permanent failure |
| 🔴 **Critical** | `ErrorHandler()` | Catch async errors (slow consumer, etc.) |
| 🟡 **Important** | `UserCredentials()` | Secure authentication |
| 🟡 **Important** | `Secure()` / `RootCAs()` | TLS encryption |
| 🟡 **Important** | `MaxReconnects()` | Control retry behavior |
| 🟡 **Important** | `RetryOnFailedConnect()` | Handle startup races |
| 🟢 **Optional** | `ReconnectJitter()` | Avoid thundering herd |
| 🟢 **Optional** | `DrainTimeout()` | Graceful shutdown |
| 🟢 **Optional** | `LameDuckModeHandler()` | Server graceful shutdown |

---

### Options by Category

| Category | Count | Options |
|----------|-------|---------|
| **Connection** | 4 | `Name`, `Timeout`, `NoEcho`, `DontRandomize` |
| **TLS/Security** | 5 | `Secure`, `TLSHandshakeFirst`, `ClientCert`, `RootCAs`, `ClientTLSConfig` |
| **Authentication** | 6 | `UserInfo`, `Token`, `Nkey`, `UserCredentials`, `UserCredentialsBytes`, `UserJWT` |
| **Reconnection** | 7 | `MaxReconnects`, `ReconnectWait`, `NoReconnect`, `ReconnectJitter`, `CustomReconnectDelay`, `ReconnectBufSize`, `RetryOnFailedConnect` |
| **Event Handlers** | 8 | `DisconnectErrHandler`, `ReconnectHandler`, `ClosedHandler`, `ErrorHandler`, `ConnectHandler`, `DiscoveredServersHandler`, `LameDuckModeHandler`, `DisconnectHandler` |
| **Performance** | 5 | `FlusherTimeout`, `PingInterval`, `MaxPingsOutstanding`, `SubChanLen`, `DrainTimeout` |
| **Advanced** | 12 | `Dialer`, `SetCustomDialer`, `Compression`, `ProxyPath`, `SkipHostLookup`, `SkipSubjectValidation`, `CustomInboxPrefix`, `InProcessServer`, `SyncQueueLen`, `NoCallbacksAfterClientClose`, `PermissionErrOnSubscribe`, `WebsocketConnectionHeaders`, `WebsocketConnectionHeadersHandler` |

---

## Additional Resources

- [NATS Go Client Documentation](https://pkg.go.dev/github.com/nats-io/nats.go)
- [NATS Go Github](https://github.com/nats-io/nats.go)

---

**Version:** NATS Go Client v1.31.0+  
**Last Updated:** January 2026