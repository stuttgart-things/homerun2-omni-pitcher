# Homerun2 Omni Pitcher

An HTTP gateway microservice that receives pitch requests and enqueues them to Redis Streams.

## Overview

Omni Pitcher is part of the homerun2 platform. It provides a simple HTTP API that accepts JSON payloads and publishes them as messages to a Redis Stream for downstream processing.

## Quick Start

```bash
# Set required environment variables
export REDIS_ADDR=redis-stack.homerun2.svc.cluster.local
export REDIS_PORT=6379
export REDIS_STREAM=messages
export AUTH_TOKEN=your-secret-token

# Run locally
go run main.go
```

## API Endpoints

| Endpoint  | Method | Description                          |
|-----------|--------|--------------------------------------|
| `/pitch`  | POST   | Enqueue a message to Redis Streams   |
| `/pitch/grafana` | POST | Accept Grafana webhook alerts |
| `/pitch/github` | POST | Accept GitHub webhook events |
| `/health` | GET    | Health check                         |

## Authentication

All `/pitch` requests require a Bearer token:

```bash
curl -X POST http://localhost:8080/pitch \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"key": "value"}'
```

## Architecture

- **Go** stdlib `net/http` - HTTP server with graceful shutdown
- **Redis Streams** - Message queue via `homerun-library`
- **RediSearch** - Optional full-text indexing for analytics (dual-write, enabled via `REDIS_SEARCH_INDEX`)
- **ko** - Container image builds (distroless)
- **KCL** - Kubernetes manifest generation
- **Dagger** - CI/CD pipeline functions
