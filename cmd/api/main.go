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

	_ "github.com/lib/pq"
	"greenlight.andreyklimov.net/internal/data"
	"greenlight.andreyklimov.net/internal/jsonlog"
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
	// Добавляем новую структуру limiter, содержащую поля для количества запросов в секунду,
	// максимального числа запросов в очереди (burst) и булево поле, которое можно использовать
	// для включения/отключения ограничения запросов.
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
}

// Измените поле logger, чтобы оно имело тип *jsonlog.Logger вместо *log.Logger.
type application struct {
	config config
	logger *jsonlog.Logger
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
	// Создаем флаги командной строки для чтения значений настроек в структуру config.
	// Обратите внимание, что по умолчанию для параметра 'enabled' установлено значение true.
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")
	flag.Parse()

	// Инициализируйте новый jsonlog.Logger, который записывает все сообщения
	// *уровня INFO и выше* в стандартный поток вывода.
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(cfg)
	if err != nil {
		// Используйте метод PrintFatal(), чтобы записать сообщение об ошибке
		// с уровнем FATAL и завершить работу. У нас нет дополнительных параметров
		// для включения в запись лога, поэтому мы передаем nil как второй параметр.
		logger.PrintFatal(err, nil)
	}
	defer db.Close()

	// Аналогично, используем метод PrintInfo() для записи сообщения уровня INFO.
	logger.PrintInfo("database connection pool established", nil)

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.port),
		Handler: app.routes(),
		// Создается новый экземпляр Go log.Logger с помощью log.New(),
		// передавая кастомный Logger в качестве первого параметра.
		// Пустая строка и 0 указывают, что экземпляр log.Logger
		// не должен использовать префикс или какие-либо флаги.
		ErrorLog:     log.New(logger, "", 0),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Снова используем метод PrintInfo() для записи сообщения "starting server"
	// на уровне INFO. Но на этот раз передаем карту с дополнительными параметрами
	// (операционная среда и адрес сервера) в качестве последнего параметра.
	logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env":  cfg.env,
	})

	err = srv.ListenAndServe()

	// Используйте метод PrintFatal() для логирования ошибки и завершения работы.
	logger.PrintFatal(err, nil)
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
