// batch_handlers.go contains handlers for batch operations on todos.
package main

import (
	"errors"
	"net/http"
)

func (app *APIApplication) deleteTodosBatch(w http.ResponseWriter, r *http.Request) {
	var input struct {
		IDs []string `json:"ids"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if len(input.IDs) == 0 {
		app.badRequestResponse(w, r, errors.New("no IDs provided"))
		return
	}

	var outcome = make(map[int]string)
	var processed []int
	success := true

	for i := range input.IDs {
		id, err := app.parseID(input.IDs[i])
		if err != nil {
			outcome[i] = err.Error()
			success = false
			continue
		}

		err = app.Models.Todos.Delete(id)
		if err != nil {
			outcome[i] = err.Error()
			success = false
			continue
		}

		processed = append(processed, i)
	}

	status := http.StatusOK
	if !success {
		status = http.StatusBadRequest
	}

	app.writeJSON(w, status, envelope{"success": success, "processed": processed, "outcome": outcome}, nil)
}

func (app *APIApplication) updateTodosBatch(w http.ResponseWriter, r *http.Request) {

}
