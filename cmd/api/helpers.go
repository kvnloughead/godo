package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	validator "github.com/kvnloughead/godo/internal"
)

// writeJSON marshals the data into JSON, then prepares and sends the response.
// The response is sent with
//
//  1. The "Content-type: application/json" header.
//  2. The status code that was supplied as an argument.
//
// Errors are simply returned to the caller.
func (app *APIApplication) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	// Marshal data map into JSON for the response, indenting for readability.
	js, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	// Loop through the headers map and add each one to the ResponseWriter's header map. If headers is nil, this loop will simply be skipped.
	for k, v := range headers {
		w.Header()[k] = v
	}

	// Add Content-type header and status code. Then write JSON to response, appending a newline for QOL.
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(status)
	w.Write(append(js, '\n'))

	return nil
}

// readJSON decodes a requests body to the target destination. If the target destination is not a
// non-nil pointer, panic will ensue. Only a single JSON value per request is
// accepted.
//
// The following errors are caught and responded to specifically.
//
//  1. In most cases, general syntax errors will result in a json.SyntaxError.
//     In this case, we return a message with the offset of the error.
//
//  2. In some cases, syntax errors will result in an io.ErrUnexpectedEOF. Here,
//     we return a message with no offset. See the open issue for details
//     https://github.com/golang/go/issues/25956
//
//  3. If the request includes data of an incorrect type, this results in a
//     *json.UnmarshalTypeError. In these cases we return a message indicating
//     the offending field, if possible. Otherwise, a generic message.
//
//  4. An empty body results in an io.EOF error, which are caught and responded
//     to appropriately.
//
//  5. If an unknown field is in the request body, an error will be returned.
//
//  6. If dst is anything but a non-nil pointer, then json.Decode returns a
//     json.InvalidUnmarshalError. In this case, we panic, rather than returning
//     an error to the handler, because to do otherwise would require excessive
//     error handling in all of our handlers.
//
// All other errors are returned as-is.
func (app *APIApplication) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	// Restrict size of request bodyy to 1MB.
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshallTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshallTypeError):
			if unmarshallTypeError.Field != "" {
				return fmt.Errorf("body contains JSON of incorrect type for field %q", unmarshallTypeError.Field)
			}
			return fmt.Errorf("body contains JSON of an incorrect type (at character %d)", unmarshallTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("request body must not be empty")

		// It is necessary to check the error for the prefix currently, but there
		// is an open issue to make this a separate error in the future.
		// https://github.com/golang/go/issues/29035
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown field %s)", fieldName)

		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not exceed %d bytes", maxBytesError.Limit)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	// To prevent multiple values from being provided in a request, we decode it
	// a second time and check for an io.EOF.
	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		return errors.New("body must contain only a single JSON value")
	}

	return nil
}

// envelope is a type used for wrapping JSON responses to ensure a consistent
// response structure. It is a map with string keys and values of any type.
//
// It is commonly in handlers and middleware to wrap responses. For example,
// error responses are typically wrapped like this:
//
//	envelope{"error": "detailed error message"}
type envelope map[string]any

// readIdParam reads an ID param from the request context and parses it as an
// int64. If the ID doesn't parse to a positive integer, an error is returned.
func (app *APIApplication) readIdParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("ID must be a positive integer")
	}

	return id, nil
}

// readQueryString returns the value of the key in the provided query string
// map. If the value is an empty string, the default value is returned instead.
func (app *APIApplication) readQueryString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}
	return s
}

// readQueryCSV reads a CSV field from the query string argument. The CSV
// string is then split into a slice of strings and returned. If the field
// is empty, the default value is returned.
func (app *APIApplication) readQueryCSV(qs url.Values, key string, defaultValue []string) []string {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}
	return strings.Split(s, ",")
}

// readQueryInt reads an integer valued field from the query string argument.
// If the field is empty, the default value is returned. If the field can't be
// converted to an integer, the default value is returned, and an error is
// added to the validator instance.
func (app *APIApplication) readQueryInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return i
}

// readQueryBool reads a boolean valued field from the query string argument.
// If the field is empty, the default value is returned. If the field can't be
// converted to a boolean, the default value is returned, and an error is
// added to the validator instance.
func (app *APIApplication) readQueryBool(qs url.Values, key string, defaultValue bool, v *validator.Validator) bool {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	b, err := strconv.ParseBool(s)
	if err != nil {
		v.AddError(key, "must be a boolean value")
		return defaultValue
	}

	return b
}

// The background method launches a background goroutine. This goroutine
// recovers from panics, logging the resulting errors with app.Logger, and
// calls the function argument.
//
// Goroutines are tracked via the app.WG WaitGroup instance, and this counter
// is checked before shutting down the application. See app.serve() for details.
func (app *APIApplication) background(fn func()) {
	// Increment WaitGroup counter.
	app.WG.Add(1)
	go func() {
		// Decrement WaitGroup counter after completion.
		defer app.WG.Done()

		defer func() {
			if err := recover(); err != nil {
				app.Logger.Error(fmt.Sprintf("%v", err))
			}
		}()

		fn()
	}()
}
