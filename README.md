# Execution-Engine

A distributed sandboxed code execution platform built with Go and Docker.

Execution-Engine safely runs untrusted user code inside isolated Docker containers with strict resource limits, execution timeouts, stdin support, and runtime isolation.

---

# Features

## Current Features

* Sandboxed code execution using Docker
* Dynamic Python code execution
* Secure temporary file generation
* Execution timeout protection
* CPU and memory limits
* Disabled container networking
* stdin support
* stdout/stderr capture
* Automatic temp file cleanup
* Modular execution architecture
* REST API built with Gin

---

# Architecture

```text
Client
  ↓
Go API Server
  ↓
Execution Layer
  ↓
Docker Sandbox
  ↓
Python Runtime
  ↓
Captured Output
  ↓
JSON Response
```

---

# Tech Stack

## Backend

* Go
* Gin

## Sandbox Runtime

* Docker
* Python 3.11

## Planned Infrastructure

* Redis
* Distributed Workers
* WebSockets
* PostgreSQL

---

# Security Features

The execution engine applies multiple isolation and resource control mechanisms:

* Docker container isolation
* Network disabled using `--network=none`
* Memory limit using `--memory=128m`
* CPU limit using `--cpus=0.5`
* Execution timeout protection
* Automatic temporary file cleanup

---

# API

## Execute Code

### Endpoint

```http
POST /execute
```

### Request Body

```json
{
  "code": "name = input()\nprint('hello', name)",
  "input": "Gautam\n"
}
```

### Response

```json
{
  "output": "hello Gautam\n"
}
```

---

# Local Setup

## Clone Repository

```bash
git clone <repo-url>
cd Execution-Engine
```

---

## Install Go Dependencies

```bash
cd api
go mod tidy
```

---

## Build Docker Runtime

### Python Runtime

```bash
cd docker-images/python
docker build -t execution-python .
```

---

## Run API Server

```bash
cd api
go run main.go
```

Server runs on:

```text
http://localhost:8080
```

---

# Example Usage

```bash
curl -X POST http://localhost:8080/execute \
-H "Content-Type: application/json" \
-d '{"code":"print(\"hello from execution engine\")"}'
```

---

# Project Structure

```text
Execution-Engine/
├── api/
│   ├── executor/
│   │   └── executor.go
│   ├── main.go
│   └── go.mod
│
├── docker-images/
│   └── python/
│       └── Dockerfile
│
├── frontend/
├── worker/
└── .gitignore
```

---

# Roadmap

## Near-Term

* JavaScript runtime support
* C++ compilation pipeline
* Redis-backed job queue
* Distributed worker architecture
* Real-time execution logs

## Long-Term

* Kubernetes autoscaling
* Multi-user execution system
* Judge/testcase engine
* Execution history
* Authentication
* Rate limiting
* Metrics dashboard

---

# Important Concepts Implemented

* Process orchestration
* Containerized execution
* Runtime isolation
* Secure code execution
* Resource-constrained sandboxing
* Temporary execution artifacts
* stdin/stdout process piping
* Timeout-based execution cancellation

---

# Recommended Additional Files

## .gitignore

```gitignore
# Go
bin/
*.exe
*.out

# Node
node_modules/

# macOS
.DS_Store

# Environment files
.env

# Temporary files
tmp/
```

---

## Future Repo Improvements

* Add GitHub Actions CI
* Add API tests
* Add Swagger/OpenAPI docs
* Add Docker Compose setup
* Add architecture diagrams
* Add benchmarks and stress tests
