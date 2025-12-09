# Kubernetes Deployment

This directory contains Kubernetes manifests for deploying the URL Status Checker to a local kind cluster.

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) (Kubernetes in Docker)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)

### Install kind

**macOS:**
```bash
brew install kind
```

**Linux:**
```bash
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind
```

### Install kubectl

**macOS:**
```bash
brew install kubectl
```

**Linux:**
```bash
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/
```

## Quick Start

### 1. Build the Docker Image

```bash
# From project root
task docker
```

### 2. Create kind Cluster

```bash
kind create cluster --config deployments/kubernetes/kind-config.yaml
```

This creates a cluster named `urlchecker-demo` with port mappings for:
- 8080 → URL Checker API
- 9090 → Prometheus UI

### 3. Load Docker Image into kind

```bash
kind load docker-image urlchecker:latest --name urlchecker-demo
```

### 4. Deploy to Kubernetes

```bash
kubectl apply -f deployments/kubernetes/
```

### 5. Verify Deployment

```bash
# Check pods are running
kubectl get pods -n url-checker

# Check services
kubectl get svc -n url-checker

# View logs
kubectl logs -n url-checker -l app=urlchecker -f
```

### 6. Access the Application

- **URL Checker API**: http://localhost:8080
- **Prometheus**: http://localhost:9090

Test the API:
```bash
curl -X POST http://localhost:8080/api/v1/check \
  -H "Content-Type: application/json" \
  -d '{
    "urls": ["https://google.com", "https://github.com"],
    "timeout": "5s",
    "max_workers": 10
  }'
```

View metrics:
```bash
curl http://localhost:8080/metrics
```

## Management Commands

### View Resources

```bash
# All resources in namespace
kubectl get all -n url-checker

# Deployments
kubectl get deployments -n url-checker

# Pods with wide output
kubectl get pods -n url-checker -o wide
```

### Logs

```bash
# Follow logs from all urlchecker pods
kubectl logs -n url-checker -l app=urlchecker -f

# Logs from specific pod
kubectl logs -n url-checker <pod-name>

# Previous pod logs (if crashed)
kubectl logs -n url-checker <pod-name> --previous
```

### Scaling

```bash
# Scale to 5 replicas
kubectl scale deployment urlchecker -n url-checker --replicas=5

# Verify scaling
kubectl get pods -n url-checker -w
```

### Update Application

```bash
# Rebuild image
task docker

# Reload into kind
kind load docker-image urlchecker:latest --name urlchecker-demo

# Restart deployment to use new image
kubectl rollout restart deployment/urlchecker -n url-checker

# Watch rollout status
kubectl rollout status deployment/urlchecker -n url-checker
```

### Port Forwarding (Alternative Access)

If you didn't use the kind-config with port mappings:

```bash
# Forward URL Checker
kubectl port-forward -n url-checker svc/urlchecker 8080:8080

# Forward Prometheus
kubectl port-forward -n url-checker svc/prometheus 9090:9090
```

## Troubleshooting

### Pods not starting

```bash
# Describe pod for events
kubectl describe pod -n url-checker <pod-name>

# Check if image is loaded
docker exec -it urlchecker-demo-control-plane crictl images | grep urlchecker
```

### Image pull errors

Make sure you loaded the image into kind:
```bash
kind load docker-image urlchecker:latest --name urlchecker-demo
```

### Health check failures

The deployment expects `/health` endpoint. If it doesn't exist, remove the probes:
```bash
kubectl edit deployment urlchecker -n url-checker
# Remove livenessProbe and readinessProbe sections
```

## Cleanup

### Delete Deployment

```bash
kubectl delete -f deployments/kubernetes/
```

### Delete Cluster

```bash
kind delete cluster --name urlchecker-demo
```

## Architecture

The deployment includes:

- **Namespace**: `url-checker` - Isolates resources
- **Deployment**: 2 replicas of urlchecker for HA
- **Service**: NodePort service exposing port 8080 → 30080
- **Prometheus**: Metrics collection with pod autodiscovery
- **ConfigMap**: Prometheus configuration

## Production Considerations

For production deployments, consider:

1. **Ingress Controller**: Use nginx-ingress or Traefik instead of NodePort
2. **TLS/SSL**: Add cert-manager for automatic certificate management
3. **Persistent Storage**: Use PersistentVolumes for Prometheus data
4. **Resource Limits**: Adjust CPU/memory based on load testing
5. **HPA**: Add HorizontalPodAutoscaler for auto-scaling
6. **Network Policies**: Restrict pod-to-pod communication
7. **RBAC**: Define proper service accounts and roles
8. **Monitoring**: Add Grafana with dashboards
9. **Secrets**: Use Kubernetes Secrets or external secret managers
10. **Multi-zone**: Deploy across multiple availability zones
