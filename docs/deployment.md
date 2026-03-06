# Deployment

## Kubernetes Manifests (KCL)

Manifests are generated using KCL in the `kcl/` directory. The modular structure:

| File               | Resource        |
|--------------------|-----------------|
| `schema.k`         | Config schema with validation |
| `labels.k`         | Common labels and config instantiation |
| `namespace.k`      | Namespace       |
| `serviceaccount.k` | ServiceAccount  |
| `configmap.k`      | ConfigMap       |
| `secret.k`         | Secrets (auth token, redis password) |
| `deploy.k`         | Deployment      |
| `service.k`        | Service         |
| `ingress.k`        | Ingress (optional) |
| `httproute.k`      | HTTPRoute via Gateway API (optional) |
| `main.k`           | Entry point     |

## Render Manifests

```bash
# Using Taskfile
task render-manifests

# Using KCL directly
kcl kcl/main.k -Y tests/kcl-deploy-profile.yaml
```

## Configuration Options

Override via KCL profile file or CLI options:

```yaml
config:
  image: ghcr.io/stuttgart-things/homerun2-omni-pitcher:v1.0.0
  namespace: homerun2
  ingressEnabled: true
  ingressHost: pitcher.example.com
  httpRouteEnabled: false
  redisAddr: redis.default.svc.cluster.local
  redisPort: "6379"
  redisStream: messages
  authToken: my-secret-token
  redisPassword: redis-pass
```

## Kustomize OCI Pipeline

Releases push a kustomize base as an OCI artifact:

```bash
# Pull the base
oras pull ghcr.io/stuttgart-things/homerun2-omni-pitcher-kustomize:v1.0.0

# Apply with overlays
kubectl apply -k .
```

## Container Image

Built with [ko](https://ko.build/) using a distroless base image (`cgr.dev/chainguard/static:latest`):

```bash
# Build locally
ko build .

# Build via Taskfile
task build-scan-image-ko
```
