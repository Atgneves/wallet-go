# RP Wallet - Go Implementation

A high-performance digital wallet management service built with Go, featuring MongoDB for data persistence and Apache Kafka for asynchronous transaction processing. This implementation maintains clean architecture principles with domain-driven design.

## 🏗️ Architecture

This Go implementation follows the same architectural patterns as the original Java Spring Boot version while leveraging Go's performance advantages:

- **Hexagonal Architecture** with clear domain boundaries
- **MongoDB** for reliable data persistence
- **Apache Kafka** for asynchronous transaction processing
- **RESTful API** with comprehensive error handling and validation
- **Health monitoring** for all critical components
- **Concurrent processing** with wallet-level locking for transaction safety

## ✨ Features

- ✅ **Wallet Management**: Create, update, activate/deactivate, and query wallets
- ✅ **Asynchronous Transactions**: Deposit, withdraw, and transfer operations via Kafka
- ✅ **Real-time Balance Tracking**: Immediate balance updates with complete transaction history
- ✅ **Operation Audit Trail**: Comprehensive logging of all wallet operations
- ✅ **Daily Transaction Summaries**: Aggregate transaction reports by date
- ✅ **Concurrency Control**: Wallet-level locking prevents race conditions
- ✅ **Business Rule Validation**: Insufficient funds, inactive/blocked wallet checks
- ✅ **Health Monitoring**: MongoDB and Kafka connectivity monitoring
- ✅ **Error Handling**: Proper HTTP status codes with detailed error messages

## Project Structure

```
wallet-go/
├── cmd/
│   └── api/
│       └── main.go              # Application entrypoint
├── internal/
│   ├── wallet/                  # Wallet Domain
│   │   ├── handler.go           # HTTP handlers (REST controllers)
│   │   ├── service.go           # Business logic implementation
│   │   ├── store.go             # MongoDB repository
│   │   ├── types.go             # Domain models and DTOs
│   │   └── validator.go         # Business rule validation
│   ├── operation/               # Operation Domain
│   │   ├── handler.go           # Operation REST endpoints
│   │   ├── service.go           # Operation business logic
│   │   ├── store.go             # Operation data access
│   │   └── types.go             # Operation models and enums
│   ├── health/                  # Health Check Domain
│   │   ├── handler.go           # Health check endpoints
│   │   ├── service.go           # Health check logic
│   │   └── types.go             # Health status models
│   ├── shared/                  # Shared Infrastructure
│   │   ├── config/              # Configuration management
│   │   ├── database/            # MongoDB client
│   │   ├── kafka/               # Kafka producer/consumer
│   │   ├── middleware/          # HTTP middlewares
│   │   ├── errors/              # Custom error types
│   │   └── utils/               # Utilities (locking, etc.)
│   └── router/                  # HTTP router configuration
├── pkg/                         # Shared packages (if needed)
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
├── docker-compose.yml           # Development environment
├── Dockerfile                   # Container build
├── Makefile                     # Development commands
└── README.md                    # This documentation
```
## 🚀 Quick Start

### Prerequisites

- **Go 1.21+** - [Download](https://golang.org/dl/)
- **Docker & Docker Compose** - [Install Docker](https://docs.docker.com/get-docker/)
- **Make** (optional) - For convenient command execution

### Creating the MongoDB container using Docker Compose


### *OBS*:   If necessary, edit etc/hosts adding the hostnames for localhost

```sudo  nano /etc/hosts```

```127.0.0.1 mongo-primary```

```127.0.0.1 mongo-secondary1```

```127.0.0.1 mongo-secondary2```



#### 1. Launch containers
```docker-compose up -d```

#### 2. Wait for containers to be ready
```sleep 30```

#### 3. Initialize replica set
```docker exec -i mongo-primary mongosh < init-replica.js```

## Access Services

- **API**: http://localhost:8080/api
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **Health Check**: http://localhost:8080/health
- **Kafka UI**: http://localhost:8090

## 📚 API Reference

### 💳 Wallet Management

| Method | Endpoint | Description | Request Body |
|--------|----------|-------------|--------------|
| `GET` | `/wallet` | List all wallets | - |
| `GET` | `/wallet/{id}` | Get wallet by ID | - |
| `POST` | `/wallet` | Create new wallet | `{"customerId": "string"}` |
| `PATCH` | `/wallet/{id}` | Update wallet status | `{"active": bool, "blocked": bool}` |

#### Example: Create Wallet
```bash
curl -X POST http://localhost:8080/wallet \
  -H "Content-Type: application/json" \
  -d '{"customerId": "customer-123"}'
```

### 💰 Transaction Operations (Asynchronous)

| Method | Endpoint | Description | Request Body |
|--------|----------|-------------|--------------|
| `POST` | `/wallet/{id}/deposit` | Deposit funds | `{"amountInCents": number}` |
| `POST` | `/wallet/{id}/withdraw` | Withdraw funds | `{"amountInCents": number}` |
| `POST` | `/wallet/{id}/transfer` | Transfer funds | `{"amountInCents": number, "walletDestinationId": "uuid"}` |

#### Example: Deposit Funds
```bash
curl -X POST http://localhost:8080/wallet/{wallet-id}/deposit \
  -H "Content-Type: application/json" \
  -d '{"amountInCents": 10000}'
```

#### Example: Transfer Funds
```bash
curl -X POST http://localhost:8080/wallet/{source-wallet-id}/transfer \
  -H "Content-Type: application/json" \
  -d '{
    "amountInCents": 5000,
    "walletDestinationId": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

### 📊 Operation History & Reports

| Method | Endpoint | Description | Query Parameters |
|--------|----------|-------------|------------------|
| `GET` | `/operations` | List operations | `walletId`, `from`, `to` |
| `GET` | `/operations/{id}` | Get operation details | - |
| `GET` | `/operations/daily-summary` | Daily summary | `walletId`, `date` |
| `GET` | `/operations/daily-summary-details` | Detailed daily summary | `walletId`, `date` |

#### Example: Get Operation History
```bash
curl "http://localhost:8080/operations?walletId={uuid}&from=2024-01-01&to=2024-01-31"
```

#### Example: Daily Summary
```bash
curl "http://localhost:8080/operations/daily-summary?walletId={uuid}&date=2024-01-15"
```

### 🏥 Health Monitoring

| Method | Endpoint | Description | Response |
|--------|----------|-------------|----------|
| `GET` | `/health` | Basic health status | `{"status": "UP/DOWN"}` |
| `GET` | `/health/details` | Detailed health info | Component-level status |

## Transaction Flow

### Synchronous Operations
1. Wallet CRUD operations
2. Operation queries
3. Daily summaries

### Asynchronous Operations (via Kafka)
1. **Deposit/Withdraw/Transfer** requests sent to respective Kafka topics
2. **Kafka consumers** process messages and execute business logic
3. **Database transactions** ensure consistency
4. **Wallet locking** prevents concurrent modification issues

## Monitoring and Health

The application provides comprehensive health checks:

- **Database Health**: MongoDB connection and replica set status
- **Kafka Health**: Kafka cluster connectivity
- **Combined Health**: Overall application status

## Copyright (c) 2025 Alan Neves

> **All rights reserved.**  
>  
> This software is the confidential and proprietary information of **Alan Neves**.  
> Unauthorized copying of this file, via any medium, is strictly prohibited.  
>  
> **Project:** *Wallet API* (Go)  
> **Description:** Digital wallet management API.


