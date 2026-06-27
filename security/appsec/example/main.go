// Пример использования пакета appsec.
//
// Задача: веб-приложение должно безопасно обрабатывать пользовательский ввод:
// - предотвращать открытые редиректы,
// - очищать HTML от опасных тегов (XSS),
// - подписывать URL-параметры для защиты от подделки,
// - проверять идентификаторы и числовые значения.
//
// Типичный сценарий: после успешного входа пользователя перенаправляют на
// страницу, указанную в параметре `return_url`. Сервер должен проверить,
// что URL ведёт только на доверенный домен, прежде чем выполнить редирект.
// Одновременно отображаемое имя пользователя проходит через санитизацию,
// чтобы исключить XSS. Ссылки на конфиденциальные операции (сброс пароля)
// подписываются HMAC, чтобы злоумышленник не мог подменить параметры.
package main

import (
	"fmt"
	"log"

	"github.com/system-highload-architect/go-solutions/security/appsec"
)

func main() {
	// 1. Безопасный редирект.
	redirectURL := "https://evil.com/phishing"
	allowedHosts := []string{"example.com", "www.example.com"}
	safe, err := appsec.SafeRedirect(redirectURL, allowedHosts)
	if err != nil {
		fmt.Printf("Редирект на %q заблокирован: %v\n", redirectURL, err)
	} else {
		fmt.Printf("Редирект разрешён: %s\n", safe)
	}

	// Разрешённый относительный редирект.
	relative := "/dashboard"
	safeRel, err := appsec.SafeRedirect(relative, allowedHosts)
	if err != nil {
		fmt.Printf("Относительный редирект заблокирован: %v\n", err)
	} else {
		fmt.Printf("Относительный редирект разрешён: %s\n", safeRel)
	}

	// 2. Санитизация HTML (защита от XSS).
	dirtyHTML := `<p>Привет, <b>Мир</b>!</p><script>alert('xss')</script>`
	cleanHTML := appsec.SanitizeHTML(dirtyHTML)
	fmt.Printf("До санитизации: %s\n", dirtyHTML)
	fmt.Printf("После санитизации: %s\n", cleanHTML)

	// 3. Подпись URL-параметров (HMAC).
	baseURL := "https://example.com/reset-password"
	params := map[string]string{
		"user_id": "42",
		"token":   "abc123",
	}
	secret := []byte("super-secret-key")
	signedURL, err := appsec.SignURLParams(baseURL, params, secret)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Подписанный URL: %s\n", signedURL)

	// Проверка подписи.
	verifiedParams, err := appsec.VerifyURLParams(signedURL, secret)
	if err != nil {
		fmt.Printf("Подпись недействительна: %v\n", err)
	} else {
		fmt.Println("Подпись верна, параметры:")
		for key, values := range verifiedParams {
			fmt.Printf("  %s = %s\n", key, values[0])
		}
	}

	// Попытка подменить параметр (без подписи).
	tamperedURL := "https://example.com/reset-password?user_id=99&token=abc123&sig=invalidsig"
	_, err = appsec.VerifyURLParams(tamperedURL, secret)
	if err != nil {
		fmt.Printf("Поддельный URL отклонён: %v\n", err)
	}

	// 4. Валидация пользовательского ввода.
	fmt.Printf("ValidID 'alice_123' -> %v\n", appsec.ValidID("alice_123"))
	fmt.Printf("ValidID '<script>' -> %v\n", appsec.ValidID("<script>"))
	fmt.Printf("ValidNumber '8080' -> %v\n", appsec.ValidNumber("8080"))
	fmt.Printf("ValidNumber '80a80' -> %v\n", appsec.ValidNumber("80a80"))
}
