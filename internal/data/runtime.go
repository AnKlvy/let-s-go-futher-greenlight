package data

import (
	"errors" // Новый импорт
	"fmt"
	"strconv"
	"strings" // Новый импорт
)

// Определяем ошибку, которую наш метод UnmarshalJSON() может вернуть, если мы не смогли успешно разобрать
// или преобразовать JSON-строку.
var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

// Объявляем пользовательский тип Runtime, который имеет базовый тип int32 (такой же, как и у нашего
// поля структуры Movie).
type Runtime int32

// Реализуем метод MarshalJSON() для типа Runtime, чтобы он удовлетворял интерфейсу json.Marshaler.
// Этот метод должен возвращать закодированное в JSON значение для времени фильма
// (в нашем случае он вернет строку в формате "<runtime> mins").
func (r Runtime) MarshalJSON() ([]byte, error) {
	// Формируем строку, содержащую время фильма в нужном формате.
	jsonValue := fmt.Sprintf("%d mins", r)
	// Используем функцию strconv.Quote() для обертывания строки в двойные кавычки.
	// Это необходимо, чтобы значение считалось корректной *JSON-строкой*.
	quotedJSONValue := strconv.Quote(jsonValue)
	// Преобразуем строку с кавычками в срез байтов и возвращаем его.
	return []byte(quotedJSONValue), nil
}

// Реализуем метод UnmarshalJSON() для типа Runtime, чтобы он удовлетворял интерфейсу json.Unmarshaler.
// ВАЖНО: Поскольку UnmarshalJSON() должен изменять получателя (наш тип Runtime), мы должны использовать
// указатель на получателя, чтобы это работало правильно. В противном случае мы будем изменять только копию
// (которая затем будет удалена после завершения работы метода).
func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {
	// Мы ожидаем, что входящее JSON-значение будет строкой в формате
	// "<runtime> mins", и первое, что нам нужно сделать, — это удалить окружающие
	// двойные кавычки из этой строки. Если мы не можем их убрать, то возвращаем
	// ошибку ErrInvalidRuntimeFormat.
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// Разделяем строку, чтобы выделить часть, содержащую число.
	parts := strings.Split(unquotedJSONValue, " ")

	// Проверяем, соответствует ли строка ожидаемому формату.
	// Если нет, возвращаем ошибку ErrInvalidRuntimeFormat.
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	// Преобразуем строку с числом в int32. Если это не удается, снова возвращаем
	// ошибку ErrInvalidRuntimeFormat.
	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// Преобразуем int32 в тип Runtime и присваиваем это получателю.
	// Обратите внимание, что мы используем оператор * для разыменования получателя
	// (который является указателем на тип Runtime), чтобы задать значение указателя.
	*r = Runtime(i)
	return nil
}
