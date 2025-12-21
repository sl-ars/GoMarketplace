# Rate Limiting

## Overview

GoMarketplace implements a Redis-based sliding window rate limiting system to protect the API from abuse, brute force attacks, and ensure fair usage across all clients.

## Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   HTTP Request  │────▶│  Rate Limit     │────▶│   Handler       │
│                 │     │  Middleware     │     │                 │
└─────────────────┘     └────────┬────────┘     └─────────────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │     Redis       │
                        │  (Sorted Sets)  │
                        └─────────────────┘
```

## Algorithm

We use the **Sliding Window** algorithm implemented with Redis Sorted Sets:

1. Each request is stored with its timestamp as score
2. Old entries outside the window are removed
3. Count of entries in window determines if request is allowed
4. Provides accurate rate limiting without the burst issues of fixed windows

### Why Sliding Window?

| Algorithm | Pros | Cons |
|-----------|------|------|
| Fixed Window | Simple | Allows 2x burst at window boundaries |
| Sliding Log | Accurate | Memory intensive |
| **Sliding Window** | Accurate, efficient | Slightly more complex |
| Token Bucket | Smooth | Harder to implement distributed |

## Rate Limit Tiers

### Authentication Endpoints (Strict)
```
Limit: 5 requests per minute
Endpoints: /api/auth/login, /api/auth/register, /api/auth/refresh
Rationale: Prevent brute force and credential stuffing attacks
```

### Public Endpoints
```
Limit: 60 requests per minute
Endpoints: /api/products, /api/products/{id}
Rationale: Allow reasonable browsing while preventing scraping
```

### Authenticated Endpoints (Standard)
```
Limit: 120 requests per minute
Endpoints: All other /api/* endpoints
Rationale: Higher limit for authenticated users
```

### Excluded Endpoints
```
/api/health - Health checks (monitoring)
/api/webhook/stripe - Stripe handles its own retries
```

## Configuration

### Environment Variables

```env
# Enable/disable rate limiting
RATELIMIT_ENABLED=true

# Requests per minute for auth endpoints (login, register)
RATELIMIT_AUTH_REQUESTS=5

# Requests per minute for public endpoints
RATELIMIT_PUBLIC_REQUESTS=60

# Requests per minute for authenticated endpoints
RATELIMIT_STANDARD_REQUESTS=120

# Trusted proxy IPs/CIDRs (comma-separated)
# REQUIRED for X-Forwarded-For to be trusted!
RATELIMIT_TRUSTED_PROXIES=10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,127.0.0.1
```

### Preset Configurations

```go
// Available in pkg/ratelimit/limiter.go
var (
    AuthLimit     = Config{Requests: 5, Window: time.Minute}
    PublicLimit   = Config{Requests: 60, Window: time.Minute}
    StandardLimit = Config{Requests: 120, Window: time.Minute}
    StrictLimit   = Config{Requests: 10, Window: time.Minute}
    RelaxedLimit  = Config{Requests: 300, Window: time.Minute}
)
```

## Response Headers

Every API response includes rate limit information:

```http
X-RateLimit-Limit: 120
X-RateLimit-Remaining: 115
X-RateLimit-Reset: 1703185200
```

| Header | Description |
|--------|-------------|
| `X-RateLimit-Limit` | Maximum requests allowed in the window |
| `X-RateLimit-Remaining` | Requests remaining in current window |
| `X-RateLimit-Reset` | Unix timestamp when the window resets |

## Rate Limited Response

When rate limit is exceeded, the API returns:

```http
HTTP/1.1 429 Too Many Requests
Content-Type: application/json
Retry-After: 45
X-RateLimit-Limit: 5
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1703185245

{
  "success": false,
  "message": "Too many requests",
  "error": {
    "code": 429,
    "details": "Rate limit exceeded. Please slow down and try again later."
  }
}
```

## Key Generation Strategy

Rate limit keys are generated based on:

### Unauthenticated Requests
```
Key: ratelimit:ip:{client_ip}:{endpoint}
Example: ratelimit:ip:192.168.1.1:/api/auth/login
```

### Authenticated Requests
```
Key: ratelimit:user:{user_id}:{endpoint}
Example: ratelimit:user:12345:/api/orders
```

### Secure IP Detection

**IMPORTANT**: IP detection is security-critical. Naive implementations can be bypassed!

#### The Problem

Attackers can set `X-Forwarded-For: fake-ip` headers to bypass rate limiting:

```bash
# Attacker bypasses rate limit by faking different IPs
curl -H "X-Forwarded-For: 1.1.1.1" https://api.example.com/login
curl -H "X-Forwarded-For: 2.2.2.2" https://api.example.com/login
# ... infinite requests with different fake IPs
```

#### The Solution

We ONLY trust `X-Forwarded-For` / `X-Real-IP` headers if:
1. The direct connection (`RemoteAddr`) comes from a **trusted proxy**
2. The trusted proxy is explicitly configured via `RATELIMIT_TRUSTED_PROXIES`

```
Direct Request (no proxy):
  RemoteAddr: 203.0.113.50 → Use this IP (cannot be spoofed)

Request via Trusted Proxy:
  RemoteAddr: 10.0.0.1 (trusted)
  X-Forwarded-For: 203.0.113.50, 10.0.0.1
  → Walk backwards, find first untrusted IP: 203.0.113.50

Request via Untrusted Proxy (attacker):
  RemoteAddr: 203.0.113.50 (NOT trusted)
  X-Forwarded-For: fake-ip (spoofed)
  → Ignore headers, use RemoteAddr: 203.0.113.50
```

#### Configuration Examples

```env
# Docker/Kubernetes internal networks
RATELIMIT_TRUSTED_PROXIES=10.0.0.0/8,172.16.0.0/12,192.168.0.0/16

# Single nginx proxy
RATELIMIT_TRUSTED_PROXIES=192.168.1.100

# AWS ALB IP ranges (check AWS docs for current ranges)
RATELIMIT_TRUSTED_PROXIES=10.0.0.0/8

# Cloudflare (use their published IP ranges)
RATELIMIT_TRUSTED_PROXIES=173.245.48.0/20,103.21.244.0/22,...
```

#### No Trusted Proxies = Maximum Security

If `RATELIMIT_TRUSTED_PROXIES` is empty, headers are **never trusted**:

```
WARN: No trusted proxies configured - X-Forwarded-For headers will be ignored
```

This is the safest default for applications directly exposed to the internet.

## Usage Examples

### Apply to Router (Global)

```go
// All routes under this router get rate limiting
api := r.PathPrefix("/api").Subrouter()
api.Use(middleware.RateLimitMiddleware(rlConfig, ratelimit.StandardLimit))
```

### Apply to Subrouter (Group)

```go
// Auth routes get stricter limits
authRouter := api.PathPrefix("/auth").Subrouter()
authRouter.Use(middleware.RateLimitByEndpoint(rlConfig, ratelimit.AuthLimit))
```

### Apply to Single Handler

```go
// Single endpoint with custom limit
api.HandleFunc("/sensitive-operation",
    middleware.RateLimitHandler(rlConfig, ratelimit.Config{
        Requests: 3,
        Window:   time.Hour,
    }, sensitiveHandler))
```

### Custom Rate Limit Configuration

```go
// Define custom limits
customLimit := ratelimit.Config{
    Requests: 10,
    Window:   5 * time.Minute,
}

router.Use(middleware.RateLimitMiddleware(rlConfig, customLimit))
```

## Adding Rate Limiting to New Endpoints

### Option 1: Automatic (Recommended)

New endpoints under `/api/*` automatically inherit the global rate limit:

```go
api.HandleFunc("/new-feature", newFeatureHandler).Methods("GET")
// Automatically gets StandardLimit (120/min)
```

### Option 2: Custom Limit for Sensitive Operations

```go
// Create subrouter with stricter limit
sensitive := api.PathPrefix("/admin").Subrouter()
sensitive.Use(middleware.RateLimitMiddleware(rlConfig, ratelimit.StrictLimit))
sensitive.HandleFunc("/dangerous-action", dangerousHandler)
```

### Option 3: Per-Endpoint Limit

```go
api.HandleFunc("/expensive-operation",
    middleware.RateLimitHandler(rlConfig, ratelimit.Config{
        Requests: 5,
        Window:   time.Minute,
    }, expensiveHandler))
```

## Redis Storage

### Key Structure
```
marketplace:ratelimit:ip:192.168.1.1:global
marketplace:ratelimit:user:12345:/api/orders
```

### Data Structure
Redis Sorted Set with:
- **Member**: Request timestamp (nanoseconds)
- **Score**: Same timestamp for range queries

### TTL
Keys automatically expire after `Window + 1 second`

## Monitoring

### Logs

Rate limit events are logged:

```json
{
  "level": "warn",
  "msg": "Rate limit exceeded",
  "key": "ratelimit:ip:192.168.1.1:/api/auth/login",
  "time": "2024-01-15T10:30:00Z"
}
```

### Redis Commands for Debugging

```bash
# Check current request count for a key
redis-cli ZCARD "marketplace:ratelimit:ip:192.168.1.1:global"

# View all rate limit keys
redis-cli KEYS "marketplace:ratelimit:*"

# Clear rate limit for specific IP
redis-cli DEL "marketplace:ratelimit:ip:192.168.1.1:global"
```

## Security Considerations

1. **IP Spoofing Prevention**: Headers are ONLY trusted from configured proxies
   - Set `RATELIMIT_TRUSTED_PROXIES` to your proxy IPs/CIDRs
   - Without this, `X-Forwarded-For` is ignored (safest default)
   
2. **Trusted Proxy Configuration**:
   ```env
   # Common setups:
   # Docker: RATELIMIT_TRUSTED_PROXIES=172.16.0.0/12
   # Kubernetes: RATELIMIT_TRUSTED_PROXIES=10.0.0.0/8
   # AWS ALB: Use AWS published IP ranges
   # Cloudflare: Use Cloudflare published IP ranges
   ```

3. **Distributed Attacks**: Redis-based limiting works across all application instances

4. **Bypass Prevention**: Rate limiting is applied before authentication middleware

5. **Fail Open**: On Redis errors, requests are allowed (logged for monitoring)

6. **Never Trust User Input**: The rightmost non-proxy IP in X-Forwarded-For is used

## Best Practices

1. **Start Conservative**: Begin with stricter limits, relax based on usage patterns
2. **Monitor**: Watch for 429 responses in production
3. **Document**: Include rate limits in API documentation
4. **Communicate**: Return `Retry-After` header to help clients
5. **Differentiate**: Use different limits for different endpoint sensitivity levels

