package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

type ActivationResponse struct {
	User struct {
		ID        int64  `json:"id"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		Activated bool   `json:"activated"`
		CreatedAt string `json:"created_at"`
	} `json:"user"`
	Message string `json:"message"`
}

var (
	activationToken string
)

// activateCmd activates a user's account using the token received via email
// during registration. An account must be activated before it can be used to
// manage todos.
var activateCmd = &cobra.Command{
	Use:   "activate <token>",
	Short: "Activate a new user's account",
	Long: `
Activate a user's account with the given token. For example:

	godo activate 1234567890

The activation token is sent to the user in a welcome email received when
they register.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		activationToken = args[0]
		url := app.Config.APIBaseURL + "/users/activation"

		// Define error handler
		handleError := func(msg string, err error) error {
			app.Logger.Error(msg,
				"error", err,
				"method", http.MethodPut,
				"url", url)
			fmt.Println("Error: Activation failed. Check logs for details.")
			return err
		}

		// Prepare JSON payload
		payload := map[string]any{
			"token": activationToken,
		}

		// Log the request (excluding password)
		app.Logger.Info("sending activation request",
			"url", url,
			"token", activationToken)

		// Create request
		req, err := app.createJSONRequest(http.MethodPut, url, payload)
		if err != nil {
			handleError("failed to create request", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		// Send request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			handleError("failed to send request", err)
			return
		}
		defer resp.Body.Close()

		// Read response body and log it
		body, err := app.readResponse(resp, handleError)
		if err != nil {
			return
		}

		// Handle different status codes
		switch resp.StatusCode {
		case http.StatusAccepted:
			var activationResp ActivationResponse
			if err := json.Unmarshal(body, &activationResp); err != nil {
				app.Logger.Error("failed to unmarshal response",
					"error", err,
					"body", string(body))
				fmt.Println("Error: Failed to parse server response")
				return
			}
			fmt.Printf("\nActivation successful for %s!\n",
				activationResp.User.Email)

		case http.StatusUnprocessableEntity:
			// Handle validation errors (including duplicate email)
			var errorResp struct {
				Error map[string]string `json:"error"`
			}
			if err := json.Unmarshal(body, &errorResp); err != nil {
				app.Logger.Error("failed to unmarshal error response",
					"error", err,
					"body", string(body))
				fmt.Println("Error: Failed to parse server error response")
				return
			}
			// Print each validation error
			fmt.Println("\nRegistration failed:")
			for field, message := range errorResp.Error {
				fmt.Printf("- %s: %s\n", field, message)
			}

		default:
			app.Logger.Error("unexpected status code",
				"status", resp.Status,
				"body", string(body))
			fmt.Printf("\nError: Unexpected response from server (status %s)\n", resp.Status)
		}
	},
}

func init() {
	rootCmd.AddCommand(activateCmd)
}
