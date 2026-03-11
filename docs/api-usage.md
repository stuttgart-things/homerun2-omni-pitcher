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

## POST /pitch/grafana

Accept Grafana webhook contact point payloads and enqueue each alert as a message.

### Request

Configure a [Grafana webhook contact point](https://grafana.com/docs/grafana/latest/alerting/configure-notifications/manage-contact-points/integrations/webhook-notifier/) with the URL `https://<pitcher-host>/pitch/grafana` and add the `Authorization: Bearer $AUTH_TOKEN` header.

Grafana sends payloads automatically when alerts fire or resolve. Each alert is mapped to a `homerun.Message`:

| Grafana field | Message field |
|---|---|
| `labels.alertname` | `title` |
| `annotations.summary` / `description` | `message` |
| `labels.severity` + `status` | `severity` |
| `startsAt` | `timestamp` |
| `receiver` | `system` |
| `dashboardURL` / `panelURL` / `generatorURL` | `url` |
| remaining labels | `tags` (comma-separated) |

### Response (Success)

```json
{
  "status": "success",
  "message": "2 of 2 alerts enqueued",
  "results": [
    {"objectId": "...", "streamId": "messages", "status": "success", "message": "Alert abc123 enqueued"}
  ],
  "errors": []
}
```

### Response (Partial failure)

If some alerts fail to enqueue but others succeed, the endpoint returns `200` with error details:

```json
{
  "status": "success",
  "message": "1 of 2 alerts enqueued",
  "results": [...],
  "errors": ["alert xyz: redis connection refused"]
}
```

### Response (Total failure)

If all alerts fail, the endpoint returns `503`:

```json
{
  "status": "error",
  "message": "Failed to enqueue all alerts",
  "errors": ["alert xyz: redis connection refused"]
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
| `REDIS_SEARCH_INDEX` | (empty)                                 | RediSearch index name (enables dual-write) |
| `AUTH_TOKEN`    | (required)                                   | Bearer token for auth    |
| `REDIS_PASSWORD`| (optional)                                   | Redis password           |
