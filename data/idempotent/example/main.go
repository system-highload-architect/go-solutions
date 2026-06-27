// Пример использования пакета idempotent.
//
// Задача: платёжный шлюз получает запросы на списание средств.
// Из‑за сетевых сбоев клиент может отправить один и тот же запрос дважды.
// Ключ идемпотентности гарантирует, что повторный запрос будет отклонён.
package main

import (
	"fmt"
	"time"

	"github.com/system-highload-architect/go-solutions/data/idempotent"
)

func main() {
	// Создаём хранилище с TTL 10 секунд (для быстрой демонстрации).
	store := idempotent.NewStore[string](10 * time.Second)
	defer store.Stop()

	paymentKey := "payment-2026-001"

	// Первый запрос (успешен).
	if store.Check(paymentKey) {
		fmt.Println("Платёж обработан")
	}

	// Повторный запрос с тем же ключом (отклонён).
	if !store.Check(paymentKey) {
		fmt.Println("Повторный платёж отклонён (дубликат)")
	}

	// Ждём истечения TTL.
	fmt.Println("Ждём 11 секунд...")
	time.Sleep(11 * time.Second)

	// После истечения TTL ключ снова становится новым.
	if store.Check(paymentKey) {
		fmt.Println("Платёж обработан после истечения TTL")
	}
}
