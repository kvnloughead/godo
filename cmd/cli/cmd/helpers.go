package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	validator "github.com/kvnloughead/godo/internal"
	"github.com/kvnloughead/godo/internal/data"
)

// ReadTokenFromFile attempts to read the contents of the authentication token
// from a file /home/username/.config/godo/.token. If the file exists and
// contains a potentially valid token string, this string is returned.
// Otherwise, an error is returned.
func (app *CLIApplication) ReadTokenFromFile() (string, error) {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	tokenFile := filepath.Join(homeDir, ".config/godo", ".token")

	if _, err := os.Stat(tokenFile); os.IsNotExist(err) {
		return "", fmt.Errorf("token file doesn't exist: %w", err)
	}

	tokenBytes, err := os.ReadFile(tokenFile)
	if err != nil {
		return "", fmt.Errorf("failed to read token file: %w", err)
	}

	token := string(tokenBytes)

	v := validator.New()
	data.ValidateTokenPlaintext(v, token)
	if !v.Valid() {
		return "", fmt.Errorf("invalid token format: %w", err)
	}

	return token, nil
}
