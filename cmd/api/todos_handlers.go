package main

import (
	"errors"
	"fmt"
	"net/http"

	validator "github.com/kvnloughead/godo/internal"
	"github.com/kvnloughead/godo/internal/data"
)

// listTodos handles GET requests to the /v1/todos endpoint.
//
// Various options for filtering, sorting, and pagination are available. See
// TodoModel.GetAll for details.
func (app *APIApplication) listTodos(w http.ResponseWriter, r *http.Request) {
	// input is an anonymous struct intended to store the query params for
	// filtering, sorting, and pagination.
	var input struct {
		Text     string
		Contexts []string
		Projects []string

		// TODO - implement additional query params for filtering.
		// Priority rune
		// Completed bool
		data.Filters
	}

	v := validator.New()
	qs := r.URL.Query()

	// Read query params into the input struct, setting reasonable defaults if
	// any are omitted, and validating the values that should be integers.
	input.Text = app.readQueryString(qs, "text", "")
	input.Contexts = app.readQueryCSV(qs, "contexts", []string{})
	input.Projects = app.readQueryCSV(qs, "projects", []string{})
	input.Filters.Page = app.readQueryInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readQueryInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readQueryString(qs, "sort", "id")
	input.Filters.SortSafelist = []string{"id", "text", "-id", "-text"}

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	data.ValidateFilters(v, input.Filters)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	todos, paginationData, err := app.Models.Todos.GetAll(
		input.Text,
		app.contextGetUser(r).ID,
		input.Contexts,
		input.Projects,
		input.Filters,
	)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(
		w,
		http.StatusOK,
		envelope{"todos": todos, "paginationData": paginationData},
		nil,
	)

	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// createTodo handles POST requests to the /v1/todos endpoint. The request
// body is decoded by the app.readJSON helper. See that function for details
// about error handling.
//
// Request bodies are validated by ValidateTodo. A failedValidationResponse
// error is sent if one or more fields fails validation.
func (app *APIApplication) createTodo(w http.ResponseWriter, r *http.Request) {
	// Struct to store the data from the response's body. The struct's fields must
	// be exported to use it with json.NewDecoder.
	var input struct {
		Text      string   `json:"text"`
		Contexts  []string `json:"contexts"`
		Projects  []string `json:"projects"`
		Priority  string   `json:"priority"`
		Completed bool     `json:"completed"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	todo := &data.Todo{
		Text:      input.Text,
		UserID:    app.contextGetUser(r).ID,
		Contexts:  input.Contexts,
		Projects:  input.Projects,
		Priority:  input.Priority,
		Completed: input.Completed,
	}

	v := validator.New()
	data.ValidateTodo(v, todo)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.Models.Todos.Insert(todo)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Specify the API location of the created resource.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/todos/%d", todo.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"todo": todo}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// getTodo handles GET requests to the /v1/todos/:id endpoint. If there is a
// todo item with matching ID and userID it will be sent in the response.
//
// If not, a 404 Not Found response is sent.
func (app *APIApplication) getTodo(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIdParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	userID := app.contextGetUser(r).ID

	todo, err := app.Models.Todos.GetTodoIfOwned(id, userID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"todo": todo}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// updateTodo handles PATCH requests to the /v1/todos/:id endpoint. It's body
// should contain one or more todo fields to be modified. Partial updates are
// supported.
//
// If fields are omitted in the request body, or if they are given a null value
// they will be unchanged.
//
// Only todo items with matching ID and userID can be updated.
func (app *APIApplication) updateTodo(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIdParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	userID := app.contextGetUser(r).ID

	todo, err := app.Models.Todos.GetTodoIfOwned(id, userID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// input is a struct to store the JSON values from the request body. We use
	// pointers to facilitate partial updates. If a value is not provided, the
	// pointer will be nil, and we can leave the corresponding field unchanged.
	var input struct {
		Text      *string   `json:"text"`
		Contexts  *[]string `json:"contexts"`
		Projects  *[]string `json:"projects"`
		Priority  *string   `json:"priority"`
		Completed *bool     `json:"completed"`
	}

	// Read JSON from request body into the input struct.
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// If the input field isn't nil, update the corresponding field in the record.
	if input.Text != nil {
		todo.Text = *input.Text
	}
	if input.Contexts != nil {
		todo.Contexts = *input.Contexts
	}
	if input.Projects != nil {
		todo.Projects = *input.Projects
	}
	if input.Priority != nil {
		todo.Priority = *input.Priority
	}
	if input.Completed != nil {
		todo.Completed = *input.Completed
	}

	// Validate the updated todo record, or return a 422 response.
	v := validator.New()
	data.ValidateTodo(v, todo)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Pass updated todo record to Todos.Update().
	err = app.Models.Todos.Update(todo)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Write updated JSON to response.
	err = app.writeJSON(w, http.StatusOK, envelope{"todo": todo}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

// deleteTodo handles requests to DELETE /v1/todos/:id. If it finds a
// document with the supplied ID it removes it from the database and sends a
// JSON response: { "message": "todo successfully deleted" }
//
// If the document is not found, a 404 response is sent.
func (app *APIApplication) deleteTodo(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIdParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Delete record or send an error response.
	err = app.Models.Todos.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "todo successfuly deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
