package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFixturesPresent is a smoke test that the detection fixtures exist. Detector
// assertions against their contents live with the detectors (M1).
func TestFixturesPresent(t *testing.T) {
	want := []string{
		"docker-compose-fastapi-react",
		"node-vite",
		"python-fastapi",
		"monorepo",
		"invalid-compose",
		"missing-env",
		"unknown-project",
	}
	for _, name := range want {
		dir := filepath.Join("testdata", name)
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("fixture %q missing: %v", name, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("fixture %q is not a directory", name)
		}
	}
}
