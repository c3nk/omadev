package cmd

import (
	"context"
	"io"
	"path/filepath"
	"strings"

	"github.com/c3nk/omadev/internal/checks"
	"github.com/c3nk/omadev/internal/compose"
	"github.com/c3nk/omadev/internal/exec"
	"github.com/c3nk/omadev/internal/project"
)

// runOverview is the bare `omadev` command: an informational overview (D1). It never
// offers to start the environment. Outside a project it falls back to help.
func runOverview(ctx context.Context, start string, w io.Writer, noColor bool, e exec.Executor, helpFallback func() error) error {
	abs, err := filepath.Abs(start)
	if err != nil {
		return err
	}
	if _, ok := project.FindRoot(abs); !ok {
		return helpFallback()
	}

	pr, err := prepare(start, w, noColor, e)
	if err != nil {
		return err
	}
	defer pr.close()
	p := pr.p

	p.Println("OMADEV")
	p.Println("")
	p.Println("Project")
	p.Println("  " + pr.proj.Name)
	p.Println("")

	if stack := stackSummary(pr.proj); stack != "" {
		p.Println("Stack")
		p.Println("  " + stack)
		p.Println("")
	}

	if !pr.isComposeProject() {
		p.Println("Run 'omadev inspect' for details.")
		return nil
	}

	p.Println("Environment")
	p.Println("  Docker Compose")
	p.Println("")

	// Best-effort status: query only if Docker is reachable. Never offer to start.
	dc := checks.Docker(ctx, e)
	if dc.Status == checks.DockerNeedsPrivilege {
		pr.comp.SetSudo(true)
	}
	if dc.Status != checks.DockerUnavailable {
		if res, perr := pr.comp.PS(ctx); perr == nil {
			statuses := compose.ParsePS(res.Stdout)
			if runningCount(statuses) > 0 {
				pr.renderStatuses(statuses)
				p.Println("")
				p.Println("Commands")
				p.Println("  omadev logs")
				p.Println("  omadev down")
				return nil
			}
		}
	}

	p.Println("Environment is not running.")
	p.Println("  Start it with: omadev up")
	return nil
}

// stackSummary builds a short "React + FastAPI + PostgreSQL"-style line from the
// headline technologies actually detected.
func stackSummary(proj project.Project) string {
	headline := []string{"React", "Vue", "Svelte", "Next.js", "FastAPI", "Django", "Flask", "PostgreSQL", "Go", "Rust"}
	present := map[string]bool{}
	for _, t := range gatherTechnologies(proj) {
		present[t] = true
	}
	var parts []string
	for _, h := range headline {
		if present[h] {
			parts = append(parts, h)
		}
	}
	return strings.Join(parts, " + ")
}
