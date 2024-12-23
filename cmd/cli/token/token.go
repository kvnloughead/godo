// Package token provides a manager for the token file that stores the token to
// authenticate HTTP requests to the API made from the CLI.
//
// The token is stored in the user's home directory in a file named ".token" or
// ".token.dev" depending on the environment.

package token

import (
	"os"
	"path/filepath"
)

const (
	defaultTokenFile = ".token"
	devTokenFile     = ".token.dev"
)

// Manager is a struct that manages the token file.
type Manager struct {
	configDir string // The directory where the token file is stored.
	isDev     bool   // Whether the token is for development.
}

// NewManager creates a new Manager.
func NewManager(configDir string, isDev bool) *Manager {
	return &Manager{
		configDir: configDir,
		isDev:     isDev,
	}
}

// TokenFile returns the path to the token file.
func (m *Manager) TokenFile() string {
	if m.isDev {
		return filepath.Join(m.configDir, devTokenFile)
	}
	return filepath.Join(m.configDir, defaultTokenFile)
}

// SaveToken saves the authentication token to the token file.
func (m *Manager) SaveToken(token string) error {
	return os.WriteFile(m.TokenFile(), []byte(token), 0600)
}

// LoadToken loads the authentication token from the token file.
func (m *Manager) LoadToken() (string, error) {
	data, err := os.ReadFile(m.TokenFile())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// DeleteToken deletes the authentication token from the token file.
func (m *Manager) DeleteToken() error {
	return os.Remove(m.TokenFile())
}
