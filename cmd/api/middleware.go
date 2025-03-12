package main

import (
	"fmt"
	"net/http"
	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Определяется отложенная функция, которая всегда выполнится в случае паники,
		// поскольку Go разворачивает стек вызовов.
		defer func() {
			// Встроенная функция recover используется для проверки, произошла ли паника.
			if err := recover(); err != nil {
				// Если паника произошла, устанавливается заголовок "Connection: close" в ответе.
				// Это сигнализирует серверу Go о необходимости закрыть текущее соединение
				// после отправки ответа.
				w.Header().Set("Connection", "close")

				// Значение, возвращаемое recover(), имеет тип any, поэтому оно приводится
				// к error с помощью fmt.Errorf(), а затем передается в вспомогательную
				// функцию serverErrorResponse(). В свою очередь, она записывает ошибку
				// с уровнем ERROR в наш кастомный логгер и отправляет клиенту ответ
				// с кодом 500 Internal Server Error.
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	// Инициализируем новый ограничитель скорости, который разрешает в среднем 2 запроса в секунду,
	// с максимальным "всплеском" в 4 запроса.
	limiter := rate.NewLimiter(2, 4)

	// Функция, которую мы возвращаем, является замыканием, которое "захватывает" переменную limiter.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Вызываем limiter.Allow(), чтобы проверить, разрешён ли запрос. Если нет,
		// то вызываем вспомогательную функцию rateLimitExceededResponse(),
		// чтобы вернуть ответ 429 Too Many Requests (мы создадим эту функцию позже).
		if !limiter.Allow() {
			app.rateLimitExceededResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
