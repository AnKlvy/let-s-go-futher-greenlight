package main

import (
	"errors"
	"fmt"
	"net/http"

	"greenlight.andreyklimov.net/internal/data"
	"greenlight.andreyklimov.net/internal/validator"
)

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Определяем структуру input для хранения входных данных JSON.
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	// Считываем JSON-запрос и записываем данные в структуру input.
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Создаем структуру Movie и заполняем ее значениями из input.
	// Обратите внимание, что переменная movie является указателем на структуру Movie.
	movie := &data.Movie{
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	// Создаем новый валидатор и проверяем корректность данных.
	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Вызываем метод Insert() у модели movies, передавая указатель на валидированную структуру movie.
	// Этот метод создаст запись в базе данных и обновит структуру movie сгенерированными значениями.
	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// При отправке HTTP-ответа мы добавляем заголовок Location, указывая клиенту URL-адрес
	// созданного ресурса. Для этого создаем пустой map http.Header и устанавливаем Location.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	// Отправляем JSON-ответ с кодом 201 Created, включая в тело ответа данные о фильме
	// и заголовок Location.
	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Call the Get() method to fetch the data for a specific movie. We also need to
	// use the errors.Is() function to check if it returns a data.ErrRecordNotFound
	// error, in which case we send a 404 Not Found response to the client.
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Извлекаем идентификатор фильма из URL.
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Получаем существующую запись фильма из базы данных. Если запись не найдена,
	// отправляем клиенту ответ с кодом 404 Not Found.
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Объявляем структуру input для хранения ожидаемых данных от клиента.
	var input struct {
		Title   string        `json:"title"`
		Year    int32         `json:"year"`
		Runtime data.Runtime  `json:"runtime"`
		Genres  []string      `json:"genres"`
	}

	// Считываем JSON-данные из тела запроса в структуру input.
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Копируем значения из тела запроса в соответствующие поля структуры movie.
	movie.Title = input.Title
	movie.Year = input.Year
	movie.Runtime = input.Runtime
	movie.Genres = input.Genres

	// Проверяем обновленные данные фильма, отправляем клиенту ответ 422 Unprocessable Entity,
	// если проверка не пройдена.
	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Передаем обновленную запись фильма в метод Update().
	err = app.models.Movies.Update(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Отправляем обновленную запись фильма в JSON-ответе.
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Извлекаем ID фильма из URL.
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Удаляем фильм из базы данных, отправляя клиенту ответ 404 Not Found,
	// если соответствующая запись не найдена.
	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Возвращаем статус 200 OK вместе с сообщением об успешном удалении.
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "фильм успешно удален"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
