// Package logger builds the application's zerolog logger.
package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// New returns a configured zerolog.Logger. When pretty is true a human-readable
// console writer is used (development); otherwise structured JSON (production).
func New(level string, pretty bool) zerolog.Logger {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)
	zerolog.TimeFieldFormat = time.RFC3339

	var w io.Writer = os.Stdout
	if pretty {
		w = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	}
	return zerolog.New(w).With().Timestamp().Logger()
}
