package telemetry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// LokiLogger sends logs to Loki and stdout
type LokiLogger struct {
	lokiURL     string
	serviceName string
	client      *http.Client
	mu          sync.Mutex
	logLevel    string
}

// NewLokiLogger creates a new logger instance
func NewLokiLogger(lokiURL, serviceName string) *LokiLogger {
	if lokiURL == "" {
		lokiURL = "http://localhost:3100"
	}

	return &LokiLogger{
		lokiURL:     lokiURL,
		serviceName: serviceName,
		client:      &http.Client{Timeout: 5 * time.Second},
		logLevel:    "info",
	}
}

type lokiPushRequest struct {
	Streams []lokiStream `json:"streams"`
}

type lokiStream struct {
	Stream map[string]string `json:"stream"`
	Values [][]string        `json:"values"`
}

// Info logs info message
func (l *LokiLogger) Info(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.log("info", msg)
}

// Error logs error message
func (l *LokiLogger) Error(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.log("error", msg)
}

// Fatal logs error message and exits
func (l *LokiLogger) Fatal(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.log("fatal", msg)
	os.Exit(1)
}

func (l *LokiLogger) log(level, msg string) {
	// Print to stdout
	log.Printf("[%s] %s", level, msg)

	// Send to Loki asynchronously
	go func() {
		if err := l.sendToLoki(level, msg); err != nil {
			// Don't log error to avoid infinite loop, just print to stderr
			fmt.Fprintf(os.Stderr, "Failed to send log to Loki: %v\n", err)
		}
	}()
}

func (l *LokiLogger) sendToLoki(level, msg string) error {
	timestamp := fmt.Sprintf("%d", time.Now().UnixNano())

	payload := lokiPushRequest{
		Streams: []lokiStream{
			{
				Stream: map[string]string{
					"service":   l.serviceName,
					"app":       l.serviceName,
					"namespace": "banking-system",
					"level":     level,
					"source":    l.serviceName,
				},
				Values: [][]string{
					{timestamp, msg},
				},
			},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/loki/api/v1/push", l.lokiURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := l.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GinMiddleware returns a gin middleware that logs requests to Loki
func (l *LokiLogger) GinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		if raw != "" {
			path = path + "?" + raw
		}

		// Create log message
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// Structure the log message as JSON for Grafana parsing
		logEntry := map[string]interface{}{
			"status":     statusCode,
			"method":     method,
			"path":       path,
			"ip":         clientIP,
			"latency_ms": latency.Milliseconds(),
			"user_agent": c.Request.UserAgent(),
			"error":      errorMessage,
			"msg":        "http_request",
		}

		// Log to stdout for debugging
		if statusCode >= 500 {
			log.Printf("[ERROR] %d | %s | %s | %s | %s", statusCode, method, path, clientIP, errorMessage)
		} else {
			log.Printf("[INFO] %d | %s | %s | %s", statusCode, method, path, clientIP)
		}

		// Send structured log to Loki
		go func() {
			jsonBytes, _ := json.Marshal(logEntry)
			if err := l.sendToLoki("info", string(jsonBytes)); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to send request log to Loki: %v\n", err)
			}
		}()
	}
}
