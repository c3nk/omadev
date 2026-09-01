// Package ui is the thin terminal-output layer. It adds color and status markers on
// a TTY and degrades to plain ASCII when output is redirected or color is disabled.
// There is no TUI framework; this indirection just leaves room for one later.
package ui

import (
	"fmt"
	"io"
	"os"
)

const (
	ansiReset  = "\x1b[0m"
	ansiGreen  = "\x1b[32m"
	ansiYellow = "\x1b[33m"
)

// Printer writes user-facing output, with or without color.
type Printer struct {
	w     io.Writer
	color bool
}

// New builds a Printer for w. Color is enabled only when w is a terminal, the
// NO_COLOR environment variable is unset, and noColor is false.
func New(w io.Writer, noColor bool) *Printer {
	return &Printer{w: w, color: colorEnabled(w, noColor)}
}

// Color reports whether this Printer emits color.
func (p *Printer) Color() bool { return p.color }

// Println writes a plain line.
func (p *Printer) Println(text string) {
	fmt.Fprintln(p.w, text)
}

// OK writes a success line (green check on a TTY, "ok" in plain mode).
func (p *Printer) OK(text string) {
	p.marker(ansiGreen, "✓", "ok", text)
}

// Warn writes a warning line (yellow mark on a TTY, "warn" in plain mode).
func (p *Printer) Warn(text string) {
	p.marker(ansiYellow, "⚠", "warn", text)
}

// Active writes a running-service line (green dot on a TTY, "*" in plain mode).
func (p *Printer) Active(text string) {
	p.marker(ansiGreen, "●", "*", text)
}

func (p *Printer) marker(color, symbol, plain, text string) {
	if p.color {
		fmt.Fprintf(p.w, "%s%s%s %s\n", color, symbol, ansiReset, text)
		return
	}
	fmt.Fprintf(p.w, "%s %s\n", plain, text)
}

func colorEnabled(w io.Writer, noColor bool) bool {
	if noColor || os.Getenv("NO_COLOR") != "" {
		return false
	}
	return isTerminal(w)
}

// isTerminal reports whether w is a character device (a terminal), using only the
// standard library.
func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
