// Пример использования пакета timewheel.
//
// Задача: реализовать планировщик напоминаний. Пользователь может
// установить таймер на определённое время, и когда оно наступает,
// система отправляет уведомление. Таймеры можно продлевать и отменять.
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/system-highload-architect/go-solutions/data/timewheel"
)

type Reminder struct {
	UserID  string
	Message string
}

func main() {
	// Создаём колесо с тиком 1 секунда и 60 слотами (покрывает 60 секунд).
	tw := timewheel.New[string, *Reminder](1*time.Second, 60,
		timewheel.WithExpireCallback[string, *Reminder](func(task timewheel.Task[string, *Reminder]) {
			fmt.Printf("[НАПОМИНАНИЕ] %s (пользователь %s)\n", task.Value.Message, task.Value.UserID)
		}),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запускаем автоматический тик.
	tw.Start(ctx)
	defer tw.Stop()

	// Устанавливаем таймер на 3 секунды.
	reminder1 := &Reminder{UserID: "user1", Message: "Проверить почту"}
	tw.Add("timer1", reminder1, 3*time.Second)
	fmt.Println("Установлен таймер на 3 секунды для user1")

	// Устанавливаем таймер на 5 секунд.
	reminder2 := &Reminder{UserID: "user2", Message: "Сделать звонок"}
	tw.Add("timer2", reminder2, 5*time.Second)
	fmt.Println("Установлен таймер на 5 секунд для user2")

	// Продлеваем таймер timer1 ещё на 4 секунды (теперь сработает через 7 секунд от старта).
	time.Sleep(2 * time.Second)
	tw.MoveValue("timer1", reminder1, 4*time.Second)
	fmt.Println("Таймер timer1 продлён на 4 секунды")

	// Отменяем таймер timer2.
	time.Sleep(2 * time.Second)
	tw.Remove("timer2")
	fmt.Println("Таймер timer2 отменён")

	// Ждём, чтобы увидеть срабатывание timer1.
	time.Sleep(5 * time.Second)
	fmt.Println("Завершение примера")
}
