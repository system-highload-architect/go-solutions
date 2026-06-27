// Пример использования пакета radixsort.
//
// Задача: RTB-аукцион получил список ставок от рекламодателей.
// Нужно быстро отсортировать ставки по возрастанию и одновременно
// запомнить исходные позиции для определения победителя.
package main

import (
	"fmt"

	"github.com/system-highload-architect/go-solutions/math/radixsort"
)

func main() {
	// Ставки в копейках от разных рекламодателей.
	bids := []int64{150, 200, 100, 300, 250}
	indices := make([]int, len(bids))
	for i := range indices {
		indices[i] = i
	}

	fmt.Println("До сортировки:")
	for i, bid := range bids {
		fmt.Printf("  Рекламодатель %d: %d коп.\n", i, bid)
	}

	// Сортируем ставки и синхронно переставляем индексы.
	radixsort.SortInt64WithIndices(bids, indices)

	fmt.Println("\nПосле сортировки (по возрастанию):")
	for i, bid := range bids {
		fmt.Printf("  Позиция %d: %d коп. (рекламодатель %d)\n", i, bid, indices[i])
	}

	// Победитель — последний (максимальная ставка).
	winnerIdx := indices[len(indices)-1]
	fmt.Printf("\nПобедитель: рекламодатель %d со ставкой %d коп.\n", winnerIdx, bids[len(bids)-1])
}
