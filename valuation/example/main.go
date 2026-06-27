// Пример использования пакета valuation.
//
// Задача: RTB-система должна оценить ставку для показа рекламы.
// Даны координаты пользователя и билборда, признаки пользователя,
// базовая ставка. Скорер вычисляет ценность показа и оптимальную ставку.
package main

import (
	"fmt"

	"github.com/system-highload-architect/go-solutions/geo/geospatial"
	"github.com/system-highload-architect/go-solutions/math/fixedpoint"
	"github.com/system-highload-architect/go-solutions/valuation"
)

func main() {
	// Модель LTV
	ltv := valuation.NewLTVModel([]float64{0.5, 0.2, -0.1, 0.3})
	// Оценка ценности показа
	imp := valuation.NewImpressionValue(0.05, 0.1, 100.0)
	// Гео-фактор с затуханием 500 м
	gf := valuation.NewGeoFactor(500.0)
	// Win-rate модель (опционально)
	wrm := valuation.NewWinRateModel(-1.5, 0.001)

	scorer := valuation.NewScorer(ltv, imp, gf, wrm)

	userPos := geospatial.Point{Lat: 55.7558, Lng: 37.6173}
	targetPos := geospatial.Point{Lat: 55.7600, Lng: 37.6200}
	features := []float64{0.8, 0.2, -0.5, 0.7}
	baseBid := fixedpoint.New(150, 2) // 1.50 руб

	score, optBid, err := scorer.Score(userPos, targetPos, features, baseBid)
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}
	fmt.Printf("Ценность показа: %.4f\n", score)
	fmt.Printf("Оптимальная ставка: %s руб.\n", optBid.String())
}
