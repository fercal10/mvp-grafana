# Prometheus Integration - Setup Guide

## Overview

Prometheus has been successfully integrated into the Bank API project, completing the observability stack with metrics collection alongside existing logs (Loki) and traces (Tempo).

## What Was Added

### 1. Go Application Changes

#### Dependencies
- Added `github.com/prometheus/client_golang v1.20.5` to `go.mod`

#### New Files
- `pkg/telemetry/metrics.go` - Prometheus metrics configuration and middleware

#### Modified Files
- `cmd/server/main.go` - Added Prometheus middleware and `/metrics` endpoint
- `internal/service/account_service.go` - Added business metrics for account operations
- `internal/service/transfer_service.go` - Added business metrics for transfer operations

### 2. Infrastructure Configuration

#### Docker Compose
- Added Prometheus service in `docker-compose.yml`
- Port: `9090:9090`
- Volume: `prometheus-data`

#### Kubernetes
Created new directory `k8s/prometheus/` with:
- `configmap.yaml` - Prometheus configuration
- `deployment.yaml` - Prometheus deployment with resource limits
- `service.yaml` - ClusterIP and NodePort services
- `pvc.yaml` - PersistentVolumeClaim for data storage

#### Configuration Files
- `config/prometheus.yaml` - Prometheus scraping configuration
- `config/grafana/datasources.yaml` - Added Prometheus datasource
- `k8s/grafana/configmap.yaml` - Added Prometheus datasource for K8s

### 3. Documentation Updates
- Updated `README.md` with Prometheus information
- Updated `ARCHITECTURE.md` with detailed architecture diagrams

## Metrics Exposed

### HTTP Metrics (Automatic)
- `http_requests_total{method, endpoint, status}` - Total HTTP requests
- `http_request_duration_seconds{method, endpoint, status}` - Request duration histogram
- `http_request_size_bytes{method, endpoint}` - Request size histogram
- `http_response_size_bytes{method, endpoint, status}` - Response size histogram

### Business Metrics
- `bank_accounts_total` - Total accounts created (Counter)
- `bank_transfers_total{status}` - Total transfers by status (Counter)
- `bank_transfer_amount_total` - Total amount transferred (Counter)
- `bank_account_balance{account_number}` - Current account balance (Gauge)

## How to Use

### Docker Compose

1. Start all services:
```bash
docker-compose up -d
```

2. Access Prometheus UI:
```
http://localhost:9090
```

3. Access metrics endpoint:
```
curl http://localhost:8080/metrics
```

4. View in Grafana:
```
http://localhost:3000 (admin/admin)
```

### Kubernetes

1. Deploy Prometheus:
```bash
kubectl apply -f k8s/prometheus/
```

2. Access Prometheus (NodePort):
```
http://localhost:30900
```

Or use port-forward:
```bash
kubectl port-forward -n banking-system svc/prometheus 9090:9090
```

## Example Queries

### PromQL Examples

#### Request Rate
```promql
rate(http_requests_total[5m])
```

#### Request Rate by Endpoint
```promql
sum by (endpoint) (rate(http_requests_total[5m]))
```

#### Error Rate
```promql
sum(rate(http_requests_total{status=~"5.."}[5m]))
```

#### Request Duration 95th Percentile
```promql
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))
```

#### Total Transfers (Success vs Failed)
```promql
sum by (status) (bank_transfers_total)
```

#### Total Amount Transferred
```promql
bank_transfer_amount_total
```

#### Account Balance
```promql
bank_account_balance{account_number="ACC001"}
```

## Configuration Details

### Scraping Configuration
- **Interval**: 15 seconds
- **Retention**: 7 days
- **Storage**: 10GB max

### Targets
Prometheus scrapes metrics from:
- Bank API (`bank-api:8080/metrics`)
- Prometheus itself (`localhost:9090`)
- Grafana (`grafana:3000`)
- Loki (`loki:3100`)
- Tempo (`tempo:3200`)

## Integration with Grafana

Prometheus is configured as a datasource in Grafana with:
- URL: `http://prometheus:9090`
- HTTP Method: POST
- Exemplar support enabled (linked to Tempo traces)

You can now create dashboards in Grafana that combine:
- **Metrics** from Prometheus
- **Logs** from Loki
- **Traces** from Tempo

## Monitoring Best Practices

1. **Request Monitoring**: Track request rate, duration, and errors
2. **Business Metrics**: Monitor account creations and transfer volumes
3. **Alerting**: Set up alerts for error rates and latency thresholds
4. **Capacity Planning**: Use metrics to plan infrastructure scaling

## Next Steps

Consider implementing:
1. **Grafana Dashboards**: Create custom dashboards combining Prometheus, Loki, and Tempo
2. **Alerting**: Configure Prometheus AlertManager for notifications
3. **Service Level Objectives (SLOs)**: Define SLIs and SLOs based on metrics
4. **Long-term Storage**: Configure remote storage (Thanos, Cortex) for production

## Troubleshooting

### Metrics Not Appearing
1. Check if Prometheus is scraping the Bank API:
   ```
   http://localhost:9090/targets
   ```

2. Verify metrics endpoint is accessible:
   ```bash
   curl http://localhost:8080/metrics
   ```

### Prometheus Not Starting
1. Check logs:
   ```bash
   # Docker Compose
   docker-compose logs prometheus
   
   # Kubernetes
   kubectl logs -n banking-system -l app=prometheus
   ```

2. Verify configuration:
   ```bash
   # Validate Prometheus config
   docker run --rm -v $(pwd)/config/prometheus.yaml:/etc/prometheus/prometheus.yaml prom/prometheus:latest promtool check config /etc/prometheus/prometheus.yaml
   ```

## Architecture Summary

```
Bank API (/metrics)
    │
    ▼
Prometheus (scrape every 15s)
    │
    ▼
Grafana (query with PromQL)
    │
    └─► Dashboards (Metrics + Logs + Traces)
```

Complete observability stack:
- **Metrics**: Prometheus
- **Logs**: Loki + Promtail
- **Traces**: Tempo + OpenTelemetry
- **Visualization**: Grafana
