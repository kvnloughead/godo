package main

import (
	"errors"
	"net/http"
	"time"

	validator "github.com/kvnloughead/godo/internal"
	"github.com/kvnloughead/godo/internal/data"
)

// registerUser handles POST requests to the /v1/users endpoint. The request
// body is decoded by the app.readJSON helper. See that function for details
// about error handling.
//
// The request body must contain a name, email, and password. Request bodies
// are validated by data.ValidateUser. A failedValidationResponse error is sent
// if one or more fields fails validation, or if the email is a duplicate. A
// hash is generated from the plaintext password via bcrypt and stored in the
// database.
//
// On successful registration, a token is generated securely and encrypted with
// SHA-256. This token is sent to the user in a a welcome email via app.Mailer,
// with instructions on how to activate the account.
func (app *APIApplication) registerUser(w http.ResponseWriter, r *http.Request) {
	// Struct to store the data from the responses body. The struct's fields must
	// be exported to use it with json.NewDecoder.
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Copy info from request body into a new user struct. Below, we set the
	// password by calling the user struct's Password.Set method.
	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Validate user. Email uniqueness is checked on attempted insert.
	v := validator.New()
	data.ValidateUser(v, user)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Insert new record into DB, if possible.
	err = app.Models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Create activation token and add to database.
	token, err := app.Models.Tokens.New(user.ID, 72*time.Hour, data.Activation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Grant user the "todos:read" permission.
	err = app.Models.Permissions.AddForUser(user.ID, data.TodosRead)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Lauch goroutine to send a welcome email.
	app.background(func() {
		data := struct {
			Token *data.Token
			User  *data.User
		}{
			Token: token,
			User:  user,
		}
		err = app.Mailer.Send(user.Email, "user_welcome.tmpl", data)
		if err != nil {
			app.Logger.Error(err.Error())
		}
	})

	// Write JSON response.
	err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *APIApplication) activateUser(w http.ResponseWriter, r *http.Request) {
	// Retrieve token from body of request and validate it.
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	data.ValidateTokenPlaintext(v, input.TokenPlaintext)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Attempt to get the corresponding user.
	user, err := app.Models.Users.GetForToken(
		data.Activation,
		input.TokenPlaintext,
	)
	if err != nil {
		switch {
		// If user can't be found, the token must be invalid or expired.
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired token")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// If user was found, activate them and update the record.
	user.Activated = true
	err = app.Models.Users.Update(user)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Delete all activation tokens for the user.
	err = app.Models.Tokens.DeleteAllForUser(data.Activation, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Grant activated user "todos:write" permission.
	err = app.Models.Permissions.AddForUser(user.ID, data.TodosWrite)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	env := envelope{"message": "user successfully activated", "user": user}
	err = app.writeJSON(w, http.StatusAccepted, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
