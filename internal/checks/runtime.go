// Package checks verifies host prerequisites. It only reads state (e.g. runs
// `docker info`); it never modifies groups, permissions, or security settings.
package checks

import (
	"context"
	"strings"

	"github.com/c3nk/omadev/internal/exec"
)

// DockerStatus classifies Docker availability.
type DockerStatus int

const (
	// DockerAvailable means `docker info` succeeded without elevation.
	DockerAvailable DockerStatus = iota
	// DockerNeedsPrivilege means Docker is present but access was denied; elevation
	// (or a rootless/docker-group setup the user controls) is required.
	DockerNeedsPrivilege
	// DockerUnavailable means Docker is not installed or the daemon is unreachable.
	DockerUnavailable
)

// DockerCheck is the result of probing Docker.
type DockerCheck struct {
	Status DockerStatus
	Detail string
}

// RequiresSudo reports whether commands should be elevated.
func (c DockerCheck) RequiresSudo() bool { return c.Status == DockerNeedsPrivilege }

// Docker probes `docker info` without sudo and classifies the outcome (D4). It never
// elevates privileges or changes any setting itself.
func Docker(ctx context.Context, e exec.Executor) DockerCheck {
	res, err := e.Capture(ctx, exec.Command{Name: "docker", Args: []string{"info"}})
	if err != nil {
		return DockerCheck{Status: DockerUnavailable, Detail: "the docker command could not be started"}
	}
	if res.ExitCode == 0 {
		return DockerCheck{Status: DockerAvailable}
	}
	detail := strings.TrimSpace(res.Stderr)
	if isPermissionDenied(res.Stderr) {
		return DockerCheck{Status: DockerNeedsPrivilege, Detail: detail}
	}
	return DockerCheck{Status: DockerUnavailable, Detail: detail}
}

func isPermissionDenied(stderr string) bool {
	return strings.Contains(strings.ToLower(stderr), "permission denied")
}
