// Пример использования пакета regression.
//
// Задача: предсказать стоимость дома по его площади с помощью линейной регрессии.
// Даны исторические данные, мы обучаем модель и делаем прогноз для нового дома.
package main

import (
	"fmt"

	"github.com/system-highload-architect/go-solutions/math/regression"
)

func main() {
	// Исторические данные: площадь (м²) и цена (тыс. руб.)
	areas := [][]float64{{30}, {40}, {50}, {60}, {70}, {80}}
	prices := []float64{3000, 3500, 4200, 4800, 5500, 6100}

	// Обучаем линейную модель
	model, err := regression.TrainLinear(areas, prices)
	if err != nil {
		panic(err)
	}

	// Прогноз для дома площадью 55 м²
	newArea := []float64{55}
	predicted := model.Predict(newArea)
	fmt.Printf("Прогнозируемая цена: %.0f тыс. руб.\n", predicted)

	// Логистическая регрессия: предсказание вероятности клика по объявлению
	clicks := [][]float64{
		{0.1, 0.5}, // маленькая картинка, низкая релевантность
		{0.9, 0.8}, // большая картинка, высокая релевантность
		{0.5, 0.3},
		{0.8, 0.9},
	}
	labels := []float64{0, 1, 0, 1} // 0 — не кликнул, 1 — кликнул

	logModel, err := regression.TrainLogistic(clicks, labels, 0.1, 1000)
	if err != nil {
		panic(err)
	}

	newAd := []float64{0.7, 0.6}
	prob := logModel.PredictProb(newAd)
	fmt.Printf("Вероятность клика: %.2f\n", prob)
}
