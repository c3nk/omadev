package project

import (
	"os"
	"path/filepath"
)

// composeFiles are the recognized Docker Compose filenames.
var composeFiles = []string{
	"compose.yaml", "compose.yml",
	"docker-compose.yaml", "docker-compose.yml",
}

// projectMarkers indicate a project root when no Compose file is present.
var projectMarkers = []string{
	"package.json", "pyproject.toml", "requirements.txt", "Pipfile",
	"go.mod", "Cargo.toml", "mise.toml", ".tool-versions",
}

// FindRoot locates the project root by walking upward from start. It returns the
// nearest ancestor directory (including start) that contains a recognized Compose
// file; if the chain has no Compose file, it returns the nearest directory with any
// project marker. The walk stops at a directory containing .git (the repository
// boundary) and never ascends above the filesystem root. It never searches
// downward. It returns ("", false) when nothing is found within the boundary.
func FindRoot(start string) (string, bool) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", false
	}

	markerDir := ""
	for {
		if hasAnyFile(dir, composeFiles) {
			return dir, true
		}
		if markerDir == "" && hasAnyFile(dir, projectMarkers) {
			markerDir = dir
		}

		atGitBoundary := isDir(filepath.Join(dir, ".git"))
		parent := filepath.Dir(dir)
		if atGitBoundary || parent == dir {
			break
		}
		dir = parent
	}

	if markerDir != "" {
		return markerDir, true
	}
	return "", false
}

func hasAnyFile(dir string, names []string) bool {
	for _, name := range names {
		info, err := os.Stat(filepath.Join(dir, name))
		if err == nil && !info.IsDir() {
			return true
		}
	}
	return false
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
