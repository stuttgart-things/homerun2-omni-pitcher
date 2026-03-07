# homerun2-omni-pitcher

A Go web service that provides an HTTP API for pitching messages to Redis Streams using the [homerun-library](https://github.com/stuttgart-things/homerun-library).

## Dev Mode (no Redis)

Run the pitcher without Redis by using file mode — messages are written as JSON lines to a local file instead of being enqueued into Redis Streams:

```bash
PITCHER_MODE=file AUTH_TOKEN=test go run .
```

Pitched messages are appended to `pitched.log` by default. Override with `PITCHER_FILE`:

```bash
PITCHER_MODE=file PITCHER_FILE=my-output.log AUTH_TOKEN=test go run .
```

Test it:

```bash
# Health check
curl http://localhost:8080/health

# Pitch a message
curl -X POST http://localhost:8080/pitch \
  -H "Authorization: Bearer test" \
  -H "Content-Type: application/json" \
  -d '{"title": "test", "message": "hello from dev mode"}'

# View pitched messages
cat pitched.log | jq .
```

## Deployment

```bash
helmfile apply -f \
git::https://github.com/stuttgart-things/helm.git@database/redis-stack.yaml.gotmpl \
--state-values-set storageClass=openebs-hostpath \
--state-values-set password="<REPLACE>" \
--state-values-set namespace=homerun2
```


## Features

- RESTful HTTP API for sending messages to Redis Streams
- Message queuing via Redis Streams with JSON storage
- Environment-based configuration
- Health check endpoint
- Simple curl-based interface

## Configuration

The service is configured via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8080` |
| `PITCHER_MODE` | Backend mode: `redis` or `file` | `redis` |
| `PITCHER_FILE` | Output file path (only when `PITCHER_MODE=file`) | `pitched.log` |
| `REDIS_ADDR` | Redis server address | `localhost` |
| `REDIS_PORT` | Redis server port | `6379` |
| `REDIS_PASSWORD` | Redis password | (empty) |
| `REDIS_STREAM` | Redis stream name | `messages` |
| `AUTH_MODE` | Auth mode: `token` or `jwt` | `token` |
| `AUTH_TOKEN` | Bearer token (when `AUTH_MODE=token`) | (required) |
| `JWT_JWKS_URL` | JWKS endpoint URL (when `AUTH_MODE=jwt`) | (required) |
| `JWT_ISSUER` | Expected JWT issuer (when `AUTH_MODE=jwt`) | (empty) |
| `JWT_AUDIENCE` | Expected JWT audience (when `AUTH_MODE=jwt`) | (empty) |
| `LOG_FORMAT` | Log format: `json` or `text` | `json` |
| `LOG_LEVEL` | Log level: `debug`, `info`, `warn`, `error` | `info` |

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
  -H "Authorization: Bearer <YOUR_AUTH_TOKEN>" \
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

## API Endpoints

- `GET /health` - Health check endpoint
- `POST /pitch` - Submit a message (to Redis Streams or file, depending on `PITCHER_MODE`). Requires `Authorization: Bearer <token>` header.

## Examples

A test script is provided in `examples/test-api.sh` to demonstrate API usage:

```bash
# Start the service first
./homerun2-omni-pitcher &

# Run the test script
./examples/test-api.sh
```

The script tests various scenarios including validation errors and successful message submissions.

## License

See [LICENSE](LICENSE) file.
