"""
Schema definition for homerun2-omni-pitcher deployment configuration
"""

schema OmniPitcher:
    """Configuration for homerun2-omni-pitcher Kubernetes deployment"""

    # Metadata
    name: str = "homerun2-omni-pitcher"
    namespace: str = "homerun2"

    # Image
    # Placeholder for local renders only — release overrides this with the
    # released version via kustomize-image-key. Points at :main so a local
    # render tracks main; :latest now tracks the newest release (#165).
    image: str = "ghcr.io/stuttgart-things/homerun2-omni-pitcher:main"
    imagePullPolicy: str = "IfNotPresent"

    # Replicas and resources
    replicas: int = 1
    cpuRequest: str = "100m"
    cpuLimit: str = "500m"
    memoryRequest: str = "128Mi"
    memoryLimit: str = "512Mi"

    # Service
    serviceType: str = "ClusterIP"
    servicePort: int = 80
    containerPort: int = 8080

    # Ingress
    ingressEnabled: bool = False
    ingressClassName: str = "nginx"
    ingressHost: str = "homerun2-omni-pitcher.example.com"
    ingressTlsEnabled: bool = False
    ingressTlsSecretName: str = "homerun2-omni-pitcher-tls"
    ingressAnnotations: {str:str} = {}

    # HTTPRoute (Gateway API)
    httpRouteEnabled: bool = False
    httpRouteParentRefName: str = ""
    httpRouteParentRefNamespace: str = ""
    httpRouteHostname: str = ""
    httpRouteAnnotations: {str:str} = {}

    # Redis
    redisAddr: str = "redis-stack.homerun2.svc.cluster.local"
    redisPort: int | str = "6379"
    redisStream: str = "messages"

    # Secrets
    authToken: str = ""
    redisPassword: str = ""

    # Config-driven stream routing (#105). The YAML content of the routing
    # file; empty leaves routing off, which is the legacy single-stream path
    # where REDIS_STREAM takes every message.
    #
    # Set, this renders a ConfigMap, mounts it, and points ROUTES_CONFIG at it
    # — without which the feature is in the binary but unreachable from a
    # kustomize deployment.
    #
    #     routesConfig: """
    #     streams: [messages, tabletennis]
    #     default_stream: messages
    #     routes:
    #       - match: { system: tabletennis }
    #         stream: tabletennis
    #     """
    routesConfig: str = ""

    # Where the routing file is mounted. ROUTES_CONFIG is set to
    # <routesPath>/routes.yaml.
    routesPath: str = "/config/routing"

    # Extra environment variables
    extraEnvVars: {str:str} = {}

    # Labels and annotations
    labels: {str:str} = {}
    annotations: {str:str} = {}

    check:
        replicas >= 0, "replicas must be >= 0"
        servicePort >= 1 and servicePort <= 65535, "servicePort must be valid"
        containerPort >= 1 and containerPort <= 65535, "containerPort must be valid"
        name != "", "name must not be empty"
        namespace != "", "namespace must not be empty"
