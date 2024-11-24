package injector

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config is a struct containing configuration settings. These settings are
// specified as CLI flags when application starts, and have defaults provided
// in case they are omitted.
type Config struct {
	Port int
	Env  string

	// Sends full stack trace of server errors in response.
	Debug BoolFlag

	// Provides verbose logging and responses in some situations. Currently only
	// middleware.logRequest makes use of this.
	Verbose BoolFlag
	DB      DatabaseConfig

	// Limiter is a struct containing configuration for our rate Limiter.
	Limiter struct {
		RPS     float64 // Requests per second. Defaults to 2.
		Burst   int     // Max request in burst. Defaults to 4.
		Enabled bool    // Defaults to true.
	}

	// SMTP is a struct containing configuration for our SMTP server.
	SMTP struct {
		Host     string
		Port     int
		Username string
		Password string
		Sender   string
	}

	// cfg.Cors is a struct containing a string slice of trusted origins.
	// If	the slice is empty, CORS will be enabled for all origins.
	Cors struct {
		TrustedOrigins []string
	}

	APIBaseURL string
}

// DatabaseConfig is a struct that stores database configuration. The DSN field
// will be necessary to connect to the database, and will be pulled from a .env
// file if there is one.
type DatabaseConfig struct {
	DSN          string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  time.Duration
}

// BoolFlag is a struct to store boolean flags. It implements the Set method
// which is called when the flags are parsed. If a flag has been passed at the
// command line the isSet field will be set to true. This can be used to
// distinguish between a default 'false' value and an unset flag.
type BoolFlag struct {
	// If isSet is false, the flag has not been set.
	isSet bool

	// The value of the flag. If isSet is false, then this will be the default.
	value bool
}

// The Set method is called whenever flag.Parse is called. If the string
// argument can be converted into a bool, then this bool is set as the
// BoolFlag's value and isSet is set to true.
func (b *BoolFlag) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}

	b.isSet = true
	b.value = v
	return nil
}

func (b *BoolFlag) String() string {
	return fmt.Sprintf("%v", b.value)
}

// loadIntFromEnvOrFlag loads an integer valued config option and assigns it to
// the target int. This function should be called after flags are parsed with
// flag.Parse.
//
// It then checks if the target has the default value. If not, no action is
// taken, because the flags should override environmental variables. If it still
// has the default value, the function checks for a matching environmental
// variable. If it exists and can be converted into an integer, it is assigned
// to the target.
func loadIntFromEnvOrFlag(target *int, defaultVal int, envKey string) {
	if *target == defaultVal {
		if envVar, ok := os.LookupEnv(envKey); ok {
			val, err := strconv.Atoi(envVar)
			if err == nil {
				*target = val
			}
		}
	}
}

// loadDurationFromEnvOrFlag loads a time.Duration valued config option and
// assigns it to the target. This function should be called after flags are
// parsed with flag.Parse.
//
// It then checks if the target has the default value. If not, no action is
// taken, because the flags should override environmental variables. If it still
// has the default value, the function checks for a matching environmental
// variable. If it exists and can be converted into a time.Duration, it is
// assigned to the target.
func loadDurationFromEnvOrFlag(target *time.Duration, defaultVal time.Duration, envKey string) {
	if *target == defaultVal {
		if envVar, ok := os.LookupEnv(envKey); ok {
			val, err := time.ParseDuration(envVar)
			if err == nil {
				*target = val
			}
		}
	}
}

// loadDefaultlessStringSetting loads a setting that doesn't provide a functioning
// default value.
//
// target should be a pointer to a string that has already been attached to a
// call to flag.StringVar.
func loadDefaultlessStringSetting(target *string, envKey string) {
	if *target == "" {
		*target = os.Getenv(envKey)
	}
}

// getModulePathAndName retrieves the Go module path and name.
func getModulePathAndName() (string, string, error) {
	cmd := exec.Command("go", "list", "-m")
	output, err := cmd.Output()
	if err != nil {
		return "", "", err
	}

	modulePath := strings.TrimSpace(string(output))
	moduleParts := strings.Split(modulePath, "/")
	moduleName := moduleParts[len(moduleParts)-1]

	return modulePath, moduleName, nil
}

