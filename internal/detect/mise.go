package detect

import (
	"regexp"
	"strings"
)

// MiseDetector reads mise.toml and .tool-versions for runtime version hints. It is
// informational in v0.1 and never installs runtimes.
type MiseDetector struct{}

func (MiseDetector) Name() string { return "mise" }

// miseToolRE matches simple `name = "version"` lines (mise.toml [tools] entries).
var miseToolRE = regexp.MustCompile(`^\s*([A-Za-z0-9_.-]+)\s*=\s*["']([^"']+)["']`)

func (d MiseDetector) Detect(ctx Context) ([]Finding, error) {
	var out []Finding
	out = append(out, d.fromToolVersions(ctx)...)
	out = append(out, d.fromMiseToml(ctx)...)
	return out, nil
}

func (d MiseDetector) fromToolVersions(ctx Context) []Finding {
	text, err := readText(ctx.FS, ".tool-versions")
	if err != nil {
		return nil
	}
	var out []Finding
	for _, line := range strings.Split(text, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		out = append(out, runtimeFinding(d.Name(), fields[0], fields[1], ".tool-versions"))
	}
	return out
}

func (d MiseDetector) fromMiseToml(ctx Context) []Finding {
	text, err := readText(ctx.FS, "mise.toml")
	if err != nil {
		return nil
	}
	var out []Finding
	inTools := false
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[") {
			inTools = trimmed == "[tools]"
			continue
		}
		if !inTools {
			continue
		}
		if m := miseToolRE.FindStringSubmatch(line); m != nil {
			out = append(out, runtimeFinding(d.Name(), m[1], m[2], "mise.toml"))
		}
	}
	return out
}

func runtimeFinding(detector, name, version, source string) Finding {
	return Finding{Kind: KindRuntime, Detector: detector, Value: runtimeLabel(name), Data: map[string]string{
		dataVersion: version, dataSource: source,
	}}
}

func runtimeLabel(name string) string {
	switch strings.ToLower(name) {
	case "node", "nodejs":
		return "Node"
	case "python":
		return "Python"
	case "go", "golang":
		return "Go"
	case "ruby":
		return "Ruby"
	case "rust":
		return "Rust"
	default:
		return name
	}
}
