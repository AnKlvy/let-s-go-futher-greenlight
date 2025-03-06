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

// Добавляем поля maxOpenConns, maxIdleConns и maxIdleTime для хранения
// параметров конфигурации пула подключений.
type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

// Определяем структуру application, которая будет содержать конфигурацию и логгер.
type application struct {
	config config
	logger *log.Logger
}

func main() {
	var cfg config
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")

	// Читаем настройки пула соединений из флагов командной строки.
	// Обрати внимание на используемые значения по умолчанию.
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")

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
		Handler:      app.routes(),                 // Подключаем маршруты (функция будет определена позже).
		IdleTimeout:  time.Minute,                  // Максимальное время простоя соединения.
		ReadTimeout:  10 * time.Second,             // Таймаут на чтение запроса.
		WriteTimeout: 30 * time.Second,             // Таймаут на отправку ответа.
	}

	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)

	// Запускаем сервер и, если возникнет ошибка, логируем её.
	err = srv.ListenAndServe()
	logger.Fatal(err)
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	// Set the maximum number of open (in-use + idle) connections in the pool. Note that
	// passing a value less than or equal to 0 will mean there is no limit.
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	// Set the maximum number of idle connections in the pool. Again, passing a value
	// less than or equal to 0 will mean there is no limit.
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	// Use the time.ParseDuration() function to convert the idle timeout duration string
	// to a time.Duration type.
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	// Set the maximum idle timeout.
	db.SetConnMaxIdleTime(duration)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}
