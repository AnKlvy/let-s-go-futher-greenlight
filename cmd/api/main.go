package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	"greenlight.andreyklimov.net/internal/data"
	"log"
	"net/http"
	"os"
	"time"
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

// Добавляем поле models для хранения нашей новой структуры Models.
type application struct {
	config config
	logger *log.Logger
	models data.Models
}

func main() {
	var cfg config
	flag.IntVar(&cfg.port, "port", 4000, "API server port")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgreSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()
	logger.Printf("database connection pool established")

	// Используем функцию data.NewModels() для инициализации структуры Models,
	// передавая в нее пул соединений в качестве параметра.
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
	err = srv.ListenAndServe()
	logger.Fatal(err)
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// Устанавливаем максимальное количество открытых (используемых + свободных) соединений в пуле.
	// Если передано значение меньше или равное 0, ограничение не устанавливается.
	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	// Устанавливаем максимальное количество свободных соединений в пуле.
	// Если передано значение меньше или равное 0, ограничение не устанавливается.
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	// Используем функцию time.ParseDuration() для преобразования строки с таймаутом простоя
	// в тип time.Duration.
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	// Устанавливаем максимальное время простоя соединений.
	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Проверяем соединение с базой данных.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}
	return db, nil
}
