package metrics

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type promProvider struct {
	registry *prometheus.Registry
}

func newPrometheusProvider() Provider {
	reg := prometheus.NewRegistry()
	return &promProvider{registry: reg}
}

func (p *promProvider) NewCounter(name, help string, labels []string) Counter {
	cv := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: name,
		Help: help,
	}, labels)
	p.registry.MustRegister(cv)
	return &promCounter{cv: cv}
}

func (p *promProvider) NewHistogram(name, help string, buckets []float64, labels []string) Histogram {
	hv := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    name,
		Help:    help,
		Buckets: buckets,
	}, labels)
	p.registry.MustRegister(hv)
	return &promHistogram{hv: hv}
}

func (p *promProvider) Handler() http.Handler {
	return promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{})
}

func (p *promProvider) Shutdown(ctx context.Context) error {
	return nil
}

type promCounter struct {
	cv *prometheus.CounterVec
}

func (c *promCounter) Inc(labelVals ...string) {
	c.cv.WithLabelValues(labelVals...).Inc()
}

type promHistogram struct {
	hv *prometheus.HistogramVec
}

func (h *promHistogram) Observe(val float64, labelVals ...string) {
	h.hv.WithLabelValues(labelVals...).Observe(val)
}
