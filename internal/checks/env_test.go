package checks

import (
	"os"
	"strings"
	"testing"
)

func TestEnv_MissingEnvFixture(t *testing.T) {
	fsys := os.DirFS("../../testdata/missing-env")
	c := Env(fsys)

	if !c.HasExample || c.HasEnv {
		t.Fatalf("expected example present and .env absent, got %+v", c)
	}
	if !c.MissingEnv() {
		t.Error("MissingEnv should be true")
	}

	want := map[string]bool{"DATABASE_URL": true, "SECRET_KEY": true}
	for _, name := range c.Vars {
		delete(want, name)
	}
	if len(want) != 0 {
		t.Errorf("missing variable names: %v (got %v)", want, c.Vars)
	}

	// Values must never leak into the reported names.
	for _, name := range c.Vars {
		if strings.ContainsAny(name, "=:/") {
			t.Errorf("variable name %q looks like it carries a value", name)
		}
	}
}

func TestEnvNames_IgnoresCommentsAndValues(t *testing.T) {
	content := "# comment\n\nexport API_URL=http://localhost:8000\nSECRET=supersecret\nBAD\n"
	got := envNames(content)
	want := []string{"API_URL", "SECRET"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("envNames = %v, want %v", got, want)
	}
}
