// Пример использования пакета logger.
//
// Задача: создать логгер для веб‑сервиса, который пишет в JSON‑формате
// с уровнем INFO и автоматически добавляет имя сервиса в каждую запись.
package main

import (
	"log/slog"

	"github.com/system-highload-architect/go-solutions/logger"
)

func main() {
	// Создаём логгер с уровнем info, форматом json и атрибутом service=myapp.
	log := logger.New("info", "json", slog.String("service", "myapp"))

	log.Info("server started", "port", 8080)
	log.Debug("this will not appear") // уровень debug ниже info
	log.Error("something went wrong", "error", "connection refused")
}
