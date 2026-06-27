// Пример использования пакета geoip.
//
// Задача: веб‑сервис определяет страну и город посетителя по его IP‑адресу,
// чтобы адаптировать контент (язык, валюта) и записывать географию в логи.
// Используется база MaxMind GeoLite2 (файл GeoLite2‑City.mmdb).
//
// Примечание: для работы примера необходимо скачать базу GeoLite2‑City.mmdb
// с сайта MaxMind (бесплатно, требуется регистрация) и положить в папку с примером.
package main

import (
	"fmt"
	"os"

	"github.com/system-highload-architect/go-solutions/security/geoip"
)

func main() {
	// Путь к файлу базы GeoLite2‑City.
	dbPath := "GeoLite2-City.mmdb"
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		fmt.Printf("Файл базы %q не найден.\n", dbPath)
		fmt.Println("Скачайте GeoLite2‑City.mmdb с https://dev.maxmind.com/geoip/geolite2-free-geolocation-data")
		return
	}

	// Открываем базу.
	db, err := geoip.New(dbPath)
	if err != nil {
		fmt.Println("Ошибка открытия базы:", err)
		return
	}
	defer db.Close()

	// Примеры IP‑адресов.
	ips := []string{
		"8.8.8.8",       // Google DNS (США)
		"77.88.55.80",   // Яндекс (Россия)
		"151.101.1.140", // Fastly (Европа)
	}

	for _, ip := range ips {
		result, err := db.Lookup(ip)
		if err != nil {
			fmt.Printf("IP %s: ошибка — %v\n", ip, err)
			continue
		}
		fmt.Printf("IP %s:\n", ip)
		fmt.Printf("  Страна: %s\n", result.Country)
		if result.City != "" {
			fmt.Printf("  Город:  %s\n", result.City)
		}
		fmt.Printf("  Координаты: %.4f, %.4f\n", result.Lat, result.Lng)
	}
}
