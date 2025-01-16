package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// Create a type that matches our expected response structure
type batchResponse struct {
	Success bool     `json:"success"`
	Results []result `json:"results"`
	Error   string   `json:"error,omitempty"`
}

func TestDeleteTodosBatch(t *testing.T) {
	app := newTestApplication(t)

	tests := []struct {
		name       string
		input      map[string][]string
		wantStatus int
		wantBody   envelope
	}{
		{
			name:       "Valid batch delete",
			input:      map[string][]string{"ids": {"1", "2", "3"}},
			wantStatus: http.StatusOK,
			wantBody: envelope{
				"success": true,
				"results": []result{
					{ID: "1", Success: true},
					{ID: "2", Success: true},
					{ID: "3", Success: true},
				},
			},
		},
		{
			name:       "Empty IDs list",
			input:      map[string][]string{"ids": {}},
			wantStatus: http.StatusBadRequest,
			wantBody: envelope{
				"error":   "no IDs provided",
				"success": false,
				"results": []result{},
			},
		},
		{
			name:       "Invalid ID format",
			input:      map[string][]string{"ids": {"1", "invalid", "3"}},
			wantStatus: http.StatusBadRequest,
			wantBody: envelope{
				"success": false,
				"results": []result{
					{ID: "1", Success: true},
					{ID: "invalid", Success: false, Error: "invalid ID"},
					{ID: "3", Success: true},
				},
			},
		},
		{
			name:       "Not found ID",
			input:      map[string][]string{"ids": {"1", "999", "3"}},
			wantStatus: http.StatusBadRequest,
			wantBody: envelope{
				"success": false,
				"results": []result{
					{ID: "1", Success: true},
					{ID: "999", Success: false, Error: "not found"},
					{ID: "3", Success: true},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert the input map to JSON
			inputJSON, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatal(err)
			}

			// Create a new request with the input JSON
			req := httptest.NewRequest(http.MethodDelete, "/v1/batch/todos", bytes.NewBuffer(inputJSON))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			app.deleteTodosBatch(rr, req)

			if status := rr.Code; status != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.wantStatus)
			}

			// Decode the response into the expected response structure
			var got batchResponse
			err = json.NewDecoder(rr.Body).Decode(&got)
			if err != nil {
				t.Fatal(err)
			}

			// Convert the response to an envelope for comparison
			gotEnvelope := envelope{
				"success": got.Success,
				"results": got.Results,
			}
			if got.Error != "" {
				gotEnvelope["error"] = got.Error
			}

			if !reflect.DeepEqual(gotEnvelope, tt.wantBody) {
				t.Errorf("handler returned wrong body\ngot: %#v\nwant: %#v",
					gotEnvelope, tt.wantBody)
			}
		})
	}
}
