package exit

import (
	"errors"
	"fmt"
	"testing"
)

func TestCodeOf(t *testing.T) {
	base := errors.New("boom")
	cases := []struct {
		name string
		err  error
		want Code
	}{
		{"nil is OK", nil, OK},
		{"generic non-coded error", base, Generic},
		{"unsupported", New(Unsupported, base), Unsupported},
		{"missing prerequisite", New(MissingPrereq, base), MissingPrereq},
		{"invalid config", New(InvalidConfig, base), InvalidConfig},
		{"canceled", New(Canceled, base), Canceled},
		{"exec failure", New(ExecFailure, base), ExecFailure},
		{"wrapped coded error is found", fmt.Errorf("context: %w", New(InvalidConfig, base)), InvalidConfig},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := CodeOf(c.err); got != c.want {
				t.Errorf("CodeOf(%v) = %d, want %d", c.err, got, c.want)
			}
		})
	}
}

func TestErrorUnwrap(t *testing.T) {
	base := errors.New("root cause")
	err := New(ExecFailure, base)
	if !errors.Is(err, base) {
		t.Errorf("expected wrapped error to unwrap to the base error")
	}
	if err.Error() != "root cause" {
		t.Errorf("Error() = %q, want %q", err.Error(), "root cause")
	}
}
