# Factory Check-in API - Release 1.0.0

## Overview

The Factory Check-in API is a production-ready backend service for managing employee clock-in/clock-out events in factory environments. It features asynchronous processing, external system integrations, and a clean layered architecture.

**Repository:** github.com/omaaartamer/factory-checkin-api
**Go Version:** 1.23.0 (Toolchain: 1.24.2)
**License:** MIT

---

## Key Features

- **Smart Dual-Endpoint Detection**: Single endpoint automatically detects check-in vs check-out based on active session state
- **Work Session Management**: Automatic hours calculation when employees clock out
- **Asynchronous Processing**: Background task processing via RabbitMQ for non-blocking operations
- **External Integrations**: Legacy API for labor cost reporting and email notifications
- **Multi-Employee Support**: Concurrent session handling for multiple employees
- **Data Persistence**: Complete audit trail with PostgreSQL

---

## Architecture

```
HTTP Clients (Card Readers, Mobile Apps, Admin Panel)
                    |
        +-----------------------+
        |  Gin HTTP Server      | <- handler/handler.go
        |  (REST API Layer)     |
        +-----------+-----------+
                    |
        +-----------------------+
        |  CheckinService       | <- service/service.go
        |  (Business Logic)     |
        +--------+--------+-----+
                 |        |
    +------------+--+  +--+-------------+
    |  Repository   |  |  RabbitMQ      |
    |  (Data Access)|  |  (Message      |
    |               |  |   Broker)      |
    +-------+-------+  +-------+--------+
            |                  |
    +-------+-------+  +-------+--------+
    |  PostgreSQL   |  |  Background    |
    |  Database     |  |  Workers       |
    +---------------+  +-------+--------+
                               |
                 +-------------+-------------+
                 | Email Service &           |
                 | Legacy API Client         |
                 +---------------------------+
```

---

## Project Structure

```
factory-checkin-api/
├── cmd/
│   └── server/
│       └── main.go                 - Application entry point
├── internal/
│   ├── handler/
│   │   └── handler.go              - REST API endpoints
│   ├── model/
│   │   └── models.go               - Data structures
│   ├── repository/
│   │   └── repository.go           - Database abstraction
│   ├── service/
│   │   └── service.go              - Business logic
│   ├── queue/
│   │   └── rabbitmq_queue.go       - Message broker implementation
│   ├── worker/
│   │   └── worker.go               - Background task processor
│   ├── email/
│   │   └── email.go                - Email service (mock)
│   └── legacy/
│       └── client.go               - Legacy API client (mock)
├── pkg/
│   └── config/
│       └── config.go               - Configuration management
├── scripts/
│   ├── reset-and-test.sh           - Infrastructure reset
│   └── complete-test.sh            - Test suite
├── docker-compose.yml              - Container orchestration
├── test-fresh.sh                   - Fresh test runner
├── go.mod                          - Module definition
└── go.sum                          - Dependency checksums
```

---

## API Endpoints

### Health Check
```
GET /health
```
Returns service availability status.

**Response:**
```json
{
  "success": true,
  "message": "Factory Check-in API is running",
  "version": "1.0.0"
}
```

### Check-in / Check-out
```
POST /api/v1/checkin
```
Unified endpoint for both operations. Automatically detects action based on active session state.

**Request:**
```json
{
  "employee_id": "EMP001"
}
```

**Response (Check-in):**
```json
{
  "success": true,
  "message": "Successfully checked in",
  "event_type": "checkin",
  "timestamp": "2025-12-31T14:30:00Z"
}
```

**Response (Check-out):**
```json
{
  "success": true,
  "message": "Successfully checked out",
  "event_type": "checkout",
  "timestamp": "2025-12-31T18:30:00Z",
  "hours_worked": 4.0
}
```

### Employee Status
```
GET /api/v1/employee/:id/status
```
Returns the current work status of an employee.

**Response (Checked In):**
```json
{
  "success": true,
  "status": "checked_in",
  "session": {
    "id": 1,
    "employee_id": "EMP001",
    "checkin_time": "2025-12-31T14:30:00Z",
    "status": "active"
  }
}
```

**Response (Not Checked In):**
```json
{
  "success": true,
  "status": "not_checked_in",
  "message": "Employee is not currently checked in"
}
```

### Queue Status
```
GET /api/v1/queue/status
```
Monitor background task queue.

**Response:**
```json
{
  "success": true,
  "queue": {
    "pending_messages": 5,
    "timestamp": "2025-12-31T18:35:00Z"
  }
}
```

---

## Database Schema

### checkin_events
| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL | Primary key |
| employee_id | VARCHAR(50) | Employee identifier |
| event_type | VARCHAR(20) | 'checkin' or 'checkout' |
| timestamp | TIMESTAMP WITH TIME ZONE | Event occurrence time |
| created_at | TIMESTAMP WITH TIME ZONE | Record creation time |

