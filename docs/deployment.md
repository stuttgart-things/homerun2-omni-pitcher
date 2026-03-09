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

## Flux App Deployment

The recommended way to deploy the full homerun2 stack (Redis Stack + omni-pitcher + core-catcher) is via the [homerun2 Flux app](https://github.com/stuttgart-things/flux/tree/main/apps/homerun2). It uses Kustomize Components to deploy all services into a shared namespace with a single Flux Kustomization.

```yaml
---
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: homerun2-flux
  namespace: flux-system
spec:
  interval: 1h
  retryInterval: 1m
  timeout: 5m
  sourceRef:
    kind: GitRepository
    name: flux-apps
  path: ./apps/homerun2
  prune: true
  wait: true
  postBuild:
    substitute:
      HOMERUN2_NAMESPACE: homerun2-flux
      HOMERUN2_OMNI_PITCHER_VERSION: v1.2.0
      HOMERUN2_OMNI_PITCHER_HOSTNAME: pitcher
      HOMERUN2_CORE_CATCHER_VERSION: v0.5.0
      HOMERUN2_CORE_CATCHER_KUSTOMIZE_VERSION: v0.5.0-web
      HOMERUN2_CORE_CATCHER_HOSTNAME: catcher
      GATEWAY_NAME: my-gateway
      GATEWAY_NAMESPACE: default
      DOMAIN: my-cluster.example.com
      HOMERUN2_REDIS_VERSION: "17.1.4"
      HOMERUN2_REDIS_STORAGE_CLASS: nfs4-csi
      HOMERUN2_REDIS_STORAGE_SIZE: 8Gi
    substituteFrom:
      - kind: Secret
        name: homerun2-flux-secrets
```

The omni-pitcher component patches the KCL base to:

- Override the container image tag
- Override the Redis connection to point to the co-deployed redis-stack
- Patch the Redis password secret with the correct credentials
- Replace the KCL-generated Ingress with a Gateway API HTTPRoute (custom hostname)

See the [Flux app README](https://github.com/stuttgart-things/flux/tree/main/apps/homerun2) for all substitution variables and a complete example.

## Container Image

Built with [ko](https://ko.build/) using a distroless base image (`cgr.dev/chainguard/static:latest`):

```bash
# Build locally
ko build .

# Build via Taskfile
task build-scan-image-ko
```
