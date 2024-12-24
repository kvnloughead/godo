/*
Copyright © 2024 Kevin Loughead <kvnloughead@gmail.com>
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"syscall"

	"github.com/kvnloughead/godo/cmd/cli/token"
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

// authCommand authenticates a user and saves authentication token to the
// .config/godo/.token file. Email and password can be provided via flags. If
// not provided, the user will be prompted for them securely.
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
		handleError := func(msg string, err error) {
			app.handleAuthenticationError(msg, err,
				"method", http.MethodPost,
				"url", url)
		}
		// Prepare JSON payload from args
		payload := map[string]string{
			"email":    email,
			"password": password,
		}

		// Marshal payload to JSON
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			app.handleAuthenticationError("Failed to marshal JSON", err)
			return
		}

		// Create request
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonPayload))
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

		// Handle response
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			handleError("Failed to read response body", err)
			return
		}

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
		isDev := strings.Contains(app.Config.APIBaseURL, "localhost")
		tokenManager := token.NewManager(filepath.Join(os.Getenv("HOME"), ".config/godo"), isDev)
		if err := tokenManager.SaveToken(authToken); err != nil {
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
