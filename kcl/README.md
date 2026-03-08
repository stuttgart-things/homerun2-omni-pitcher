# homerun2-omni-pitcher / KCL Deployment

KCL-based Kubernetes manifests for homerun2-omni-pitcher. Renders Namespace, ServiceAccount, ConfigMap, Secrets, Deployment, Service, and optionally Ingress or HTTPRoute.

## Render Manifests

### Via Dagger (recommended)

```bash
# render with a profile file
dagger call -m github.com/stuttgart-things/dagger/kcl@v0.82.0 run \
  --source kcl \
  --parameters-file tests/kcl-movie-scripts-profile.yaml \
  export --path /tmp/rendered-omni-pitcher.yaml

# render with inline parameters
dagger call -m github.com/stuttgart-things/dagger/kcl@v0.82.0 run \
  --source kcl \
  --parameters 'config.image=ghcr.io/stuttgart-things/homerun2-omni-pitcher:v1.1.2,config.namespace=homerun2' \
  export --path /tmp/rendered-omni-pitcher.yaml
```

### Via kcl CLI

```bash
kcl run kcl/main.k \
  -D 'config.image=ghcr.io/stuttgart-things/homerun2-omni-pitcher:v1.1.2' \
  -D 'config.namespace=homerun2'
```

## Deploy to Cluster

```bash
# render + apply
dagger call -m github.com/stuttgart-things/dagger/kcl@v0.82.0 run \
  --source kcl \
  --parameters-file tests/kcl-movie-scripts-profile.yaml \
  export --path /tmp/rendered-omni-pitcher.yaml

kubectl apply -f /tmp/rendered-omni-pitcher.yaml
```

## Profile Parameters

| Parameter | Default | Description |
|---|---|---|
| `config.image` | `ghcr.io/stuttgart-things/homerun2-omni-pitcher:latest` | Container image |
| `config.namespace` | `homerun2` | Target namespace |
| `config.replicas` | `1` | Replica count |
| `config.serviceType` | `ClusterIP` | Service type |
| `config.servicePort` | `80` | Service port |
| `config.containerPort` | `8080` | Container port |
| `config.redisAddr` | `redis-stack.homerun2.svc.cluster.local` | Redis host |
| `config.redisPort` | `6379` | Redis port |
| `config.redisStream` | `messages` | Redis stream name |
| `config.authToken` | *(empty)* | Bearer auth token (creates Secret if set) |
| `config.redisPassword` | *(empty)* | Redis password (creates Secret if set) |
| `config.ingressEnabled` | `false` | Enable Ingress |
| `config.ingressHost` | `homerun2-omni-pitcher.example.com` | Ingress hostname |
| `config.ingressTlsEnabled` | `false` | Enable Ingress TLS |
| `config.httpRouteEnabled` | `false` | Enable HTTPRoute (Gateway API) |
| `config.httpRouteParentRefName` | *(empty)* | Gateway name |
| `config.httpRouteParentRefNamespace` | *(empty)* | Gateway namespace |
| `config.httpRouteHostname` | *(empty)* | HTTPRoute hostname |

## Example Profiles

### movie-scripts cluster (HTTPRoute + redis-stack)

```yaml
---
config.image: ghcr.io/stuttgart-things/homerun2-omni-pitcher:v1.1.2
config.namespace: homerun2
config.ingressEnabled: false
config.httpRouteEnabled: true
config.httpRouteParentRefName: movie-scripts2-gateway
config.httpRouteParentRefNamespace: default
config.httpRouteHostname: homerun2-omni-pitcher.movie-scripts2.sthings-vsphere.labul.sva.de
config.redisAddr: redis-stack.redis-stack.svc.cluster.local
config.redisPort: "6379"
config.redisStream: messages
config.authToken: <your-token>
config.redisPassword: <your-password>
```

### dev cluster (Ingress + local Redis)

```yaml
---
config.image: ghcr.io/stuttgart-things/homerun2-omni-pitcher:latest
config.namespace: homerun2
config.ingressEnabled: true
config.ingressHost: homerun2-omni-pitcher.example.com
config.ingressTlsEnabled: true
config.ingressAnnotations:
  cert-manager.io/cluster-issuer: cluster-issuer-approle
config.redisAddr: redis-stack.homerun2.svc.cluster.local
config.redisPort: "6379"
config.redisStream: messages
config.authToken: changeme
config.redisPassword: changeme
```
