// Пример использования пакета lru.
//
// Задача: сервис кэширует результаты дорогих вычислений. Размер кэша ограничен 3 записями.
// При добавлении 4‑го элемента самый старый автоматически удаляется, и вызывается
// callback-логирование, чтобы отслеживать вытеснения.
package main

import (
	"fmt"

	"github.com/system-highload-architect/go-solutions/data/lru"
)

func main() {
	// Создаём LRU‑кеш ёмкостью 3 с колбэком на вытеснение.
	cache := lru.New[string, string](3,
		lru.WithEvictCallback[string, string](func(key string, value string) {
			fmt.Printf("[EVICTED] Ключ %q со значением %q удалён\n", key, value)
		}),
	)

	// Заполняем кеш.
	cache.Set("a", "Alice")
	cache.Set("b", "Bob")
	cache.Set("c", "Charlie")

	fmt.Println("После добавления трёх элементов:")
	printKeys(cache, "a", "b", "c")

	// Добавляем четвёртый – самый старый ("a") должен быть вытеснен.
	cache.Set("d", "Diana")
	fmt.Println("\nПосле добавления 'd':")
	printKeys(cache, "a", "b", "c", "d")

	// Продлеваем жизнь ключу "b" (перемещаем в голову).
	cache.Extend("b")
	fmt.Println("\nПосле продления 'b':")
	printKeys(cache, "a", "b", "c", "d")

	// Добавляем ещё один – теперь вытеснится "c".
	cache.Set("e", "Eve")
	fmt.Println("\nПосле добавления 'e':")
	printKeys(cache, "a", "b", "c", "d", "e")
}

func printKeys(c *lru.Cache[string, string], keys ...string) {
	for _, k := range keys {
		if v, ok := c.Get(k); ok {
			fmt.Printf("  %s -> %s\n", k, v)
		}
	}
}
