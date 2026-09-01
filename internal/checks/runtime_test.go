package checks

import (
	"context"
	"testing"

	"github.com/c3nk/omadev/internal/exec"
)

func TestDocker(t *testing.T) {
	ctx := context.Background()

	cases := []struct {
		name   string
		result exec.Result
		want   DockerStatus
	}{
		{"available", exec.Result{ExitCode: 0}, DockerAvailable},
		{"permission denied", exec.Result{ExitCode: 1, Stderr: "Got permission denied while trying to connect to the Docker daemon socket"}, DockerNeedsPrivilege},
		{"daemon down", exec.Result{ExitCode: 1, Stderr: "Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?"}, DockerUnavailable},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &exec.Fake{Results: map[string]exec.Result{"docker info": c.result}}
			got := Docker(ctx, f)
			if got.Status != c.want {
				t.Errorf("status = %d, want %d (detail: %q)", got.Status, c.want, got.Detail)
			}
		})
	}
}

func TestRequiresSudo(t *testing.T) {
	if !(DockerCheck{Status: DockerNeedsPrivilege}).RequiresSudo() {
		t.Error("NeedsPrivilege must require sudo")
	}
	if (DockerCheck{Status: DockerAvailable}).RequiresSudo() {
		t.Error("Available must not require sudo")
	}
}
