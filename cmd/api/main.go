package main

import (
	"context"
	"database/sql"
	"expvar"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/kvnloughead/godo/internal/data"
	"github.com/kvnloughead/godo/internal/mailer"
	"github.com/kvnloughead/godo/internal/vcs"
	_ "github.com/lib/pq"
)

var (
	version = vcs.Version()
)

// The application struct is used for dependency injection.
type application struct {
	config Config
	logger *slog.Logger
	models data.Models
	mailer mailer.Mailer

	// The WaitGroup instance allows us to track goroutines in progress, to
	// prevent shutdown until they are all completed. No need for initialization,
	// the zero-valued sync.WaitGroup is useable, with counter set to 0.
	wg sync.WaitGroup
}

func main() {
	// Parse CLI flags into config struct (to be added to dependencies).
	var cfg = LoadConfig()

	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	// If -version flag is set, display version and exit.
	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		os.Exit(0)
	}

	// Create structured logger (to be added to dependencies).
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Open database connection.
	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("database connection pool established")

	// Set additional debug variables, accessible at GET /debug/vars.
	setDebugVars(db)

	app := application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.Username,
			cfg.SMTP.Password, cfg.SMTP.Sender),
	}

	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}

// openDB creates an sql.DB connection pool for the supplied DSN and returns it.
// If a connection can't be established within 5 seconds, an error is returned.
func openDB(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DB.DSN)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.DB.MaxOpenConns)
	db.SetMaxIdleConns(cfg.DB.MaxIdleConns)
	db.SetConnMaxIdleTime(cfg.DB.MaxIdleTime)

	// Create a context with an empty parent context and a 5s timeout deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt to connect to the database within the 5s lifetime of the context.
	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// The setDebugVars method publishes additional data to expvar handler. Debug
// variables are available at GET /debug/vars. Data published:
//
//   - version: the API's version number
//   - timestamp: a Unix timestamp
//   - gouroutines: the number of current goroutines running
//   - database: the result of db.Stats()
func setDebugVars(db *sql.DB) {
	expvar.NewString("version").Set(version)
	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))
	expvar.Publish("goroutine", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))
}
