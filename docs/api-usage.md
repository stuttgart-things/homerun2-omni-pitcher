# API Usage

## POST /pitch

Enqueue a message to Redis Streams.

### Request

```bash
curl -X POST http://localhost:8080/pitch \
  -H "Authorization: Bearer $AUTH_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "test-message",
    "repository": "my-repo",
    "payload": {"key": "value"}
  }'
```

### Response (Success)

```json
{
  "status": "success",
  "message": "Message enqueued",
  "objectID": "1234567890-0"
}
```

### Response (Error)

```json
{
  "status": "error",
  "message": "Failed to enqueue message to Redis"
}
```

### Authentication Errors

Missing or invalid token returns `401 Unauthorized`:

```json
{
  "status": "error",
  "message": "Missing Authorization header"
}
```

## GET /health

Returns `200 OK` when the service is running.

### Response

```json
{
  "status": "ok"
}
```

## Environment Variables

| Variable        | Default                                      | Description              |
|-----------------|----------------------------------------------|--------------------------|
| `PORT`          | `8080`                                       | HTTP server port         |
| `REDIS_ADDR`    | `redis-stack.homerun2.svc.cluster.local`     | Redis server address     |
| `REDIS_PORT`    | `6379`                                       | Redis server port        |
| `REDIS_STREAM`  | `messages`                                   | Redis stream name        |
| `AUTH_TOKEN`    | (required)                                   | Bearer token for auth    |
| `REDIS_PASSWORD`| (optional)                                   | Redis password           |
