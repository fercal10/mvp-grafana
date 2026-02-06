package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tribal/bank-api/internal/handlers"
	"github.com/tribal/bank-api/internal/repository"
	"github.com/tribal/bank-api/internal/service"
	"github.com/tribal/bank-api/pkg/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

const (
	serviceName    = "transfers-api"
	serviceVersion = "1.0.0"
)

func main() {
	ctx := context.Background()

	// Setup Loki logger
	lokiURL := os.Getenv("LOKI_ENDPOINT")
	logger := telemetry.NewLokiLogger(lokiURL, serviceName)

	// Setup OpenTelemetry
	shutdown, err := telemetry.SetupOpenTelemetry(ctx, serviceName, serviceVersion)
	if err != nil {
		logger.Error("Warning: Failed to setup OpenTelemetry: %v", err)
	} else {
		defer func() {
			if err := shutdown(ctx); err != nil {
				logger.Error("Error shutting down OpenTelemetry: %v", err)
			}
		}()
	}

	// Get database path from environment or use default
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/bank.db"
	}

	// Initialize repository
	repo, err := repository.NewRepository(dbPath)
	if err != nil {
		logger.Fatal("Failed to initialize repository: %v", err)
	}

	// Initialize services and handlers
	transferService := service.NewTransferService(repo)
	transferHandler := handlers.NewTransferHandler(transferService)
	transactionHandler := handlers.NewTransactionHandler(serviceName)

	// Setup Gin router
	router := gin.Default()

	// Add custom Loki logging middleware
	router.Use(logger.GinMiddleware())

	// Add Prometheus middleware
	router.Use(telemetry.PrometheusMiddleware())

	// Add OpenTelemetry middleware
	router.Use(otelgin.Middleware(serviceName))

	// Metrics endpoint for Prometheus
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check endpoint
	router.GET("/health", transactionHandler.HealthCheck)
	router.GET("/ready", transactionHandler.HealthCheck)

	// API routes
	api := router.Group("/api")
	{
		api.POST("/transfers", transferHandler.CreateTransfer)
		api.GET("/transfers/:id", transferHandler.GetTransfer)
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}
