package main

import (
	"flag"
	"os"
	"testing"

	"github.com/go-playground/assert/v2"
)

// TestLoadConfig tests loading configuration via environment variables and
// flags.
func TestLoadConfig(t *testing.T) {
	// Clear environment before running tests
	os.Clearenv()

	tests := []struct {
		name           string
		envVars        map[string]string
		args           []string
		expectedConfig Config
		expectedError  bool
	}{
		{
			name: "Environmental Variables Only",
			envVars: map[string]string{
				"DB_DSN":  "postgres://testuser:password@localhost/testdb",
				"PORT":    "8080",
				"VERBOSE": "true",
			},
			args: []string{},
			expectedConfig: Config{
				DB:      DatabaseConfig{DSN: "postgres://testuser:password@localhost/testdb"},
				Port:    8080,
				Verbose: BoolFlag{isSet: false, value: true},
			},
			expectedError: false,
		},
		{
			name:    "Flags Only",
			envVars: map[string]string{},
			args: []string{
				"-db-dsn", "postgres://testuser:password@localhost/testdb",
				"-port", "8080",
				"-verbose", "true",
			},
			expectedConfig: Config{
				DB:      DatabaseConfig{DSN: "postgres://testuser:password@localhost/testdb"},
				Port:    8080,
				Verbose: BoolFlag{isSet: true, value: true},
			},
			expectedError: false,
		},
		{
			name: "Flags Override Environmental Variables",
			envVars: map[string]string{
				"DB_DSN":  "postgres://testuser:password@localhost/ENV_DB",
				"PORT":    "5555",
				"VERBOSE": "true",
			},
			args: []string{
				"-db-dsn", "postgres://testuser:password@localhost/testdb",
				"-port", "8080",
				"-verbose", "false",
			},
			expectedConfig: Config{
				DB:      DatabaseConfig{DSN: "postgres://testuser:password@localhost/testdb"},
				Port:    8080,
				Verbose: BoolFlag{isSet: true, value: false},
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environmental variables.
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Reset the flag set before each set of tests.
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Set os.Args[0] to "cmd", and then append the args to os.Args.
			os.Args = append([]string{"cmd"}, tt.args...)

			var cfg = LoadConfig()

			assert.Equal(t, cfg.DB.DSN, tt.expectedConfig.DB.DSN)
			assert.Equal(t, cfg.Port, tt.expectedConfig.Port)
			assert.Equal(t, cfg.Verbose, tt.expectedConfig.Verbose)

			// Cleanup by unsetting environment variables
			for key := range tt.envVars {
				os.Unsetenv(key)
			}
		})
	}
}
