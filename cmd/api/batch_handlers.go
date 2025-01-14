// batch_handlers.go contains handlers for batch operations on todos.
package main

import (
	"net/http"
)

func (app *APIApplication) deleteTodosBatch(w http.ResponseWriter, r *http.Request) {
	var input struct {
		IDs []string `json:"ids"`
	}

	var outcome = make(map[int]string)
	var processed []int
	success := true

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

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

	app.writeJSON(w, http.StatusOK, envelope{"success": success, "processed": processed, "outcome": outcome}, nil)
}

func (app *APIApplication) updateTodosBatch(w http.ResponseWriter, r *http.Request) {

}
