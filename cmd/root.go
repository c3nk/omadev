// Package cmd wires the omadev command-line interface. Command handlers stay thin
// and delegate to internal services; this file is the composition root.
package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/c3nk/omadev/internal/exit"
	"github.com/c3nk/omadev/internal/logging"
	"github.com/spf13/cobra"
)

// version is the build-stamped version of the binary. It is overridden at release
// time via -ldflags "-X github.com/c3nk/omadev/cmd.version=...".
var version = "0.0.0-dev"

// newRootCmd builds the root command. Bare `omadev` will eventually print a project
// overview; until the inspect command lands it prints help.
func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "omadev",
		Short: "A developer experience CLI designed for Omarchy",
		Long: "Omadev inspects a repository, explains how its development environment runs,\n" +
			"and safely starts and inspects it. It understands your environment; it does\n" +
			"not replace Docker, Compose, mise, or your project's own scripts.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true, // Execute owns error presentation
		// Configure logging from the global verbosity flags before any command runs.
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			verbose, _ := cmd.Flags().GetBool("verbose")
			debug, _ := cmd.Flags().GetBool("debug")
			slog.SetDefault(logging.New(os.Stderr, logging.Options{Verbose: verbose, Debug: debug}))
			return nil
		},
		// Until the inspect overview lands, bare `omadev` prints help. This also makes
		// the root command runnable so `--help` renders the full usage and flags.
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// Global flags. No -v shorthand for --verbose: cobra reserves -v for --version.
	root.PersistentFlags().Bool("verbose", false, "enable verbose (info-level) logging")
	root.PersistentFlags().Bool("debug", false, "enable debug-level logging (never prints secrets)")
	root.PersistentFlags().Bool("no-color", false, "disable colored output")

	root.AddCommand(newInspectCmd())

	return root
}

// Execute runs the root command and exits the process with the mapped exit code.
// Commands present their own user-facing output for controlled outcomes (an
// *exit.Error); any other error is unexpected and is printed here.
func Execute() {
	err := newRootCmd().Execute()
	if err != nil {
		var coded *exit.Error
		if !errors.As(err, &coded) {
			fmt.Fprintln(os.Stderr, "Error:", err)
		}
	}
	os.Exit(int(exit.CodeOf(err)))
}
