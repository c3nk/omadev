package compose

import (
	"context"
	"errors"
	"testing"

	"github.com/c3nk/omadev/internal/exec"
	"github.com/c3nk/omadev/internal/exit"
)

func TestCommandBuilders(t *testing.T) {
	c := New(&exec.Fake{}, "/repo")

	if got := c.UpCmd().String(); got != "docker compose up -d" {
		t.Errorf("UpCmd = %q", got)
	}
	if got := c.DownCmd().String(); got != "docker compose down" {
		t.Errorf("DownCmd = %q", got)
	}
	if got := c.LogsCmd().String(); got != "docker compose logs -f" {
		t.Errorf("LogsCmd = %q", got)
	}
	if got := c.DownCmd(); contains(got.Args, "--volumes") || contains(got.Args, "-v") {
		t.Error("down must never remove volumes")
	}
}

func TestSudoPrefix(t *testing.T) {
	c := New(&exec.Fake{}, "/repo")
	c.SetSudo(true)
	if got := c.UpCmd().String(); got != "sudo docker compose up -d" {
		t.Errorf("with sudo, UpCmd = %q", got)
	}
}

func TestUpDelegates(t *testing.T) {
	f := &exec.Fake{}
	c := New(f, "/repo")
	if _, err := c.Up(context.Background(), exec.Stdio{}); err != nil {
		t.Fatal(err)
	}
	last, _ := f.Last()
	if exec.Key(last) != "docker compose up -d" || last.Dir != "/repo" {
		t.Errorf("Up delegated %q in %q", exec.Key(last), last.Dir)
	}
}

func TestCheckAvailable(t *testing.T) {
	ctx := context.Background()

	v2 := &exec.Fake{Results: map[string]exec.Result{"docker compose version": {ExitCode: 0}}}
	if err := CheckAvailable(ctx, v2); err != nil {
		t.Errorf("v2 present should be available, got %v", err)
	}

	legacy := &exec.Fake{
		Default: exec.Result{ExitCode: 1},
		Results: map[string]exec.Result{"docker-compose version": {ExitCode: 0}},
	}
	err := CheckAvailable(ctx, legacy)
	var coded *exit.Error
	if !errors.As(err, &coded) || coded.Code != exit.MissingPrereq {
		t.Errorf("legacy-only should be MissingPrereq, got %v", err)
	}

	none := &exec.Fake{Default: exec.Result{ExitCode: 1}}
	if err := CheckAvailable(ctx, none); err == nil {
		t.Error("no compose should be a prerequisite error")
	}
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}
