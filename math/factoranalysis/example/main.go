// Пример использования пакета factoranalysis.
//
// Задача: у нас есть данные о клиентах интернет-магазина (траты, частота покупок,
// время на сайте). Мы хотим сжать эти три признака в два главных компонента,
// чтобы визуализировать сегменты клиентов на плоскости.
package main

import (
	"fmt"

	"github.com/system-highload-architect/go-solutions/math/factoranalysis"
)

func main() {
	// Данные: 5 клиентов, 3 признака (траты, частота, время).
	data := [][]float64{
		{100, 5, 30},
		{120, 7, 40},
		{80, 3, 20},
		{200, 10, 60},
		{150, 8, 50},
	}

	// Обучаем PCA, оставляем 2 главные компоненты.
	pca, err := factoranalysis.TrainPCA(data, 2)
	if err != nil {
		panic(err)
	}

	// Смотрим, сколько дисперсии объясняют компоненты.
	ratios := pca.ExplainedVarianceRatio()
	fmt.Printf("Доля объяснённой дисперсии:\n")
	for i, r := range ratios {
		fmt.Printf("  Компонента %d: %.2f\n", i+1, r)
	}

	// Трансформируем исходные данные в 2D.
	transformed := pca.Transform(data)
	fmt.Println("\nКоординаты в пространстве главных компонент:")
	for i, row := range transformed {
		fmt.Printf("  Клиент %d: (%.2f, %.2f)\n", i+1, row[0], row[1])
	}

	// Восстанавливаем данные обратно (с потерями).
	reconstructed := pca.InverseTransform(transformed)
	fmt.Println("\nВосстановленные данные:")
	for i, row := range reconstructed {
		fmt.Printf("  Клиент %d: %.0f руб., %.0f покупок, %.0f мин.\n", i+1, row[0], row[1], row[2])
	}
}
