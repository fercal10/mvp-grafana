# Bank API - Banking System with OpenTelemetry, Grafana & Full Observability Stack

Sistema de API bancaria simple construido con Go y Gin, integrado con OpenTelemetry para observabilidad y un stack completo de Grafana para visualización de métricas, logs y trazas.

## Características

- **API REST** con endpoints para gestión de cuentas y transferencias
- **OpenTelemetry** integrado para trazas distribuidas
- **Prometheus** para recolección y almacenamiento de métricas
- **Grafana** para visualización de datos
- **Loki** para agregación y consulta de logs
- **Tempo** para almacenamiento y consulta de trazas
- **Promtail** para recolección de logs
- **SQLite** como base de datos (persistente)
- Sin autenticación ni validaciones (API simple para demostración)

## Arquitectura

```
┌─────────────┐
│   Cliente   │
└──────┬──────┘
       │
       v
┌─────────────────────────────────────────────┐
│         Bank API (Gin + Go)                 │
│  ┌──────────────────────────────────────┐   │
│  │   OpenTelemetry + Prometheus         │   │
│  │   Middleware                         │   │
│  └──────────────────────────────────────┘   │
│    │ Traces │ Logs │ Metrics (/metrics)     │
└────┼────────┼───────┼────────────────────────┘
     │        │       │
     v        v       v
┌─────────┐ ┌──────────┐ ┌────────────┐
│  Tempo  │ │ Promtail │ │ Prometheus │
└────┬────┘ └────┬─────┘ └─────┬──────┘
     │           │              │
     v           v              v
┌────────────────────────────────────┐
│            Grafana                 │
│  (Visualización de Logs, Traces    │
│   y Métricas)                      │
└────────────────┬───────────────────┘
                 │
                 v
            ┌─────────┐
            │  Loki   │
            └─────────┘
```

## Endpoints de la API

### Cuentas

- `GET /api/accounts` - Listar todas las cuentas
- `GET /api/accounts/:id` - Obtener cuenta por ID
- `POST /api/accounts` - Crear nueva cuenta
- `GET /api/accounts/:id/transactions` - Listar transacciones de una cuenta

### Transferencias

- `POST /api/transfers` - Realizar transferencia entre cuentas
- `GET /api/transfers/:id` - Obtener información de una transferencia

### Health Check

- `GET /health` - Estado de salud de la API
- `GET /ready` - Estado de preparación de la API

### Métricas

- `GET /metrics` - Endpoint de métricas Prometheus

## Estructura del Proyecto

La API se despliega como dos microservicios (accounts-api y transfers-api) que comparten la misma base de datos.

```
.
├── cmd/
│   ├── accounts-api/
│   │   └── main.go                 # Microservicio de cuentas (puerto 8080)
│   └── transfers-api/
│       └── main.go                 # Microservicio de transferencias (puerto 8081)
├── internal/
│   ├── handlers/                   # Handlers HTTP
│   ├── models/                     # Modelos de datos
│   ├── repository/                 # Acceso a datos (SQLite)
│   └── service/                    # Lógica de negocio
├── pkg/
│   └── telemetry/                  # Configuración OpenTelemetry
├── k8s/                            # Manifiestos Kubernetes
│   ├── grafana/
│   ├── loki/
│   ├── tempo/
│   ├── promtail/
│   └── prometheus/
├── config/                         # Configuración para docker-compose
├── migrations/                     # Migraciones SQL
├── Dockerfile
├── docker-compose.yml
└── README.md
```

## Despliegue

### Opción 1: Kubernetes (Producción)

#### Prerrequisitos

- Kubernetes cluster local (minikube, kind, Docker Desktop)
- kubectl configurado
- Docker para construir la imagen

#### Pasos

1. **Construir las imágenes Docker de ambos microservicios:**

```bash
docker build -f Dockerfile.accounts -t accounts-api:latest .
docker build -f Dockerfile.transfers -t transfers-api:latest .
```

O usar el script de despliegue (construye y despliega):

```bash
./scripts/deploy-k8s.sh
```

2. **Cargar las imágenes en el cluster (si usas kind o minikube):**

```bash
# Para kind
kind load docker-image accounts-api:latest
kind load docker-image transfers-api:latest

# Para minikube
minikube image load accounts-api:latest
minikube image load transfers-api:latest
```

