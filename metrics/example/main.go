// Пример использования пакета metrics.
//
// Задача: сервис должен вести учёт количества запросов и задержек.
// В production метрики отправляются через OTLP в коллектор,
// а в разработке доступны на /metrics в формате Prometheus.
package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/system-highload-architect/go-solutions/metrics"
)

func main() {
	// Инициализируем метрики (false = Prometheus, true = OTLP)
	if err := metrics.Init(context.Background(), "myapp", false); err != nil {
		log.Fatal(err)
	}
	defer metrics.Shutdown(context.Background())

	// Регистрируем метрики
	reqCounter := metrics.NewCounter("requests_total", "Total requests", []string{"method"})
	latencyHist := metrics.NewHistogram("request_latency_seconds", "Request latency",
		[]float64{0.01, 0.05, 0.1, 0.5, 1}, []string{})

	// Имитируем обработку запросов
	go func() {
		for {
			reqCounter.Inc("GET")
			latencyHist.Observe(0.23)
			time.Sleep(500 * time.Millisecond)
		}
	}()

	// Отдаём метрики по /metrics
	http.Handle("/metrics", metrics.Handler())
	log.Println("Listening on :2112")
	http.ListenAndServe(":2112", nil)
}
