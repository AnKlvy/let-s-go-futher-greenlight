package data

import (
	"greenlight.andreyklimov.net/internal/validator"
	"math"
	"strings"
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

// Определяем новую структуру Metadata для хранения метаданных пагинации.
type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

// Функция calculateMetadata() вычисляет соответствующие метаданные пагинации
// на основе общего количества записей, текущей страницы и размера страницы.
// Обратите внимание, что значение последней страницы вычисляется с помощью
// функции math.Ceil(), которая округляет число вверх до ближайшего целого.
// Например, если всего 12 записей и размер страницы 5, то последняя страница
// будет вычисляться как math.Ceil(12/5) = 3.
func calculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		// Если записей нет, возвращаем пустую структуру Metadata.
		return Metadata{}
	}
	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}
