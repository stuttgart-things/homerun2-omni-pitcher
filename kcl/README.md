# homerun2-omni-pitcher KCL Module

This KCL module generates Kubernetes manifests for the homerun2-omni-pitcher service using the official [k8s KCL module](https://artifacthub.io/packages/kcl/kcl-module/k8s).

## Overview

The module creates the following Kubernetes resources:
- ServiceAccount with security hardening
- ConfigMap for application configuration
- Secret for authentication token
- Secret for Redis connection details
- Deployment with security contexts and resource limits
- Service (ClusterIP)
- Ingress with TLS and nginx annotations

## Prerequisites

- KCL CLI installed ([installation guide](https://kcl-lang.io/docs/user_docs/getting-started/install))
- Access to a Kubernetes cluster
- kubectl configured

## Installation

The module dependencies are already configured in `kcl.mod`. To install/update them:

```bash
kcl mod download
```

## Configuration

Configuration can be provided in three ways (in order of precedence):

### 1. Using Command-Line Flags (Highest Priority)

```bash
kcl run main.k \
--format yaml \
-D hostname=api \
-D redisPassword="Test123" \ # pragma: allowlist secret
-D domain=mycompany.io \
-D clusterIssuer="letsencrypt-prod" \
-D namespace=staging \
| yq eval -P '.items[]' - \
| sed '/^apiVersion:/s/^/---\n/' \
| tail -n +2
```
