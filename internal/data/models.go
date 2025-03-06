package data

import (
	"database/sql"
	"errors"
)

// Определяем пользовательскую ошибку ErrRecordNotFound. Мы будем возвращать ее из метода Get(),
// если попытаемся найти фильм, которого нет в базе данных.
var (
	ErrRecordNotFound = errors.New("record not found")
)

// Создаем структуру Models, которая оборачивает MovieModel. Позже мы добавим сюда и другие модели,
// такие как UserModel и PermissionModel, по мере разработки.
type Models struct {
	Movies MovieModel
}

// Для удобства мы также добавляем метод New(), который возвращает структуру Models
// с инициализированным MovieModel.
func NewModels(db *sql.DB) Models {
	return Models{
		Movies: MovieModel{DB: db},
	}
}