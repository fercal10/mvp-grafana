# Arquitectura del Sistema

## Visión General

El sistema consiste en una API bancaria simple desarrollada en Go con Gin, completamente instrumentada con OpenTelemetry y Prometheus, con un stack completo de observabilidad basado en Grafana para visualización de métricas, logs y trazas.

## Componentes

### 1. Bank API (Aplicación Principal)

**Tecnologías:**
- Go 1.22
- Gin (framework HTTP)
- GORM (ORM)
- SQLite (base de datos)
- OpenTelemetry SDK

**Responsabilidades:**
- Gestionar cuentas bancarias
- Procesar transferencias
- Registrar transacciones
- Emitir trazas, logs y métricas instrumentados

**Estructura de Código:**
```
cmd/server/main.go           # Punto de entrada, configuración de servidor
internal/
  handlers/                  # Handlers HTTP (capa de presentación)
  models/                    # Modelos de dominio
  repository/                # Acceso a datos (SQLite)
  service/                   # Lógica de negocio
pkg/telemetry/               # Configuración OpenTelemetry
```

### 2. Prometheus (Métricas)

**Función:**
- Recolectar métricas de la Bank API vía scraping del endpoint `/metrics`
- Almacenar métricas con series temporales
- Proveer API de consulta para Grafana

**Configuración:**
- Retención: 7 días
- Scrape interval: 15 segundos
- Storage: Filesystem local

**Métricas Recolectadas:**
- Métricas HTTP: requests totales, duración, tamaño
- Métricas de negocio: cuentas creadas, transferencias, balances

### 3. OpenTelemetry (Trazas)

**Componentes:**
- **TracerProvider**: Gestión de trazas distribuidas
- **OTLP Exporter**: Exportador HTTP hacia Tempo
- **Gin Middleware**: Instrumentación automática de requests HTTP

**Datos Exportados:**
- **Traces**: Cada request HTTP genera un trace con spans
- **Spans**: Operaciones de servicio (CreateAccount, CreateTransfer, etc.)
- **Attributes**: Metadata de operaciones (account_id, amount, etc.)

### 4. Tempo (Almacenamiento de Trazas)

**Función:**
- Recibir trazas vía OTLP (puerto 4318 HTTP, 4317 gRPC)
- Almacenar trazas en formato local
- Proveer API de consulta para Grafana

**Endpoints:**
- 3200: API HTTP para consultas
- 4317: Receptor OTLP gRPC
- 4318: Receptor OTLP HTTP

### 5. Loki (Almacenamiento de Logs)

**Función:**
- Recibir logs de Promtail
- Indexar y almacenar logs
- Proveer API de consulta (LogQL)

**Configuración:**
- Retención: 7 días (168h)
- Storage: Filesystem local
- Schema: v11 con boltdb-shipper

### 6. Promtail (Recolector de Logs)

**Función:**
- Scraping de logs de contenedores Docker
- Scraping de logs de pods Kubernetes
- Envío de logs a Loki con labels

**Labels Agregados:**
- `namespace`: Namespace de Kubernetes
- `app`: Nombre de la aplicación
- `pod`: Nombre del pod
- `container`: Nombre del contenedor

### 7. Grafana (Visualización)

**Datasources:**
- **Prometheus**: Para consultar métricas
- **Loki**: Para consultar logs
- **Tempo**: Para consultar trazas

**Dashboard Pre-configurado:**
- Request rate por endpoint
- Status codes distribution
- Recent logs panel
- Transfer operations logs
- Error logs filtered

## Flujo de Datos

### Request HTTP Normal

```
1. Cliente → Bank API (POST /api/transfers)
2. Gin Middleware → OpenTelemetry crea trace + Prometheus registra métricas
3. Handler → Service → Repository
4. Cada capa agrega spans al trace y actualiza métricas
5. OpenTelemetry exporta trace a Tempo
6. Gin escribe logs → stdout → Promtail → Loki
7. Prometheus scrape endpoint /metrics cada 15s
8. Respuesta al cliente
```

### Consulta en Grafana

