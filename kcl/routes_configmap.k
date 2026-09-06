"""
Routing ConfigMap for homerun2-omni-pitcher

Holds the stream routing rules (#105) mounted at `routesPath`. Only rendered
when routesConfig is set; without it the deployment keeps the single-stream
path and never mounts anything.
"""

import labels

config = labels.config

routesConfigMap = {
    apiVersion: "v1"
    kind: "ConfigMap"
    metadata: {
        name: "{}-routes".format(config.name)
        namespace: config.namespace
        labels: labels.commonLabels
    }
    data: {
        "routes.yaml": config.routesConfig
    }
} if config.routesConfig else None
