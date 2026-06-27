// Пример использования пакета breaker.
//
// Задача: сервис вызывает внешнее API, которое может быть недоступно.
// Circuit Breaker предотвращает каскадные отказы, быстро возвращая ошибку,
// когда API не отвечает в течение длительного времени.
package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/system-highload-architect/go-solutions/net/breaker"
)

func main() {
	// Создаём Circuit Breaker: разрывается после 2 ошибок,
	// остаётся открытым 3 секунды, затем пробует Half‑Open.
	cb := breaker.New("external-api", 2, 3*time.Second)

	// Имитация нестабильного внешнего сервиса.
	callExternal := func() error {
		// В 70% случаев возвращаем ошибку.
		if time.Now().UnixNano()%10 < 7 {
			return errors.New("service unavailable")
		}
		return nil
	}

	for i := 0; i < 10; i++ {
		err := cb.Execute(context.Background(), callExternal)
		switch {
		case err == nil:
			fmt.Printf("Запрос %d: успех\n", i)
		case errors.Is(err, breaker.ErrCircuitOpen):
			fmt.Printf("Запрос %d: цепь разомкнута, пропускаем\n", i)
		default:
			fmt.Printf("Запрос %d: ошибка (%v)\n", i, err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}
