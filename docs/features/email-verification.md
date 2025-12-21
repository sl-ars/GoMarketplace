# Email Verification & Password Reset

## Overview

GoMarketplace implements a secure email verification and password reset system using **RabbitMQ for async email delivery**. This ensures reliability, scalability, and decoupling of email sending from the main application flow.

## Features

- **Email Verification** after registration
- **Verification Code** (6-digit) for mobile-friendly verification
- **Password Reset** via secure email links
- **Secure Tokens** with SHA-256 hashing (tokens never stored in plain text)
- **Configurable TTLs** (24h for verification, 1h for password reset)
- **Beautiful HTML Email Templates**
- **Async Email Delivery** via RabbitMQ (reliable, retry-capable)

## Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   User Action   │────▶│   Auth Service  │────▶│  Email Publisher│
│  (Register/     │     │  (Generate      │     │  (RabbitMQ)     │
│   Reset Pass)   │     │   Token)        │     │                 │
└─────────────────┘     └────────┬────────┘     └────────┬────────┘
                                 │                       │
                                 ▼                       ▼
                        ┌─────────────────┐     ┌─────────────────┐
                        │   PostgreSQL    │     │  Email Worker   │
                        │ (Store Hash)    │     │  (Consumer)     │
                        └─────────────────┘     └────────┬────────┘
                                                         │
                                                         ▼
                                                ┌─────────────────┐
                                                │   SMTP Server   │
                                                │  (Gmail/SES/    │
                                                │   SendGrid)     │
                                                └─────────────────┘
```

## RabbitMQ Email Queue

### Why RabbitMQ?

| Benefit | Description |
|---------|-------------|
| **Reliability** | If SMTP fails, messages stay in queue for retry |
| **Decoupling** | Main app doesn't block on email sending |
| **Scalability** | Run multiple email workers |
| **Resilience** | App continues working if email service is down |
| **Observability** | Queue depth shows email backlog |

### Queue Configuration

```
Exchange: emails.exchange (topic)
Queue: emails
Routing Keys:
  - email.verification
  - email.password_reset
  - email.welcome
  - email.order_confirm

Features:
  - Durable queue (survives broker restart)
  - Message TTL: 24 hours
  - Dead letter exchange for failed messages
  - Manual acknowledgment (at-least-once delivery)
```

### Running the Email Worker

```bash
# Start the email worker (dedicated process)
go run cmd/email_worker/main.go --config ./configs/.env

# Or with Docker
docker-compose up email-worker
```

Add to `docker-compose.yml`:

```yaml
email-worker:
  build: .
  command: ["./email_worker", "--config", "/app/configs/.env"]
  depends_on:
    - rabbitmq
  env_file:
    - ./configs/.env
  restart: unless-stopped
```

## Security Model

### Token Security

1. **Plain Text Token** → Sent to user via email
2. **SHA-256 Hash** → Stored in database
3. **Validation** → Hash comparison (never store plain tokens)

```
User receives: abc123...xyz (URL-safe base64, 32 bytes)
Database stores: sha256(abc123...xyz) = 64-char hex hash
```

### Why This Approach?

- Even if database is compromised, tokens cannot be recovered
- Same security model as password storage
- Prevents token reuse attacks

## API Endpoints

### Email Verification

#### 1. Verify Email (Token from URL)
```http
POST /api/auth/verify-email
Content-Type: application/json

{
  "token": "ABC123...XYZ"
}
```

Or via URL query parameter:
```http
GET /api/auth/verify-email?token=ABC123...XYZ
```

#### 2. Verify Email (Email + Code)
```http
POST /api/auth/verify-email
Content-Type: application/json

{
  "email": "user@example.com",
  "code": "123456"
}
```

#### 3. Resend Verification Email
```http
POST /api/auth/resend-verification
Content-Type: application/json

{
  "email": "user@example.com"
}
```

### Password Reset

#### 1. Request Password Reset
```http
POST /api/auth/forgot-password
Content-Type: application/json

{
  "email": "user@example.com"
}
```

#### 2. Reset Password
```http
POST /api/auth/reset-password
Content-Type: application/json

{
  "token": "ABC123...XYZ",
  "new_password": "newSecurePassword123"
}
```

## Configuration

### Environment Variables

```env
# Enable/disable email sending
EMAIL_ENABLED=true

# SMTP Configuration
EMAIL_HOST=smtp.gmail.com
EMAIL_PORT=587
EMAIL_USERNAME=your-email@gmail.com
EMAIL_PASSWORD=your-app-password
EMAIL_FROM=noreply@gomarketplace.com
EMAIL_FROM_NAME=GoMarketplace
EMAIL_USE_TLS=true

# Auth Settings
AUTH_BASE_URL=https://yourapp.com
AUTH_REQUIRE_EMAIL_VERIFIED=false

# RabbitMQ (required for email queue)
RABBIT_MQ_URL=amqp://guest:guest@localhost:5672/
```

### Gmail Setup

For Gmail, you need an **App Password**:

1. Enable 2-Factor Authentication on your Google account
2. Go to Google Account → Security → App passwords
3. Create a new app password for "Mail"
4. Use this password in `EMAIL_PASSWORD`

### SendGrid/Mailgun

For production, consider using a dedicated email service:

```env
# SendGrid (via SMTP)
EMAIL_HOST=smtp.sendgrid.net
EMAIL_PORT=587
EMAIL_USERNAME=apikey
EMAIL_PASSWORD=SG.your-api-key

# Mailgun (via SMTP)
EMAIL_HOST=smtp.mailgun.org
EMAIL_PORT=587
EMAIL_USERNAME=postmaster@your-domain.mailgun.org
EMAIL_PASSWORD=your-mailgun-password
```

## Database Schema

Migration adds these columns to `users` table:

```sql
-- Email verification
email_verified BOOLEAN DEFAULT FALSE
email_verified_at TIMESTAMP
verification_token_hash VARCHAR(64)
verification_token_expires_at TIMESTAMP
verification_code VARCHAR(6)

