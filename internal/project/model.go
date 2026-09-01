// Package project holds the normalized representation of a repository that
// detectors produce and the rest of omadev consumes. Types here carry data only;
// detection and execution live elsewhere.
package project

// Confidence expresses how sure detection is about a project. The numeric order is
// meaningful: Low < Medium < High, so gating can compare (e.g. c >= ConfidenceHigh).
type Confidence int

const (
	ConfidenceLow Confidence = iota
	ConfidenceMedium
	ConfidenceHigh
)

func (c Confidence) String() string {
	switch c {
	case ConfidenceHigh:
		return "HIGH"
	case ConfidenceMedium:
		return "MEDIUM"
	case ConfidenceLow:
		return "LOW"
	default:
		return "UNKNOWN"
	}
}

// ExecutionStrategy is how a project is meant to run. v0.1 can execute only Compose.
type ExecutionStrategy int

const (
	ExecutionNone ExecutionStrategy = iota
	ExecutionCompose
)

func (e ExecutionStrategy) String() string {
	switch e {
	case ExecutionCompose:
		return "Docker Compose"
	case ExecutionNone:
		return "None"
	default:
		return "Unknown"
	}
}

// Project is the normalized model of a repository.
type Project struct {
	Name              string
	Root              string // absolute path to the project root
	Components        []Component
	Services          []Service
	Runtimes          []Runtime
	URLs              []URL
	ExecutionStrategy ExecutionStrategy
	Confidence        Confidence
	Notes             []string // e.g. ambiguity notes, "compose.override.yaml merged"
}

// Component is a detected application part: a directory and its technologies.
type Component struct {
	Name       string   // e.g. "frontend"
	Path       string   // relative to Root
	Technology []string // e.g. ["React", "TypeScript", "Vite"]
}

// Service is a Docker Compose service.
type Service struct {
	Name      string
	Image     string
	Ports     []Port // published ports only
	HasHealth bool   // a Compose healthcheck is defined
	Role      string // "app"/"api"/"database" when evidence supports it, else ""
}

// Port is a published container port mapping.
type Port struct {
	Published int
	Target    int
}

// Runtime is a runtime version hint (informational in v0.1).
type Runtime struct {
	Name    string // "Node", "Python"
	Version string // "24", "3.13"
	Source  string // "mise.toml", ".tool-versions", "engines"
}

// URL is an evidence-backed address derived from a published port.
type URL struct {
	Label string // "app", "api"
	URL   string // e.g. http://localhost:5173
}
