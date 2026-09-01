// Package cmd wires the omadev command-line interface. Command handlers stay thin
// and delegate to internal services; this file is the composition root.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// version is the build-stamped version of the binary. It is overridden at release
// time via -ldflags "-X github.com/c3nk/omadev/cmd.version=...".
var version = "0.0.0-dev"

// newRootCmd builds the root command. Bare `omadev` will eventually print a project
// overview; until the inspect command lands it prints help.
func newRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "omadev",
		Short: "A developer experience CLI designed for Omarchy",
		Long: "Omadev inspects a repository, explains how its development environment runs,\n" +
			"and safely starts and inspects it. It understands your environment; it does\n" +
			"not replace Docker, Compose, mise, or your project's own scripts.",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: false,
		// Until the inspect overview lands, bare `omadev` prints help. This also makes
		// the root command runnable so `--help` renders the full usage and flags.
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
}

// Execute runs the root command. Exit-code mapping is refined in a later change.
func Execute() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
