// Пример использования пакета sampler.
//
// Задача: сервис обрабатывает тысячи запросов в секунду.
// Чтобы не перегружать систему логирования, мы хотим записывать
// подробный лог только для 10% запросов.
package main

import (
	"fmt"

	"github.com/system-highload-architect/go-solutions/data/sampler"
)

func main() {
	// Создаём сэмплер с вероятностью 10%.
	s := sampler.NewSampler(0.1)

	requests := []string{
		"GET /home", "POST /login", "GET /profile",
		"GET /home", "GET /about", "POST /login",
		"GET /profile", "GET /home", "POST /login",
		"GET /home", "GET /about", "GET /profile",
	}

	fmt.Println("Обработка запросов (логируем ~10%):")
	for _, req := range requests {
		if s.Sample() {
			fmt.Printf("  [LOG] %s — записан в подробный лог\n", req)
		}
		// Здесь происходит обычная обработка запроса...
	}
	fmt.Println("Обработка завершена")
}