-- Password reset
reset_token_hash VARCHAR(64)
reset_token_expires_at TIMESTAMP
```

## Token TTLs

| Token Type | TTL | Rationale |
|------------|-----|-----------|
| Email Verification | 24 hours | Users may not check email immediately |
| Password Reset | 1 hour | Security-sensitive, should be used quickly |

## Email Templates

### Verification Email

Beautiful HTML email with:
- Welcome message
- **Clickable button** with verification link
- **6-digit code** for manual entry
- Expiration notice
- Security footer

### Password Reset Email

Professional HTML email with:
- Clear call-to-action button
- Direct reset link
- Security warning if not requested
- Expiration notice

## Flow Diagrams

### Registration Flow

```
1. User registers → Creates account (email_verified=false)
2. Generate verification token + 6-digit code
3. Store SHA-256 hash in database
4. Publish to RabbitMQ "emails" queue
5. Email worker processes message
6. SMTP sends email to user
7. User clicks link OR enters code
8. Validate token, mark email_verified=true
9. User can now log in (if AUTH_REQUIRE_EMAIL_VERIFIED=true)
```

### Password Reset Flow

```
1. User clicks "Forgot Password"
2. Enter email address
3. Generate reset token (valid 1 hour)
4. Store SHA-256 hash in database
5. Publish to RabbitMQ "emails" queue (priority: 8)
6. Email worker processes message
7. SMTP sends email to user
8. User clicks link, enters new password
9. Validate token, update password hash
10. Clear reset token from database
11. User logs in with new password
```

## Email Worker Details

### Message Processing

```go
// Message format in queue
type EmailMessage struct {
    Type      string                 // "email.verification", "email.password_reset"
    To        string                 // Recipient email
    Subject   string                 // Email subject
    Data      map[string]interface{} // Template data
    Priority  int                    // 0-9, higher = more important
    Timestamp int64                  // When queued
}
```

### Retry Logic

| Scenario | Action |
|----------|--------|
| SMTP success | Acknowledge message |
| SMTP failure | Nack + requeue for retry |
| Invalid message | Reject (no requeue) |
| Template error | Reject (no requeue) |

### Monitoring

```bash
# Check queue depth
rabbitmqctl list_queues name messages

# Check for dead letters
rabbitmqctl list_queues -p / name messages_ready | grep dlx
```

## Security Considerations

### 1. Information Disclosure Prevention

```go
// Always return success, even if email doesn't exist
httpx.WriteSuccess(w, http.StatusOK, 
    "If your email is registered, you will receive an email", nil)
```

### 2. Rate Limiting

Auth endpoints have strict rate limiting (5 requests/minute):
- Prevents brute-force token guessing
- Prevents email bombing

### 3. Token Invalidation

- Verification token cleared after successful verification
- Reset token cleared after password change
- Expired tokens automatically rejected

### 4. Secure Token Generation

```go
// 32 bytes of cryptographically secure random data
randomBytes := make([]byte, 32)
rand.Read(randomBytes)
token := base64.URLEncoding.EncodeToString(randomBytes)
```

## Testing Locally

### Without Email (Development)

Set `EMAIL_ENABLED=false` and check logs for verification tokens:

```bash
# Logs will show:
INFO Verification email would be sent (email disabled)
     token=ABC123...XYZ code=123456
```

### With RabbitMQ (Staging/Production)

1. Start RabbitMQ: `docker-compose up rabbitmq`
2. Configure SMTP settings in `.env`
3. Set `EMAIL_ENABLED=true`
4. Start email worker: `go run cmd/email_worker/main.go`
5. Register a user
6. Check RabbitMQ management UI for queued messages
7. Email worker processes and sends email

### RabbitMQ Management UI

Access at `http://localhost:15672` (guest/guest):
- View queue depth
- Monitor message rates
- Inspect dead letters
- Purge queues for testing

## Implementation Details

### Packages

| Package | Purpose |
|---------|---------|
| `pkg/email/` | SMTP sender, HTML templates |
| `pkg/token/` | Secure token generation, hashing |
| `internal/messagebus/` | RabbitMQ publishers, message types |
| `cmd/email_worker/` | Email worker process |

### Key Files

- `pkg/email/email.go` - Email sending and templates
- `pkg/token/token.go` - Token generation and validation
- `internal/messagebus/email_events.go` - Email publisher, message types
- `cmd/email_worker/main.go` - Email worker entry point
- `internal/usecases/auth.go` - Business logic
- `internal/services/auth_service.go` - Orchestration
- `internal/deliveries/http/auth/auth_handler.go` - HTTP handlers

## Troubleshooting

### Emails Not Sending

1. Check `EMAIL_ENABLED=true`
2. Verify email worker is running
3. Check RabbitMQ queue for pending messages
4. Check email worker logs for SMTP errors
5. For Gmail, ensure App Password is used

### Messages Stuck in Queue

1. Check email worker is running: `ps aux | grep email_worker`
2. Check RabbitMQ connection: `rabbitmqctl list_connections`
3. Check consumer: `rabbitmqctl list_consumers`

### Token Expired

- Verification: 24-hour window
- Password reset: 1-hour window
- User can request new token via resend/forgot-password

### Email Already Verified

Gracefully handled - returns success message without error.

### Dead Letter Queue Growing

Check DLQ for failed messages:
```bash
rabbitmqctl list_queues -p / | grep dlx
```

Common causes:
- Invalid email addresses
- Template errors
- SMTP authentication failures
