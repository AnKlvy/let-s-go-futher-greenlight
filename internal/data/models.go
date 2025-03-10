package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict = errors.New("edit conflict")
)

type Models struct {
	// Устанавливаем поле Movies как интерфейс, содержащий методы, которые должны поддерживать
	// как 'реальная' модель, так и мок-модель.
	Movies interface {
		Insert(movie *Movie) error
		Get(id int64) (*Movie, error)
		Update(movie *Movie) error
		Delete(id int64) error
		GetAll (title string, genres []string, filters Filters) ([]*Movie, error)
	}
}

// Создаем вспомогательную функцию, которая возвращает экземпляр Models, содержащий только мок-модели.
func NewMockModels() Models {
	return Models{
		Movies: MockMovieModel{},
	}
}

// Для удобства мы также добавляем метод New(), который возвращает структуру Models
// с инициализированным MovieModel.
func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
	}
}
