import k8s.api.core.v1 as corev1
import k8s.api.apps.v1 as appsv1
import k8s.api.networking.v1 as networkingv1
import k8s.apimachinery.pkg.apis.meta.v1 as metav1
import regex
import datetime
import crypto

# Helper function to generate a random token from timestamp and hash
# Note: This generates a deterministic token based on timestamp. For production,
# set authToken explicitly via -D flag or kcl.yaml
_generateToken = lambda {
    _timestamp = str(datetime.now())
    _seed = "homerun2-omni-pitcher-${_timestamp}"
    _hash = crypto.sha256(_seed)
    # Take first 32 characters of the hash
    _hash[:32]
}

# Configuration variables - can be overridden via -D flag
hostname = option("hostname") or "homerun2-omni-pitcher"
domain = option("domain") or "example.com"
clusterIssuer = option("clusterIssuer") or "cluster-issuer-approle"
authToken = option("authToken") or _generateToken()
redisAddress = option("redisAddress") or "redis-stack.homerun2.svc.cluster.local:6379"
redisPassword = option("redisPassword") or "CHANGE-ME-TO-ACTUAL-PASSWORD"

# Common labels
appName = "homerun2-omni-pitcher"
namespace = option("namespace") or "homerun2"
fqdn = "${hostname}.${domain}"

commonLabels = {
    app = appName
    "app.kubernetes.io/name" = appName
    "app.kubernetes.io/component" = "api"
    "app.kubernetes.io/part-of" = "homerun2"
}

# ServiceAccount
serviceAccount = corev1.ServiceAccount {
    metadata = metav1.ObjectMeta {
        name = appName
        namespace = namespace
        labels = commonLabels
    }
    automountServiceAccountToken = False
}

# ConfigMap
configMap = corev1.ConfigMap {
    metadata = metav1.ObjectMeta {
        name = "${appName}-config"
        namespace = namespace
        labels = commonLabels
    }
    data = {
        LOG_LEVEL = "info"
    }
}

# Secret - Token
secretToken = corev1.Secret {
    metadata = metav1.ObjectMeta {
        name = "${appName}-token"
        namespace = namespace
        labels = commonLabels
    }
    type = "Opaque"
    stringData = {
        "auth-token" = authToken
    }
}

# Secret - Redis
secretRedis = corev1.Secret {
    metadata = metav1.ObjectMeta {
        name = "${appName}-redis"
        namespace = namespace
        labels = commonLabels
    }
    type = "Opaque"
    stringData = {
        address = redisAddress
        password = redisPassword
    }
}

