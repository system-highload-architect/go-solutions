// Package metrics provides a unified interface for application metrics with
// support for OpenTelemetry (OTLP) and Prometheus backends.
//
// Example usage:
//
//	package main
//
//	import (
//	    "context"
//	    "log"
//	    "github.com/system-highload-architect/go-solutions/metrics"
//	)
//
//	func main() {
//	    // Инициализируем метрики с OTLP-экспортёром
//	    if err := metrics.Init(context.Background(), "myapp", true); err != nil {
//	        log.Fatal(err)
//	    }
//	    defer metrics.Shutdown(context.Background())
//
//	    // Создаём счётчик
//	    reqCounter := metrics.NewCounter("requests_total", "Total requests", []string{"method"})
//	    reqCounter.Inc("GET")
//	    reqCounter.Inc("POST")
//
//	    // Создаём гистограмму
//	    latencyHist := metrics.NewHistogram("request_latency_seconds", "Request latency",
//	        []float64{0.01, 0.05, 0.1, 0.5, 1}, []string{})
//	    latencyHist.Observe(0.23)
//	}
package metrics

import (
	"context"
	"net/http"
)

// Counter is a monotonically increasing metric.
type Counter interface {
	Inc(labelVals ...string)
}

// Histogram records the distribution of values.
type Histogram interface {
	Observe(val float64, labelVals ...string)
}

// Provider abstracts the underlying metrics backend.
type Provider interface {
	NewCounter(name, help string, labels []string) Counter
	NewHistogram(name, help string, buckets []float64, labels []string) Histogram
	Handler() http.Handler // returns an HTTP handler for exposing metrics (e.g., /metrics)
	Shutdown(ctx context.Context) error
}

// ----------------------------------------------------------------------------
// Global provider (initialised once at startup)
// ----------------------------------------------------------------------------
var globalProvider Provider = &noopProvider{}

// Init initialises the global metrics provider.
//   - serviceName: name of the service (used in OTLP resource)
//   - useOTLP: if true, OpenTelemetry OTLP exporter is used; otherwise Prometheus text format.
func Init(ctx context.Context, serviceName string, useOTLP bool) error {
	if useOTLP {
		p, err := newOTLPProvider(ctx, serviceName)
		if err != nil {
			return err
		}
		globalProvider = p
	} else {
		globalProvider = newPrometheusProvider()
	}
	return nil
}

// NewCounter creates a new counter using the global provider.
func NewCounter(name, help string, labels []string) Counter {
	return globalProvider.NewCounter(name, help, labels)
}

// NewHistogram creates a new histogram using the global provider.
func NewHistogram(name, help string, buckets []float64, labels []string) Histogram {
	return globalProvider.NewHistogram(name, help, buckets, labels)
}

// Handler returns the HTTP handler for exposing metrics.
func Handler() http.Handler {
	return globalProvider.Handler()
}

// Shutdown gracefully stops the metrics provider.
func Shutdown(ctx context.Context) error {
	return globalProvider.Shutdown(ctx)
}

// ----------------------------------------------------------------------------
// No-op provider (used before Init)
// ----------------------------------------------------------------------------
type noopProvider struct{}

func (n *noopProvider) NewCounter(name, help string, labels []string) Counter { return &noopCounter{} }
func (n *noopProvider) NewHistogram(name, help string, buckets []float64, labels []string) Histogram {
	return &noopHistogram{}
}
func (n *noopProvider) Handler() http.Handler              { return http.NotFoundHandler() }
func (n *noopProvider) Shutdown(ctx context.Context) error { return nil }

type noopCounter struct{}

func (n *noopCounter) Inc(labelVals ...string) {}

type noopHistogram struct{}

func (n *noopHistogram) Observe(val float64, labelVals ...string) {}
