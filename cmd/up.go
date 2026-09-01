package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/c3nk/omadev/internal/compose"
	"github.com/c3nk/omadev/internal/exec"
	"github.com/c3nk/omadev/internal/exit"
	"github.com/c3nk/omadev/internal/plan"
	"github.com/c3nk/omadev/internal/project"
	"github.com/c3nk/omadev/internal/ui"
	"github.com/spf13/cobra"
)

func newUpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "up [path]",
		Short: "Start the development environment after showing the plan and confirming",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			noColor, _ := cmd.Flags().GetBool("no-color")
			return runUp(cmd.Context(), argOrDot(args), cmd.OutOrStdout(), cmd.InOrStdin(), noColor, exec.OS{})
		},
	}
}

func runUp(ctx context.Context, start string, w io.Writer, in io.Reader, noColor bool, e exec.Executor) error {
	pr, err := prepare(start, w, noColor, e)
	if err != nil {
		return err
	}
	defer pr.close()

	// Only start when confident it is a Compose project (D10, D11).
	if !pr.isComposeProject() || pr.proj.Confidence != project.ConfidenceHigh {
		pr.p.Warn("Omadev could not confidently determine how to start this project.")
		pr.p.Println("  Run 'omadev inspect' to see what was detected.")
		return exit.New(exit.Unsupported, errors.New("cannot start: low confidence or no compose strategy"))
	}

	if err := pr.ensureDocker(ctx, e); err != nil {
		return err
	}

	reportMissingEnv(pr)

	plan.Render(w, plan.ComposeUp(pr.comp))

	ok, err := ui.Confirm(in, w, "Continue?")
	if err != nil {
		return err
	}
	if !ok {
		pr.p.Println("Cancelled. Nothing was started.")
		return exit.New(exit.Canceled, errors.New("cancelled by user"))
	}

	code, err := pr.comp.Up(ctx, exec.Stdio{In: in, Out: w, Err: w})
	if err != nil {
		return err
	}
	if code != 0 {
		return exit.New(exit.ExecFailure, fmt.Errorf("docker compose up exited with status %d", code))
	}

	// Verify: re-query service state rather than trusting exit code 0 alone (D6).
	pr.p.Println("")
	if res, perr := pr.comp.PS(ctx); perr == nil {
		pr.renderStatuses(compose.ParsePS(res.Stdout))
	}
	pr.p.Println("")
	pr.p.Println("Environment started.")
	return nil
}
