# Quick Start Guide

## Opción 1: Docker Compose (Más Rápido)

### 1. Iniciar todos los servicios

```bash
docker-compose up -d
```

### 2. Esperar a que los servicios estén listos (30-60 segundos)

```bash
docker-compose ps
```

### 3. Probar la API

```bash
# Crear cuenta
curl -X POST http://localhost:8080/api/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_number": "ACC001", "initial_balance": 1000.0}'

# Crear segunda cuenta
curl -X POST http://localhost:8080/api/accounts \
  -H "Content-Type: application/json" \
  -d '{"account_number": "ACC002", "initial_balance": 500.0}'

# Hacer transferencia
curl -X POST http://localhost:8080/api/transfers \
  -H "Content-Type: application/json" \
  -d '{"from_account_number": "ACC001", "to_account_number": "ACC002", "amount": 100.0, "description": "Test"}'

# Ver cuentas
curl http://localhost:8080/api/accounts
```

### 4. Ver en Grafana

Abrir http://localhost:3000
- Usuario: `admin`
- Contraseña: `admin`

En el menú lateral izquierdo:
1. Ir a "Dashboards"
2. Buscar "Bank API Monitoring"
3. Ver logs, trazas y métricas en tiempo real

### 5. Detener

```bash
docker-compose down
```

---

## Opción 2: Kubernetes

### 1. Construir imágenes de ambos microservicios

```bash
docker build -f Dockerfile.accounts -t accounts-api:latest .
docker build -f Dockerfile.transfers -t transfers-api:latest .
```

O usar el script de despliegue (construye y despliega):

```bash
./scripts/deploy-k8s.sh
```

### 2. Cargar imágenes al cluster (si usas kind)

```bash
kind load docker-image accounts-api:latest
kind load docker-image transfers-api:latest
```

### 3. Desplegar

```bash
make k8s-deploy
```

O manualmente:

```bash
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/loki/
kubectl apply -f k8s/tempo/
kubectl apply -f k8s/promtail/
kubectl apply -f k8s/grafana/
kubectl apply -f k8s/
```

### 4. Verificar

```bash
kubectl get pods -n banking-system
```

### 5. Acceder a servicios

```bash
# API (NodePort)
curl http://localhost:30080/health

# Grafana (NodePort)
open http://localhost:30300
```

### 6. Probar API

```bash
./scripts/test-api.sh http://localhost:30080
```

### 7. Limpiar

```bash
kubectl delete namespace banking-system
```

---

## Makefile Commands

```bash
# Ver todos los comandos disponibles
make help

# Compilar localmente
make build

# Ejecutar localmente
make run

# Docker Compose
make compose-up      # Iniciar servicios
make compose-down    # Detener servicios
make compose-logs    # Ver logs

# Kubernetes
make k8s-deploy      # Desplegar a K8s
make k8s-status      # Ver estado
make k8s-logs        # Ver logs
make k8s-delete      # Eliminar deployment

# Testing
make test-api        # Probar API local/compose
make test-api-k8s    # Probar API en K8s

# Build
make docker-build    # Construir imagen Docker
make clean           # Limpiar build artifacts
```

---

## Endpoints de la API

| Método | Endpoint | Descripción |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/ready` | Readiness check |
| GET | `/api/accounts` | Listar cuentas |
| GET | `/api/accounts/:id` | Obtener cuenta |
| POST | `/api/accounts` | Crear cuenta |
| GET | `/api/accounts/:id/transactions` | Transacciones de cuenta |
| POST | `/api/transfers` | Crear transferencia |
| GET | `/api/transfers/:id` | Obtener transferencia |

---

## URLs de Servicios

### Docker Compose
- **API**: http://localhost:8080
- **Grafana**: http://localhost:3000 (admin/admin)
- **Loki**: http://localhost:3100
- **Tempo**: http://localhost:3200

### Kubernetes (NodePort)
- **API**: http://localhost:30080
- **Grafana**: http://localhost:30300 (admin/admin)

---

## Troubleshooting

### Docker Compose

**Problema**: Servicios no inician
```bash
docker-compose logs
docker-compose down -v
docker-compose up -d
```

**Problema**: Puerto ocupado
```bash
# Cambiar puertos en docker-compose.yml
```

### Kubernetes

**Problema**: Pods no inician
```bash
kubectl describe pod -n banking-system <pod-name>
kubectl logs -n banking-system <pod-name>
```

**Problema**: Imagen no encontrada
```bash
# Verificar que las imágenes estén cargadas
docker images | grep -E "accounts-api|transfers-api"

# Recargar imágenes
kind load docker-image accounts-api:latest
kind load docker-image transfers-api:latest
# o
minikube image load accounts-api:latest
minikube image load transfers-api:latest
```

---

## Próximos Pasos

1. **Explorar Grafana**: Ver dashboard con métricas y logs
2. **Probar Trazas**: Hacer varias transferencias y ver trazas en Tempo
3. **Ver Logs**: Usar Loki para buscar logs específicos
4. **Personalizar**: Modificar dashboard, agregar más métricas

Para más detalles, ver [README.md](README.md)
