package data

import (
	"greenlight.andreyklimov.net/internal/validator"
)

// Добавляем поле SortSafelist для хранения поддерживаемых значений сортировки.
type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafelist []string
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
