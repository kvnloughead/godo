// batch_handlers.go contains handlers for batch operations on todos.
package main

import (
	"net/http"
)

type result struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func (app *APIApplication) deleteTodosBatch(w http.ResponseWriter, r *http.Request) {
	var input struct {
		IDs []string `json:"ids"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		app.writeJSON(w, http.StatusBadRequest, envelope{
			"error":   err.Error(),
			"success": false,
			"results": []result{},
		}, nil)
		return
	}

	if len(input.IDs) == 0 {
		app.writeJSON(w, http.StatusBadRequest, envelope{
			"error":   "no IDs provided",
			"success": false,
			"results": []result{},
		}, nil)
		return
	}

	success := true
	var results []result

	handleError := func(result *result, err error, msg string) {
		result.Success = false
		result.Error = msg
		app.logError(r, err.Error())
		results = append(results, *result)
		success = false
	}

	for i := range input.IDs {
		res := result{ID: input.IDs[i], Success: true}

		id, err := app.parseID(input.IDs[i])
		if err != nil {
			handleError(&res, err, "invalid ID")
			continue
		}

		err = app.Models.Todos.Delete(id)
		if err != nil {
			handleError(&res, err, "not found")
			continue
		}

		results = append(results, res)
	}

	status := http.StatusOK
	if !success {
		status = http.StatusBadRequest
	}

	app.writeJSON(w, status, envelope{"success": success, "results": results}, nil)
}

func (app *APIApplication) updateTodosBatch(w http.ResponseWriter, r *http.Request) {

}
