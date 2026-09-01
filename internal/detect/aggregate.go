package detect

import (
	"strconv"

	"github.com/c3nk/omadev/internal/project"
)

// Finding Data keys used by detectors and read by the aggregator.
const (
	dataImage         = "image"
	dataHealth        = "health"        // "true" when a Compose healthcheck exists
	dataRole          = "role"          // "app"/"api"/"database" when evidenced
	dataService       = "service"       // owning service name (for KindPort)
	dataPublished     = "published"     // published port (for KindPort)
	dataTarget        = "target"        // target port (for KindPort)
	dataVersion       = "version"       // runtime version (for KindRuntime)
	dataSource        = "source"        // where a runtime version came from
	dataComponent     = "component"     // component path (for KindTechnology)
	dataComponentName = "componentName" // display name for a component
	dataAmbiguous     = "ambiguous"     // "true" on a KindWarning that blocks HIGH
)

// Aggregate composes findings into a normalized project model and computes overall
// confidence. It is where a future .omdev.yaml override layer would merge (D2).
func Aggregate(root, name string, findings []Finding) project.Project {
	p := project.Project{Name: name, Root: root}

	serviceIdx := map[string]int{}   // service name -> index in p.Services
	componentIdx := map[string]int{} // component path -> index in p.Components
	ambiguous := false

	// Pass 1: services, runtimes, components, execution strategy, warnings.
	for _, f := range findings {
		switch f.Kind {
		case KindExecution:
			p.ExecutionStrategy = project.ExecutionCompose
		case KindService:
			serviceIdx[f.Value] = len(p.Services)
			p.Services = append(p.Services, project.Service{
				Name:      f.Value,
				Image:     f.Data[dataImage],
				HasHealth: f.Data[dataHealth] == "true",
				Role:      f.Data[dataRole],
			})
		case KindRuntime:
			p.Runtimes = append(p.Runtimes, project.Runtime{
				Name:    f.Value,
				Version: f.Data[dataVersion],
				Source:  f.Data[dataSource],
			})
		case KindTechnology:
			path := f.Data[dataComponent]
			if path == "" {
				path = "."
			}
			i, ok := componentIdx[path]
			if !ok {
				i = len(p.Components)
				componentIdx[path] = i
				cname := f.Data[dataComponentName]
				if cname == "" {
					cname = componentName(path, name)
				}
				p.Components = append(p.Components, project.Component{Name: cname, Path: path})
			}
			p.Components[i].Technology = append(p.Components[i].Technology, f.Value)
		case KindWarning:
			p.Notes = append(p.Notes, f.Value)
			if f.Data[dataAmbiguous] == "true" {
				ambiguous = true
			}
		}
	}

	// Pass 2: ports attach to their service, and each published port yields a URL.
	for _, f := range findings {
		if f.Kind != KindPort {
			continue
		}
		pub, _ := strconv.Atoi(f.Data[dataPublished])
		target, _ := strconv.Atoi(f.Data[dataTarget])
		if i, ok := serviceIdx[f.Data[dataService]]; ok {
			p.Services[i].Ports = append(p.Services[i].Ports, project.Port{Published: pub, Target: target})
		}
		if pub > 0 {
			p.URLs = append(p.URLs, project.URL{
				Label: f.Data[dataService],
				URL:   "http://localhost:" + strconv.Itoa(pub),
			})
		}
	}

	p.Confidence = confidenceFor(p, ambiguous)
	return p
}

// confidenceFor implements the confidence rules (D9/D10). HIGH requires an
// unambiguous Compose strategy with at least one service; ambiguity blocks HIGH.
func confidenceFor(p project.Project, ambiguous bool) project.Confidence {
	compose := p.ExecutionStrategy == project.ExecutionCompose
	if compose && len(p.Services) > 0 && !ambiguous {
		return project.ConfidenceHigh
	}
	if compose || len(p.Components) > 0 || len(p.Runtimes) > 0 || len(p.Services) > 0 {
		return project.ConfidenceMedium
	}
	return project.ConfidenceLow
}

func componentName(path, projectName string) string {
	if path == "." || path == "" {
		return projectName
	}
	return path
}
