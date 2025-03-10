package data

import (
	"context"
	"database/sql"
	"errors"
	"github.com/lib/pq"
	"greenlight.andreyklimov.net/internal/validator"
	"time"
)

// Определяем структуру MovieModel, которая содержит пул соединений с базой данных.
type MovieModel struct {
	DB *sql.DB
}

func (m MovieModel) Insert(movie *Movie) error {
	query := `
    INSERT INTO movies (title, year, runtime, genres)
    VALUES ($1, $2, $3, $4)
    RETURNING id, created_at, version`
	args := []any{movie.Title, movie.Year, movie.Runtime, pq.Array(movie.Genres)}

	// Создаём контекст с тайм-аутом 3 секунды.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Используем QueryRowContext() и передаём контекст в качестве первого аргумента.
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

func (m MovieModel) Get(id int64) (*Movie, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	// Удаляем конструкцию pg_sleep(10).
	query := `
    SELECT id, created_at, title, year, runtime, genres, version
    FROM movies
    WHERE id = $1`

	var movie Movie
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Убираем &[]byte{} из первого аргумента Scan().
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&movie.ID,
		&movie.CreatedAt,
		&movie.Title,
		&movie.Year,
		&movie.Runtime,
		pq.Array(&movie.Genres),
		&movie.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &movie, nil
}

func (m MovieModel) Update(movie *Movie) error {
	query := `
    UPDATE movies
    SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
    WHERE id = $5 AND version = $6
    RETURNING version`
	args := []any{
		movie.Title,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		movie.ID,
		movie.Version,
	}

	// Создаём контекст с тайм-аутом 3 секунды.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Используем QueryRowContext() и передаём контекст в качестве первого аргумента.
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m MovieModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `
    DELETE FROM movies
    WHERE id = $1`

	// Создаём контекст с тайм-аутом 3 секунды.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Используем ExecContext() и передаём контекст в качестве первого аргумента.
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// Создаем новый метод GetAll(), который возвращает срез фильмов. Хотя мы
// пока не используем их, мы настроили его так, чтобы он принимал различные параметры фильтрации.
func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, error) {
	// Формируем SQL-запрос для получения всех записей о фильмах.
	query := `
		SELECT id, created_at, title, year, runtime, genres, version
		FROM movies
		ORDER BY id`

	// Создаем контекст с тайм-аутом в 3 секунды.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Используем QueryContext() для выполнения запроса. Это возвращает sql.Rows с результатами.
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	// Важно: откладываем вызов rows.Close(), чтобы убедиться, что resultset будет закрыт перед выходом из GetAll().
	defer rows.Close()

	// Инициализируем пустой срез для хранения данных о фильмах.
	movies := []*Movie{}

	// Используем rows.Next для перебора строк в результате запроса.
	for rows.Next() {
		// Инициализируем пустую структуру Movie для хранения данных об отдельном фильме.
		var movie Movie

		// Считываем значения из строки в структуру Movie. Обратите внимание, что
		// для поля genres мы используем адаптер pq.Array().
		err := rows.Scan(
			&movie.ID,
			&movie.CreatedAt,
			&movie.Title,
			&movie.Year,
			&movie.Runtime,
			pq.Array(&movie.Genres),
			&movie.Version,
		)
		if err != nil {
			return nil, err
		}

		// Добавляем структуру Movie в срез.
		movies = append(movies, &movie)
	}

	// После завершения итерации по rows.Next() вызываем rows.Err(),
	// чтобы получить любую ошибку, возникшую во время итерации.
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Если все прошло успешно, возвращаем срез фильмов.
	return movies, nil
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

func (m MockMovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, error) {
	return nil, nil
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
