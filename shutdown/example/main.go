package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/system-highload-architect/go-solutions/logger"
	"github.com/system-highload-architect/go-solutions/shutdown"
)

// Пример демонстрирует приоритетное завершение через Stop.
// Задача: веб-сервис получил фатальную ошибку БД и должен немедленно
// закрыть соединения и остановить сервер, не дожидаясь сигнала ОС.
func main() {
	log := logger.New("info", "json", slog.String("service", "webapp"))
	srv := &http.Server{Addr: ":8080"}
	db, _ := sql.Open("postgres", "...")

	mgr := shutdown.NewManager(30 * time.Second)
	mgr.SetLogger(log)

	// Приоритет 0 – остановить сервер как можно быстрее,
	// чтобы не принимать новые запросы.
	mgr.Add("http_server", 0, func(ctx context.Context) error {
		log.Info("shutting down HTTP server")
		return srv.Shutdown(ctx)
	}, 5*time.Second)

	// Приоритет 1 – закрыть базу данных после остановки сервера.
	mgr.Add("database", 1, func(ctx context.Context) error {
		log.Info("closing database")
		return db.Close()
	}, 3*time.Second)

	go srv.ListenAndServe()

	// Имитация фатальной ошибки через 2 секунды.
	go func() {
		time.Sleep(2 * time.Second)
		log.Error("database connection lost, initiating shutdown")
		if err := mgr.Stop(); err != nil {
			log.Error("shutdown failed", "error", err)
		}
	}()

	// Обычно ждём сигнал ОС, но здесь для примера просто ждём 5 секунд.
	time.Sleep(5 * time.Second)
}
