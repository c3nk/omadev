package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/c3nk/omadev/internal/exec"
	"github.com/c3nk/omadev/internal/exit"
	"github.com/c3nk/omadev/internal/plan"
	"github.com/c3nk/omadev/internal/ui"
	"github.com/spf13/cobra"
)

func newDownCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "down [path]",
		Short: "Stop the development environment (data is preserved)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			noColor, _ := cmd.Flags().GetBool("no-color")
			return runDown(cmd.Context(), argOrDot(args), cmd.OutOrStdout(), cmd.InOrStdin(), noColor, exec.OS{})
		},
	}
}

func runDown(ctx context.Context, start string, w io.Writer, in io.Reader, noColor bool, e exec.Executor) error {
	pr, err := prepare(start, w, noColor, e)
	if err != nil {
		return err
	}
	defer pr.close()

	if !pr.isComposeProject() {
		pr.p.Warn("This is not a Docker Compose project; nothing to stop.")
		return exit.New(exit.Unsupported, errors.New("not a compose project"))
	}
	if err := pr.ensureDocker(ctx, e); err != nil {
		return err
	}

	plan.Render(w, plan.ComposeDown(pr.comp))

	ok, err := ui.Confirm(in, w, "Continue?")
	if err != nil {
		return err
	}
	if !ok {
		pr.p.Println("Cancelled. Nothing was stopped.")
		return exit.New(exit.Canceled, errors.New("cancelled by user"))
	}

	code, err := pr.comp.Down(ctx, exec.Stdio{In: in, Out: w, Err: w})
	if err != nil {
		return err
	}
	if code != 0 {
		return exit.New(exit.ExecFailure, fmt.Errorf("docker compose down exited with status %d", code))
	}

	pr.p.Println("")
	pr.p.Println("Environment stopped. Development data was preserved.")
	return nil
}
