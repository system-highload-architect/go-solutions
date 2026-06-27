// Пример использования пакета device.
//
// Задача: сервер получает HTTP‑запросы с заголовком User‑Agent.
// Необходимо определить тип устройства, ОС и браузер для:
//   - адаптации интерфейса (мобильная / десктопная версия),
//   - сбора аналитики по устройствам пользователей,
//   - фильтрации ботов (исключать из статистики, применять другие лимиты).
//
// Пакет device выполняет парсинг без аллокаций, поэтому его можно
// вызывать на каждый запрос даже в высоконагруженных системах.
//
// Примечание: можно добавлять собственные правила через RegisterDevice,
// если стандартных классификаторов недостаточно.
package main

import (
	"fmt"

	"github.com/system-highload-architect/go-solutions/security/device"
)

func main() {
	// Примеры реальных User‑Agent.
	uas := []string{
		// Десктопный Chrome на Windows.
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		// Мобильный Safari на iPhone.
		"Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
		// Планшет на Android (содержит "Tablet").
		"Mozilla/5.0 (Linux; Android 13; SM‑T870) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36 Tablet",
		// Бот Googlebot.
		"Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)",
	}

	for _, ua := range uas {
		info := device.Parse(ua)
		fmt.Printf("User‑Agent: %s\n", ua)
		fmt.Printf("  Тип устройства: %s\n", typeToString(info.Type))
		fmt.Printf("  ОС:            %s\n", osToString(info.OS))
		fmt.Printf("  Браузер:       %s\n\n", browserToString(info.Browser))
	}

	// Пример расширения: добавляем своё устройство.
	device.RegisterDevice("SuperConsole", device.DeviceInfo{
		Type:    device.Desktop,
		OS:      device.Linux,
		Browser: device.Chrome,
	})
	customUA := "SuperConsole/1.0 (Linux; en‑US) Chrome/120.0.0.0"
	customInfo := device.Parse(customUA)
	fmt.Printf("Кастомный User‑Agent: %s\n", customUA)
	fmt.Printf("  Тип устройства: %s\n", typeToString(customInfo.Type))
}

func typeToString(t device.Type) string {
	switch t {
	case device.Desktop:
		return "Desktop"
	case device.Mobile:
		return "Mobile"
	case device.Tablet:
		return "Tablet"
	case device.Bot:
		return "Bot"
	default:
		return "Unknown"
	}
}

func osToString(o device.OS) string {
	switch o {
	case device.Windows:
		return "Windows"
	case device.MacOS:
		return "macOS"
	case device.Linux:
		return "Linux"
	case device.Android:
		return "Android"
	case device.IOS:
		return "iOS"
	default:
		return "Unknown"
	}
}

func browserToString(b device.Browser) string {
	switch b {
	case device.Chrome:
		return "Chrome"
	case device.Safari:
		return "Safari"
	case device.Firefox:
		return "Firefox"
	case device.Edge:
		return "Edge"
	case device.Opera:
		return "Opera"
	case device.IE:
		return "Internet Explorer"
	default:
		return "Unknown"
	}
}
