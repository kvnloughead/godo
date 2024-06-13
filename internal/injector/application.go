package injector

import (
	"database/sql"
	"log/slog"
	"sync"

	"github.com/kvnloughead/godo/internal/data"
	"github.com/kvnloughead/godo/internal/mailer"
)

// The application struct is used for dependency injection.
type Application struct {
	Config Config
	Logger *slog.Logger
	Models data.Models
	Mailer mailer.Mailer

	// The WaitGroup instance allows us to track goroutines in progress, to
	// prevent shutdown until they are all completed. No need for initialization,
	// the zero-valued sync.WaitGroup is useable, with counter set to 0.
	WG sync.WaitGroup
}

func NewApplication(cfg Config, logger *slog.Logger, db *sql.DB) *Application {
	return &Application{
		Config: cfg,
		Logger: logger,
		Models: data.NewModels(db),
		Mailer: mailer.New(cfg.SMTP.Host, cfg.SMTP.Port, cfg.SMTP.Username, cfg.SMTP.Password, cfg.SMTP.Sender),
	}
}
