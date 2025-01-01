package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	validator "github.com/kvnloughead/godo/internal"
	"github.com/kvnloughead/godo/internal/data"
)

// ReadTokenFromFile attempts to read the contents of the authentication token
// from a file /home/username/.config/godo/.token. If the file exists and
// contains a potentially valid token string, this string is returned.
// Otherwise, an error is returned.
func (app *CLIApplication) ReadTokenFromFile() (string, error) {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	tokenFile := filepath.Join(homeDir, ".config/godo", ".token")

	if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
		return "", fmt.Errorf("token file doesn't exist: %w", err)
	}

	tokenBytes, err := os.ReadFile(tokenFile)
	if err != nil {
		return "", fmt.Errorf("failed to read token file: %w", err)
	}

	token := string(tokenBytes)

	v := validator.New()
	data.ValidateTokenPlaintext(v, token)
	if !v.Valid() {
		return "", fmt.Errorf("invalid token format: %w", err)
	}

	return token, nil
}

// handleError handles CLI errors by logging the error with app.Logger.Error and
// sending a user friendly message with fmt.Println.
//
// - logMsg is added as the msg field in the log.
// - stdoutMsg is printed to stdout.
// - err is added as the error field in the log.
// - fields is variadic and is added as additional fields in the log.
func (app *CLIApplication) handleError(logMsg, stdoutMsg string, err error, fields ...any) {
	// Convert fields to []any for slog.Error
	logFields := make([]any, len(fields)+2) // +2 for error field
	copy(logFields, fields)
	// Add error at the end
	logFields[len(fields)] = "error"
	logFields[len(fields)+1] = err

	app.Logger.Error(logMsg, logFields...)
	fmt.Println(stdoutMsg)
}

// handleAuthenticationError handles authentication related errors. It calls
// handleError with the appropriate log message, error message, and additional
// fields.
func (app *CLIApplication) handleAuthenticationError(logMsg string, err error, fields ...any) {
	app.handleError(logMsg,
		"\nError: failed to authenticate. \nCheck `~/.config/godo/logs` for details.\n",
		err,
		fields...)
}

// createJSONRequest creates a new HTTP request with the given method, URL, and
// payload. It sets the Content-Type header to "application/json" and the
// Authorization header to the token.
//
// It also logs the request method, url, and payload. If any additional string
// arguments are provided (i.e. excludeFields), they are removed from the
// payload before logging.
func (app *CLIApplication) createJSONRequest(method, url string, payload map[string]any, excludeFields ...string) (*http.Request, error) {
	// Log the request (omitting sensitive fields)
	if payload != nil {
		logPayload := make(map[string]any)
		for k, v := range payload {
			logPayload[k] = v
		}
		for _, field := range excludeFields {
			delete(logPayload, field)
		}
		app.Logger.Info("sending request",
			"method", method,
			"url", url,
			"payload", logPayload)
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// readResponse reads the response body and logs the response's method,
// url, status, and body. If the body isn't valid JSON, it logs the body as a
// string.
func (app *CLIApplication) readResponse(resp *http.Response, handleError func(string, error) error) ([]byte, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, handleError("failed to read response body", err)
	}

	// Try to parse as JSON first
	var responseBody any
	if json.Valid(body) {
		var data interface{} // use interface{} instead of map to handle arrays and nested structures
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, handleError("failed to parse JSON response", err)
		}
		responseBody = data
	} else {
		responseBody = string(body)
	}

	app.Logger.Info("received response",
		"method", resp.Request.Method,
		"url", resp.Request.URL,
		"status", resp.Status,
		"body", responseBody)

	return body, nil
}

// readTodoListResponse is logs condensed data from the GET /v1/todos endpoint.
// Specifically, it logs the pagination data and the number of todos. The actual
// todos are not logged.
func (app *CLIApplication) readTodoListResponse(resp *http.Response, handleError func(string, error) error) ([]byte, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, handleError("failed to read response body", err)
	}

	if json.Valid(body) {
		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			return nil, handleError("failed to parse JSON response", err)
		}

		// Create condensed version for logging
		logData := map[string]interface{}{
			"pagination": data["paginationData"],
			"todo_count": len(data["todos"].([]interface{})),
		}

		app.Logger.Info("received todos",
			"method", resp.Request.Method,
			"url", resp.Request.URL,
			"status", resp.Status,
			"summary", logData)
	}

	return body, nil
}
