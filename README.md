# homerun2-omni-pitcher

A Go web service that provides an HTTP API for pitching messages to Redis Streams using the [homerun-library](https://github.com/stuttgart-things/homerun-library).

## Features

- RESTful HTTP API for sending messages to Redis Streams
- Message queuing via Redis Streams with JSON storage
- Environment-based configuration
- Health check endpoint
- Simple curl-based interface

## Installation

```bash
go build -o homerun2-omni-pitcher
```

## Configuration

The service is configured via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8080` |
| `REDIS_ADDR` | Redis server address | `localhost` |
| `REDIS_PORT` | Redis server port | `6379` |
| `REDIS_PASSWORD` | Redis password | (empty) |
| `REDIS_STREAM` | Redis stream name | `messages` |

## Usage

### Start the service

```bash
export REDIS_ADDR=localhost
export REDIS_PORT=6379
export REDIS_STREAM=messages
./homerun2-omni-pitcher
```

### Health Check

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "healthy",
  "time": "2025-11-15T15:30:00Z"
}
```

### Pitch a Message

Send a POST request to `/pitch` with a JSON payload:

```bash
curl -X POST http://localhost:8080/pitch \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Deployment Notification",
    "message": "Service xyz deployed successfully",
    "severity": "success",
    "author": "ci-pipeline",
    "system": "demo-system",
    "tags": "deployment,production,success",
    "assigneeaddress": "ops-team@example.com",
    "assigneename": "Ops Team",
    "artifacts": "docker://registry.example.com/xyz:1.0.0",
    "url": "http://example.com/deployment/xyz"
  }'
```

Response:
```json
{
  "objectId": "550e8400-e29b-41d4-a716-446655440000-demo-system",
  "streamId": "messages",
  "status": "success",
  "message": "Message successfully enqueued"
}
```

### Message Fields

| Field | Required | Description | Default |
|-------|----------|-------------|---------|
| `title` | Yes | Short title of the message | - |
| `message` | Yes | The actual message content | - |
| `severity` | No | Severity level (info, warning, error, success) | `info` |
| `author` | No | Creator of the message | `unknown` |
| `timestamp` | No | ISO-8601 timestamp | Current time |
| `system` | No | Originating system | `homerun2-omni-pitcher` |
| `tags` | No | Comma-separated list of tags | - |
| `assigneeaddress` | No | Email or address of the assignee | - |
| `assigneename` | No | Name of the assignee | - |
| `artifacts` | No | Related artifacts (e.g., container image) | - |
| `url` | No | Related URL | - |

## Docker

### Using docker-compose (recommended)

The easiest way to run the service with Redis:

```bash
docker-compose up
```

This will start both Redis and the homerun2-omni-pitcher service.

### Building and running manually

```bash
docker build -t homerun2-omni-pitcher .
docker run -p 8080:8080 \
  -e REDIS_ADDR=redis \
  -e REDIS_PORT=6379 \
  -e REDIS_STREAM=messages \
  homerun2-omni-pitcher
```

## API Endpoints

- `GET /health` - Health check endpoint
- `POST /pitch` - Submit a message to Redis Streams

## Examples

A test script is provided in `examples/test-api.sh` to demonstrate API usage:

```bash
# Start the service first (with docker-compose or directly)
docker-compose up -d

# Run the test script
./examples/test-api.sh
```

The script tests various scenarios including validation errors and successful message submissions.

## License

See [LICENSE](LICENSE) file.

