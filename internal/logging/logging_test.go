package logging

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestLevelFor(t *testing.T) {
	cases := []struct {
		name string
		opts Options
		want slog.Level
	}{
		{"default is warn", Options{}, slog.LevelWarn},
		{"verbose is info", Options{Verbose: true}, slog.LevelInfo},
		{"debug is debug", Options{Debug: true}, slog.LevelDebug},
		{"debug wins over verbose", Options{Verbose: true, Debug: true}, slog.LevelDebug},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := LevelFor(c.opts); got != c.want {
				t.Errorf("LevelFor(%+v) = %v, want %v", c.opts, got, c.want)
			}
		})
	}
}

func TestNewRespectsLevel(t *testing.T) {
	var buf bytes.Buffer
	log := New(&buf, Options{}) // default: warn

	log.Info("info message")
	if buf.Len() != 0 {
		t.Errorf("info message should be suppressed at default level, got: %q", buf.String())
	}

	log.Warn("warn message")
	if !strings.Contains(buf.String(), "warn message") {
		t.Errorf("warn message should be emitted at default level, got: %q", buf.String())
	}
}
