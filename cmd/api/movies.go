package main

import (
	"fmt"
	"net/http"
	"time"

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
		// Use the new notFoundResponse() helper.
		app.notFoundResponse(w, r)
		return
	}
	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Casablanca",
		Runtime:   102,
		Genres:    []string{"drama", "romance", "war"},
		Version:   1,
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		// Use the new serverErrorResponse() helper.
		app.serverErrorResponse(w, r, err)
	}
}
