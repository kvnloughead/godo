package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type authResponse struct {
	AuthenticationToken struct {
		Token  string `json:"token"`
		Expiry string `json:"expiry"`
	} `json:"authentication_token"`
}

var (
	email    string
	password string
)

// authCmd authenticates a user and creates a token for subsequent API requests.
// The token is stored in the user's config directory and is valid for 14 days
// in production or 28 days in development.
var authCmd = &cobra.Command{
	Use:   "auth [-e email] [-p password]",
	Short: "Authenticate with the Godo API",
	Long: `
Authenticates a godo user's account. This will create an authentication token that will be used for subsequent commands. The token is valid for 14 days in production and 28 days in development mode.

If email or password are not provided via flags, you will be prompted for them.
The password will not be displayed when typed.

Examples:

    # Authenticate with flags
    godo auth -e user@example.com -p mypassword

    # Authenticate with prompts
    godo auth

    # Authenticate with email flag only (will prompt for password)
    godo auth -e user@example.com

Only an activated user can be authenticated. Run 'godo activate -h' for more information.`,
	Run: func(cmd *cobra.Command, args []string) {
		// If email wasn't provided via flag, prompt for it
		if email == "" {
			fmt.Print("Enter email: ")
			fmt.Scanln(&email)
		}

		// If password wasn't provided via flag, prompt for it securely
		if password == "" {
			fmt.Print("Enter password: ")
			bytePassword, err := term.ReadPassword(int(syscall.Stdin))
			if err != nil {
				app.handleAuthenticationError("Failed to read password", err)
				return
			}
			fmt.Println() // Add newline after password input
			password = string(bytePassword)
		}
		// Create request url
		url := app.Config.APIBaseURL + "/tokens/authentication"

		// Define a helper function that captures the parameters that are common to
		// all errors
		handleError := func(msg string, err error) error {
			app.handleAuthenticationError(msg, err,
				"method", http.MethodPost,
				"url", url)
			return err
		}
		// Prepare JSON payload from args
		payload := map[string]any{
			"email":    email,
			"password": password,
		}

		// Create request
		req, err := app.createJSONRequest(http.MethodPost, url, payload)
		if err != nil {
			handleError("Failed to create request", err)
			return
		}
		req.Header.Set("Context-Type", "application/json")

		// Send request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			handleError("Failed to send request", err)
			return
		}

		// Read response body and log it
		body, err := app.readResponse(resp, handleError)
		if err != nil {
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			handleError(fmt.Sprintf("Failed to authenticate: %s", string(body)), nil)
			return
		}

		// Unmarshal response
		var authResp authResponse
		err = json.Unmarshal(body, &authResp)
		if err != nil {
			handleError("Failed to unmarshal response", err)
			return
		}

		// Retrieve token from response
		if authResp.AuthenticationToken.Token == "" {
			handleError("Token not found in response", nil)
			return
		}
		authToken := authResp.AuthenticationToken.Token

		// Save token securely using token manager
		if err := app.TokenManager.SaveToken(authToken); err != nil {
			handleError("Failed to save token", err)
			return
		}
		fmt.Println("Authentication successful and token saved")
	},
}

func init() {
	authCmd.Flags().StringVarP(&email, "email", "e", "", "Email")
	authCmd.Flags().StringVarP(&password, "password", "p", "", "Password")
	rootCmd.AddCommand(authCmd)
}
