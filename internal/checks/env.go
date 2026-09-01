package checks

import (
	"io/fs"
	"strings"
)

// EnvCheck reports environment-file prerequisites without ever reading secret values.
type EnvCheck struct {
	HasEnv     bool
	HasExample bool
	Vars       []string // variable NAMES implied by .env.example (never their values)
}

// MissingEnv reports whether an example exists but the real .env does not.
func (c EnvCheck) MissingEnv() bool { return c.HasExample && !c.HasEnv }

// Env inspects .env and .env.example at the root of fsys. It records only the
// variable names from .env.example; it never reads or returns any value.
func Env(fsys fs.FS) EnvCheck {
	c := EnvCheck{}
	if info, err := fs.Stat(fsys, ".env"); err == nil && !info.IsDir() {
		c.HasEnv = true
	}
	if data, err := fs.ReadFile(fsys, ".env.example"); err == nil {
		c.HasExample = true
		c.Vars = envNames(string(data))
	}
	return c
}

// envNames extracts variable names (the text left of the first '='), discarding
// every value. Comments and blank lines are ignored.
func envNames(content string) []string {
	seen := map[string]bool{}
	var out []string
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			continue
		}
		name := strings.TrimSpace(line[:eq]) // left of '=' only; the value is discarded
		if name != "" && !seen[name] {
			seen[name] = true
			out = append(out, name)
		}
	}
	return out
}
