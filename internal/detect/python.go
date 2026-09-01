package detect

import (
	"regexp"
	"sort"
)

// PythonDetector inspects Python dependency files. To avoid adding a TOML parser it
// scans file contents for evidence (dependency names, requires-python); it never
// executes Python tooling. It also emits FastAPI/pytest (rather than separate
// detectors).
type PythonDetector struct{}

func (PythonDetector) Name() string { return "python" }

var pythonFiles = []string{"pyproject.toml", "requirements.txt", "Pipfile"}
var lockManagers = map[string]string{"uv.lock": "uv", "poetry.lock": "poetry"}
var requiresPythonRE = regexp.MustCompile(`requires-python\s*=\s*["'][^0-9]*([0-9]+\.[0-9]+)`)

type pyEvidence struct {
	content string
	manager string
	version string
}

func (d PythonDetector) Detect(ctx Context) ([]Finding, error) {
	dirs := map[string]*pyEvidence{}
	ev := func(dir string) *pyEvidence {
		if dirs[dir] == nil {
			dirs[dir] = &pyEvidence{}
		}
		return dirs[dir]
	}

	for _, name := range pythonFiles {
		for _, p := range findFilesNamed(ctx.FS, name) {
			e := ev(dirOf(p))
			if text, err := readText(ctx.FS, p); err == nil {
				e.content += "\n" + text
				if name == "Pipfile" && e.manager == "" {
					e.manager = "pipenv"
				}
				if m := requiresPythonRE.FindStringSubmatch(text); m != nil && e.version == "" {
					e.version = m[1]
				}
			}
		}
	}
	for lock, manager := range lockManagers {
		for _, p := range findFilesNamed(ctx.FS, lock) {
			ev(dirOf(p)).manager = manager
		}
	}

	var out []Finding
	for _, dir := range sortedKeys(dirs) {
		e := dirs[dir]
		out = append(out, techFinding(d.Name(), "Python", dir))

		manager := e.manager
		if manager == "" {
			manager = "pip"
		}
		out = append(out, techFinding(d.Name(), manager, dir))

		for _, dep := range []struct{ needle, label string }{
			{"fastapi", "FastAPI"}, {"django", "Django"}, {"flask", "Flask"}, {"pytest", "pytest"},
		} {
			if containsFold(e.content, dep.needle) {
				out = append(out, techFinding(d.Name(), dep.label, dir))
			}
		}

		if e.version != "" {
			out = append(out, Finding{Kind: KindRuntime, Detector: d.Name(), Value: "Python", Data: map[string]string{
				dataVersion: e.version, dataSource: "pyproject.toml",
			}})
		}
	}
	return out, nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
