# RP Wallet - Go Implementation

A high-performance wallet management service built with Go, responsible for managing user wallet monetary transactions.

## Architecture

This Go implementation maintains the same architecture as the original Java Spring Boot version:

- **Clean Architecture** with domain-driven design
- **MongoDB** with replica set for data persistence
- **Apache Kafka** for asynchronous transaction processing
- **RESTful API** with comprehensive error handling
- **Health checks** for MongoDB and Kafka
- **Swagger documentation** for API exploration

## Features

- ✅ **Wallet Management**: Create, update, and query wallets
- ✅ **Asynchronous Transactions**: Deposit, withdraw, and transfer via Kafka
- ✅ **Balance Tracking**: Real-time balance updates with transaction history
- ✅ **Operation History**: Complete audit trail of all operations
- ✅ **Daily Summaries**: Aggregate transaction data by date
- ✅ **Concurrency Control**: Wallet-level locking for transaction safety
- ✅ **Health Monitoring**: MongoDB and Kafka health indicators
- ✅ **Error Handling**: Comprehensive exception handling with proper HTTP status codes

## Project Structure

```
.
├── cmd/                    # Application entrypoints
├── configs/                # Configuration files
├── internal/
│   ├── api/               # API layer (controllers, DTOs)
│   │   ├── operation/
│   │   └── wallet/
│   ├── config/            # Configuration management
│   ├── domain/            # Business logic layer
│   │   ├── health/
│   │   ├── operation/
│   │   └── wallet/
│   └── infrastructure/    # External dependencies
│       ├── database/
│       ├── kafka/
│       ├── middleware/
│       └── router/
├── pkg/                   # Shared packages
│   └── logger/
├── docker-compose.yml
├── Dockerfile
├── Makefile
└── README.md
```
## Access Services

- **API**: http://localhost:8080/api
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **Health Check**: http://localhost:8080/health
- **Kafka UI**: http://localhost:8090

## API Endpoints

### Wallets

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/wallet` | List all wallets |
| GET | `/api/wallet/{id}` | Get wallet by ID |
| POST | `/api/wallet` | Create new wallet |
| PATCH | `/api/wallet/{id}` | Update wallet status |
| POST | `/api/wallet/{id}/deposit` | Deposit funds (async) |
| POST | `/api/wallet/{id}/withdraw` | Withdraw funds (async) |
| POST | `/api/wallet/{id}/transfer` | Transfer funds (async) |

### Operations

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/operations` | List operations with filters |
| GET | `/operations/{id}` | Get operation by ID |
| GET | `/operations/daily-summary` | Get daily summary |
| GET | `/operations/daily-summary-details` | Get detailed daily summary |

### Health

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Basic health check |
| GET | `/health/details` | Detailed health check |


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
> **Project:** Wallet API (Go)  
> **Description:** Digital wallet management API.

