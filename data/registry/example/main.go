// Пример использования пакета registry.
//
// Задача: шлюз JSON‑RPC должен направлять вызовы к разным обработчикам
// в зависимости от имени метода. Реестр хранит строго типизированные
// обработчики и позволяет быстро их вызывать.
package main

import (
	"context"
	"fmt"

	"github.com/system-highload-architect/go-solutions/data/registry"
)

func main() {
	// Создаём реестр, где ключ — строка (имя метода),
	// запрос — строка, ответ — строка.
	reg := registry.New[string, string, string]()

	// Регистрируем обработчики.
	reg.Register("auction.bid", func(ctx context.Context, req string) (string, error) {
		return fmt.Sprintf("bid placed: %s", req), nil
	})
	reg.Register("accounting.debit", func(ctx context.Context, req string) (string, error) {
		return fmt.Sprintf("debited: %s", req), nil
	})

	// Диспетчеризуем вызовы.
	resp, err := reg.Dispatch(context.Background(), "auction.bid", "device-123")
	if err != nil {
		fmt.Println("Ошибка:", err)
	} else {
		fmt.Println(resp)
	}

	resp, err = reg.Dispatch(context.Background(), "accounting.debit", "campaign-1")
	if err != nil {
		fmt.Println("Ошибка:", err)
	} else {
		fmt.Println(resp)
	}

	// Попытка вызвать незарегистрированный метод.
	_, err = reg.Dispatch(context.Background(), "unknown.method", "data")
	if err != nil {
		fmt.Println("Ожидаемая ошибка:", err)
	}
}
