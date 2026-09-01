package detect

import (
	"io/fs"
	"path"
	"strings"
)

// findFilesNamed walks the tree rooted at ctx.FS and returns the relative paths of
// files with the given base name, skipping IgnoredDirs. The walk never leaves the
// root (the filesystem is confined) and never descends into ignored directories.
func findFilesNamed(fsys fs.FS, name string) []string {
	var out []string
	_ = fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries rather than aborting
		}
		if d.IsDir() {
			if p != "." && IgnoredDirs[d.Name()] {
				return fs.SkipDir
			}
			return nil
		}
		if d.Name() == name {
			out = append(out, p)
		}
		return nil
	})
	return out
}

// readText reads a file's contents as a string.
func readText(fsys fs.FS, p string) (string, error) {
	b, err := fs.ReadFile(fsys, p)
	return string(b), err
}

// dirOf returns the directory of a file path relative to the root, using "." for
// files at the root.
func dirOf(filePath string) string {
	dir := path.Dir(filePath)
	if dir == "" {
		return "."
	}
	return dir
}

// containsFold reports whether s contains substr, case-insensitively.
func containsFold(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
