// Package exec is the subprocess boundary. Commands are built as structured argv
// and never passed through a shell (no shell interpolation of repository-derived
// values). It offers two modes: Capture (run to completion, return output — used to
// verify state) and Stream (wire child stdio to the terminal — used for logs and
// interactive output). A fake implementation covers both for tests.
package exec

import (
	"context"
	"io"
)

// Command is a structured subprocess invocation.
type Command struct {
	Name string   // e.g. "docker"
	Args []string // e.g. ["compose", "up", "-d"]
	Dir  string   // working directory
	Env  []string // explicit environment; nil inherits the current process env
}

// Result is the outcome of a captured command. A non-zero ExitCode is not a Go
// error: the command ran but returned non-zero. Err is returned only when the
// command could not be started.
type Result struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

// Stdio wires a streamed command's standard streams.
type Stdio struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer
}

// Executor runs commands. Implementations must build argv without a shell.
type Executor interface {
	// Capture runs c to completion and returns its output and exit code.
	Capture(ctx context.Context, c Command) (Result, error)
	// Stream connects c's stdio to sio and returns its exit code.
	Stream(ctx context.Context, c Command, sio Stdio) (int, error)
}
