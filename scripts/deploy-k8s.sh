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

# Helm repos for LGTM stack (Grafana, Loki, Prometheus, Tempo)
echo "📦 Adding Helm repos..."
helm repo add grafana https://grafana.github.io/helm-charts 
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts 
helm repo update

echo "Deploying Loki (Helm)..."
helm upgrade --install loki grafana/loki -n banking-system -f k8s/helm/loki-values.yaml 

echo "Deploying Tempo (Helm)..."
helm upgrade --install tempo grafana/tempo -n banking-system -f k8s/helm/tempo-values.yaml 

echo "Deploying Prometheus (Helm)..."
helm upgrade --install prometheus prometheus-community/prometheus -n banking-system -f k8s/helm/prometheus-values.yaml 

echo "Deploying Grafana (Helm)..."
helm upgrade --install grafana grafana/grafana -n banking-system -f k8s/helm/grafana-values.yaml  

echo "Deploying Bank API microservices..."
kubectl apply -f k8s/accounts-api/
kubectl apply -f k8s/transfers-api/

echo "⏳ Waiting for pods to be ready..."
kubectl wait --for=condition=ready pod -l app=accounts-api -n banking-system --timeout=120s
kubectl wait --for=condition=ready pod -l app=transfers-api -n banking-system --timeout=120s

# Exponer puertos en localhost (port-forward; en k3d/kind NodePort no siempre llega al host)
echo "🔌 Exponiendo puertos en localhost..."
pkill -f "port-forward.*banking-system.*30080"
pkill -f "port-forward.*banking-system.*30081"
pkill -f "port-forward.*banking-system.*30300"
nohup kubectl port-forward -n banking-system svc/accounts-api 30080:8080 
nohup kubectl port-forward -n banking-system svc/transfers-api 30081:8081 
nohup kubectl port-forward -n banking-system svc/grafana 30300:80 
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
