package detect

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/c3nk/omadev/internal/project"
)

type stubDetector struct {
	name     string
	findings []Finding
}

func (s stubDetector) Name() string { return s.name }
func (s stubDetector) Detect(Context) ([]Finding, error) {
	return s.findings, nil
}

func TestRegistryRun(t *testing.T) {
	a := stubDetector{name: "a", findings: []Finding{{Kind: KindTechnology, Detector: "a", Value: "React", Confidence: project.ConfidenceMedium}}}
	b := stubDetector{name: "b", findings: []Finding{{Kind: KindTechnology, Detector: "b", Value: "FastAPI", Confidence: project.ConfidenceMedium}}}

	reg := NewRegistry(a)
	reg.Register(b)

	got, err := reg.Run(Context{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got[0].Value != "React" || got[1].Value != "FastAPI" {
		t.Fatalf("unexpected findings: %+v", got)
	}
}

func TestOpenContextReadsWithinRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module x\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx, r, err := OpenContext(root)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	data, err := fs.ReadFile(ctx.FS, "go.mod")
	if err != nil {
		t.Fatalf("reading a file within root should succeed: %v", err)
	}
	if string(data) != "module x\n" {
		t.Errorf("unexpected content: %q", data)
	}
}

func TestOpenContextRefusesSymlinkEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink semantics differ on Windows")
	}
	base := t.TempDir()
	root := filepath.Join(base, "repo")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	// A secret file outside the root, and a symlink inside the root pointing to it.
	secret := filepath.Join(base, "secret.txt")
	if err := os.WriteFile(secret, []byte("top secret"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(secret, filepath.Join(root, "escape")); err != nil {
		t.Fatal(err)
	}

	ctx, r, err := OpenContext(root)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	if _, err := fs.ReadFile(ctx.FS, "escape"); err == nil {
		t.Fatal("reading through an escaping symlink must fail")
	} else if errors.Is(err, fs.ErrNotExist) {
		// Acceptable: escape is reported as not-exist rather than leaking content.
		t.Logf("escape reported as not-exist: %v", err)
	}
}
