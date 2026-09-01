package cmd

import (
	"context"
	"errors"
	"io"

	"github.com/c3nk/omadev/internal/compose"
	"github.com/c3nk/omadev/internal/exec"
	"github.com/c3nk/omadev/internal/exit"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status [path]",
		Short: "Show current service state and ports (read-only)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			noColor, _ := cmd.Flags().GetBool("no-color")
			return runStatus(cmd.Context(), argOrDot(args), cmd.OutOrStdout(), noColor, exec.OS{})
		},
	}
}

func runStatus(ctx context.Context, start string, w io.Writer, noColor bool, e exec.Executor) error {
	pr, err := prepare(start, w, noColor, e)
	if err != nil {
		return err
	}
	defer pr.close()

	if !pr.isComposeProject() {
		pr.p.Warn("This is not a Docker Compose project.")
		return exit.New(exit.Unsupported, errors.New("not a compose project"))
	}
	if err := pr.ensureDocker(ctx, e); err != nil {
		return err
	}

	res, err := pr.comp.PS(ctx)
	if err != nil {
		return err
	}

	pr.p.Println("Project: " + pr.proj.Name)
	pr.p.Println("")
	pr.renderStatuses(compose.ParsePS(res.Stdout))
	return nil
}
