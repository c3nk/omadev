package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"

	"github.com/c3nk/omadev/internal/checks"
	"github.com/c3nk/omadev/internal/compose"
	"github.com/c3nk/omadev/internal/detect"
	"github.com/c3nk/omadev/internal/exec"
	"github.com/c3nk/omadev/internal/exit"
	"github.com/c3nk/omadev/internal/project"
	"github.com/c3nk/omadev/internal/ui"
)

// prepared holds the shared setup the Compose commands need.
type prepared struct {
	p     *ui.Printer
	proj  project.Project
	comp  *compose.Compose
	fsys  fs.FS
	root  string
	close func() error
}

// prepare detects the project at start and builds the shared services. The caller
// must call close when done.
func prepare(start string, w io.Writer, noColor bool, e exec.Executor) (*prepared, error) {
	p := ui.New(w, noColor)

	abs, err := filepath.Abs(start)
	if err != nil {
		return nil, err
	}
	root, ok := project.FindRoot(abs)
	if !ok {
		p.Warn("No recognized project was found here.")
		return nil, exit.New(exit.Unsupported, errors.New("no recognized project"))
	}

	ctx, r, err := detect.OpenContext(root)
	if err != nil {
		return nil, err
	}
	findings, err := detect.DefaultRegistry().Run(ctx)
	if err != nil {
		r.Close()
		return nil, err
	}

	return &prepared{
		p:     p,
		proj:  detect.Aggregate(root, filepath.Base(root), findings),
		comp:  compose.New(e, root),
		fsys:  ctx.FS,
		root:  root,
		close: r.Close,
	}, nil
}

// isComposeProject reports whether execution can proceed via Compose.
func (pr *prepared) isComposeProject() bool {
	return pr.proj.ExecutionStrategy == project.ExecutionCompose
}

// ensureDocker verifies Docker and applies privilege detection (D4). It returns a
// prerequisite error when Docker is unavailable, and it never changes any setting.
func (pr *prepared) ensureDocker(ctx context.Context, e exec.Executor) error {
	dc := checks.Docker(ctx, e)
	switch dc.Status {
	case checks.DockerUnavailable:
		pr.p.Warn("Docker is not available.")
		if dc.Detail != "" {
			pr.p.Println("  " + dc.Detail)
		}
		pr.p.Println("  Install or start Docker, then try again.")
		return exit.New(exit.MissingPrereq, errors.New("docker unavailable"))
	case checks.DockerNeedsPrivilege:
		pr.comp.SetSudo(true)
		pr.p.Println("Docker requires elevated access; commands will use sudo.")
	}
	return compose.CheckAvailable(ctx, e)
}

// renderStatuses prints per-service state and published ports.
func (pr *prepared) renderStatuses(statuses []compose.ServiceStatus) {
	if len(statuses) == 0 {
		pr.p.Println("No services are running.")
		return
	}
	pr.p.Println("Services")
	for _, s := range statuses {
		pr.p.Println(fmt.Sprintf("  %-10s %s", s.State, s.Name))
	}
	ports := false
	for _, s := range statuses {
		if len(s.Ports) > 0 {
			ports = true
		}
	}
	if ports {
		pr.p.Println("")
		pr.p.Println("Ports")
		for _, s := range statuses {
			for _, port := range s.Ports {
				pr.p.Println(fmt.Sprintf("  %-10s %d", s.Name, port))
			}
		}
	}
}

// reportMissingEnv warns when .env is missing, listing only the variable names it
// can infer. It never reads values and never creates or modifies .env.
func reportMissingEnv(pr *prepared) {
	envc := checks.Env(pr.fsys)
	if !envc.MissingEnv() {
		return
	}
	pr.p.Warn(".env not found (.env.example is present)")
	if len(envc.Vars) > 0 {
		pr.p.Println("  Possible required variables:")
		for _, v := range envc.Vars {
			pr.p.Println("    " + v)
		}
	}
	pr.p.Println("")
}

// argOrDot returns the first positional argument, or "." when none was given.
func argOrDot(args []string) string {
	if len(args) == 1 {
		return args[0]
	}
	return "."
}

func runningCount(statuses []compose.ServiceStatus) int {
	n := 0
	for _, s := range statuses {
		if s.State == compose.StateRunning || s.State == compose.StateHealthy {
			n++
		}
	}
	return n
}
