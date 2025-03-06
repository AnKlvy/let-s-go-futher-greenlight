package main

import (
	"context"     
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	// Импортируем драйвер pq, чтобы он зарегистрировался в database/sql.
	// Используем пустой идентификатор `_`, чтобы компилятор Go не жаловался на неиспользуемый импорт.
	_ "github.com/lib/pq"
)

const version = "1.0.0"

// Определяем структуру config для хранения настроек приложения.
type config struct {
	port int
	env  string
	db   struct {
		dsn string
	}
}

// Определяем структуру application, которая будет содержать конфигурацию и логгер.
type application struct {
	config config
	logger *log.Logger
}

func main() {
	// Создаём переменную конфигурации cfg.
	var cfg config
	
	// Читаем параметры командной строки (флаги).
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	// Читаем DSN для базы данных из командной строки, если флаг не указан, используем значение по умолчанию.
	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://greenlight:pa55word@localhost/greenlight?sslmode=disable", "PostgreSQL DSN")
	flag.Parse()
	
	// Создаём логгер, который будет писать в stdout.
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	
	// Открываем соединение с базой данных через openDB().
	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err) // Завершаем работу, если не удалось подключиться к базе.
	}
	defer db.Close() // Закрываем соединение при завершении работы.

	logger.Printf("database connection pool established") // Логируем успешное подключение.

	// Создаём экземпляр application.
	app := &application{
		config: cfg,
		logger: logger,
	}

	// Настраиваем HTTP-сервер.
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port), // Указываем порт из конфигурации.
		Handler:      app.routes(), // Подключаем маршруты (функция будет определена позже).
		IdleTimeout:  time.Minute, // Максимальное время простоя соединения.
		ReadTimeout:  10 * time.Second, // Таймаут на чтение запроса.
		WriteTimeout: 30 * time.Second, // Таймаут на отправку ответа.
	}

	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)

	// Запускаем сервер и, если возникнет ошибка, логируем её.
	err = srv.ListenAndServe()
	logger.Fatal(err)
}

// Функция openDB() создаёт пул подключений к базе данных.
func openDB(cfg config) (*sql.DB, error) {
	// Открываем соединение с базой данных через sql.Open().
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// Создаём контекст с таймаутом 5 секунд.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Проверяем соединение с базой через PingContext().
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	// Возвращаем пул подключений.
	return db, nil
}
