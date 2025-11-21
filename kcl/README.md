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

## Usage

### Generate YAML Output

```bash
kcl run --format yaml \
| yq eval -P '.items[]' - \
| awk 'BEGIN{doc=""} /^apiVersion: /{if(doc!=""){print "---";} doc=1} {print}'
```