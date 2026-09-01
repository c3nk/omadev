package cmd

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/c3nk/omadev/internal/detect"
	"github.com/c3nk/omadev/internal/exit"
	"github.com/c3nk/omadev/internal/project"
	"github.com/c3nk/omadev/internal/ui"
	"github.com/spf13/cobra"
)

func newInspectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect [path]",
		Short: "Inspect the project and report how its environment runs (read-only)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			start := "."
			if len(args) == 1 {
				start = args[0]
			}
			noColor, _ := cmd.Flags().GetBool("no-color")
			return runInspect(start, cmd.OutOrStdout(), noColor)
		},
	}
}

// runInspect detects the project at start and renders a read-only overview. It never
// starts or changes anything.
func runInspect(start string, w io.Writer, noColor bool) error {
	p := ui.New(w, noColor)

	absStart, err := filepath.Abs(start)
	if err != nil {
		return err
	}

	root, ok := project.FindRoot(absStart)
	if !ok {
		p.Warn("No recognized project was found here.")
		p.Println("")
		p.Println("Omadev inspects repositories that use Docker Compose or a known project layout.")
		return exit.New(exit.Unsupported, errors.New("no recognized project"))
	}

	ctx, r, err := detect.OpenContext(root)
	if err != nil {
		return err
	}
	defer r.Close()

	findings, err := detect.DefaultRegistry().Run(ctx)
	if err != nil {
		return err
	}

	proj := detect.Aggregate(root, filepath.Base(root), findings)
	renderInspect(p, proj)

	switch {
	case detect.HasInvalidConfig(findings):
		return exit.New(exit.InvalidConfig, errors.New("invalid project configuration"))
	case isEmptyProject(proj):
		return exit.New(exit.Unsupported, errors.New("no recognized project"))
	default:
		return nil
	}
}

func renderInspect(p *ui.Printer, proj project.Project) {
	p.Println("Project: " + proj.Name)
	p.Println("")

	if techs := gatherTechnologies(proj); len(techs) > 0 {
		p.Println("Detected")
		for _, t := range techs {
			p.OK(t)
		}
		p.Println("")
	}

	if len(proj.Runtimes) > 0 {
		p.Println("Runtime")
		for _, rt := range proj.Runtimes {
			p.Println(fmt.Sprintf("  %-10s %s", rt.Name, rt.Version))
		}
		p.Println("")
	}

	if len(proj.Services) > 0 {
		p.Println("Services")
		for _, s := range proj.Services {
			p.Println("  " + s.Name)
		}
		p.Println("")
	}

	if proj.ExecutionStrategy == project.ExecutionCompose {
		p.Println("Execution")
		p.Println("  " + proj.ExecutionStrategy.String())
		p.Println("")
		p.Println("Commands")
		p.Println("  Up     docker compose up -d")
		p.Println("  Down   docker compose down")
		p.Println("  Logs   docker compose logs -f")
		p.Println("")
	}

	if len(proj.Notes) > 0 {
		p.Println("Warnings")
		for _, n := range proj.Notes {
			p.Warn(n)
		}
		p.Println("")
	}

	p.Println("Confidence: " + proj.Confidence.String())
}

func gatherTechnologies(proj project.Project) []string {
	seen := map[string]bool{}
	var out []string
	for _, c := range proj.Components {
		for _, t := range c.Technology {
			if !seen[t] {
				seen[t] = true
				out = append(out, t)
			}
		}
	}
	return out
}

func isEmptyProject(proj project.Project) bool {
	return len(proj.Components) == 0 &&
		len(proj.Services) == 0 &&
		len(proj.Runtimes) == 0 &&
		proj.ExecutionStrategy == project.ExecutionNone
}
