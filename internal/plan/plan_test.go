package plan

import (
	"bytes"
	"strings"
	"testing"

	"github.com/c3nk/omadev/internal/compose"
	"github.com/c3nk/omadev/internal/exec"
)

func TestComposeUpPlan(t *testing.T) {
	c := compose.New(&exec.Fake{}, "/repo")
	p := ComposeUp(c)

	if p.Strategy != "Docker Compose" {
		t.Errorf("strategy = %q", p.Strategy)
	}
	if len(p.Actions) == 0 || p.Actions[0].Kind != ActionExec {
		t.Fatalf("expected first action to be exec, got %+v", p.Actions)
	}
	if p.Actions[0].Command.String() != "docker compose up -d" {
		t.Errorf("first command = %q", p.Actions[0].Command.String())
	}
}

func TestRenderShowsCommand(t *testing.T) {
	c := compose.New(&exec.Fake{}, "/repo")
	var buf bytes.Buffer
	Render(&buf, ComposeUp(c))

	out := buf.String()
	for _, want := range []string{"Execution strategy:", "Docker Compose", "Plan", "docker compose up -d"} {
		if !strings.Contains(out, want) {
			t.Errorf("rendered plan missing %q\n---\n%s", want, out)
		}
	}
}

func TestComposeDownPreservesVolumes(t *testing.T) {
	c := compose.New(&exec.Fake{}, "/repo")
	p := ComposeDown(c)
	if strings.Contains(p.Actions[0].Command.String(), "-v") || strings.Contains(p.Actions[0].Command.String(), "volumes") {
		t.Error("down plan must not remove volumes")
	}
}
