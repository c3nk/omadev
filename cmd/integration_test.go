//go:build integration

// These tests exercise the real `docker compose` binary and run only under the
// `integration` build tag (a separate CI job), never in the default unit-test run.
package cmd

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"

	"github.com/c3nk/omadev/internal/exec"
)

const integrationFixture = "../testdata/integration"

func dockerAvailable(ctx context.Context) bool {
	res, err := exec.OS{}.Capture(ctx, exec.Command{Name: "docker", Args: []string{"info"}})
	return err == nil && res.ExitCode == 0
}

func TestIntegration_UpStatusDown(t *testing.T) {
	ctx := context.Background()
	if !dockerAvailable(ctx) {
		t.Skip("docker is not available")
	}

	// Always attempt to stop the environment, even if the test fails midway.
	t.Cleanup(func() {
		_ = runDown(ctx, integrationFixture, io.Discard, strings.NewReader("y\n"), true, exec.OS{})
	})

	var out bytes.Buffer
	if err := runUp(ctx, integrationFixture, &out, strings.NewReader("y\n"), true, exec.OS{}); err != nil {
		t.Fatalf("up failed: %v\n%s", err, out.String())
	}

	out.Reset()
	if err := runStatus(ctx, integrationFixture, &out, true, exec.OS{}); err != nil {
		t.Fatalf("status failed: %v", err)
	}
	if !strings.Contains(out.String(), "app") {
		t.Errorf("status should list the app service:\n%s", out.String())
	}

	if err := runDown(ctx, integrationFixture, io.Discard, strings.NewReader("y\n"), true, exec.OS{}); err != nil {
		t.Fatalf("down failed: %v", err)
	}
}
