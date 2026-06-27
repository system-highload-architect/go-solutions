package metrics

import (
	"context"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

type otelProvider struct {
	meter  metric.Meter
	pusher *sdkmetric.MeterProvider
}

func newOTLPProvider(ctx context.Context, serviceName string) (Provider, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	exporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return nil, err
	}

	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(exporter,
				sdkmetric.WithInterval(10*time.Second),
			),
		),
	)
	otel.SetMeterProvider(provider)

	return &otelProvider{
		meter:  provider.Meter(serviceName),
		pusher: provider,
	}, nil
}

func (p *otelProvider) NewCounter(name, help string, labels []string) Counter {
	c, _ := p.meter.Int64Counter(name, metric.WithDescription(help))
	return &otelCounter{c: c, labels: labels}
}

func (p *otelProvider) NewHistogram(name, help string, buckets []float64, labels []string) Histogram {
	h, _ := p.meter.Float64Histogram(name, metric.WithDescription(help))
	return &otelHistogram{h: h, labels: labels}
}

func (p *otelProvider) Handler() http.Handler {
	return http.NotFoundHandler()
}

func (p *otelProvider) Shutdown(ctx context.Context) error {
	return p.pusher.Shutdown(ctx)
}

type otelCounter struct {
	c      metric.Int64Counter
	labels []string
}

func (c *otelCounter) Inc(labelVals ...string) {
	attrs := make([]attribute.KeyValue, len(labelVals))
	for i, v := range labelVals {
		attrs[i] = attribute.String(c.labels[i], v)
	}
	c.c.Add(context.Background(), 1, metric.WithAttributes(attrs...))
}

type otelHistogram struct {
	h      metric.Float64Histogram
	labels []string
}

func (h *otelHistogram) Observe(val float64, labelVals ...string) {
	attrs := make([]attribute.KeyValue, len(labelVals))
	for i, v := range labelVals {
		attrs[i] = attribute.String(h.labels[i], v)
	}
	h.h.Record(context.Background(), val, metric.WithAttributes(attrs...))
}
