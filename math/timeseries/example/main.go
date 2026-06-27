// Пример использования пакета timeseries.
//
// Задача: сервис доставки еды хочет спрогнозировать количество заказов
// на следующие 3 дня на основе данных за последние 2 недели.
// Используется тройное экспоненциальное сглаживание Хольт‑Уинтерса.
package main

import (
	"fmt"

	"github.com/system-highload-architect/go-solutions/math/timeseries"
)

func main() {
	// Количество заказов за последние 14 дней (ежедневные данные).
	orders := []float64{
		120, 135, 148, 155, 162, 170, 175,
		180, 178, 185, 190, 195, 200, 210,
	}

	// Параметры модели (можно подобрать автоматически, здесь заданы вручную).
	params := timeseries.HoltWintersParams{
		Alpha:  0.5, // сглаживание уровня
		Beta:   0.3, // сглаживание тренда
		Gamma:  0.2, // сглаживание сезонности
		Period: 7,   // недельная сезонность
	}

	// Прогноз на 3 дня вперёд.
	horizon := 3
	forecast, err := timeseries.HoltWintersForecast(orders, horizon, params)
	if err != nil {
		fmt.Println("Ошибка прогноза:", err)
		return
	}

	fmt.Printf("Прогноз заказов на %d дня(ей):\n", horizon)
	for i, v := range forecast {
		fmt.Printf("  День %d: %.0f заказов\n", i+1, v)
	}

	// Также вычислим скользящее среднее для сглаживания графика.
	sma := timeseries.SimpleMovingAverage(orders, 3)
	fmt.Printf("\n3‑дневное скользящее среднее (последние 5 значений):\n")
	for i := len(sma) - 5; i < len(sma); i++ {
		fmt.Printf("  %.0f\n", sma[i])
	}
}