```
1. Usuario abre Grafana Dashboard
2. Grafana consulta Prometheus (métricas últimos 5 minutos)
3. Grafana consulta Loki (logs últimos 5 minutos)
4. Grafana consulta Tempo (trazas correlacionadas)
5. Dashboard muestra:
   - Métricas en tiempo real (request rate, latencia, errores)
   - Logs en tiempo real
   - Gráficas de métricas de negocio
   - Trazas individuales clickeables
```

## Diagrama de Arquitectura Detallado

```
┌─────────────────────────────────────────────────────────────┐
│                         Cliente                              │
│                    (curl, Postman, etc)                      │
└────────────────────────┬────────────────────────────────────┘
                         │ HTTP REST
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                       Bank API                               │
│  ┌───────────────────────────────────────────────────────┐  │
│  │           Gin Router + Middleware                      │  │
│  │  ┌────────────────────────────────────────────────┐   │  │
│  │  │  Prometheus Middleware                         │   │  │
│  │  │  - Registra métricas HTTP                      │   │  │
│  │  └────────────────────────────────────────────────┘   │  │
│  │  ┌────────────────────────────────────────────────┐   │  │
│  │  │  OpenTelemetry Gin Middleware (otelgin)        │   │  │
│  │  │  - Crea trace por request                      │   │  │
│  │  │  - Inyecta contexto en handlers                │   │  │
│  │  └────────────────────────────────────────────────┘   │  │
│  └───────────────────────────────────────────────────────┘  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │             Endpoint /metrics                         │  │
│  │  - Expone métricas Prometheus                         │  │
│  └───────────────────────────────────────────────────────┘  │
│  ┌───────────────────────────────────────────────────────┐  │
│  │                    Handlers                           │  │
│  │  - accounts.go (GET, POST accounts)                   │  │
│  │  - transfers.go (POST transfers)                      │  │
│  │  - transactions.go (health checks)                    │  │
│  └────────────────────┬──────────────────────────────────┘  │
│                       │                                      │
│  ┌────────────────────▼──────────────────────────────────┐  │
│  │                   Services                            │  │
│  │  - account_service.go                                 │  │
│  │  - transfer_service.go                                │  │
│  │  (Lógica de negocio + Spans + Métricas)              │  │
│  └────────────────────┬──────────────────────────────────┘  │
│                       │                                      │
│  ┌────────────────────▼──────────────────────────────────┐  │
│  │                 Repository                            │  │
│  │  - sqlite.go (GORM + SQLite)                          │  │
│  └───────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌───────────────────────────────────────────────────────┐  │
│  │         OpenTelemetry SDK                             │  │
│  │  - TracerProvider                                     │  │
│  │  - OTLP Exporter (HTTP)                               │  │
│  └────────────────────┬──────────────────────────────────┘  │
└─────────────────────────┼────────────────────────────────────┘
                         │ OTLP/HTTP (4318)
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                        Tempo                                 │
│  - Recibe trazas OTLP                                        │
│  - Almacena en /tmp/tempo/traces                             │
│  - API de consulta en puerto 3200                            │
└─────────────────────────┬───────────────────────────────────┘
                         │
                         │ Query API
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                       Grafana                                │
│  ┌────────────────────────────────────────────────────────┐ │
│  │             Datasource: Prometheus                     │ │
│  │  - Consulta métricas con PromQL                       │ │
│  │  - Visualiza series temporales                        │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │             Datasource: Tempo                          │ │
│  │  - Consulta trazas                                     │ │
│  │  - Correlaciona con logs y métricas                   │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │             Datasource: Loki                           │ │
│  │  - Consulta logs con LogQL                            │ │
│  │  - Filtra por labels                                  │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │           Dashboard: Bank API Monitoring               │ │
│  │  - Request rate (Prometheus)                          │ │
│  │  - Request duration (Prometheus)                      │ │
│  │  - Business metrics (Prometheus)                      │ │
│  │  - Logs panel (Loki)                                  │ │
│  │  - Error logs (Loki)                                  │ │
│  │  - Traces (Tempo)                                     │ │
│  └────────────────────────────────────────────────────────┘ │
└────┬────────────────────┬───────────────────────────────────┘
     │ PromQL            │ LogQL Queries
     │                   │
     ▼                   ▼
┌──────────────┐  ┌─────────────────────────────────────────┐
│  Prometheus  │  │              Loki                        │
│              │  │  - Almacena logs indexados               │
│  - Scrape    │  │  - Storage: /tmp/loki                    │
│    /metrics  │  │  - Retención: 7 días                     │
│    cada 15s  │  └─────────────────────┬───────────────────┘
│  - Retención │                        ▲
│    7 días    │                        │ Push API (3100)
│  - Storage:  │                        │
│    /promethe │  ┌─────────────────────────────────────────┐
│    us        │  │              Promtail                    │
└──────────────┘  │  - Scraping de logs de contenedores      │
                  │  - Labels: namespace, app, pod           │
                  │  - Push a Loki                           │
                  └─────────────────────────────────────────┘
```

