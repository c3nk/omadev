package detect

import (
	"encoding/json"
	"io/fs"
)

// NodeDetector inspects package.json files. It also emits React/TypeScript/Vite/
// Vitest technologies (rather than separate detectors) so each file is read once. It
// never runs a package manager.
type NodeDetector struct{}

func (NodeDetector) Name() string { return "node" }

type packageJSON struct {
	Scripts         map[string]string `json:"scripts"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

func (d NodeDetector) Detect(ctx Context) ([]Finding, error) {
	var out []Finding
	for _, p := range findFilesNamed(ctx.FS, "package.json") {
		component := dirOf(p)

		raw, err := fs.ReadFile(ctx.FS, p)
		if err != nil {
			continue
		}
		var pkg packageJSON
		if err := json.Unmarshal(raw, &pkg); err != nil {
			out = append(out, note(d.Name(), "package.json at "+p+" could not be parsed"))
			continue
		}

		node := Finding{Kind: KindTechnology, Detector: d.Name(), Value: "Node.js", Data: map[string]string{dataComponent: component}}
		if dev := pkg.Scripts["dev"]; dev != "" {
			node.Data["devScript"] = dev
		}
		out = append(out, node)

		if pm := nodePackageManager(ctx.FS, component); pm != "" {
			out = append(out, techFinding(d.Name(), pm, component))
		}
		for _, name := range []string{"react", "vue", "svelte", "vite", "typescript", "vitest", "next"} {
			if hasDep(pkg, name) {
				out = append(out, techFinding(d.Name(), nodeTechLabel(name), component))
			}
		}
	}
	return out, nil
}

func hasDep(pkg packageJSON, name string) bool {
	if _, ok := pkg.Dependencies[name]; ok {
		return true
	}
	_, ok := pkg.DevDependencies[name]
	return ok
}

// nodePackageManager infers the package manager from a lockfile in the component dir.
func nodePackageManager(fsys fs.FS, component string) string {
	locks := map[string]string{
		"pnpm-lock.yaml":    "pnpm",
		"yarn.lock":         "yarn",
		"package-lock.json": "npm",
		"bun.lockb":         "bun",
	}
	for lock, pm := range locks {
		p := lock
		if component != "." {
			p = component + "/" + lock
		}
		if info, err := fs.Stat(fsys, p); err == nil && !info.IsDir() {
			return pm
		}
	}
	return ""
}

func nodeTechLabel(dep string) string {
	switch dep {
	case "react":
		return "React"
	case "vue":
		return "Vue"
	case "svelte":
		return "Svelte"
	case "vite":
		return "Vite"
	case "typescript":
		return "TypeScript"
	case "vitest":
		return "Vitest"
	case "next":
		return "Next.js"
	default:
		return dep
	}
}

// tech is a KindTechnology finding attached to a component.
func techFinding(detector, value, component string) Finding {
	return Finding{Kind: KindTechnology, Detector: detector, Value: value, Data: map[string]string{dataComponent: component}}
}
