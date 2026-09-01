// Package detect holds the detector framework and the individual detectors. A
// detector is a pure function over a repository: it reads files and returns typed
// findings. It never mutates the repository, never executes repository scripts, and
// never makes network calls. An aggregator turns findings into the project model.
package detect

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/c3nk/omadev/internal/project"
)

// FindingKind classifies what a finding reports.
type FindingKind int

const (
	KindTechnology FindingKind = iota // e.g. "React", "FastAPI"
	KindService                       // a Docker Compose service
	KindRuntime                       // e.g. "Node 24"
	KindExecution                     // an execution strategy is available
	KindPort                          // a published port
	KindWarning                       // e.g. "compose.yaml unparseable"
)

// Finding is a single structured observation from a detector.
type Finding struct {
	Kind       FindingKind
	Detector   string
	Value      string
	Confidence project.Confidence
	Data       map[string]string // structured extras (image, path, version, ...)
}

// IgnoredDirs are directories detectors must not descend into.
var IgnoredDirs = map[string]bool{
	".git": true, "node_modules": true, ".venv": true,
	"venv": true, "dist": true, "build": true,
}

// Context is what a detector is given: the project root path and a read-only,
// symlink-escape-safe filesystem rooted there.
type Context struct {
	Root string // absolute project root path
	FS   fs.FS  // rooted at Root; cannot be used to reach files outside it
}

// Detector inspects a repository and returns findings.
type Detector interface {
	Name() string
	Detect(ctx Context) ([]Finding, error)
}

// OpenContext builds a Context for root. The returned io.Closer must be closed when
// detection is done; the filesystem is backed by an os.Root that confines all access
// to the root, refusing symlinks that resolve outside it.
func OpenContext(root string) (Context, *os.Root, error) {
	r, err := os.OpenRoot(root)
	if err != nil {
		return Context{}, nil, err
	}
	return Context{Root: root, FS: r.FS()}, r, nil
}

// Registry is an ordered set of detectors.
type Registry struct {
	detectors []Detector
}

// NewRegistry builds a registry from the given detectors.
func NewRegistry(ds ...Detector) *Registry {
	return &Registry{detectors: append([]Detector(nil), ds...)}
}

// Register appends a detector.
func (r *Registry) Register(d Detector) {
	r.detectors = append(r.detectors, d)
}

// Detectors returns the registered detectors.
func (r *Registry) Detectors() []Detector {
	return append([]Detector(nil), r.detectors...)
}

// Run executes every detector against ctx and returns the combined findings. The
// result is order-independent across detectors; findings are appended in registry
// order for stable output.
func (r *Registry) Run(ctx Context) ([]Finding, error) {
	var all []Finding
	for _, d := range r.detectors {
		found, err := d.Detect(ctx)
		if err != nil {
			return nil, fmt.Errorf("detector %s: %w", d.Name(), err)
		}
		all = append(all, found...)
	}
	return all, nil
}
