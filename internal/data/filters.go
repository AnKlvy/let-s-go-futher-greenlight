package data

import (
	"greenlight.andreyklimov.net/internal/validator"
	"strings" // Новый импорт
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
}

func (f Filters) limit() int {
	return f.PageSize
}
func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

// Проверяем, соответствует ли переданное значение Sort одному из допустимых значений,
// и если да, извлекаем имя столбца, удаляя ведущий знак минуса (если он есть).
func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafelist {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}
	panic("unsafe sort parameter: " + f.Sort)
}

// Возвращает направление сортировки ("ASC" или "DESC") в зависимости от
// префиксного символа в поле Sort.
func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}
	return "ASC"
}

func ValidateFilters(v *validator.Validator, f Filters) {
	// Проверяем, что параметры page и page_size содержат допустимые значения.
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "must be a maximum of 10 million")
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")

	// Проверяем, что параметр sort соответствует значению из safelist.
	v.Check(validator.PermittedValue(f.Sort, f.SortSafelist...), "sort", "invalid sort value")
}
