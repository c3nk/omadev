package ui

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrinterPlainWhenNotTerminal(t *testing.T) {
	var buf bytes.Buffer
	p := New(&buf, false) // a *bytes.Buffer is not a terminal

	if p.Color() {
		t.Fatal("color must be disabled when the writer is not a terminal")
	}

	p.OK("Docker Compose")
	p.Warn(".env not found")
	out := buf.String()

	if strings.Contains(out, "\x1b[") {
		t.Errorf("plain output must not contain ANSI escape codes, got: %q", out)
	}
	if !strings.Contains(out, "ok Docker Compose") {
		t.Errorf("expected plain ok line, got: %q", out)
	}
	if !strings.Contains(out, "warn .env not found") {
		t.Errorf("expected plain warn line, got: %q", out)
	}
}

func TestConfirm(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{"empty defaults to yes", "\n", true},
		{"y is yes", "y\n", true},
		{"yes is yes", "yes\n", true},
		{"uppercase Y is yes", "Y\n", true},
		{"n is no", "n\n", false},
		{"uppercase N is no", "N\n", false},
		{"no is no", "no\n", false},
		{"garbage is no", "maybe\n", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var out bytes.Buffer
			got, err := Confirm(strings.NewReader(c.input), &out, "Continue?")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != c.want {
				t.Errorf("Confirm(%q) = %v, want %v", c.input, got, c.want)
			}
			if !strings.Contains(out.String(), "[Y/n]") {
				t.Errorf("prompt should show [Y/n], got: %q", out.String())
			}
		})
	}
}
