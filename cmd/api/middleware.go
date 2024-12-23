package main

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	validator "github.com/kvnloughead/godo/internal"
	"github.com/kvnloughead/godo/internal/data"

	"github.com/google/uuid"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

// recoverPanic is a middleware that catches all panics in a handler chain.
// When a panic is caught, it is handled by
//  1. Setting the "Connection: close" header, to instruct go to shut down the
//     server after sending the response.
//  2. Sending a 500 Internal Server Error response containing the error from
//     the recovered panic.
func (app *APIApplication) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				// err has type any so must be converted to error
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// rateLimit is a middleware that limits the number of requests to an average of
// 2 per second per IP address, with bursts of up to 4 seconds.
//
// If an X-Forwarded-For or X-Real-IP header is found, the IP is taken from
// there. Otherwise it is taken from r.RemoteAddr.
//
// If the limit is exceeded, a 429 Too Many Request response is sent to the
// client.
func (app *APIApplication) rateLimit(next http.Handler) http.Handler {
	// Struct client contains data corresponding to a client IP. It has a rate
	// limiter property, and a lastSeen property used to remove unused clients
	// from the clients map.
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)

		// Metrics
		rateLimitExceeded = expvar.NewInt("rate_limit_exceeded_total")
		currentClients    = expvar.NewInt("rate_limit_current_clients")
	)

	// Start background goroutine to remove old entries from the clients map.
	go func() {
		for {
			time.Sleep(time.Minute)

			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			// Update the current number of clients.
			currentClients.Set(int64(len(clients)))

			mu.Unlock()
		}
	}()

	// addRateLimitHeaders adds the rate limit headers to the response.
	addRateLimitHeaders := func(w http.ResponseWriter, remaining float64) {
		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(app.Config.Limiter.Burst))
		w.Header().Set("X-RateLimit-Remaining", strconv.FormatFloat(remaining, 'f', 0, 64))
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Second).Unix(), 10))
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.Config.Limiter.Enabled {
			// Get IP address. If an X-Forwarded-For or X-Real-IP header is found, the
			// IP is taken from there. Otherwise it is taken from r.RemoteAddr.
			ip := realip.FromRequest(r)
			mu.Lock()

			// If no limiter exists for current IP, add it to the map of clients.
			if _, ok := clients[ip]; !ok {
				limiter := rate.NewLimiter(
					rate.Limit(app.Config.Limiter.RPS),
					app.Config.Limiter.Burst,
				)
				clients[ip] = &client{limiter: limiter}
			}

			clients[ip].lastSeen = time.Now()

			// If the client's limiter doesn't allow the request, increment the
			// rateLimitExceeded counter and send a 429 response.
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				addRateLimitHeaders(w, 0) // 0 remaining tokens
				rateLimitExceeded.Add(1)
				app.Logger.Info("rate limit exceeded",
					"ip", ip,
					"limit", app.Config.Limiter.RPS,
					"burst", app.Config.Limiter.Burst,
				)
				app.rateLimitExceededReponse(w, r)
				return
			}

			// We can't defer unlocking this mutext, because it wouldn't occur until
			// all downstream handlers have returned.
			remaining := clients[ip].limiter.Tokens()
			addRateLimitHeaders(w, remaining)
			mu.Unlock()

		}
		next.ServeHTTP(w, r)
	})
}

// The authenticate middleware authenticates a user based on the token provided
// in the authorization header. The header should be of the form "Bearer
// <token>". The token should be 26 bytes long.
//
// 401 Unauthorized responses are sent if the authorization header is
// malformed, if the token is invalid, or if a user record corresponding to the
// token isn't found.
//
// If everything checks out, the user's data is added to the request context.
// Otherwise, the anonymous user is added to the request context.
func (app *APIApplication) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The "Vary: Authorization" header indicates to caches that the response
		// may vary based on the value of the request's Authorization header.
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader == "" {
			// If there is no authorization header, add anonymous user to the context.
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		// Split the header and return a 401 if not in the format "Bearer <token>".
		parts := strings.Split(authorizationHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := parts[1]

		// Validate that the token is 26 bytes long.
		v := validator.New()
		data.ValidateTokenPlaintext(v, token)
		if !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// Get user from DB. If record isn't found we send a 401 response.
		user, err := app.Models.Users.GetForToken(data.Authentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		// Add user to request context and call the next handler.
		r = app.contextSetUser(r, user)
		next.ServeHTTP(w, r)
	})
}

