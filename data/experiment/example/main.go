// Пример использования пакета experiment.
//
// Задача: интернет‑магазин тестирует новый алгоритм рекомендаций.
// 20% пользователей должны видеть новый алгоритм («new_recs»),
// остальные 80% — старый. Распределение детерминировано:
// один и тот же пользователь всегда попадает в одну группу.
package main

import (
	"fmt"

	"github.com/system-highload-architect/go-solutions/data/experiment"
)

func main() {
	// Создаём эксперимент с 20% трафика на новую версию.
	flags := map[string]float64{
		"new_recs": 0.2,
	}
	exp := experiment.New(flags)

	// Список пользователей (email).
	users := []string{
		"alice@example.com",
		"bob@example.com",
		"charlie@example.com",
		"diana@example.com",
		"eve@example.com",
	}

	fmt.Println("Распределение пользователей по группам:")
	for _, user := range users {
		if exp.IsInExperiment(user, "new_recs") {
			fmt.Printf("  %s → новая версия рекомендаций\n", user)
		} else {
			fmt.Printf("  %s → старая версия\n", user)
		}
	}

	// Проверяем, что один и тот же пользователь всегда в одной группе.
	user := "alice@example.com"
	fmt.Printf("\nПовторная проверка для %s: ", user)
	if exp.IsInExperiment(user, "new_recs") {
		fmt.Println("новая версия")
	} else {
		fmt.Println("старая версия")
	}
}
