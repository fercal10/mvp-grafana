package telemetry

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "Size of HTTP requests in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "endpoint"},
	)

	httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "Size of HTTP responses in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "endpoint", "status"},
	)

	// Business metrics
	BankAccountsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "bank_accounts_total",
			Help: "Total number of bank accounts created",
		},
	)

	BankTransfersTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "bank_transfers_total",
			Help: "Total number of transfers processed",
		},
		[]string{"status"}, // success, failed
	)

	BankTransferAmountTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "bank_transfer_amount_total",
			Help: "Total amount transferred",
		},
	)

	BankAccountBalance = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bank_account_balance",
			Help: "Current balance of bank accounts",
		},
		[]string{"account_number"},
	)
)

// PrometheusMiddleware is a Gin middleware that records HTTP metrics
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Get request size
		requestSize := computeApproximateRequestSize(c.Request)

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()
		
		// Get response info
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		endpoint := c.FullPath()
		
		// If endpoint is empty (404), use the request path
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		// Get response size
		responseSize := float64(c.Writer.Size())

		// Record metrics
		httpRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
		httpRequestDuration.WithLabelValues(method, endpoint, status).Observe(duration)
		httpRequestSize.WithLabelValues(method, endpoint).Observe(float64(requestSize))
		
		if responseSize > 0 {
			httpResponseSize.WithLabelValues(method, endpoint, status).Observe(responseSize)
		}
	}
}

// computeApproximateRequestSize computes the approximate size of the request
func computeApproximateRequestSize(r *http.Request) int {
	s := 0
	if r.URL != nil {
		s += len(r.URL.Path)
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}
	return s
}

// RecordTransfer records a transfer metric
func RecordTransfer(amount float64, success bool) {
	status := "success"
	if !success {
		status = "failed"
	}
	BankTransfersTotal.WithLabelValues(status).Inc()
	
	if success {
		BankTransferAmountTotal.Add(amount)
	}
}

// RecordAccountCreation records an account creation metric
func RecordAccountCreation() {
	BankAccountsTotal.Inc()
}

// UpdateAccountBalance updates the balance gauge for an account
func UpdateAccountBalance(accountNumber string, balance float64) {
	BankAccountBalance.WithLabelValues(accountNumber).Set(balance)
}
