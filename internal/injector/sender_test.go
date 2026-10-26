package injector

import (
	"flag"
	"os"
	"testing"
)

// TestSMTPSender tests how the SMTP sender is chosen outside production.
func TestSMTPSender(t *testing.T) {
	tests := []struct {
		name     string
		sender   string
		expected string
	}{
		{
			name:     "SMTP_SENDER is used when set",
			sender:   "Test <test@example.com>",
			expected: "Test <test@example.com>",
		},
		{
			name:     "Fixed sender is used without SMTP_SENDER or the Go toolchain",
			sender:   "",
			expected: devSMTPSender,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DB_DSN", "postgres://testuser:password@localhost/testdb")
			t.Setenv("SMTP_SENDER", tt.sender)
			// An empty PATH makes `go list -m` fail, as in the runtime image.
			t.Setenv("PATH", "")

			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			os.Args = []string{"cmd"}

			cfg := LoadConfig()

			if cfg.SMTP.Sender != tt.expected {
				t.Errorf("got sender %q, want %q", cfg.SMTP.Sender, tt.expected)
			}
		})
	}
}
