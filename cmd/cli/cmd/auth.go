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
		// Prepare JSON payload from args
		payload := map[string]string{
			"email":    email,
			"password": password,
		}
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			app.handleAuthenticationError("Failed to marshal JSON", err)
			return
		}

		// Create request
		url := app.Config.APIBaseURL + "/v1/tokens/authentication"
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonPayload))
		if err != nil {
			app.handleAuthenticationError("Failed to create request", err)
			return
		}
		req.Header.Set("Context-Type", "application/json")

		// Send request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			app.handleAuthenticationError("Failed to send request", err)
			return
		}

		// Handle response
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			app.handleAuthenticationError(fmt.Sprintf("Failed to create token for %s", email), nil)
			return
		}

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			app.handleAuthenticationError("Failed to read response body", err)
			return
		}

		// Unmarshal response
		var authResp authResponse
		err = json.Unmarshal(body, &authResp)
		if err != nil {
			app.handleAuthenticationError("Failed to unmarshal response", err)
			return
		}

		// Retrieve token from response
		if authResp.AuthenticationToken.Token == "" {
			app.handleAuthenticationError("Token not found in response", err)
			return
		}
		token := authResp.AuthenticationToken.Token

		// Save token securely
		homedir, err := os.UserHomeDir()
		if err != nil {
			app.handleAuthenticationError("Failed to get home directory", err)
			return
		}
		configDir := filepath.Join(homedir, ".config/godo")
		os.MkdirAll(configDir, 0755)

		tokenFile := filepath.Join(configDir, ".token")
		if err := os.WriteFile(tokenFile, []byte(token), 0600); err != nil {
			app.handleAuthenticationError("Failed to save token", err)
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
