# Execution-Engine

A distributed, containerized code execution platform built with Go, Redis, Docker, and distributed workers.

Execution-Engine safely runs untrusted user code inside isolated Docker sandboxes with strict CPU/memory limits, execution timeouts, asynchronous job processing, persistent execution tracking, and realtime execution streaming infrastructure.

---

# Features

## Secure Sandboxed Execution

- Isolated Docker runtime execution
- CPU and memory limits
- Disabled container networking
- Temporary execution file isolation
- Automatic cleanup
- Timeout-based execution cancellation
- stdin support
- stdout/stderr capture

---

# Distributed Architecture

- Redis-backed distributed job queue
- Asynchronous execution workers
- Persistent execution result tracking
- Distributed worker execution model
- Multi-service Docker Compose orchestration
- WebSocket-based realtime streaming infrastructure
- Event-driven execution pipeline

---

# Supported Languages

- Python 3.11
- JavaScript (Node.js)
- C++

---

# Architecture

```text
                ┌──────────────────┐
                │      Client      │
                └────────┬─────────┘
                         │
                         ▼
                ┌──────────────────┐
                │     Go API       │
                │   (Gin Server)   │
                └────────┬─────────┘
                         │
                         ▼
                ┌──────────────────┐
                │   Redis Queue    │
                │ execution_queue  │
                └────────┬─────────┘
                         │
                         ▼
                ┌──────────────────┐
                │ Distributed      │
                │ Worker Service   │
                └────────┬─────────┘
                         │
                         ▼
                ┌──────────────────┐
                │ Docker Sandbox   │
                │ Runtime Engine   │
                └────────┬─────────┘
                         │
                         ▼
                ┌──────────────────┐
                │ Language Runtime │
                │ Python / JS / C++│
                └────────┬─────────┘
                         │
                         ▼
                ┌──────────────────┐
                │ Persistent       │
                │ Execution State  │
                └──────────────────┘
```

---

# Execution Flow

```text
Client Request
    ↓
API creates execution job
    ↓
Job pushed to Redis queue
    ↓
Worker consumes job
    ↓
Worker launches isolated Docker container
    ↓
Code executes inside sandbox
    ↓
Worker stores execution result
    ↓
Client fetches execution status/result
```

---

# Tech Stack

## Backend

- Go
- Gin
- Redis
- Gorilla WebSocket

## Infrastructure

- Docker
- Docker Compose
- Distributed Workers

## Language Runtimes

- Python 3.11
- Node.js
- GCC / G++

---

# Security Features

The execution engine applies multiple isolation and security mechanisms:

## Container Isolation

- Docker sandbox execution
- Independent runtime containers
- Automatic container cleanup

## Resource Limits

```bash
--memory=128m
--cpus=0.5
```

## Network Isolation

```bash
--network=none
```

Prevents:
- internet access
- port scanning
- external API abuse
- malicious outbound traffic

## Timeout Protection

All executions are forcibly terminated after timeout expiration.

## Filesystem Isolation

- Temporary execution artifacts
- Automatic cleanup
- No persistent runtime state

---

# API

## Execute Code

### Endpoint

```http
POST /execute
```

### Request

```json
{
  "language": "python",
  "code": "print('hello world')",
  "input": ""
}
```

### Response

```json
{
  "job_id": "uuid",
  "message": "job queued"
}
```

---

## Fetch Execution Result

### Endpoint

```http
GET /result/:id
```

### Response

```json
{
  "id": "uuid",
  "status": "completed",
  "output": "hello world\n",
  "error": ""
}
```

---

# Execution States

```text
queued
running
completed
failed
```

---

# WebSocket Streaming

Execution-Engine includes realtime execution streaming infrastructure using:

- Redis Pub/Sub
- Gorilla WebSockets
- Event-driven worker communication

## WebSocket Endpoint

```text
/ws/:job_id
```

---

# Local Development Setup

## Clone Repository

```bash
git clone https://github.com/GautamKappagal/Execution-Engine.git

cd Execution-Engine
```

---

# Build Runtime Images

## Python Runtime

```bash
cd docker-images/python
docker build -t execution-python .
```

---

## JavaScript Runtime

```bash
cd docker-images/javascript
docker build -t execution-javascript .
```

---

## C++ Runtime

```bash
cd docker-images/cpp
docker build -t execution-cpp .
```

---

# Run Entire Distributed System

From project root:

```bash
docker compose up --build
```

This starts:
- API service
- Redis service
- Distributed worker service

---

# Example Usage

## Execute Python

```bash
curl -X POST http://localhost:8080/execute \
-H "Content-Type: application/json" \
-d '{
  "language":"python",
  "code":"print(\"hello from execution engine\")"
}'
```

---

## Fetch Result

```bash
curl http://localhost:8080/result/<JOB_ID>
```

---

# Project Structure

```text
Execution-Engine/
│
├── api/
│   ├── executor/
│   │   ├── python.go
│   │   ├── javascript.go
│   │   ├── cpp.go
│   │   ├── types.go
│   │   └── registry.go
│   │
│   ├── models/
│   │   ├── execution_job.go
│   │   └── execution_result.go
│   │
│   ├── redis/
│   │   └── redis.go
│   │
│   ├── main.go
│   ├── Dockerfile
│   └── go.mod
│
├── worker/
│   ├── main.go
│   ├── Dockerfile
│   └── go.mod
│
├── docker-images/
│   ├── python/
│   ├── javascript/
│   └── cpp/
│
├── docker-compose.yml
├── go.work
├── README.md
└── .gitignore
```

---

# Important Systems Concepts Implemented

- Distributed worker architecture
- Queue-based execution systems
- Containerized sandboxing
- Docker socket delegation
- Asynchronous job processing
- Persistent execution tracking
- Redis Pub/Sub
- Event-driven communication
- Runtime isolation
- Resource-constrained execution
- Temporary execution artifacts
- WebSocket infrastructure
- Multi-service orchestration
- Monorepo Go workspace architecture

---

# Current Limitations

- Final output streaming only
- No authentication
- No rate limiting
- No persistent database
- No execution history UI
- No autoscaling workers yet

---

# Roadmap

## Near-Term

- Incremental realtime log streaming
- Frontend playground UI
- Monaco editor integration
- Execution history
- PostgreSQL persistence
- Rate limiting
- API authentication

---

## Advanced Infrastructure

- Kubernetes deployment
- Autoscaling workers
- Priority queues
- Dead-letter queues
- Metrics dashboard
- Worker heartbeat monitoring
- Distributed tracing

---

# Future Security Improvements

- seccomp syscall filtering
- User namespace isolation
- Read-only root filesystem
- Non-root execution users
- Process count limits
- Filesystem quotas