## Modelos de Datos

### Account
```go
type Account struct {
    ID            uint
    AccountNumber string
    Balance       float64
    CreatedAt     time.Time
    UpdatedAt     time.Time
}
```

### Transfer
```go
type Transfer struct {
    ID            uint
    FromAccountID uint
    ToAccountID   uint
    Amount        float64
    Description   string
    CreatedAt     time.Time
}
```

### Transaction
```go
type Transaction struct {
    ID          uint
    AccountID   uint
    Type        TransactionType  // deposit, withdrawal, transfer
    Amount      float64
    Reference   string
    Description string
    CreatedAt   time.Time
}
```

## Seguridad y Limitaciones

### Decisiones de Diseño (Simplificación)
- ❌ No hay autenticación
- ❌ No hay autorización
- ❌ Validaciones mínimas
- ❌ Sin rate limiting
- ❌ Sin encriptación de datos sensibles

### Para Producción se Necesitaría
- ✅ Autenticación JWT/OAuth2
- ✅ Autorización basada en roles
- ✅ Validaciones completas de datos
- ✅ Rate limiting
- ✅ HTTPS/TLS
- ✅ Encriptación de datos en reposo
- ✅ Auditoría completa
- ✅ Base de datos PostgreSQL/MySQL
- ✅ Alta disponibilidad (múltiples replicas)
- ✅ Backup y disaster recovery

## Escalabilidad

### Actual (Single Instance)
- 1 replica de Bank API
- SQLite (single file)
- Adecuado para: Demo, desarrollo, pruebas

### Para Escalar
1. **Horizontal Scaling**
   - Múltiples replicas de Bank API
   - Load balancer (Ingress en K8s)
   - Database: PostgreSQL/MySQL con connection pooling

2. **Observabilidad Escalada**
   - Tempo con S3/GCS backend
   - Loki con object storage
   - Prometheus con almacenamiento remoto (Thanos, Cortex)
   - Grafana con autenticación y RBAC

3. **Storage**
   - PersistentVolumes en K8s
   - Database managed service (RDS, CloudSQL)
   - Object storage para backups

## Despliegue en Diferentes Entornos

### Desarrollo (Docker Compose)
- Todo en un solo docker-compose.yml
- Volúmenes locales
- Networking simplificado

### Staging/Producción (Kubernetes)
- Namespaces separados
- PersistentVolumeClaims
- ConfigMaps y Secrets
- Resource limits y requests
- Health checks y readiness probes
- Rolling updates

## Monitoreo y Alerting

### Métricas Disponibles (Prometheus)

#### Métricas HTTP
- `http_requests_total` - Total de requests por método, endpoint y status code
- `http_request_duration_seconds` - Duración de requests HTTP (histograma)
- `http_request_size_bytes` - Tamaño de requests HTTP
- `http_response_size_bytes` - Tamaño de responses HTTP

#### Métricas de Negocio
- `bank_accounts_total` - Total de cuentas bancarias creadas
- `bank_transfers_total` - Total de transferencias (por status: success/failed)
- `bank_transfer_amount_total` - Monto total transferido
- `bank_account_balance` - Balance actual por cuenta (gauge)

### Logs Estructurados
- Formato JSON
- Trace ID en cada log
- Niveles: INFO, WARN, ERROR
- Contextual information

### Trazas Distribuidas
- Full request lifecycle
- Service-to-database latency
- Error tracking
- Performance bottlenecks identification