// LoadConfig loads the configuration, returning the resulting Config struct.
// In development mode, it loads from .env file and detects Go module
// information.
//
// In production, it skips .env loading and module detection.
//
// Configuration is loaded in the following order:
//
// 1. Default values
// 2. Environment variables (including .env file in development)
// 3. Command line flags (these take highest precedence)
//
// The -db-dsn flag must be provided either as an environmental variable or
// flag, as it has no default value.
func LoadConfig() Config {
	env := os.Getenv("ENV")
	var modulePath, moduleName string

	// Only load .env file and get module info in non-production environments
	if env != "production" {
		err := godotenv.Load()
		if err != nil {
			log.Print("Error loading .env file:", err)
		}

		modulePath, moduleName, err = getModulePathAndName()
		if err != nil {
			log.Print("Error getting Go module path and name:", err)
		}
	}

	var cfg Config

	flag.IntVar(&cfg.Port, "port", 4000, "The port to run the app on.")
	flag.StringVar(&cfg.Env,
		"env",
		"development",
		"Environment (development|staging|production)")
	flag.Var(&cfg.Debug, "debug", "Run in debug mode")
	flag.Var(&cfg.Verbose, "verbose", "Provide verbose logging")

	// Read DB-related settings from CLI flags.
	flag.StringVar(&cfg.DB.DSN, "db-dsn", "", "Postgresql DSN")
	flag.IntVar(&cfg.DB.MaxOpenConns, "db-max-open-conns", 25, "Postgresql max open connections")
	flag.IntVar(&cfg.DB.MaxIdleConns, "db-max-idle-conns", 25, "Postgresql max idle connections")
	flag.DurationVar(&cfg.DB.MaxIdleTime, "db-max-idle-time", 15*time.Minute, "Postgresql max connection idle time")

	// Read SMTP related settings from CLI flags. The defaults are derived from
	// the Mailtrap server we are using for testing.
	flag.StringVar(&cfg.SMTP.Host, "smtp-host", "sandbox.smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.SMTP.Port, "smtp-port", 25, "SMTP server port")
	flag.StringVar(&cfg.SMTP.Username, "smtp-username", "", "SMTP username")
	flag.StringVar(&cfg.SMTP.Password, "smtp-password", "", "SMTP password")

	if env != "production" {
		flag.StringVar(&cfg.SMTP.Sender, "smtp-sender", fmt.Sprintf("%s <no-reply@%s>", moduleName, modulePath), "SMTP sender")
	} else {
		flag.StringVar(&cfg.SMTP.Sender, "smtp-sender", fmt.Sprintf("%s <no-reply@%s>", moduleName, modulePath), "SMTP sender")
		// CLI related settings.
		flag.StringVar(&cfg.APIBaseURL, "api-base-url", "http://localhost:4000", "Base url that API runs on")
	}

	flag.Parse()

	// Load settings that don't have defaults provided. Suitable values must be
	// provided as either flags or environmental variables.
	loadDefaultlessStringSetting(&cfg.DB.DSN, "DB_DSN")
	loadDefaultlessStringSetting(&cfg.SMTP.Username, "SMTP_USERNAME")
	loadDefaultlessStringSetting(&cfg.SMTP.Password, "SMTP_PASSWORD")
	loadDefaultlessStringSetting(&cfg.SMTP.Sender, "SMTP_SENDER")

	// Load integer and duration valued configuration options.
	loadIntFromEnvOrFlag(&cfg.Port, 4000, "PORT")
	loadIntFromEnvOrFlag(&cfg.DB.MaxOpenConns, 25, "DB_MAX_OPEN_CONNS")
	loadIntFromEnvOrFlag(&cfg.DB.MaxIdleConns, 25, "DB_MAX_IDLE_CONNS")
	loadDurationFromEnvOrFlag(&cfg.DB.MaxIdleTime, 15*time.Minute, "DB_MAX_IDLE_TIME")

	// Load Boolean valued configuration options.
	if !cfg.Verbose.isSet {
		cfg.Verbose.value = os.Getenv("VERBOSE") == "true"
	}
	if !cfg.Debug.isSet {
		cfg.Debug.value = os.Getenv("DEBUG") == "true"
	}

	return cfg
}
