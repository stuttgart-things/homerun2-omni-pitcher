# homerun2-omni-pitcher

A Go HTTP microservice that accepts JSON messages via `POST /pitch` and enqueues them into Redis Streams using the [homerun-library](https://github.com/stuttgart-things/homerun-library).

[![Build & Test](https://github.com/stuttgart-things/homerun2-omni-pitcher/actions/workflows/build-test.yaml/badge.svg)](https://github.com/stuttgart-things/homerun2-omni-pitcher/actions/workflows/build-test.yaml)
[![Pages](https://stuttgart-things.github.io/homerun2-omni-pitcher/)](https://stuttgart-things.github.io/homerun2-omni-pitcher/)

## API Endpoints

| Endpoint | Method | Auth | Description |
|----------|--------|------|-------------|
| `/health` | `GET` | None | Health check (returns version, commit, date) |
| `/pitch` | `POST` | Bearer token or JWT | Submit a message to Redis Streams or file |

<details>
<summary><b>Pitch a message</b></summary>

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

</details>

<details>
<summary><b>Message fields</b></summary>

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

</details>

## Deployment

<details>
<summary><b>Download binary from releases</b></summary>

Pre-built binaries are available for Linux and macOS (amd64/arm64) on the [Releases](https://github.com/stuttgart-things/homerun2-omni-pitcher/releases) page.

```bash
# Download latest release (example: linux amd64)
VERSION=$(gh release view --json tagName -q .tagName)
curl -Lo homerun2-omni-pitcher.tar.gz \
  "https://github.com/stuttgart-things/homerun2-omni-pitcher/releases/download/${VERSION}/homerun2-omni-pitcher_${VERSION#v}_linux_amd64.tar.gz"

tar xzf homerun2-omni-pitcher.tar.gz
chmod +x homerun2-omni-pitcher

# Run with Redis
export REDIS_ADDR=localhost REDIS_PORT=6379 REDIS_STREAM=messages AUTH_TOKEN=mysecret
./homerun2-omni-pitcher
```

</details>

<details>
<summary><b>Container image (ko / ghcr.io)</b></summary>

The container image is built with [ko](https://ko.build) on top of `cgr.dev/chainguard/static` and published to GitHub Container Registry.

```bash
# Pull the image
docker pull ghcr.io/stuttgart-things/homerun2-omni-pitcher:<tag>

# Run with Docker
docker run -p 8080:8080 \
  -e REDIS_ADDR=redis -e REDIS_PORT=6379 \
  -e REDIS_STREAM=messages -e AUTH_TOKEN=mysecret \
  ghcr.io/stuttgart-things/homerun2-omni-pitcher:<tag>
```

</details>

<details>
<summary><b>Deploy to Kubernetes with KCL</b></summary>

KCL manifests in `kcl/` are the source of truth for Kubernetes deployment. The modular KCL modules cover: `deploy.k`, `service.k`, `ingress.k`, `secret.k`, `configmap.k`, `serviceaccount.k`, `namespace.k`, `httproute.k`.

**Render manifests locally:**

```bash
# Render with kcl CLI
kcl run kcl/ -Y tests/kcl-deploy-profile.yaml

# Render via Dagger (non-interactive)
task render-manifests-quick

# Render via Dagger (interactive, prompts for source/profile/output)
task render-manifests
```

**Deploy to a cluster using the Dagger `kubernetes-deployment` blueprint:**

```bash
# Deploy with defaults (uses Taskfile vars)
task deploy-kcl

# Deploy with custom parameters
task deploy-kcl \
  OCI_SOURCE=ghcr.io/stuttgart-things/homerun2-omni-pitcher-kustomize \
  PARAMETERS='namespace=homerun2' \
  NAMESPACE=homerun2 \
  KUBECONFIG=~/.kube/movie-scripts
```

This uses the [`kubernetes-deployment`](https://github.com/stuttgart-things/blueprints) Dagger module (v1.68.0) which renders KCL from an OCI source and applies it to the target cluster.

**Build and push kustomize base as OCI artifact:**

```bash
# Render kustomize base from KCL
task render-kustomize-base

# Push to OCI registry (interactive, prompts for tag)
task push-kustomize-base
```

</details>

<details>
<summary><b>Deploy Redis (prerequisite)</b></summary>

```bash
helmfile apply -f \
  git::https://github.com/stuttgart-things/helm.git@database/redis-stack.yaml.gotmpl \
  --state-values-set storageClass=openebs-hostpath \
  --state-values-set password="<REPLACE>" \
  --state-values-set namespace=homerun2
```

</details>

## Development

<details>
<summary><b>Dev mode (no Redis needed)</b></summary>

Run the pitcher without Redis using file mode — messages are written as JSON lines to a local file:

```bash
PITCHER_MODE=file AUTH_TOKEN=test go run .
```

Messages are appended to `pitched.log` by default. Override with `PITCHER_FILE`:

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

</details>

<details>
<summary><b>Project structure</b></summary>

```
main.go                    # Entrypoint, routing, graceful shutdown
internal/
  banner/                  # Animated startup banner (Bubble Tea)
  config/                  # Env-based config loading, slog setup
  handlers/                # HTTP handlers (pitch, health)
  middleware/              # Auth (token + JWT/JWKS), request logging
  models/                  # Response structs
  pitcher/                 # Pitcher interface (Redis + File backends)
kcl/                       # Kubernetes manifests (modular KCL)
dagger/                    # CI functions (Lint, Build, Test, Scan)
.ko.yaml                   # ko build configuration
Taskfile.yaml              # Task runner
```

</details>

<details>
<summary><b>Configuration reference</b></summary>

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8080` |
| `PITCHER_MODE` | Backend mode: `redis` or `file` | `redis` |
| `PITCHER_FILE` | Output file path (file mode only) | `pitched.log` |
| `REDIS_ADDR` | Redis server address | `localhost` |
| `REDIS_PORT` | Redis server port | `6379` |
| `REDIS_PASSWORD` | Redis password | (empty) |
| `REDIS_STREAM` | Redis stream name | `messages` |
| `AUTH_MODE` | Auth mode: `token` or `jwt` | `token` |
| `AUTH_TOKEN` | Bearer token (token mode) | (required) |
| `JWT_JWKS_URL` | JWKS endpoint URL (jwt mode) | (required) |
| `JWT_ISSUER` | Expected JWT issuer (jwt mode) | (optional) |
| `JWT_AUDIENCE` | Expected JWT audience (jwt mode) | (optional) |
| `LOG_FORMAT` | Log format: `json` or `text` | `json` |
| `LOG_LEVEL` | Log level: `debug`, `info`, `warn`, `error` | `info` |
| `LOG_HEALTH_CHECKS` | Log `/health` probe requests (`true`/`false`) | `false` |

</details>

<details>
<summary><b>CI/CD and release process</b></summary>

Releases are fully automated via GitHub Actions and [semantic-release](https://semantic-release.gitbook.io/).

**Workflow chain on merge to `main`:**

1. **Build, Push & Scan Container Image** — builds the container image with ko, pushes to `ghcr.io`, and scans with Trivy
2. **Release** (triggered on successful image build) — runs semantic-release which:
   - Analyzes commit messages using [conventional commits](https://www.conventionalcommits.org/)
   - `fix:` → patch bump (e.g., 1.1.2 → 1.1.3)
   - `feat:` → minor bump (e.g., 1.1.2 → 1.2.0)
   - Creates a GitHub release with changelog
   - Stages the container image to `ghcr.io` with the release tag
   - Pushes the kustomize base as OCI artifact to `ghcr.io/stuttgart-things/homerun2-omni-pitcher-kustomize`

**Trigger a release manually:**

```bash
# Via GitHub Actions UI or CLI
task trigger-release

# Local release pipeline (interactive)
task release-local
```

**Branch naming convention:**

- `fix/<issue-number>-<short-description>` — bug fixes (patch)
- `feat/<issue-number>-<short-description>` — new features (minor)
- `test/<issue-number>-<short-description>` — test-only changes (no release)

</details>

## Testing

<details>
<summary><b>Unit tests</b></summary>

Unit tests run without Redis:

```bash
go test ./...
```

</details>

<details>
<summary><b>Integration tests (Dagger + Redis)</b></summary>

Integration tests spin up a Redis service via Dagger:

```bash
task build-test-binary
```

</details>

<details>
<summary><b>Lint</b></summary>

```bash
task lint
```

</details>

<details>
<summary><b>Build and scan container image</b></summary>

```bash
task build-scan-image-ko
```

</details>

## Links

- [GitHub Pages](https://stuttgart-things.github.io/homerun2-omni-pitcher/)
- [Releases](https://github.com/stuttgart-things/homerun2-omni-pitcher/releases)
- [Container Images](https://github.com/stuttgart-things/homerun2-omni-pitcher/pkgs/container/homerun2-omni-pitcher)
- [homerun-library](https://github.com/stuttgart-things/homerun-library)

## License

See [LICENSE](LICENSE) file.