### work_sessions
| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL | Primary key |
| employee_id | VARCHAR(50) | Employee identifier |
| checkin_time | TIMESTAMP WITH TIME ZONE | Session start |
| checkout_time | TIMESTAMP WITH TIME ZONE | Session end (nullable) |
| hours_worked | DECIMAL(5,2) | Calculated duration (nullable) |
| status | VARCHAR(20) | 'active' or 'completed' |
| created_at | TIMESTAMP WITH TIME ZONE | Creation time |
| updated_at | TIMESTAMP WITH TIME ZONE | Last update time |

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | HTTP server port |
| `DATABASE_URL` | `postgres://checkin_user:checkin_password@localhost:5432/checkin_db?sslmode=disable` | PostgreSQL connection |
| `RABBITMQ_URL` | `amqp://checkin_user:checkin_password@localhost:5672/` | RabbitMQ connection |
| `LEGACY_API_URL` | `http://localhost:9000/api/labor-cost` | Legacy system endpoint |
| `SMTP_HOST` | localhost | Email server hostname |
| `SMTP_PORT` | 587 | Email server port |
| `SMTP_USERNAME` | (empty) | Email authentication |
| `SMTP_PASSWORD` | (empty) | Email authentication |
| `MAX_RETRIES` | 5 | Max retry attempts |
| `RETRY_DELAY_SECONDS` | 30 | Delay between retries |

---

## Getting Started

### Prerequisites
- Docker and Docker Compose
- Go 1.21+
- jq (for testing)
- curl (for API testing)

### Quick Start

```bash
# Clone repository
git clone https://github.com/omaaartamer/factory-checkin-api.git
cd factory-checkin-api

# Install dependencies
go mod tidy

# Start infrastructure and run tests
./test-fresh.sh

# Or manually:
docker compose up -d      # Start PostgreSQL + RabbitMQ
sleep 20                  # Wait for services
go run cmd/server/main.go # Start API server
```

### Verification

```bash
# Health check
curl http://localhost:8080/health

# Test check-in
curl -X POST http://localhost:8080/api/v1/checkin \
  -H "Content-Type: application/json" \
  -d '{"employee_id": "EMP001"}'

# Check employee status
curl http://localhost:8080/api/v1/employee/EMP001/status

# Monitor queue
curl http://localhost:8080/api/v1/queue/status
```

### RabbitMQ Management UI
- URL: http://localhost:15672
- Credentials: `checkin_user` / `checkin_password`

---

## Technology Stack

| Component | Technology | Version |
|-----------|------------|---------|
| Language | Go | 1.23.0 |
| Web Framework | Gin Gonic | 1.11.0 |
| Database | PostgreSQL | 15 |
| Message Broker | RabbitMQ | 3 |
| SQL Library | sqlx | 1.4.0 |
| AMQP Client | streadway/amqp | 1.1.0 |
| Containerization | Docker Compose | - |

---

## Data Flow

### Check-in Flow
1. Client sends POST `/api/v1/checkin` with employee_id
2. Handler validates request
3. Service checks for active session
4. No active session found → create check-in event and work session
5. Return success response immediately

### Check-out Flow
1. Client sends POST `/api/v1/checkin` with employee_id
2. Handler validates request
3. Service finds active session
4. Create check-out event
5. Calculate hours worked
6. Update work session to completed
7. Queue async tasks (labor report, email notification)
8. Return success response immediately
9. Background worker processes queued messages asynchronously

---

## Security Considerations

### Current State
- No authentication (designed for internal factory network)
- CORS enabled for all origins
- Input validation via Gin request binding

### Production Recommendations
- Implement API key or JWT authentication
- Use secrets management for credentials
- Restrict CORS to specific origins
- Add rate limiting
- Enable HTTPS/TLS
- Implement request signing

---

## Deployment Checklist

- [ ] Environment variables configured
- [ ] Database credentials secured
- [ ] RabbitMQ credentials secured
- [ ] Legacy API endpoint verified
- [ ] SMTP configured (if using real email)
- [ ] CORS origins restricted
- [ ] Logging aggregation configured
- [ ] Monitoring/alerting setup
- [ ] Database backups configured

---

## CI/CD Pipeline

GitHub Actions workflow (`.github/workflows/go.yml`):
- **Trigger:** Push to main, pull requests to main
- **Go Version:** 1.21
- **Steps:** Checkout → Setup Go → Build → Test

---

## Known Limitations

1. **Mock Implementations**: Email and legacy API services log instead of executing real operations
2. **No Authentication**: API is open; authentication should be added for production
3. **No Dead-Letter Queue**: Failed async messages are logged but not requeued
4. **Single Worker**: Background processing uses a single goroutine

---

## Future Enhancements

- Dead-letter queue for failed messages
- Exponential backoff for retries
- Rate limiting per employee
- Webhook notifications
- Real SMTP email integration
- Actual legacy API integration
- API authentication (JWT/API keys)
- Horizontal scaling support

---

## License

MIT License - See [LICENSE](../LICENSE) for details.
