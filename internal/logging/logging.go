// Package logging builds the structured logger used across omadev. Verbosity is
// controlled by the global --verbose/--debug flags; debug output must never include
// secret values.
package logging

import (
	"io"
	"log/slog"
)

// Options controls logger construction from the global verbosity flags.
type Options struct {
	Verbose bool
	Debug   bool
}

// LevelFor returns the log level implied by the verbosity flags. The default is
// Warn; --verbose raises it to Info; --debug raises it to Debug and wins over
// --verbose.
func LevelFor(o Options) slog.Level {
	switch {
	case o.Debug:
		return slog.LevelDebug
	case o.Verbose:
		return slog.LevelInfo
	default:
		return slog.LevelWarn
	}
}

// New builds a text logger writing to w at the level implied by o.
func New(w io.Writer, o Options) *slog.Logger {
	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: LevelFor(o)}))
}
