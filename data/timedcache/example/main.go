// Пример использования пакета timedcache.
//
// Задача: веб‑приложение кэширует профили пользователей, полученные из внешнего API.
// Каждый профиль должен жить в кэше 5 минут, после чего автоматически удаляться.
// При удалении вызывается финализатор, который логирует факт истечения.
package main

import (
	"fmt"
	"time"

	"github.com/system-highload-architect/go-solutions/data/timedcache"
)

type UserProfile struct {
	Name  string
	Email string
}

func main() {
	// Создаём кэш с TTL 5 секунд (для быстрой демонстрации) и финализатором.
	cache := timedcache.New[string, *UserProfile](
		5*time.Second,
		timedcache.WithFinalizer[string, *UserProfile](func(key string, profile *UserProfile) {
			fmt.Printf("[FINALIZER] Ключ %q истёк, профиль %q удалён\n", key, profile.Name)
		}),
	)
	defer cache.Stop()

	// Добавляем профили.
	cache.Set("user:1", &UserProfile{Name: "Алексей", Email: "alex@example.com"})
	cache.Set("user:2", &UserProfile{Name: "Борис", Email: "boris@example.com"})

	// Получаем профиль до истечения TTL.
	if profile, ok := cache.Get("user:1"); ok {
		fmt.Printf("Найден: %s (%s)\n", profile.Name, profile.Email)
	}

	// Продлеваем жизнь ключу "user:2" ещё на 5 секунд.
	cache.Extend("user:2")

	// Ждём истечения TTL для "user:1".
	fmt.Println("Ждём 6 секунд...")
	time.Sleep(6 * time.Second)

	// Проверяем, что "user:1" уже удалён.
	if _, ok := cache.Get("user:1"); !ok {
		fmt.Println("Ключ \"user:1\" удалён (как и ожидалось)")
	}

	// "user:2" должен быть ещё жив.
	if profile, ok := cache.Get("user:2"); ok {
		fmt.Printf("Ключ \"user:2\" ещё жив: %s (%s)\n", profile.Name, profile.Email)
	}
}
