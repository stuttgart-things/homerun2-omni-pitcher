"""
Namespace for homerun2-omni-pitcher
"""

import labels

config = labels.config

namespace = {
    apiVersion: "v1"
    kind: "Namespace"
    metadata: {
        name: config.namespace
        labels: labels.commonLabels
    }
}
