// Пример использования config в микросервисе.
package main

import (
	"fmt"
	"log"

	"github.com/system-highload-architect/go-solutions/config"
)

type AppConfig struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
}

type ServerConfig struct {
	Port int `yaml:"port" env:"SERVER_PORT"`
}

type DatabaseConfig struct {
	DSN string `yaml:"dsn" env:"DATABASE_DSN"`
}

func main() {
	var cfg AppConfig
	if err := config.Load(&cfg,
		config.WithPath("config.yaml"),
		config.WithEnvPrefix("MYAPP_"),
	); err != nil {
		log.Fatalf("cannot load config: %v", err)
	}

	fmt.Printf("Server port: %d\n", cfg.Server.Port)
	fmt.Printf("Database DSN: %s\n", cfg.Database.DSN)
	// Далее запуск сервера...
}
