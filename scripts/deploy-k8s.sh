#!/bin/bash

set -e

echo "🚀 Deploying Bank API to Kubernetes..."

# Build Docker image
echo "📦 Building Docker image..."
docker build -t bank-api:latest .

# Cargar imagen en el cluster (kind/k3d/minikube no comparten el daemon Docker del host)
CONTEXT=$(kubectl config current-context 2>/dev/null || true)
if [[ "$CONTEXT" == *"kind"* ]]; then
  echo "📥 Loading image into kind cluster..."
  kind load docker-image bank-api:latest
elif [[ "$CONTEXT" == *"k3d"* ]]; then
  echo "📥 Loading image into k3d cluster..."
  K3D_CLUSTER="${CONTEXT#k3d-}"
  k3d image import bank-api:latest -c "$K3D_CLUSTER"
elif command -v minikube &>/dev/null && [[ "$CONTEXT" == *"minikube"* ]]; then
  echo "📥 Loading image into minikube..."
  minikube image load bank-api:latest
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

echo "Deploying Bank API..."
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml

echo "⏳ Waiting for pods to be ready..."
kubectl wait --for=condition=ready pod -l app=bank-api -n banking-system --timeout=120s

# Exponer puertos en localhost (port-forward; en k3d/kind NodePort no siempre llega al host)
echo "🔌 Exponiendo puertos en localhost..."
pkill -f "port-forward.*banking-system.*30080" 2>/dev/null || true
pkill -f "port-forward.*banking-system.*30300" 2>/dev/null || true
nohup kubectl port-forward -n banking-system svc/bank-api 30080:8080 >/dev/null 2>&1 &
nohup kubectl port-forward -n banking-system svc/grafana 30300:3000 >/dev/null 2>&1 &
sleep 1

echo "✅ Deployment complete!"
echo ""
echo "📊 Service endpoints (port-forward en localhost):"
echo "  • Bank API:  http://localhost:30080  (puerto 30080)"
echo "  • Grafana:   http://localhost:30300  (puerto 30300, usuario: admin / admin)"
echo ""
echo "🔌 Port-forwards activos. Para detenerlos:"
echo "  pkill -f 'port-forward.*banking-system'"
echo ""
echo "🔍 Check status with:"
echo "  kubectl get pods -n banking-system"
echo "  kubectl get services -n banking-system"
