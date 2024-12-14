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

	"github.com/spf13/cobra"
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

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate a terminal session.",
	Long:  `TODO - add help text`,
	Run: func(cmd *cobra.Command, args []string) {
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

		// Then use it throughout the function
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
		token := authResp.AuthenticationToken.Token

		// Save token securely
		homedir, err := os.UserHomeDir()
		if err != nil {
			handleError("Failed to get home directory", err)
			return
		}
		configDir := filepath.Join(homedir, ".config/godo")
		os.MkdirAll(configDir, 0755)

		tokenFile := filepath.Join(configDir, ".token")
		if err := os.WriteFile(tokenFile, []byte(token), 0600); err != nil {
			handleError("Failed to save token", err)
			return
		}
		fmt.Println("Authentication successful and token saved")
	},
}

func init() {
	authCmd.Flags().StringVarP(&email, "email", "e", "", "Email")
	authCmd.Flags().StringVarP(&password, "password", "p", "", "Password")
	authCmd.MarkFlagRequired("email")
	authCmd.MarkFlagRequired("password")
	rootCmd.AddCommand(authCmd)
}
