package jsonlog

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

// Определяем тип Level для представления уровня серьезности записи в журнале.
type Level int8

// Инициализируем константы, представляющие уровни серьезности. Используем iota
// как сокращение для присвоения последовательных целочисленных значений.
const (
	LevelInfo  Level = iota // Значение 0.
	LevelError              // Значение 1.
	LevelFatal              // Значение 2.
	LevelOff                // Значение 3.
)

// Возвращаем удобочитаемое строковое представление уровня серьезности.
func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""
	}
}

// Определяем собственный тип Logger. Он хранит выходное место назначения для записей,
// минимальный уровень серьезности, а также мьютекс для синхронизации записей.
type Logger struct {
	out      io.Writer
	minLevel Level
	mu       sync.Mutex
}

// Возвращаем новый экземпляр Logger, который записывает записи в журнал при уровне
// серьезности не ниже указанного.
func New(out io.Writer, minLevel Level) *Logger {
	return &Logger{
		out:      out,
		minLevel: minLevel,
	}
}

// Вспомогательные методы для записи логов с разными уровнями серьезности.
// В качестве второго параметра принимают карту с произвольными "свойствами",
// которые будут добавлены в запись лога.
func (l *Logger) PrintInfo(message string, properties map[string]string) {
	l.print(LevelInfo, message, properties)
}

func (l *Logger) PrintError(err error, properties map[string]string) {
	l.print(LevelError, err.Error(), properties)
}

func (l *Logger) PrintFatal(err error, properties map[string]string) {
	l.print(LevelFatal, err.Error(), properties)
	os.Exit(1) // При уровне FATAL также завершаем выполнение приложения.
}

// Внутренний метод print для записи логов.
func (l *Logger) print(level Level, message string, properties map[string]string) (int, error) {
	// Если уровень серьезности ниже минимального уровня логирования, просто выходим.
	if level < l.minLevel {
		return 0, nil
	}

	// Анонимная структура для хранения данных записи лога.
	aux := struct {
		Level      string            `json:"level"`
		Time       string            `json:"time"`
		Message    string            `json:"message"`
		Properties map[string]string `json:"properties,omitempty"`
		Trace      string            `json:"trace,omitempty"`
	}{
		Level:      level.String(),
		Time:       time.Now().UTC().Format(time.RFC3339),
		Message:    message,
		Properties: properties,
	}

	// Включаем стек вызовов для уровней ERROR и FATAL.
	if level >= LevelError {
		aux.Trace = string(debug.Stack())
	}

	// Переменная для хранения JSON-записи.
	var line []byte

	// Кодируем структуру в JSON. Если произошла ошибка, записываем текст ошибки.
	line, err := json.Marshal(aux)
	if err != nil {
		line = []byte(LevelError.String() + ": unable to marshal log message: " + err.Error())
	}

	// Блокируем мьютекс, чтобы избежать одновременной записи нескольких потоков.
	l.mu.Lock()
	defer l.mu.Unlock()

	// Записываем лог и добавляем перевод строки.
	return l.out.Write(append(line, '\n'))
}

// Реализуем метод Write() для соответствия интерфейсу io.Writer.
// Записывает запись с уровнем ERROR без дополнительных свойств.
func (l *Logger) Write(message []byte) (n int, err error) {
	return l.print(LevelError, string(message), nil)
}
