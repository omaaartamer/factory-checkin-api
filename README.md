# Factory Check-in API

A production-ready backend service for handling employee clock-in and clock-out events in factory environments. Features asynchronous processing, external system integration, and comprehensive testing.

## 🏗️ System Architecture
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Card Readers  │    │   Mobile Apps   │    │  Admin Panel    │
│                 │    │                 │    │                 │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────▼──────────────┐
                    │                            │
                    │        Gin HTTP Server     │
                    │     (REST API Layer)       │
                    │                            │
                    └─────────────┬──────────────┘
                                  │
                    ┌─────────────▼──────────────┐
                    │                            │
                    │      Service Layer         │
                    │   (Business Logic)         │
                    │                            │
                    └─────┬───────────────┬──────┘
                          │               │
              ┌───────────▼──────┐   ┌────▼──────────────┐
              │                  │   │                   │
              │ Repository Layer │   │   RabbitMQ Queue  │
              │  (Data Access)   │   │  (Message Broker) │
              │                  │   │                   │
              └───────────┬──────┘   └────┬──────────────┘
                          │               │
              ┌───────────▼──────┐        │
              │                  │        │
              │   PostgreSQL     │        │
              │   Database       │        │
              │                  │        │
              └──────────────────┘        │
                                          │
                            ┌─────────────▼──────────────┐
                            │                            │
                            │    Background Workers      │
                            │                            │
                            └─┬─────────────────────────┬┘
                              │                         │
                 ┌────────────▼──────┐     ┌───────────▼────────┐
                 │                   │     │                    │
                 │   Email Service   │     │  Legacy API Client │
                 │ (SMTP/Mock Email) │     │  (Labor Cost       │
                 │                   │     │   Reporting)       │
                 └───────────────────┘     └────────────────────┘
```

## ✨ Features Implemented

### Core Functionality
- **Smart Check-in/Check-out**: Single endpoint automatically detects whether employee is checking in or out
- **Work Session Tracking**: Calculates hours worked with precise timestamps
- **Multi-employee Support**: Concurrent sessions for multiple employees
- **Data Persistence**: All events and sessions stored in PostgreSQL

### Asynchronous Processing
- **Message Queue**: RabbitMQ for reliable task processing
- **Background Workers**: Process tasks without blocking user requests
- **Retry Logic**: Automatic retry for failed external API calls
- **Error Handling**: Graceful degradation when external services fail

### External Integrations
- **Legacy API Client**: Reports labor hours to company systems
- **Email Notifications**: Sends work summary to employees after checkout
- **Mock Services**: Configurable mock implementations for testing

### Production Features
- **Docker Infrastructure**: Complete containerized setup
- **Health Monitoring**: Health check endpoints and queue status
- **Configuration Management**: Environment-based configuration
- **Automated Testing**: Comprehensive test suite with fresh data resets

## 🛠️ Technology Stack

- **Backend Framework**: Go with Gin HTTP framework
- **Database**: PostgreSQL 15 with SQLX
- **Message Broker**: RabbitMQ (AMQP protocol)
- **Containerization**: Docker & Docker Compose
- **Testing**: Custom automated test suite with jq JSON parsing
- **Architecture**: Clean layered architecture (Handler → Service → Repository)

## 📁 Directory Structure
```
factory-checkin-api/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point, dependency injection
├── internal/
│   ├── handler/
│   │   └── handler.go              # HTTP handlers, REST API endpoints
│   ├── model/
│   │   └── models.go               # Data structures, API contracts
│   ├── repository/
│   │   └── repository.go           # Database access layer, CRUD operations
│   ├── service/
│   │   └── service.go              # Business logic, work session management
│   ├── queue/
│   │   └── rabbitmq_queue.go       # RabbitMQ implementation, message handling
│   ├── worker/
│   │   └── worker.go               # Background task processor
│   ├── email/
│   │   └── email.go                # Email service for employee notifications
│   └── legacy/
│       └── client.go               # Legacy API client for labor cost reporting
├── pkg/
│   └── config/
│       └── config.go               # Configuration management, environment variables
├── scripts/
│   ├── reset-and-test.sh           # Infrastructure reset script
│   └── complete-test.sh            # Comprehensive automated test suite
├── docker-compose.yml              # Container orchestration (PostgreSQL + RabbitMQ)
├── test-fresh.sh                   # One-command fresh test runner
├── go.mod                          # Go module dependencies
└── README.md                       # This file
```

### Key Files Explained

**`cmd/server/main.go`**
- Application bootstrap and dependency injection
- Initializes database, message queue, services, and HTTP server
- Graceful shutdown handling

**`internal/handler/handler.go`**
- REST API endpoints implementation
- Request validation and response formatting
- CORS middleware and error handling

**`internal/service/service.go`**
- Core business logic for check-in/check-out processing
- Work session management and hours calculation
- Asynchronous task queuing

**`internal/repository/repository.go`**
- Database abstraction layer
- Automatic table creation and migrations
- CRUD operations for events and work sessions

**`internal/queue/rabbitmq_queue.go`**
- RabbitMQ integration using AMQP protocol
- Message serialization and queue management
- Replaces custom messaging implementation for compliance

**`internal/worker/worker.go`**
- Background task processor
- Handles email notifications and legacy API calls
- Error handling and retry logic

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose
- Go 1.24+ (for development)
- jq (for testing)

### Clone and Run
```bash
# Clone the repository
git clone https://github.com/omaaartamer/factory-checkin-api.git
cd factory-checkin-api

# Install dependencies
go mod tidy

# Start infrastructure and run tests
./test-fresh.sh
```

### Manual Setup
```bash
# Start infrastructure
docker compose up -d

# Wait for services to start
sleep 20

# Run the application
go run cmd/server/main.go
```

### API Usage
```bash
# Health check
curl http://localhost:8080/health

# Employee check-in
curl -X POST http://localhost:8080/api/v1/checkin \
  -H "Content-Type: application/json" \
  -d '{"employee_id": "EMP001"}'

# Employee check-out (same endpoint automatically detects)
curl -X POST http://localhost:8080/api/v1/checkin \
  -H "Content-Type: application/json" \
  -d '{"employee_id": "EMP001"}'

# Check employee status
curl http://localhost:8080/api/v1/employee/EMP001/status

# Monitor queue
curl http://localhost:8080/api/v1/queue/status
```

## 🤖 AI Assistance Disclosure

This project was developed with ChatGPT's assistance in the following areas:

### Code Development
- **Initial Learning**: ChatGPT helped break down the task turning it into easier building blocks
- **Architecture Guidance**: Suggested clean layered architecture patterns and Go best practices (Directory structure)
- **Code Generation**: Assisted with boilerplate code only for handlers, services, and database operations
- **Debugging Support**: Helped diagnose and fix compilation errors, import issues, and configuration problems

### Testing & Quality Assurance
- **Test Suite Creation**: Completely AI designed and implemented comprehensive automated testing framework
- **Test Case Development**: Generated test scenarios for multi-employee workflows
- **Infrastructure Testing**: Created scripts for testing Docker services and health checks

### Configuration & DevOps
- **Docker Configuration**: Assisted with debugging of docker-compose.yml 
- **Environment Management**: Helped design configuration management with the go code and port management

### Documentation & Architecture
- **Code Documentation**: AI (Cursor) helped write clear comments and documentation throughout codebase
- **README Creation**: This comprehensive README was created with AI assistance but has been edited and reviewed line by line
