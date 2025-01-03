package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

type RegisterResponse struct {
	User struct {
		ID        int64  `json:"id"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		Activated bool   `json:"activated"`
	} `json:"user"`
}

var (
	name string
)

// registerCmd creates a new user account. After registration, an activation
// token is sent to the provided email address.
var registerCmd = &cobra.Command{
	Use:   "register [-e email]",
	Short: "Register a new user account",
	Long: `
Register a new user account with godo. After registration, you'll receive an email with an activation token. You must activate your account before you can use it.

If an email is not provided via flag, you will be prompted for it.

Examples:

    # Register with email flag
    godo register -e user@example.com

    # Register with prompt
    godo register

After registering, check your email for the activation token and run:

    godo activate <token>

See 'godo activate -h' for more information.`,
	Run: func(cmd *cobra.Command, args []string) {
		url := app.Config.APIBaseURL + "/users"

		// Define error handler
		handleError := func(msg string, err error) error {
			app.Logger.Error(msg,
				"error", err,
				"method", http.MethodPost,
				"url", url)
			fmt.Println("Error: Registration failed. Check logs for details.")
			return err
		}

		// Prepare JSON payload
		payload := map[string]any{
			"email":    email,
			"password": password,
			"name":     name,
		}

		// Create request. The password will be omitted from the log.
		req, err := app.createJSONRequest(http.MethodPost, url, payload, "password")
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
			var registerResp RegisterResponse
			if err := json.Unmarshal(body, &registerResp); err != nil {
				app.Logger.Error("failed to unmarshal response",
					"error", err,
					"body", string(body))
				fmt.Println("Error: Failed to parse server response")
				return
			}
			fmt.Printf("\nRegistration successful for %s!\nPlease check your email for activation instructions.\n",
				registerResp.User.Email)

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
	registerCmd.Flags().StringVarP(&email, "email", "e", "", "Email address")
	registerCmd.Flags().StringVarP(&password, "password", "p", "", "Password")
	registerCmd.Flags().StringVarP(&name, "name", "n", "", "Name")
	registerCmd.MarkFlagRequired("email")
	registerCmd.MarkFlagRequired("password")
	registerCmd.MarkFlagRequired("name")
	rootCmd.AddCommand(registerCmd)
}
