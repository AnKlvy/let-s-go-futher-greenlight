package data

import (
	"database/sql" // Новый импорт
	"greenlight.andreyklimov.net/internal/validator"
	"time"
	"github.com/lib/pq" // New import

)

// Определяем структуру MovieModel, которая содержит пул соединений с базой данных.
type MovieModel struct {
	DB *sql.DB
}

// Метод Insert() принимает указатель на структуру Movie, которая должна содержать
// данные для новой записи.
func (m MovieModel) Insert(movie *Movie) error {
	// Определяем SQL-запрос для вставки новой записи в таблицу movies и получения
	// автоматически сгенерированных данных.
	query := `
	INSERT INTO movies (title, year, runtime, genres)
	VALUES ($1, $2, $3, $4)
	RETURNING id, created_at, version`
	// Создаем срез args, содержащий значения для параметров-заполнителей из
	// структуры movie. Объявление этого среза сразу рядом с нашим SQL-запросом помогает
	// сделать более понятным *какие значения используются и где* в запросе.
	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}
	// Используем метод QueryRow() для выполнения SQL-запроса с нашим пулом соединений,
	// передавая срез args как вариативный параметр и сканируем значения id, created_at и version
	// в структуру movie.
	return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}


// Добавляем заглушку метода для получения конкретного фильма по ID.
func (m MovieModel) Get(id int64) (*Movie, error) {
	return nil, nil
}

// Добавляем заглушку метода для обновления информации о фильме.
func (m MovieModel) Update(movie *Movie) error {
	return nil
}

// Добавляем заглушку метода для удаления фильма по ID.
func (m MovieModel) Delete(id int64) error {
	return nil
}

type MockMovieModel struct{}

func (m MockMovieModel) Insert(movie *Movie) error {
	return nil
}

func (m MockMovieModel) Get(id int64) (*Movie, error) {
	return nil, nil
}

func (m MockMovieModel) Update(movie *Movie) error {
	return nil
}

func (m MockMovieModel) Delete(id int64) error {
	return nil
}

type Movie struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Title     string    `json:"title"`
	Year      int32     `json:"year,omitempty"`
	Runtime   Runtime   `json:"runtime,omitempty"`
	Genres    []string  `json:"genres,omitempty"`
	Version   int32     `json:"version"`
}

// ValidateMovie выполняет валидацию данных фильма.
func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")
	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")
	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")
}
