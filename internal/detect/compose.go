package detect

import (
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"github.com/c3nk/omadev/internal/project"
	"gopkg.in/yaml.v3"
)

// composeBaseNames and composeOverrideNames are recognized in precedence order.
var (
	composeBaseNames     = []string{"compose.yaml", "compose.yml", "docker-compose.yaml", "docker-compose.yml"}
	composeOverrideNames = []string{"compose.override.yaml", "compose.override.yml", "docker-compose.override.yaml", "docker-compose.override.yml"}
)

// ComposeDetector reads a Docker Compose file without Docker and without resolving
// .env/${VAR} interpolation (D3). It owns Compose-image-based infrastructure hints
// (e.g. PostgreSQL) so the file is parsed once.
type ComposeDetector struct{}

func (ComposeDetector) Name() string { return "compose" }

type composeFile struct {
	Services map[string]composeService `yaml:"services"`
}

type composeService struct {
	Image       string      `yaml:"image"`
	Ports       []yaml.Node `yaml:"ports"`
	Healthcheck *yaml.Node  `yaml:"healthcheck"`
	Profiles    []string    `yaml:"profiles"`
}

func (d ComposeDetector) Detect(ctx Context) ([]Finding, error) {
	bases := existingFiles(ctx.FS, composeBaseNames)
	if len(bases) == 0 {
		return nil, nil
	}
	overrides := existingFiles(ctx.FS, composeOverrideNames)

	out := []Finding{{Kind: KindTechnology, Detector: d.Name(), Value: "Docker Compose"}}

	if len(bases) > 1 {
		out = append(out, ambiguity(d.Name(), "multiple Docker Compose base files present"))
	}
	if len(overrides) > 1 {
		out = append(out, ambiguity(d.Name(), "multiple Docker Compose override files present"))
	}

	data, err := fs.ReadFile(ctx.FS, bases[0])
	if err != nil {
		return nil, err
	}

	var cf composeFile
	if err := yaml.Unmarshal(data, &cf); err != nil {
		w := ambiguity(d.Name(), "Docker Compose file "+bases[0]+" could not be parsed")
		w.Data["invalid"] = "true"
		return append(out, w), nil
	}

	if hasProfiles(cf) {
		out = append(out, ambiguity(d.Name(), "Docker Compose profiles are not supported in v0.1"))
	}
	if len(overrides) == 1 {
		out = append(out, note(d.Name(), "merged "+overrides[0]))
	}

	out = append(out, Finding{Kind: KindExecution, Detector: d.Name(), Value: "Docker Compose"})

	for _, name := range sortedServiceNames(cf.Services) {
		svc := cf.Services[name]
		sf := Finding{Kind: KindService, Detector: d.Name(), Value: name, Data: map[string]string{}}
		if svc.Image != "" {
			sf.Data[dataImage] = svc.Image
		}
		if svc.Healthcheck != nil {
			sf.Data[dataHealth] = "true"
		}
		if role := roleForImage(svc.Image); role != "" {
			sf.Data[dataRole] = role
		}
		out = append(out, sf)

		if isPostgresImage(svc.Image) {
			out = append(out, Finding{Kind: KindTechnology, Detector: d.Name(), Value: "PostgreSQL"})
		}

		for _, p := range parsePorts(svc.Ports) {
			out = append(out, Finding{Kind: KindPort, Detector: d.Name(), Data: map[string]string{
				dataService:   name,
				dataPublished: strconv.Itoa(p.published),
				dataTarget:    strconv.Itoa(p.target),
			}})
		}
	}

	return out, nil
}

func hasProfiles(cf composeFile) bool {
	for _, s := range cf.Services {
		if len(s.Profiles) > 0 {
			return true
		}
	}
	return false
}

type portMapping struct{ published, target int }

// parsePorts extracts published ports from Compose port entries. It handles the
// common short string form ("5173:5173", "127.0.0.1:8000:8000") and the long
// mapping form. Entries whose published port is not a concrete number (e.g. it uses
// ${VAR} interpolation, which is not resolved) are skipped.
func parsePorts(nodes []yaml.Node) []portMapping {
	var out []portMapping
	for i := range nodes {
		n := &nodes[i]
		switch n.Kind {
		case yaml.ScalarNode:
			if pm, ok := parseShortPort(n.Value); ok {
				out = append(out, pm)
			}
		case yaml.MappingNode:
			var long struct {
				Published int `yaml:"published"`
				Target    int `yaml:"target"`
			}
			if err := n.Decode(&long); err == nil && long.Published > 0 {
				out = append(out, portMapping{published: long.Published, target: long.Target})
			}
		}
	}
	return out
}

func parseShortPort(s string) (portMapping, bool) {
	parts := strings.Split(s, ":")
	if len(parts) < 2 {
		return portMapping{}, false // container-only port, no stable published port
	}
	pub, err := strconv.Atoi(parts[len(parts)-2])
	if err != nil {
		return portMapping{}, false // e.g. ${PORT}: not resolved
	}
	target, _ := strconv.Atoi(parts[len(parts)-1])
	return portMapping{published: pub, target: target}, true
}

var databaseImages = map[string]string{
	"postgres": "database", "mysql": "database", "mariadb": "database", "mongo": "database",
}

func roleForImage(image string) string {
	return databaseImages[imageBase(image)]
}

func isPostgresImage(image string) bool {
	return imageBase(image) == "postgres"
}

// imageBase returns the bare image name without registry, path, or tag.
func imageBase(image string) string {
	if image == "" {
		return ""
	}
	name := image
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	if i := strings.IndexAny(name, ":@"); i >= 0 {
		name = name[:i]
	}
	return name
}

func sortedServiceNames(services map[string]composeService) []string {
	names := make([]string, 0, len(services))
	for n := range services {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func existingFiles(fsys fs.FS, names []string) []string {
	var out []string
	for _, name := range names {
		if info, err := fs.Stat(fsys, name); err == nil && !info.IsDir() {
			out = append(out, name)
		}
	}
	return out
}

func ambiguity(detector, msg string) Finding {
	return Finding{Kind: KindWarning, Detector: detector, Value: msg, Confidence: project.ConfidenceLow, Data: map[string]string{dataAmbiguous: "true"}}
}

func note(detector, msg string) Finding {
	return Finding{Kind: KindWarning, Detector: detector, Value: msg, Data: map[string]string{}}
}
