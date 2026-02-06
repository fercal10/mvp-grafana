#!/bin/bash

set -e

echo "🚀 Deploying Bank API microservices to Kubernetes..."

# Build Docker images for both microservices
echo "📦 Building Docker images..."
docker build -f Dockerfile.accounts -t accounts-api:latest .
docker build -f Dockerfile.transfers -t transfers-api:latest .

# Cargar imágenes en el cluster (kind/k3d/minikube no comparten el daemon Docker del host)
CONTEXT=$(kubectl config current-context 2>/dev/null || true)
if [[ "$CONTEXT" == *"kind"* ]]; then
  echo "📥 Loading images into kind cluster..."
  kind load docker-image accounts-api:latest
  kind load docker-image transfers-api:latest
elif [[ "$CONTEXT" == *"k3d"* ]]; then
  echo "📥 Loading images into k3d cluster..."
  K3D_CLUSTER="${CONTEXT#k3d-}"
  k3d image import accounts-api:latest -c "$K3D_CLUSTER"
  k3d image import transfers-api:latest -c "$K3D_CLUSTER"
elif command -v minikube &>/dev/null && [[ "$CONTEXT" == *"minikube"* ]]; then
  echo "📥 Loading images into minikube..."
  minikube image load accounts-api:latest
  minikube image load transfers-api:latest
fi

# Apply Kubernetes manifests
echo "☸️  Applying Kubernetes manifests..."

kubectl apply -f k8s/namespace.yaml

echo "Deploying Loki..."
kubectl apply -f k8s/loki/

echo "Deploying Tempo..."
kubectl apply -f k8s/tempo/

echo "Deploying Prometheus..."
kubectl apply -f k8s/prometheus/

echo "Deploying Promtail..."
kubectl apply -f k8s/promtail/

echo "⏳ Waiting for Loki and Prometheus to be ready..."
kubectl wait --for=condition=ready pod -l app=loki -n banking-system --timeout=120s
kubectl wait --for=condition=ready pod -l app=prometheus -n banking-system --timeout=120s

echo "Deploying Grafana..."
# Solo aplicar manifests YAML (dashboard.json es JSON de Grafana, no recurso K8s)
for f in k8s/grafana/*.yaml; do
  kubectl apply -f "$f"
done

echo "Deploying Bank API microservices..."
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/configmap-transfers.yaml
kubectl apply -f k8s/pvc-data.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/deployment-transfers.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/service-transfers.yaml

echo "⏳ Waiting for pods to be ready..."
kubectl wait --for=condition=ready pod -l app=accounts-api -n banking-system --timeout=120s
kubectl wait --for=condition=ready pod -l app=transfers-api -n banking-system --timeout=120s

# Exponer puertos en localhost (port-forward; en k3d/kind NodePort no siempre llega al host)
echo "🔌 Exponiendo puertos en localhost..."
pkill -f "port-forward.*banking-system.*30080" 2>/dev/null || true
pkill -f "port-forward.*banking-system.*30081" 2>/dev/null || true
pkill -f "port-forward.*banking-system.*30300" 2>/dev/null || true
nohup kubectl port-forward -n banking-system svc/accounts-api 30080:8080 >/dev/null 2>&1 &
nohup kubectl port-forward -n banking-system svc/transfers-api 30081:8081 >/dev/null 2>&1 &
nohup kubectl port-forward -n banking-system svc/grafana 30300:3000 >/dev/null 2>&1 &
sleep 1

echo "✅ Deployment complete!"
echo ""
echo "📊 Service endpoints (port-forward en localhost):"
echo "  • Accounts API: http://localhost:30080  (puerto 30080)"
echo "  • Transfers API: http://localhost:30081  (puerto 30081)"
echo "  • Grafana:       http://localhost:30300  (puerto 30300, usuario: admin / admin)"
echo ""
echo "🔌 Port-forwards activos. Para detenerlos:"
echo "  pkill -f 'port-forward.*banking-system'"
echo ""
echo "🔍 Check status with:"
echo "  kubectl get pods -n banking-system"
echo "  kubectl get services -n banking-system"
