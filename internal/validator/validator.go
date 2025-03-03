package validator

import (
	"regexp"
)

// Объявляем регулярное выражение для проверки формата email-адресов (мы будем
// использовать его позже в книге). Если вам интересно, этот шаблон регулярного
// выражения взят с https://html.spec.whatwg.org/#valid-e-mail-address. Примечание:
// если вы читаете это в формате PDF или EPUB и не видите полный шаблон, см. примечание
// далее на странице.
var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

// Определяем новый тип Validator, который содержит карту ошибок валидации.
type Validator struct {
	Errors map[string]string
}

// New — вспомогательная функция, которая создаёт новый экземпляр Validator с пустой картой ошибок.
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid возвращает true, если в карте ошибок нет записей.
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError добавляет сообщение об ошибке в карту (только если для данного ключа еще нет записи).
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Check добавляет сообщение об ошибке в карту только в том случае, если проверка валидации не 'ok'.
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// Универсальная функция, которая возвращает true, если указанное значение присутствует в списке.
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	for i := range permittedValues {
		if value == permittedValues[i] {
			return true
		}
	}
	return false
}

// Matches возвращает true, если строковое значение соответствует указанному шаблону регулярного выражения.
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// Универсальная функция, которая возвращает true, если все значения в срезе уникальны.
func Unique[T comparable](values []T) bool {
	uniqueValues := make(map[T]bool)
	for _, value := range values {
		uniqueValues[value] = true
	}
	return len(values) == len(uniqueValues)
}
