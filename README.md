# Courier Service

A high-performance Go-based microservice for managing courier deliveries and order assignments. This service handles courier registration, delivery tracking, and order distribution with real-time monitoring and rate limiting.

## Overview

This courier service is designed to efficiently manage delivery operations with the following core responsibilities:

- **Courier Management**: Register and manage courier profiles with various transport types
- **Order Assignment**: Automatically assign orders to available couriers based on policies
- **Delivery Tracking**: Monitor delivery status and progress
- **Real-time Monitoring**: Track active deliveries with continuous monitoring
- **Event Processing**: Handle order events via Kafka for asynchronous processing
- **Rate Limiting**: Protect API endpoints with token bucket rate limiting
- **Observability**: Integrated Prometheus metrics and structured logging

## Features

- ✅ RESTful API for courier and delivery management
- ✅ gRPC protocol support for order services
- ✅ PostgreSQL persistence with migrations
- ✅ Kafka consumer for event processing
- ✅ Token bucket rate limiting
- ✅ Prometheus metrics and monitoring
- ✅ pprof profiling server support
- ✅ Docker and Docker Compose support
- ✅ Comprehensive test coverage with integration tests
- ✅ Graceful shutdown with signal handling

## Tech Stack

- **Language**: Go 1.25+
- **Web Framework**: Echo v4
- **Database**: PostgreSQL with pgx driver
- **Message Broker**: Apache Kafka (Sarama)
- **RPC**: gRPC with Protocol Buffers
- **Monitoring**: Prometheus
- **Testing**: testcontainers-go for integration tests
- **Containerization**: Docker

## Project Structure

```
.
├── cmd/
│   └── app/
│       └── main.go              # Application entry point
├── internal/
│   ├── app/
│   │   └── app.go               # Application setup and initialization
│   ├── handler/                 # HTTP request handlers
│   │   ├── courier/
│   │   ├── delivery/
│   │   └── errors/
│   ├── usecase/                 # Business logic layer
│   │   ├── courier/
│   │   ├── delivery/
│   │   └── order_event/
│   ├── repository/              # Data access layer
│   │   ├── courier/
│   │   └── delivery/
│   ├── model/                   # Domain models
│   ├── gateway/                 # External service integrations
│   │   ├── order/
│   │   └── orderhttp/
│   ├── infra/
│   │   └── postgres/            # Database connection & transactions
│   ├── transport/
│   │   └── kafka/               # Kafka consumer
│   ├── worker/                  # Background workers
│   │   ├── delivery_monitor.go
│   │   └── order_assigner.go
│   ├── ratelimit/               # Rate limiting middleware
│   ├── observability/           # Monitoring and metrics
│   ├── routes/                  # Route registration
│   ├── proto/                   # Protocol Buffer definitions
│   └── integration/             # Integration tests
├── migrations/                  # Database migrations (Goose)
├── pkg/
│   └── config/                  # Configuration management
├── docker-compose.yml           # Docker Compose setup
├── Dockerfile                   # Container image definition
└── Makefile                     # Build and development tasks
```

## Getting Started

### Prerequisites

- Go 1.25 or higher
- Docker and Docker Compose
- PostgreSQL 13+ (via Docker)
- Apache Kafka (via Docker)

### Installation

1. **Clone the repository**
```bash
git clone https://github.com/cdxy1/go-courier-service.git
cd go-courier-service
```

2. **Set up environment variables**
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. **Start infrastructure services**
```bash
make up
```

4. **Run database migrations**
```bash
make migrate
```

5. **Build the application**
```bash
make build
```

6. **Run the application**
```bash
./bin/app
```

### Quick Start with Docker

```bash
make up
make migrate
docker-compose exec app make build
docker-compose up --build
```

## Configuration

Configuration is managed via environment variables. Key settings:

```env
# Server
PORT=8080

# PostgreSQL
POSTGRES_USER=user
POSTGRES_PASSWORD=password
POSTGRES_DB=courier_service
POSTGRES_PORT=5432
POSTGRES_HOST=postgres

# Kafka
KAFKA_BROKERS=kafka:9092

# Delivery settings
DELIVERY_ON_FOOT_DURATION=60      # minutes
DELIVERY_SCOOTER_DURATION=30      # minutes
DELIVERY_CAR_DURATION=20          # minutes
DELIVERY_MONITOR_INTERVAL=30      # seconds

# Profiling
PPROF_ENABLED=false
PPROF_PORT=6060
```

## API Endpoints

### Courier Management

- `POST /couriers` - Register a new courier
- `GET /couriers/:id` - Get courier details
- `PATCH /couriers/:id` - Update courier information
- `GET /couriers/:id/assignments` - Get courier assignments count

### Delivery Management

- `POST /deliveries` - Create a new delivery
- `GET /deliveries/:id` - Get delivery details
- `PATCH /deliveries/:id` - Update delivery status

### Health Check

- `GET /health` - Application health status
- `GET /metrics` - Prometheus metrics endpoint

## Architecture

### Clean Architecture

The application follows a clean architecture pattern with clear separation of concerns:

```
┌─────────────────────────────┐
│     HTTP/gRPC Handlers      │
├─────────────────────────────┤
│       Use Cases             │
│    (Business Logic)         │
├─────────────────────────────┤
│      Repositories           │
│    (Data Access)            │
├─────────────────────────────┤
│      Database               │
│    (PostgreSQL)             │
└─────────────────────────────┘
```

### Key Components

- **Handlers**: Process HTTP/gRPC requests and convert them to use case calls
- **Use Cases**: Contain business logic and orchestrate repository access
- **Repositories**: Handle data persistence operations
- **Models**: Domain models representing core entities (Courier, Delivery, Order)
- **Workers**: Background processes for order assignment and delivery monitoring
- **Gateway**: Integration with external services

### Message Flow

1. **Order Events**: Kafka events are consumed by the `EventConsumer`
2. **Order Assignment**: `OrderAssigner` worker distributes orders to available couriers
3. **Delivery Monitoring**: `DeliveryMonitor` tracks active deliveries and updates statuses

## Development

### Running Tests

```bash
# Run all tests
make test

# Run integration tests
make test-integration

# Generate coverage report
go test -cover ./...
make cover
```

### Code Formatting

```bash
make fmt
```

### Building for Production

```bash
make build
# Binary will be available at ./bin/app
```

### Docker Build

```bash
docker build -t courier-service:latest .
```

### Profiling

Enable pprof profiling server by setting `PPROF_ENABLED=true` in your environment variables. Access profiles at `http://localhost:6060/debug/pprof/`

### Database Migrations

Migrations are managed with Goose. To create a new migration:

```bash
goose -dir ./migrations create migration_name sql
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
