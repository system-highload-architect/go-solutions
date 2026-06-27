// Пример использования пакета fixedpoint.
//
// Задача: интернет-магазин должен точно посчитать сумму корзины,
// применить скидку 15% и сконвертировать итог в доллары по курсу 0.011.
// Все расчёты ведутся в копейках, без float64, чтобы избежать ошибок округления.
package main

import (
	"fmt"

	"github.com/system-highload-architect/go-solutions/math/fixedpoint"
)

func main() {
	// Создаём цены товаров в копейках.
	apple := fixedpoint.New(8990, 2)   // 89.90 руб
	banana := fixedpoint.New(5490, 2)  // 54.90 руб
	cherry := fixedpoint.New(12990, 2) // 129.90 руб

	// Суммируем (масштабы одинаковые).
	total, err := fixedpoint.Sum([]fixedpoint.Money{apple, banana, cherry})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Сумма корзины: %s руб.\n", total.String()) // 274.70 руб.

	// Применяем скидку 15% (умножение на float с округлением).
	discount := total.MulFloat(0.85)
	fmt.Printf("После скидки 15%%: %s руб.\n", discount.String()) // 233.50 руб.

	// Конвертируем в доллары (scale=2) по курсу 0.011.
	usd := discount.Convert(0.011, 2)
	fmt.Printf("В долларах: %s USD\n", usd.String()) // 2.57 USD

	// Проверяем переполнение (для демонстрации).
	safe, err := total.AddChecked(fixedpoint.New(1000000000000000, 2))
	if err != nil {
		fmt.Println("Переполнение при сложении:", err)
	} else {
		fmt.Println("Сумма с большим числом:", safe.String())
	}
}