// The requireAuthenticatedUser middleware prevents users from accessing a
// resource unless they are authenticated. If they aren't authenticated, a 401
// response is sent.
//
// This middleware accepts and returns an http.HandlerFunc, as opposed to
// http.Handler, which allows us to wrap our individual /v1/todo** routes
// with it.
func (app *APIApplication) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := contextGet[*data.User](r, userContextKey)

		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		if !user.Activated {
			app.activationRequiredResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// The requireActivatedUser middleware prevents users from accessing a resource
// unless they are authenticated and activated. It authenticates users by
// calling app.requireAuthenticatedUser.
//
// If the user isn't authenticated, a 401 response is sent.
// If the user is authenticated, but not activated, a 403 response is sent.
//
// This middleware accepts and returns an http.HandlerFunc, as opposed to
// http.Handler, which allows us to wrap our individual /v1/todo** routes
// with it.
func (app *APIApplication) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := contextGet[*data.User](r, userContextKey)

		if !user.Activated {
			app.activationRequiredResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})

	return app.requireAuthenticatedUser(fn)
}

// The requirePermission middleware prevents users from accessing a resource
// unless they are authenticated, activated, and have the necessary permission.
// It authenticates users and checks their activation status by calling
// app.requireAuthenticatedUser.
//
// If the user isn't authenticated, a 401 response is sent.
// If the user is authenticated, but not activated, or if the user doesn't have
// the correct permissions, a 403 response is sent.
//
// This middleware accepts and returns an http.HandlerFunc, as opposed to
// http.Handler, which allows us to wrap our individual /v1/todo** routes
// with it.
func (app *APIApplication) requirePermission(permission data.PermissionCode, next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// There is no need to check IsAnonymous, this is handled by an earlier
		// middleware in the chain.
		user := contextGet[*data.User](r, userContextKey)

		permissions, err := app.Models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		if !permissions.Includes(permission) {
			app.permissionRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})

	return app.requireActivatedUser(fn)
}

// The isPreflight helper returns true if the request is preflight. A preflight
// request must
//
//   - use the OPTIONS method
//   - have an Origin header
//   - have an Access-Control-Allow-Methods header
func (app *APIApplication) isPreflight(r *http.Request) bool {
	return r.Method == http.MethodOptions &&
		r.Header.Get("Origin") != "" &&
		r.Header.Get("Access-Control-Request-Method") != ""
}

// The enableCORS middleware allows CORS from all trusted origins. Trusted
// origins must be passed as the -cors-trusted-origin flag at runtime.
//
// In the case of preflight requests, the appropriate response headers are set
// and a 200 OK response is send. We send 200 rather than 204 because some
// browsers don't support 204 No Content responses.
//
// This middleware allows the Authorization header in cross-origin requests, so
// it it critical to not set the Access-Control-Allow-Origin header to *.
func (app *APIApplication) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Tell caches that response may vary depending on the value of the
		// following request headers.
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Vary", "Access-Control-Request-Method")

		origin := r.Header.Get("Origin")

		if origin != "" {
			for i := range app.Config.Cors.TrustedOrigins {
				if app.Config.Cors.TrustedOrigins[i] == origin {
					w.Header().Set("Access-Control-Allow-Origin", origin)

					// If the request is a preflight request, set the necessary headers
					// and send a 200 OK response with no further action.
					if app.isPreflight(r) {
						w.Header().Set("Access-Control-Allow-Methods",
							"OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers",
							"Authorization, Content-Type")
						w.WriteHeader(http.StatusOK)
						return
					}

					break
				}
			}

		}
		next.ServeHTTP(w, r)
	})
}

//
// Metrics
//

// metricsResponseWriter wraps (and implements) the http.ResponseWriter
// interface, enabling the saving of status codes for tracking in the metrics
// middleware.
//
// It has an integer field for recording the response's status code, and a
// boolean field that tracks whether the response headers have been written
// already.
//
// It implements the Header, WriteHeader, and Write methods of the wrapped
// interface, as well as an Unwrap method that returns the wrapped interface.
type metricsResponseWriter struct {
	wrapped       http.ResponseWriter
	statusCode    int
	headerWritten bool
}