# Deployment
deployment = appsv1.Deployment {
    metadata = metav1.ObjectMeta {
        name = appName
        namespace = namespace
        labels = commonLabels
    }
    spec = appsv1.DeploymentSpec {
        replicas = 1
        selector = metav1.LabelSelector {
            matchLabels = {
                app = appName
            }
        }
        strategy = appsv1.DeploymentStrategy {
            type = "RollingUpdate"
            rollingUpdate = appsv1.RollingUpdateDeployment {
                maxSurge = 1
                maxUnavailable = 0
            }
        }
        template = corev1.PodTemplateSpec {
            metadata = metav1.ObjectMeta {
                labels = commonLabels
            }
            spec = corev1.PodSpec {
                securityContext = corev1.PodSecurityContext {
                    runAsNonRoot = True
                    runAsUser = 65532
                    runAsGroup = 65532
                    fsGroup = 65532
                    seccompProfile = corev1.SeccompProfile {
                        type = "RuntimeDefault"
                    }
                }
                serviceAccountName = appName
                containers = [
                    corev1.Container {
                        name = appName
                        image = "ghcr.io/stuttgart-things/homerun2-omni-pitcher/homerun2-omni-pitcher-19372b3164b7560205efc0b3c53f43c1@sha256:a5a828d899040e57f116375922f376e73decdc86302e3a9f375077b64d66b711"
                        imagePullPolicy = "IfNotPresent"
                        ports = [
                            corev1.ContainerPort {
                                name = "http"
                                containerPort = 8080
                                protocol = "TCP"
                            }
                        ]
                        env = [
                            corev1.EnvVar {
                                name = "PORT"
                                value = "8080"
                            }
                            corev1.EnvVar {
                                name = "AUTH_TOKEN"
                                valueFrom = corev1.EnvVarSource {
                                    secretKeyRef = corev1.SecretKeySelector {
                                        name = "${appName}-token"
                                        key = "auth-token"
                                    }
                                }
                            }
                            corev1.EnvVar {
                                name = "REDIS_ADDR"
                                value = "redis-stack.homerun2.svc.cluster.local"
                            }
                            corev1.EnvVar {
                                name = "REDIS_PORT"
                                value = "6379"
                            }
                            corev1.EnvVar {
                                name = "REDIS_PASSWORD"
                                valueFrom = corev1.EnvVarSource {
                                    secretKeyRef = corev1.SecretKeySelector {
                                        name = "${appName}-redis"
                                        key = "password"
                                    }
                                }
                            }
                            corev1.EnvVar {
                                name = "REDIS_STREAM"
                                value = "messages"
                            }
                        ]
                        envFrom = [
                            corev1.EnvFromSource {
                                configMapRef = corev1.ConfigMapEnvSource {
                                    name = "${appName}-config"
                                }
                            }
                        ]
                        livenessProbe = corev1.Probe {
                            httpGet = corev1.HTTPGetAction {
                                path = "/health"
                                port = "http"
                            }
                            initialDelaySeconds = 10
                            periodSeconds = 10
                            timeoutSeconds = 5
                            successThreshold = 1
                            failureThreshold = 3
                        }
                        readinessProbe = corev1.Probe {
                            httpGet = corev1.HTTPGetAction {
                                path = "/health"
                                port = "http"
                            }
                            initialDelaySeconds = 5
                            periodSeconds = 5
                            timeoutSeconds = 3
                            successThreshold = 1
                            failureThreshold = 3
                        }
                        resources = corev1.ResourceRequirements {
                            requests = {
                                cpu = "100m"
                                memory = "128Mi"
                            }
                            limits = {
                                cpu = "500m"
                                memory = "512Mi"
                            }
                        }
                        securityContext = corev1.SecurityContext {
                            allowPrivilegeEscalation = False
                            readOnlyRootFilesystem = True
                            runAsNonRoot = True
                            runAsUser = 65532
                            capabilities = corev1.Capabilities {
                                drop = ["ALL"]
                            }
                        }
                        volumeMounts = [
                            corev1.VolumeMount {
                                name = "tmp"
                                mountPath = "/tmp"
                            }
                        ]
                    }
                ]
                volumes = [
                    corev1.Volume {
                        name = "tmp"
                        emptyDir = corev1.EmptyDirVolumeSource {}
                    }
                ]
                affinity = corev1.Affinity {
                    podAntiAffinity = corev1.PodAntiAffinity {
                        preferredDuringSchedulingIgnoredDuringExecution = [
                            corev1.WeightedPodAffinityTerm {
                                weight = 100
                                podAffinityTerm = corev1.PodAffinityTerm {
                                    labelSelector = metav1.LabelSelector {
                                        matchLabels = {
                                            app = appName
                                        }
                                    }
                                    topologyKey = "kubernetes.io/hostname"
                                }
                            }
                        ]
                    }
                }
            }
        }
    }
}

# Service
service = corev1.Service {
    metadata = metav1.ObjectMeta {
        name = appName
        namespace = namespace
        labels = commonLabels
    }
    spec = corev1.ServiceSpec {
        type = "ClusterIP"
        selector = {
            app = appName
        }
        ports = [
            corev1.ServicePort {
                port = 80
                targetPort = "http"
                protocol = "TCP"
                name = "http"
            }
        ]
    }
}

# Ingress
ingress = networkingv1.Ingress {
    metadata = metav1.ObjectMeta {
        name = appName
        namespace = namespace
        labels = commonLabels
        annotations = {
            "nginx.ingress.kubernetes.io/ssl-redirect" = "true"
            "nginx.ingress.kubernetes.io/force-ssl-redirect" = "true"
            "nginx.ingress.kubernetes.io/proxy-body-size" = "10m"
            "nginx.ingress.kubernetes.io/proxy-connect-timeout" = "60"
            "nginx.ingress.kubernetes.io/proxy-send-timeout" = "60"
            "nginx.ingress.kubernetes.io/proxy-read-timeout" = "60"
            "nginx.ingress.kubernetes.io/limit-rps" = "100"
            "cert-manager.io/cluster-issuer" = clusterIssuer
            "cert-manager.io/common-name" = fqdn
        }
    }
    spec = networkingv1.IngressSpec {
        ingressClassName = "nginx"
        tls = [
            networkingv1.IngressTLS {
                hosts = [fqdn]
                secretName = "${appName}-tls"
            }
        ]
        rules = [
            networkingv1.IngressRule {
                host = fqdn
                http = networkingv1.HTTPIngressRuleValue {
                    paths = [
                        networkingv1.HTTPIngressPath {
                            path = "/"
                            pathType = "Prefix"
                            backend = networkingv1.IngressBackend {
                                service = networkingv1.IngressServiceBackend {
                                    name = appName
                                    port = networkingv1.ServiceBackendPort {
                                        name = "http"
                                    }
                                }
                            }
                        }
                    ]
                }
            }
        ]
    }
}

# Export all resources
items = [
    serviceAccount
    configMap
    secretToken
    secretRedis
    deployment
    service
    ingress
]
