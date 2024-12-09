package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

var registerCmd = &cobra.Command{
	Use:   "register -e EMAIL -p PASSWORD -n NAME",
	Short: "Register a new user",
	Long:  "Register a new user with the given email, password, and name. An activation email will be sent to the user.",
	Run: func(cmd *cobra.Command, args []string) {
		// Prepare JSON payload
		payload := map[string]string{
			"email":    email,
			"password": password,
			"name":     name,
		}
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			app.Logger.Error("failed to marshal JSON", "error", err)
			return
		}

		// Log the request (excluding password)
		app.Logger.Info("sending registration request",
			"url", app.Config.APIBaseURL+"/users",
			"email", email,
			"name", name)

		// Create request
		url := app.Config.APIBaseURL + "/users"
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonPayload))
		if err != nil {
			app.Logger.Error("failed to create request", "error", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		// Send request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			app.Logger.Error("failed to send request", "error", err)
			return
		}
		defer resp.Body.Close()

		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			app.Logger.Error("failed to read response body", "error", err)
			return
		}

		// Log the response (before processing)
		app.Logger.Info("received response",
			"status", resp.Status,
			"body", string(body))

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
