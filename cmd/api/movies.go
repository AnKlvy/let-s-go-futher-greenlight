package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

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

	// Получаем запись о фильме как обычно.
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

	// Если в запросе присутствует заголовок X-Expected-Version,
	// проверяем, что версия фильма в базе данных совпадает с указанной в заголовке.
	if r.Header.Get("X-Expected-Version") != "" {
		if strconv.FormatInt(int64(movie.Version), 32) != r.Header.Get("X-Expected-Version") {
			app.editConflictResponse(w, r)
			return
		}
	}

	// Используем указатели для полей Title, Year и Runtime.
	var input struct {
		Title   *string       `json:"title"`
		Year    *int32        `json:"year"`
		Runtime *data.Runtime `json:"runtime"`
		Genres  []string      `json:"genres"`
	}

	// Декодируем JSON как обычно.
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Если значение input.Title равно nil, значит, в JSON-запросе не было передано
	// соответствующей пары "ключ-значение" для "title". В этом случае оставляем
	// запись о фильме без изменений. В противном случае обновляем значение title.
	// Важно: так как input.Title теперь является указателем на строку, перед
	// присвоением значения в структуру фильма необходимо разыменовать указатель
	// с помощью оператора *.
	if input.Title != nil {
		movie.Title = *input.Title
	}

	// Аналогично обновляем остальные поля в структуре input.
	if input.Year != nil {
		movie.Year = *input.Year
	}
	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}
	if input.Genres != nil {
		movie.Genres = input.Genres // Для срезов разыменование не требуется.
	}

	// Валидируем обновлённую запись фильма.
	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Перехватываем ошибку ErrEditConflict и вызываем новый вспомогательный метод
	// editConflictResponse().
	err = app.models.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
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
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "movie deleted successfuly"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string
		Genres []string
		data.Filters
	}
	v := validator.New()
	qs := r.URL.Query()
	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	// Добавляем поддерживаемые значения сортировки для этого эндпоинта в safelist.
	input.Filters.SortSafelist = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}
	// Выполняем проверки валидации структуры Filters и отправляем ответ с ошибками, если необходимо.
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	fmt.Fprintf(w, "%+v\n", input)
}
