package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kvnloughead/godo/internal/data"
)

// The contextKey type is a custom string type for request context keys.
type contextKey string

var (
	userContextKey    = contextKey("user")
	requestContextKey = contextKey("requestContext")
)

// The contextSetUser method accepts a request and a user struct as arguments,
// adds the user to the request's context with a key of "user", and returns a
// copy of the request.
func (app *APIApplication) contextSetUser(r *http.Request, user *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// contextGet retrieves a value from the request context with type safety.
// Example usage:
//
//	user := contextGet[*data.User](r, userContextKey)
//	reqCtx := contextGet[*requestContext](r, requestContextKey)
func contextGet[T any](r *http.Request, key contextKey) T {
	value, ok := r.Context().Value(key).(T)
	if !ok {
		panic(fmt.Sprintf("missing or invalid type for %s in request context", key))
	}
	return value
}
