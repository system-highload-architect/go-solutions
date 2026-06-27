// Пример использования пакета ratelimit.
//
// Задача: выбрать подходящий алгоритм ограничения скорости для веб‑сервера.
// Ниже демонстрируются четыре алгоритма, их поведение и рекомендации по выбору.
//
// Алгоритмы:
// 1. Token Bucket (рекомендуемый) – гибкий, позволяет кратковременные бёрсты.
// 2. Leaky Bucket – сглаживает поток, обрабатывает запросы равномерно.
// 3. Sliding Window Log – точный, но более затратный по памяти.
// 4. Fixed Window – простой, но страдает от «эффекта границы окна».
//
// Для большинства веб‑сервисов оптимален Token Bucket: он естественно
// ограничивает среднюю скорость, но разрешает бёрсты до заданного предела.
// Leaky Bucket хорош, когда нужно гарантировать постоянную интенсивность
// обработки (например, запись в БД с фиксированной скоростью).
// Sliding Window Log полезен для строгих SLA, где важна точность до миллисекунд.
// Fixed Window стоит использовать только для черновых ограничений, так как
// допускает удвоенную нагрузку на стыке окон.
package main

import (
	"fmt"
	"time"

	"github.com/system-highload-architect/go-solutions/net/ratelimit"
)

func main() {
	fmt.Println("=== Token Bucket ===")
	fmt.Println("5 токенов/сек, бёрст 3. Позволяет отправить до 3 запросов мгновенно,")
	fmt.Println("затем ограничивает среднюю скорость.")
	tb := ratelimit.NewTokenBucket(5, 3)
	for i := 0; i < 8; i++ {
		if tb.Allow() {
			fmt.Printf("  Запрос %d: разрешён\n", i+1)
		} else {
			fmt.Printf("  Запрос %d: отклонён\n", i+1)
		}
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("\n=== Leaky Bucket ===")
	fmt.Println("5 запросов/сек, ёмкость 3. Запросы «вытекают» равномерно,")
	fmt.Println("лишние отбрасываются, если очередь заполнена.")
	lb := ratelimit.NewLeakyBucket(5, 3)
	for i := 0; i < 8; i++ {
		if lb.Allow() {
			fmt.Printf("  Запрос %d: разрешён\n", i+1)
		} else {
			fmt.Printf("  Запрос %d: отклонён\n", i+1)
		}
		time.Sleep(100 * time.Millisecond)
	}

	fmt.Println("\n=== Sliding Window Log ===")
	fmt.Println("Лимит 3 запроса за 1 секунду. Хранит точные времена последних запросов,")
	fmt.Println("отклоняет при превышении лимита в любой момент окна.")
	swl := ratelimit.NewSlidingWindowLog(3, 1*time.Second)
	for i := 0; i < 6; i++ {
		if swl.Allow() {
			fmt.Printf("  Запрос %d: разрешён\n", i+1)
		} else {
			fmt.Printf("  Запрос %d: отклонён\n", i+1)
		}
		time.Sleep(200 * time.Millisecond)
	}

	fmt.Println("\n=== Fixed Window ===")
	fmt.Println("Лимит 3 запроса в секунду. Но! В начале новой секунды счётчик сбрасывается,")
	fmt.Println("поэтому на стыке окон можно пропустить 2 * лимит запросов за короткое время.")
	fw := ratelimit.NewFixedWindow(3, 1*time.Second)
	for i := 0; i < 6; i++ {
		if fw.Allow() {
			fmt.Printf("  Запрос %d: разрешён\n", i+1)
		} else {
			fmt.Printf("  Запрос %d: отклонён\n", i+1)
		}
		time.Sleep(200 * time.Millisecond)
	}
}
