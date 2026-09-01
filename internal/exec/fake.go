package exec

import (
	"context"
	"strings"
)

// Fake is a test executor. It records every command and returns scripted results,
// so command and planner logic can be tested without a real subprocess.
type Fake struct {
	Commands   []Command         // every command received, in order
	Results    map[string]Result // keyed by Key(command); overrides Default
	Default    Result            // returned by Capture when no scripted result matches
	StreamExit int               // returned by Stream
	StreamErr  error             // returned by Stream (nil for success)
}

// Key is the lookup key for a command: its name and args joined by spaces.
func Key(c Command) string {
	return strings.TrimSpace(c.Name + " " + strings.Join(c.Args, " "))
}

func (f *Fake) Capture(_ context.Context, c Command) (Result, error) {
	f.Commands = append(f.Commands, c)
	if r, ok := f.Results[Key(c)]; ok {
		return r, nil
	}
	return f.Default, nil
}

func (f *Fake) Stream(_ context.Context, c Command, _ Stdio) (int, error) {
	f.Commands = append(f.Commands, c)
	return f.StreamExit, f.StreamErr
}

// Last returns the most recently received command, or false if none.
func (f *Fake) Last() (Command, bool) {
	if len(f.Commands) == 0 {
		return Command{}, false
	}
	return f.Commands[len(f.Commands)-1], true
}
