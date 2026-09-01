package detect

import (
	"os"
	"path/filepath"
	"testing"
)

// openTemp writes the given files (path -> content) into a temp dir and opens a
// detector Context on it.
func openTemp(t *testing.T, files map[string]string) Context {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		full := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	ctx, root, err := OpenContext(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { root.Close() })
	return ctx
}

func mustDetect(t *testing.T, d Detector, ctx Context) []Finding {
	t.Helper()
	f, err := d.Detect(ctx)
	if err != nil {
		t.Fatalf("%s detect: %v", d.Name(), err)
	}
	return f
}

func techComponent(findings []Finding, value string) (string, bool) {
	for _, f := range findings {
		if f.Kind == KindTechnology && f.Value == value {
			return f.Data[dataComponent], true
		}
	}
	return "", false
}

func TestNodeDetector_Vite(t *testing.T) {
	f := runDetector(t, "node-vite", NodeDetector{})
	for _, want := range []string{"Node.js", "npm", "Vite", "React", "TypeScript"} {
		if !has(f, KindTechnology, want) {
			t.Errorf("expected node technology %q, got %+v", want, f)
		}
	}
	for _, x := range f {
		if x.Value == "Node.js" && x.Data["devScript"] == "" {
			t.Errorf("expected a dev script to be recorded")
		}
	}
}

func TestPythonDetector_FastAPI(t *testing.T) {
	f := runDetector(t, "python-fastapi", PythonDetector{})
	for _, want := range []string{"Python", "pip", "FastAPI", "pytest"} {
		if !has(f, KindTechnology, want) {
			t.Errorf("expected python technology %q, got %+v", want, f)
		}
	}
	foundVersion := false
	for _, x := range f {
		if x.Kind == KindRuntime && x.Value == "Python" && x.Data[dataVersion] == "3.13" {
			foundVersion = true
		}
	}
	if !foundVersion {
		t.Errorf("expected Python 3.13 runtime hint, got %+v", f)
	}
}

func TestDockerfileDetector(t *testing.T) {
	ctx := openTemp(t, map[string]string{"Dockerfile": "FROM scratch\n"})
	f := mustDetect(t, DockerfileDetector{}, ctx)
	if !has(f, KindTechnology, "Dockerfile") {
		t.Errorf("expected Dockerfile finding, got %+v", f)
	}

	none := runDetector(t, "node-vite", DockerfileDetector{})
	if len(none) != 0 {
		t.Errorf("expected no Dockerfile finding for node-vite, got %+v", none)
	}
}

func TestPostgresDetector_DependencyEvidence(t *testing.T) {
	ctx := openTemp(t, map[string]string{"requirements.txt": "fastapi\npsycopg[binary]>=3\n"})
	f := mustDetect(t, PostgresDetector{}, ctx)
	if !has(f, KindTechnology, "PostgreSQL") {
		t.Errorf("expected PostgreSQL from psycopg evidence, got %+v", f)
	}
}

func TestMiseDetector(t *testing.T) {
	ctx := openTemp(t, map[string]string{
		".tool-versions": "node 24\npython 3.13\n",
		"mise.toml":      "[tools]\ngo = \"1.24\"\n",
	})
	f := mustDetect(t, MiseDetector{}, ctx)

	want := map[string]string{"Node": "24", "Python": "3.13", "Go": "1.24"}
	for _, x := range f {
		if x.Kind == KindRuntime {
			if v, ok := want[x.Value]; ok && x.Data[dataVersion] == v {
				delete(want, x.Value)
			}
		}
	}
	if len(want) != 0 {
		t.Errorf("missing runtime hints: %v (got %+v)", want, f)
	}
}
