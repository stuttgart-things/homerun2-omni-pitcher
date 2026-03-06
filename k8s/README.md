# Kubernetes Manifests for homerun2-omni-pitcher

## Overview

This directory contains Kubernetes manifests for deploying the homerun2-omni-pitcher service with production-ready security configurations.

## Files

- **deployment.yaml**: Main deployment with 2 replicas, resource limits, security contexts, and health probes
- **service.yaml**: ClusterIP service exposing port 80
- **serviceaccount.yaml**: ServiceAccount with automountServiceAccountToken disabled
- **configmap.yaml**: ConfigMap for non-sensitive configuration
- **secret-token.yaml**: Secret for authentication token
- **secret-redis.yaml**: Secret for Redis connection details
- **ingress.yaml**: Ingress resource with TLS and security headers
- **kustomization.yaml**: Kustomize configuration for easy deployment

## Security Features

### Pod Security
- **runAsNonRoot**: true (runs as user 65532)
- **readOnlyRootFilesystem**: true
- **allowPrivilegeEscalation**: false
- **capabilities**: All dropped
- **seccompProfile**: RuntimeDefault

### Resources
- **Requests**: 100m CPU, 128Mi memory
- **Limits**: 500m CPU, 512Mi memory

### High Availability
- 2 replicas with pod anti-affinity
- Rolling update strategy with zero downtime
- Liveness and readiness probes

## Prerequisites

Before deploying, you need to:

1. **Update secrets** in `secret-token.yaml` and `secret-redis.yaml`
2. **Update domain** in `ingress.yaml`
3. **Ensure Redis is available** in your cluster

## Deployment

### Quick Deploy with kubectl

```bash
# Update secrets first!
kubectl apply -k .
```

### Step-by-Step Deployment

```bash
# 1. Create namespace (optional)
kubectl create namespace homerun2

# 2. Update secrets
# Generate a secure token
openssl rand -base64 32

# Edit secret-token.yaml and secret-redis.yaml with actual values
vi secret-token.yaml
vi secret-redis.yaml

# 3. Update ingress domain
vi ingress.yaml

# 4. Apply manifests
kubectl apply -k . -n homerun2

# 5. Verify deployment
kubectl get pods -n homerun2
kubectl get svc -n homerun2
kubectl get ingress -n homerun2
```

## Configuration

### Environment Variables (ConfigMap)

Edit `configmap.yaml` to add non-sensitive configuration:

```yaml
data:
  LOG_LEVEL: "info"
  # Add more configuration here
```

### Secrets

#### Authentication Token (secret-token.yaml)

```bash
# Generate a secure token
openssl rand -base64 32

# Update secret-token.yaml
kubectl create secret generic homerun2-omni-pitcher-token \
  --from-literal=auth-token=YOUR_TOKEN_HERE \
  --dry-run=client -o yaml > secret-token.yaml
```

#### Redis Connection (secret-redis.yaml)

Update with your Redis connection details:
- `address`: Redis server address (e.g., `redis-master.redis.svc.cluster.local:6379`)
- `password`: Redis password

### Ingress

Update the following in `ingress.yaml`:
- Replace `homerun2-omni-pitcher.example.com` with your actual domain
- Uncomment cert-manager annotation if using cert-manager for TLS

## Monitoring

```bash
# Check pod status
kubectl get pods -l app=homerun2-omni-pitcher

# View logs
kubectl logs -l app=homerun2-omni-pitcher --tail=100 -f

# Check health endpoint
kubectl port-forward svc/homerun2-omni-pitcher 8080:80
curl http://localhost:8080/health
```

## Testing

```bash
# Port forward to test locally
kubectl port-forward svc/homerun2-omni-pitcher 8080:80

# Test health endpoint
curl http://localhost:8080/health

# Test pitch endpoint (requires token)
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X POST http://localhost:8080/pitch \
     -d '{"data": "test"}'
```

## Scaling

```bash
# Scale replicas
kubectl scale deployment homerun2-omni-pitcher --replicas=3

# Or edit deployment
kubectl edit deployment homerun2-omni-pitcher
```

## Testing

### Test Jobs

Three test jobs are available to validate your deployment:

#### 1. Internal Service Test (test-job.yaml)

Tests the service via ClusterIP (internal cluster communication):

```bash
# Run internal tests
kubectl apply -f test-job.yaml

# Watch test progress
kubectl logs -f job/homerun2-omni-pitcher-test

# Check test results
kubectl get jobs homerun2-omni-pitcher-test
```

Tests performed:
- ✓ Health endpoint availability
- ✓ Pitch endpoint with valid token
- ✓ Authorization (no token - should fail)
- ✓ Authorization (invalid token - should fail)
- ✓ Service DNS resolution

#### 2. External/Ingress Test (test-job-external.yaml)

Tests the service via Ingress (external access):

```bash
# Update EXTERNAL_URL in test-job-external.yaml first!
vi test-job-external.yaml

# Run external tests
kubectl apply -f test-job-external.yaml

# Watch test progress
kubectl logs -f job/homerun2-omni-pitcher-test-external
```

Tests performed:
- ✓ Health endpoint via ingress
- ✓ Pitch endpoint via ingress
- ✓ TLS certificate validation (if HTTPS)

#### 3. Load Test (test-job-load.yaml)

Runs a basic load test with 5 parallel pods, each sending 50 requests:

```bash
# Run load test (250 total requests)
kubectl apply -f test-job-load.yaml

# Watch all test pods
kubectl get pods -l app=homerun2-omni-pitcher-load-test -w

# View results from all pods
kubectl logs -l app=homerun2-omni-pitcher-load-test

# View specific pod results
kubectl logs job/homerun2-omni-pitcher-load-test
```

#### Clean Up Test Jobs

```bash
# Delete test jobs
kubectl delete job homerun2-omni-pitcher-test
kubectl delete job homerun2-omni-pitcher-test-external
kubectl delete job homerun2-omni-pitcher-load-test

# Or delete all at once
kubectl delete job -l app=homerun2-omni-pitcher
```

## Troubleshooting

```bash
# Check pod events
kubectl describe pod -l app=homerun2-omni-pitcher

# Check logs
kubectl logs -l app=homerun2-omni-pitcher --all-containers=true

# Check service endpoints
kubectl get endpoints homerun2-omni-pitcher

# Test DNS resolution
kubectl run -it --rm debug --image=busybox --restart=Never -- nslookup homerun2-omni-pitcher

# Run quick curl test
kubectl run -it --rm curl-test --image=curlimages/curl --restart=Never -- \
  curl -v http://homerun2-omni-pitcher.default.svc.cluster.local/health
```

## Updating the Image

To update to a new image version, edit `kustomization.yaml`:

```yaml
images:
- name: homerun2-omni-pitcher
  newName: ghcr.io/stuttgart-things/homerun2-omni-pitcher/homerun2-omni-pitcher-NEW-HASH
  digest: sha256:NEW-SHA256-HASH
```

Then apply:

```bash
kubectl apply -k .
kubectl rollout status deployment/homerun2-omni-pitcher
```

## Security Considerations

1. **Always update secrets** before deploying to production
2. **Use cert-manager** or provide your own TLS certificates
3. **Enable network policies** to restrict pod communication
4. **Use a secrets management solution** (Vault, Sealed Secrets, etc.) for production
5. **Review and adjust resource limits** based on your workload
6. **Enable monitoring and alerting** for production deployments
