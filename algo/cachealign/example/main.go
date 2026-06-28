// Пример использования пакета cachealign.
//
// Задача: есть таблица переходов конечного автомата (FSM).
// События поступают массово (например, 1000 в секунду).
// Нужно быстро выбирать функцию-обработчик по составному ключу (тип+состояние+событие).
//
// Пакет cachealign помогает:
//   - упаковать ключ в компактное uint32 (быстрое сравнение, мало памяти),
//   - группировать ключи в батчи, выровненные по размеру кэш-линии (64 байта),
//     чтобы при последовательном обходе минимизировать промахи кэша процессора.
//
// Преимущества:
//   - меньше cache‑misses → выше пропускная способность на больших объёмах,
//   - нет указателей в куче → меньше нагрузки на GC,
//   - батчи можно обрабатывать параллельно (errgroup) без блокировок.
//
// Недостатки:
//   - требует интернирования строк на этапе парсинга запроса,
//   - небольшое усложнение кода ради выигрыша, который заметен только на
//     действительно больших нагрузках (>100 000 событий/сек).
package main

import (
	"fmt"

	"github.com/system-highload-architect/go-solutions/algo/cachealign"
)

// ─── Константы для FSM (пример) ────────────────────────────────────
const (
	TypeStandard uint8 = iota
	TypeVIP
)

const (
	StateCreated uint8 = iota
	StateReserved
)

const (
	EventReserve uint8 = iota
	EventCancel
)

// ─── Функции-обработчики ───────────────────────────────────────────
func reserveStandard() { fmt.Println("reserve standard") }
func cancelStandard()  { fmt.Println("cancel standard") }
func reserveVIP()      { fmt.Println("reserve VIP") }
func cancelVIP()       { fmt.Println("cancel VIP") }

func main() {
	// 1. Таблица переходов (map[AlignedKey]func)
	table := map[cachealign.AlignedKey]func(){
		cachealign.MakeKey(TypeStandard, StateCreated, EventReserve): reserveStandard,
		cachealign.MakeKey(TypeStandard, StateCreated, EventCancel):  cancelStandard,
		cachealign.MakeKey(TypeVIP, StateCreated, EventReserve):      reserveVIP,
		cachealign.MakeKey(TypeVIP, StateCreated, EventCancel):       cancelVIP,
	}

	// 2. Создаём батч с ёмкостью, кратной 16 ключам (16 * 4 = 64 байта)
	batch := cachealign.NewAlignedBatch(100)

	// 3. Имитируем поток событий: добавляем ключи в батч
	batch.Append(cachealign.MakeKey(TypeStandard, StateCreated, EventReserve))
	batch.Append(cachealign.MakeKey(TypeVIP, StateCreated, EventCancel))
	batch.Append(cachealign.MakeKey(TypeStandard, StateCreated, EventCancel))
	batch.Append(cachealign.MakeKey(TypeVIP, StateCreated, EventReserve))

	// 4. Обрабатываем батч: последовательный обход с минимальными cache‑misses
	fmt.Println("Processing batch...")
	batch.Process(func(key cachealign.AlignedKey) {
		if fn, ok := table[key]; ok {
			fn()
		} else {
			fmt.Println("unknown transition")
		}
	})
}
