// Package plan builds the ordered list of actions a state-changing command will
// perform and renders it for the SHOW step, before any confirmation or execution.
package plan

import (
	"fmt"
	"io"

	"github.com/c3nk/omadev/internal/compose"
	"github.com/c3nk/omadev/internal/exec"
)

// ActionKind classifies a plan step.
type ActionKind int

const (
	// ActionExec runs a command.
	ActionExec ActionKind = iota
	// ActionVerify re-queries state after execution.
	ActionVerify
)

// Action is one step in a plan.
type Action struct {
	Title   string
	Kind    ActionKind
	Command exec.Command // set for ActionExec
}

// Plan is the ordered set of actions for a command, under one execution strategy.
type Plan struct {
	Strategy string
	Actions  []Action
}

// ComposeUp is the plan for starting a Compose environment.
func ComposeUp(c *compose.Compose) Plan {
	return Plan{
		Strategy: "Docker Compose",
		Actions: []Action{
			{Title: "Start Docker Compose environment", Kind: ActionExec, Command: c.UpCmd()},
			{Title: "Wait for services and verify status", Kind: ActionVerify},
		},
	}
}

// ComposeDown is the plan for stopping a Compose environment (volumes preserved).
func ComposeDown(c *compose.Compose) Plan {
	return Plan{
		Strategy: "Docker Compose",
		Actions: []Action{
			{Title: "Stop Docker Compose environment (data preserved)", Kind: ActionExec, Command: c.DownCmd()},
		},
	}
}

// Render writes the plan for the SHOW step. It always displays the exact command(s).
func Render(w io.Writer, p Plan) {
	fmt.Fprintf(w, "Execution strategy:\n  %s\n\n", p.Strategy)
	fmt.Fprintln(w, "Plan")
	for _, a := range p.Actions {
		if a.Kind == ActionExec {
			fmt.Fprintf(w, "  %s:\n    %s\n", a.Title, a.Command.String())
		} else {
			fmt.Fprintf(w, "  %s\n", a.Title)
		}
	}
	fmt.Fprintln(w)
}
