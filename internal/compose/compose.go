// Package compose builds and runs Docker Compose v2 commands, delegating to the real
// `docker compose` binary rather than reimplementing it. Only v2 is supported; a
// legacy-only environment is reported as a prerequisite error.
package compose

import (
	"context"
	"errors"

	"github.com/c3nk/omadev/internal/exec"
	"github.com/c3nk/omadev/internal/exit"
)

// Compose runs Docker Compose commands for a project directory.
type Compose struct {
	exec exec.Executor
	dir  string
	sudo bool // set by privilege detection when elevation is required
}

// New builds a Compose bound to an executor and project directory.
func New(e exec.Executor, dir string) *Compose {
	return &Compose{exec: e, dir: dir}
}

// SetSudo records whether commands must be prefixed with sudo.
func (c *Compose) SetSudo(sudo bool) { c.sudo = sudo }

// cmd builds a `docker compose <args>` command, prefixed with sudo when required.
func (c *Compose) cmd(args ...string) exec.Command {
	full := append([]string{"compose"}, args...)
	if c.sudo {
		return exec.Command{Name: "sudo", Args: append([]string{"docker"}, full...), Dir: c.dir}
	}
	return exec.Command{Name: "docker", Args: full, Dir: c.dir}
}

// Command builders (exposed so the planner can display the exact command).
func (c *Compose) UpCmd() exec.Command   { return c.cmd("up", "-d") }
func (c *Compose) DownCmd() exec.Command { return c.cmd("down") }
func (c *Compose) LogsCmd() exec.Command { return c.cmd("logs", "-f") }
func (c *Compose) PSCmd() exec.Command   { return c.cmd("ps", "--format", "json") }

// Up starts the environment (streamed so the user sees progress).
func (c *Compose) Up(ctx context.Context, sio exec.Stdio) (int, error) {
	return c.exec.Stream(ctx, c.UpCmd(), sio)
}

// Down stops the environment. It never removes volumes.
func (c *Compose) Down(ctx context.Context, sio exec.Stdio) (int, error) {
	return c.exec.Stream(ctx, c.DownCmd(), sio)
}

// Logs tails logs (streamed).
func (c *Compose) Logs(ctx context.Context, sio exec.Stdio) (int, error) {
	return c.exec.Stream(ctx, c.LogsCmd(), sio)
}

// PS captures service status as JSON for parsing.
func (c *Compose) PS(ctx context.Context) (exec.Result, error) {
	return c.exec.Capture(ctx, c.PSCmd())
}

// CheckAvailable verifies Docker Compose v2 is usable. If only legacy docker-compose
// (v1) is present, it returns a prerequisite error explaining that v2 is required.
func CheckAvailable(ctx context.Context, e exec.Executor) error {
	if res, err := e.Capture(ctx, exec.Command{Name: "docker", Args: []string{"compose", "version"}}); err == nil && res.ExitCode == 0 {
		return nil
	}
	if res, err := e.Capture(ctx, exec.Command{Name: "docker-compose", Args: []string{"version"}}); err == nil && res.ExitCode == 0 {
		return exit.New(exit.MissingPrereq, errors.New("only legacy docker-compose (v1) was found; Omadev requires Docker Compose v2 (docker compose)"))
	}
	return exit.New(exit.MissingPrereq, errors.New("Docker Compose v2 is not available; install Docker with the compose plugin (docker compose)"))
}
