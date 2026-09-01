package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/c3nk/omadev/internal/exit"
)

func TestInspect_ComposeProject(t *testing.T) {
	var out bytes.Buffer
	err := runInspect("../testdata/docker-compose-fastapi-react", &out, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := out.String()
	for _, want := range []string{"Docker Compose", "PostgreSQL", "FastAPI", "postgres", "Confidence: HIGH", "docker compose up -d"} {
		if !strings.Contains(got, want) {
			t.Errorf("inspect output missing %q\n---\n%s", want, got)
		}
	}
}

func TestInspect_UnknownProject(t *testing.T) {
	// Isolated temp dir with a .git boundary and nothing recognizable.
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := runInspect(dir, &out, true)

	var coded *exit.Error
	if !errors.As(err, &coded) || coded.Code != exit.Unsupported {
		t.Fatalf("expected Unsupported exit error, got %v", err)
	}
}

func TestInspect_InvalidCompose(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Unterminated quote -> parse error.
	if err := os.WriteFile(filepath.Join(dir, "compose.yaml"), []byte("services:\n  web:\n    image: \"nginx\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var out bytes.Buffer
	err := runInspect(dir, &out, true)

	var coded *exit.Error
	if !errors.As(err, &coded) || coded.Code != exit.InvalidConfig {
		t.Fatalf("expected InvalidConfig exit error, got %v", err)
	}
}
