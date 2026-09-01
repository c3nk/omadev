package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/c3nk/omadev/internal/exec"
	"github.com/c3nk/omadev/internal/exit"
	"github.com/spf13/cobra"
)

func newLogsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logs [path]",
		Short: "Tail development logs (delegates to docker compose logs -f)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			noColor, _ := cmd.Flags().GetBool("no-color")
			return runLogs(cmd.Context(), argOrDot(args), cmd.OutOrStdout(), cmd.InOrStdin(), noColor, exec.OS{})
		},
	}
}

func runLogs(ctx context.Context, start string, w io.Writer, in io.Reader, noColor bool, e exec.Executor) error {
	pr, err := prepare(start, w, noColor, e)
	if err != nil {
		return err
	}
	defer pr.close()

	if !pr.isComposeProject() {
		pr.p.Warn("This is not a Docker Compose project; no logs to show.")
		return exit.New(exit.Unsupported, errors.New("not a compose project"))
	}
	if err := pr.ensureDocker(ctx, e); err != nil {
		return err
	}

	code, err := pr.comp.Logs(ctx, exec.Stdio{In: in, Out: w, Err: w})
	if err != nil {
		return err
	}
	if code != 0 {
		return exit.New(exit.ExecFailure, fmt.Errorf("docker compose logs exited with status %d", code))
	}
	return nil
}