3. **Desplegar todos los componentes:**

```bash
# Aplicar todos los manifiestos
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/

# Aplicar componentes de observabilidad
kubectl apply -f k8s/loki/
kubectl apply -f k8s/tempo/
kubectl apply -f k8s/promtail/
kubectl apply -f k8s/prometheus/
kubectl apply -f k8s/grafana/

# Aplicar la API (configmaps, PVC compartido, deployments y services de ambos microservicios)
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/configmap-transfers.yaml
kubectl apply -f k8s/pvc-data.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/deployment-transfers.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/service-transfers.yaml
```

**Nota:** El PVC compartido (`pvc-data.yaml`) usa `ReadWriteMany` para que ambos pods monten la misma DB. Si tu cluster no tiene un storage class con RWM (p. ej. kind/minikube sin NFS), el PVC puede quedar en Pending. En ese caso, para desarrollo local puedes cambiar en `deployment.yaml` y `deployment-transfers.yaml` el volumen a `emptyDir: {}` en lugar de `persistentVolumeClaim`; entonces cada pod tendrá su propia DB y los datos no se comparten entre servicios.

4. **Verificar el despliegue:**

```bash
kubectl get pods -n banking-system
kubectl get services -n banking-system
```

5. **Acceder a los servicios:**

```bash
# Accounts API (NodePort 30080) y Transfers API (NodePort 30081)
curl http://localhost:30080/health
curl http://localhost:30081/health

# Grafana (NodePort 30300)
# Usuario: admin / Contraseña: admin
open http://localhost:30300

# Prometheus (NodePort 30900)
open http://localhost:30900
```

#### Port Forwarding (alternativa)

Si prefieres usar port-forward en lugar de NodePort:

```bash
# API
kubectl port-forward -n banking-system svc/accounts-api 8080:8080
kubectl port-forward -n banking-system svc/transfers-api 8081:8081

# Grafana
kubectl port-forward -n banking-system svc/grafana 3000:3000

# Loki
kubectl port-forward -n banking-system svc/loki 3100:3100

# Tempo
kubectl port-forward -n banking-system svc/tempo 3200:3200

# Prometheus
kubectl port-forward -n banking-system svc/prometheus 9090:9090
```

### Opción 2: Docker Compose (Desarrollo Local)

#### Prerrequisitos

- Docker
- Docker Compose
- Go 1.24+

#### Pasos

1. **Iniciar servicios de infraestructura (Loki, Tempo, Prometheus, Grafana):**

   ```bash
   # Iniciar todo excepto la API (si quieres correr la API localmente)
   docker-compose up -d loki tempo prometheus grafana promtail
   ```

2. **Ejecutar la API localmente:**

   La API está configurada para conectarse automáticamente a los servicios locales en `localhost`.

   ```bash
   # Cargar variables de entorno recomendadas (opcional)
   # export $(cat .env.example | xargs)

   # Ejecutar ambos microservicios (en dos terminales)
   PORT=8080 go run ./cmd/accounts-api
   PORT=8081 go run ./cmd/transfers-api
   ```

   La API se iniciará en `http://localhost:8080` y enviará:
   - Trazas a Tempo en `localhost:4318`
   - Logs a Loki en `http://localhost:3100`

3. **Verificar que los contenedores estén corriendo:**

   ```bash
   docker-compose ps
   ```

4. **Acceder a los servicios:**
   - API: http://localhost:8080
   - Grafana: http://localhost:3000 (admin/admin)
   - Loki: http://localhost:3100
   - Tempo: http://localhost:3200
   - Prometheus: http://localhost:9090

5. **Detener los servicios:**

   ```bash
   docker-compose down
   ```

## Uso de la API

### Crear una cuenta

```bash
curl -X POST http://localhost:8080/api/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "account_number": "ACC001",
    "initial_balance": 1000.0
  }'
```

### Listar cuentas

```bash
curl http://localhost:8080/api/accounts
```

### Obtener cuenta específica

```bash
curl http://localhost:8080/api/accounts/1
```

### Realizar transferencia

