package project

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func TestFindRoot_ComposeAtStart(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "compose.yaml"))
	write(t, filepath.Join(dir, ".git", "HEAD"))

	got, ok := FindRoot(dir)
	if !ok || got != dir {
		t.Fatalf("FindRoot = (%q, %v), want (%q, true)", got, ok, dir)
	}
}

func TestFindRoot_ComposeInParent(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "compose.yaml"))
	write(t, filepath.Join(root, ".git", "HEAD"))
	sub := filepath.Join(root, "services", "web")
	mkdir(t, sub)

	got, ok := FindRoot(sub)
	if !ok || got != root {
		t.Fatalf("FindRoot = (%q, %v), want (%q, true)", got, ok, root)
	}
}

func TestFindRoot_MarkerWhenNoCompose(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "package.json"))
	write(t, filepath.Join(dir, ".git", "HEAD"))

	got, ok := FindRoot(dir)
	if !ok || got != dir {
		t.Fatalf("FindRoot = (%q, %v), want (%q, true)", got, ok, dir)
	}
}

func TestFindRoot_NothingFound(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, ".git", "HEAD")) // repo boundary, but no compose/marker
	sub := filepath.Join(root, "empty")
	mkdir(t, sub)

	if got, ok := FindRoot(sub); ok {
		t.Fatalf("FindRoot = (%q, true), want not found", got)
	}
}

func TestFindRoot_StopsAtGitBoundary(t *testing.T) {
	outer := t.TempDir()
	write(t, filepath.Join(outer, "compose.yaml")) // above the repo boundary
	repo := filepath.Join(outer, "repo")
	write(t, filepath.Join(repo, ".git", "HEAD"))
	sub := filepath.Join(repo, "sub")
	mkdir(t, sub)

	// The compose file in `outer` must not be found: .git in `repo` bounds the walk.
	if got, ok := FindRoot(sub); ok {
		t.Fatalf("FindRoot = (%q, true), want not found (blocked by .git boundary)", got)
	}
}
