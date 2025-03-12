package main

import (
	"fmt"
	"net/http"
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