```bash
curl -X POST http://localhost:8080/api/transfers \
  -H "Content-Type: application/json" \
  -d '{
    "from_account_number": "ACC001",
    "to_account_number": "ACC002",
    "amount": 100.0,
    "description": "Pago de servicio"
  }'
```

### Ver transacciones de una cuenta

```bash
curl http://localhost:8080/api/accounts/1/transactions
```

## Grafana

### Acceso

- URL: http://localhost:3000 (docker-compose) o http://localhost:30300 (k8s)
- Usuario: `admin`
- Contraseña: `admin`

### Datasources Pre-configurados

1. **Loki** - Para consultar logs
2. **Tempo** - Para consultar trazas distribuidas
3. **Prometheus** - Para consultar métricas

### Dashboard Pre-instalado

El dashboard "Bank API Monitoring" incluye:

- **Request Rate**: Tasa de requests por endpoint
- **Recent Logs**: Logs recientes de la aplicación
- **HTTP Status Codes**: Distribución de códigos de estado
- **Transfer Operations**: Logs de operaciones de transferencia
- **Error Logs**: Filtro de logs con errores

## Observabilidad

La aplicación está completamente instrumentada con un stack de observabilidad moderno:

### OpenTelemetry
- **Trazas (Traces)**: A Tempo vía OTLP HTTP (puerto 4318)
- **Logs**: A Loki vía Promtail
- **Contexto**: Las trazas incluyen información de span para cada request HTTP

### Prometheus
La aplicación expone métricas Prometheus en el endpoint `/metrics`:

#### Métricas HTTP
- `http_requests_total` - Total de requests HTTP por método, endpoint y status
- `http_request_duration_seconds` - Duración de requests HTTP
- `http_request_size_bytes` - Tamaño de requests HTTP
- `http_response_size_bytes` - Tamaño de responses HTTP

#### Métricas de Negocio
- `bank_accounts_total` - Total de cuentas bancarias creadas
- `bank_transfers_total` - Total de transferencias procesadas (por status: success/failed)
- `bank_transfer_amount_total` - Monto total transferido
- `bank_account_balance` - Balance actual de cuentas bancarias

### Variables de Entorno

- `OTLP_ENDPOINT`: Endpoint del exportador OTLP (default: `tempo:4318`)
- `DB_PATH`: Ruta a la base de datos SQLite (default: `./data/bank.db`)
- `PORT`: Puerto de la API (default: `8080`)
- `GIN_MODE`: Modo de Gin (`debug`, `release`) (default: `release`)

## Desarrollo

### Compilar localmente

```bash
make build
# o por separado:
make build-accounts build-transfers
```

### Ejecutar localmente

Ejecuta ambos microservicios en dos terminales (comparten la misma DB_PATH):

```bash
export DB_PATH=./data/bank.db
export OTLP_ENDPOINT=localhost:4318

# Terminal 1: accounts-api
PORT=8080 go run ./cmd/accounts-api

# Terminal 2: transfers-api
PORT=8081 go run ./cmd/transfers-api
```

### Ejecutar tests

```bash
go test ./...
```

## Limpieza

### Kubernetes

```bash
kubectl delete namespace banking-system
```

### Docker Compose

```bash
docker-compose down -v  # -v elimina también los volúmenes
```

## Troubleshooting

### La API no puede conectarse a Tempo

Verifica que Tempo esté corriendo:

```bash
# Kubernetes
kubectl get pods -n banking-system | grep tempo

# Docker Compose
docker-compose ps tempo
```

### No se ven logs en Loki

Verifica que Promtail esté corriendo y pueda acceder a Loki:

```bash
# Kubernetes
kubectl logs -n banking-system -l app=promtail

# Docker Compose
docker-compose logs promtail
```

### Error de permisos en SQLite

Asegúrate de que el directorio de datos exista y tenga permisos:

```bash
mkdir -p ./data
chmod 755 ./data
```

## Características Técnicas

- **Go 1.22**
- **Gin** para el framework HTTP
- **GORM** para ORM
- **SQLite** para persistencia
- **OpenTelemetry** para trazas distribuidas
- **Prometheus** para métricas
- **Grafana** para visualización
- **Loki** para logs
- **Tempo** para trazas
- **Promtail** para recolección de logs

## Licencia

MIT

## Autor

Tribal - Banking System Demo
