package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

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
	// Определяем структуру client, которая будет содержать ограничитель скорости и время последней активности для каждого клиента.
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu sync.Mutex
		// Обновляем карту так, чтобы значения были указателями на структуру client.
		clients = make(map[string]*client)
	)

	// Запускаем фоновую горутину, которая раз в минуту удаляет старые записи из карты clients.
	go func() {
		for {
			time.Sleep(time.Minute)
			// Блокируем мьютекс, чтобы предотвратить выполнение проверок ограничителя скорости во время очистки.
			mu.Lock()
			// Проходим по всем клиентам. Если клиент не был активен в течение последних трех минут, удаляем его из карты.
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			// Важно разблокировать мьютекс после завершения очистки.
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Выполняем проверку только в том случае, если ограничение запросов включено.
		if app.config.limiter.enabled {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}
			mu.Lock()
			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					// Используем значения количества запросов в секунду и burst из структуры config.
					limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
				}
			}
			clients[ip].lastSeen = time.Now()
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}
			mu.Unlock()
		}
		next.ServeHTTP(w, r)
	})
}
