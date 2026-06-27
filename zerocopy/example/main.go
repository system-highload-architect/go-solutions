// Пример использования пакета zerocopy.
//
// Задача: написать обработчик HTTP‑запросов, который принимает JSON‑RPC,
// извлекает поле "method" без полного разбора тела и без лишних выделений
// памяти. Это типичный сценарий для API‑шлюза или прокси.
package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/system-highload-architect/go-solutions/zerocopy"
)

func main() {
	http.HandleFunc("/rpc", handleRPC)
	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", nil)
}

func handleRPC(w http.ResponseWriter, r *http.Request) {
	// 1. Берём буфер из пула (не аллоцируем новый).
	bufPtr := zerocopy.GetBytes()
	buf := *bufPtr
	defer zerocopy.PutBytes(bufPtr) // вернём буфер в пул после обработки

	// 2. Читаем тело запроса в этот буфер.
	for {
		if len(buf) == cap(buf) {
			newBuf := make([]byte, len(buf), 2*cap(buf))
			copy(newBuf, buf)
			buf = newBuf
		}
		n, err := r.Body.Read(buf[len(buf):cap(buf)])
		buf = buf[:len(buf)+n]
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, "cannot read body", http.StatusBadRequest)
			return
		}
	}
	*bufPtr = buf // обновляем указатель в пуле

	// 3. Извлекаем поле "method" без парсинга всего JSON.
	methodBytes, ok := zerocopy.GetJSONField(buf, "method")
	if !ok {
		http.Error(w, "missing method", http.StatusBadRequest)
		return
	}
	// methodBytes содержит строку в кавычках, например `"auction.bid"`.
	// Превращаем в строку без копирования:
	methodStr := zerocopy.BytesToString(methodBytes[1 : len(methodBytes)-1])

	// 4. Формируем ответ, используя тот же пул для тела ответа.
	respBufPtr := zerocopy.GetBytes()
	respBuf := *respBufPtr
	defer zerocopy.PutBytes(respBufPtr)

	respBuf = append(respBuf, `{"jsonrpc":"2.0","result":"processed `...)
	respBuf = append(respBuf, zerocopy.StringToBytes(methodStr)...)
	respBuf = append(respBuf, `"}`...)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respBuf)
	*respBufPtr = respBuf[:0] // сброс для пула
}
