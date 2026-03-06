"""
Schema definition for homerun2-omni-pitcher deployment configuration
"""

schema OmniPitcher:
    """Configuration for homerun2-omni-pitcher Kubernetes deployment"""

    # Metadata
    name: str = "homerun2-omni-pitcher"
    namespace: str = "homerun2"

    # Image
    image: str = "ghcr.io/stuttgart-things/homerun2-omni-pitcher:latest"
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
