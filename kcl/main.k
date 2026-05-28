"""
Main entry point for homerun2-omni-pitcher Kubernetes deployment
Imports all resources and exports manifests
"""

import serviceaccount
import configmap
import secret
import deploy
import service
import ingress
import httproute

# Export all manifests (filter out None values).
# Namespace is NOT emitted from the KCL profile: when multiple co-tenanted
# Applications (omni-pitcher + core-catcher + …) all ship a Namespace resource
# into the same workload namespace, Argo flags it as a SharedResource and one
# App's prune cycle deletes the ns out from under the others. The consuming
# Application sets `syncOptions: CreateNamespace=true` (and PR-preview
# namespaces are pre-created by Kyverno) so nothing here needs to emit it.
# See stuttgart-things/homerun2-omni-pitcher#132.
manifests = [
    m for m in [
        serviceaccount.serviceAccount
        configmap.configMap
        secret.secretToken
        secret.secretRedis
        deploy.deployment
        service.service
        ingress.ingress
        httproute.httproute
    ] if m
]
