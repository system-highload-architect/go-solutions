// Пример использования пакета breaker.
//
// Задача: сервис вызывает внешний платёжный шлюз, который периодически
// недоступен. Circuit Breaker предотвращает каскадные отказы и лавинообразный
// рост задержек, быстро возвращая ошибку, если шлюз стабильно не отвечает.
//
// Алгоритм:
//   - Closed: все запросы проходят. После 2 последовательных ошибок цепь
//     размыкается (Open) на 3 секунды.
//   - Open: запросы немедленно отклоняются с ErrCircuitOpen. Через 3 секунды
//     состояние меняется на Half-Open.
//   - Half-Open: разрешается один пробный запрос. Если он успешен – цепь
//     замыкается (Closed), если нет – снова Open.
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/system-highload-architect/go-solutions/net/breaker"
)

func main() {
	// Создаём Circuit Breaker для платёжного шлюза.
	cb := breaker.New("payment-gateway", 2, 3*time.Second)

	// Имитация нестабильного внешнего сервиса.
	// В начале он возвращает ошибки, потом восстанавливается.
	callCounter := 0
	callExternal := func() error {
		callCounter++
		if callCounter <= 3 {
			return errors.New("gateway timeout")
		}
		return nil // на 4-м вызове "чинится"
	}

	fmt.Println("=== Демонстрация Circuit Breaker ===\n")

	for i := 1; i <= 8; i++ {
		fmt.Printf("Запрос %d (состояние: %s):\n", i, cb.State())
		err := cb.Execute(context.Background(), callExternal)
		switch {
		case err == nil:
			fmt.Println("  ✅ Успех")
		case errors.Is(err, breaker.ErrCircuitOpen):
			fmt.Println("  ⚠️ Цепь разомкнута, запрос отклонён без вызова сервиса")
		default:
			fmt.Printf("  ❌ Ошибка: %v\n", err)
		}
		time.Sleep(1 * time.Second)
	}

	fmt.Printf("\nИтоговое состояние: %s\n", cb.State())
}