// newMetricResponseWriter returns a metricsResponseWriter struct that wraps
// the ResponseWriter that was passed as an argument. The statusCode field is
// set to http.StatusOK (200).
func newMetricResponseWriter(w http.ResponseWriter) *metricsResponseWriter {
	return &metricsResponseWriter{
		wrapped:    w,
		statusCode: http.StatusOK,
	}
}

// metricsResponseWriter.Header() calls the wrapped interface's Header() method.
func (mw *metricsResponseWriter) Header() http.Header {
	return mw.wrapped.Header()
}

// metricsResponseWriter.WriteHeader() calls the wrapped interface's
// WriteHeader() method. Then, if the headers for this response haven't yet
// been written, it records the response's status code and sets the calling
// structs headerWritten field to true.
func (mw *metricsResponseWriter) WriteHeader(statusCode int) {
	mw.wrapped.WriteHeader(statusCode)
	if !mw.headerWritten {
		mw.headerWritten = true
		mw.statusCode = statusCode
	}
}

// metricsResponseWriter.Write() calls the wrapped interface's Write() method.
// This will automatically write the headers if they haven't been written
// already, but will leave the status code at the default (200). So before
// calling the wrapped interface's Write() method, we set headerWritten to true.
func (mw *metricsResponseWriter) Write(b []byte) (int, error) {
	mw.headerWritten = true
	return mw.wrapped.Write(b)
}

// metricsResponseWriter.Unwrap() returns the wrapped http.ResponseWriter
// interface.
func (mw *metricsResponseWriter) Unwrap() http.ResponseWriter {
	return mw.wrapped
}

// The metrics middleware tracks some request specific data for sharing with
// the /debug/vars endpoint. Tracked information:
//
//   - total number of requests recieved
//   - total responses sent
//   - total processing time (in microseconds)
//   - a map of the total number responses sent for each status code
func (app *APIApplication) metrics(next http.Handler) http.Handler {
	var (
		totalRequestsRecieved           = expvar.NewInt("total_requests_recieved")
		totalResponsesSent              = expvar.NewInt("total_responses_sent")
		totalProcessingTimeMicroseconds = expvar.NewInt("total_processing_time_μs")
		totalResponsesSentByStatus      = expvar.NewMap("total_responses_sent_by_status")
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		totalRequestsRecieved.Add(1)

		mw := newMetricResponseWriter(w)
		next.ServeHTTP(mw, r)

		// Increment response counter for the response's status code, as well as the
		// counter of total responses.
		totalResponsesSentByStatus.Add(strconv.Itoa(mw.statusCode), 1)
		totalResponsesSent.Add(1)

		duration := time.Since(start).Microseconds()
		totalProcessingTimeMicroseconds.Add(duration)
	})
}

// Struct requestContext holds metadata about the current request
type requestContext struct {
	start      time.Time
	duration   time.Duration
	statusCode int
	userAgent  string
	authStatus string
	requestID  string // unique identifier for request tracing
}

// The contextualizeRequest middleware initializes a requestContext struct at
// the start of the request, and stores it in the request context. It also
// creates a response writer wrapper to capture the response's status code.
func (app *APIApplication) contextualizeRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := &requestContext{
			start:     time.Now(),
			userAgent: r.UserAgent(),
			requestID: uuid.New().String(),
		}

		app.Logger.Info("request started",
			"request_id", ctx.requestID,
			"method", r.Method,
			"uri", r.URL.RequestURI(),
		)

		// Store our request context
		r = r.WithContext(context.WithValue(r.Context(), requestContextKey, ctx))

		rw := newMetricResponseWriter(w)

		// Run all middleware
		next.ServeHTTP(rw, r)

		// Get the final request after all middleware has run
		user := contextGet[*data.User](r, userContextKey)
		authStatus := "unknown"
		if user.IsAnonymous() {
			authStatus = "anonymous"
		} else {
			authStatus = "authenticated"
		}

		ctx.duration = time.Since(ctx.start)
		ctx.statusCode = rw.statusCode
		ctx.authStatus = authStatus

		app.Logger.Info("request completed",
			"request_id", ctx.requestID,
			"duration", ctx.duration,
			"status", ctx.statusCode,
			"auth_status", ctx.authStatus,
		)
	})
}
