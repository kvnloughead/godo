package main

import (
	"expvar"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/kvnloughead/godo/internal/data"
)

// The routes function initializes and returns an http.Handler with all the
// route definitions for the application. It uses httprouter for routing
// requests to their corresponding handlers based on the HTTP method and path.
//
// The defined routes are as follows:
//
//   - GET    /v1/healthcheck   				 Show application information.
//
//   - GET    /v1/todos								   Show details of a subset of todos.
//     [permissions - todos:read]
//
//   - POST   /v1/todos								   Create a new todo.
//     [permissions - todos:write]
//
//   - GET    /v1/todos/:id	  				   Show details of a specific todo.
//     [permissions - todos:read]
//
//   - PATCH  /v1/todos/:id						   Update details of a specific todo.
//     [permissions - todos:read]
//
//   - DELETE /v1/todos/:id	  				   Delete a specific todo.
//     [permissions - todos:read]
//
//   - POST   /v1/users         				 Register a new user.
//
//   - PUT    /v1/users/activation     	 Activates a user.
//
//   - POST   /v1/tokens/activation   	 Generate a new activation token.
//
//   - POST   /v1/tokens/authentication  Generate an authentication token.
//
//   - GET    /debug/vars                Display application metrics.
//
// This function also sets up custom error handling for scenarios where no
// route is matched (404 Not Found) and when a method is not allowed for a
// given route (405 Method Not Allowed), using the custom error handlers
// defined in api/errors.go.
//
// Finally, the router is wrapped with the recoverPanic middleware to handle any
// panics that occur during request processing.
func (app *application) routes() http.Handler {
	router := httprouter.New()

	// Set custom error handlers for 404 and 405 errors.
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheck)

	// The /v1/todos endpoints require either todos:read or todos:write permission
	router.HandlerFunc(http.MethodGet, "/v1/todos", app.requirePermission(data.TodosRead, app.listTodos))
	router.HandlerFunc(http.MethodPost, "/v1/todos", app.requirePermission(data.TodosWrite, app.createTodo))
	router.HandlerFunc(http.MethodGet, "/v1/todos/:id", app.requirePermission(data.TodosRead, app.getTodo))
	router.HandlerFunc(http.MethodPatch, "/v1/todos/:id", app.requirePermission(data.TodosWrite, app.updateTodo))
	router.HandlerFunc(http.MethodDelete, "/v1/todos/:id", app.requirePermission(data.TodosWrite, app.deleteTodo))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUser)
	router.HandlerFunc(http.MethodPut, "/v1/users/activation", app.activateUser)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/activation", app.createActivationToken)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationToken)

	// Expose application metrics as a JSON response to HTTP request.
	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	middlewares := alice.New(app.logRequest, app.metrics, app.recoverPanic, app.enableCORS, app.rateLimit, app.authenticate)
	return middlewares.Then(router)
}
