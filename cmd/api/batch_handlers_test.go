package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeleteTodosBatch(t *testing.T) {
	app := newTestApplication(t)

	tests := []struct {
		name       string
		input      map[string][]string
		wantStatus int
		wantBody   map[string]interface{}
	}{
		{
			name:       "Valid batch delete",
			input:      map[string][]string{"ids": {"1", "2", "3"}},
			wantStatus: http.StatusOK,
			wantBody: map[string]interface{}{
				"success":   true,
				"processed": []int{0, 1, 2},
			},
		},
		{
			name:       "Empty IDs list",
			input:      map[string][]string{"ids": {}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "Invalid ID format",
			input:      map[string][]string{"ids": {"1", "invalid", "3"}},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputJSON, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatal(err)
			}

			req := httptest.NewRequest(http.MethodDelete, "/v1/batch/todos", bytes.NewBuffer(inputJSON))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			app.deleteTodosBatch(rr, req)

			if status := rr.Code; status != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.wantStatus)
			}

			if tt.wantBody != nil {
				var got map[string]interface{}
				err = json.NewDecoder(rr.Body).Decode(&got)
				if err != nil {
					t.Fatal(err)
				}

				// Compare specific fields
				if got["success"] != tt.wantBody["success"] {
					t.Errorf("handler returned wrong success status: got %v want %v",
						got["success"], tt.wantBody["success"])
				}

				// Compare processed arrays
				gotProcessed, ok := got["processed"].([]interface{})
				if !ok {
					t.Fatal("processed field is not an array")
				}
				wantProcessed := tt.wantBody["processed"].([]int)
				if len(gotProcessed) != len(wantProcessed) {
					t.Errorf("handler returned wrong number of processed IDs: got %d want %d",
						len(gotProcessed), len(wantProcessed))
				}
			}
		})
	}
}
