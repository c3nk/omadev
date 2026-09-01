package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/c3nk/omadev/internal/exec"
	"github.com/c3nk/omadev/internal/exit"
)

const composeFixture = "../testdata/docker-compose-fastapi-react"

// dockerReadyFake returns a fake where Docker is available and services are running.
func dockerReadyFake() *exec.Fake {
	return &exec.Fake{
		Results: map[string]exec.Result{
			"docker info":                     {ExitCode: 0},
			"docker compose version":          {ExitCode: 0},
			"docker compose ps --format json": {ExitCode: 0, Stdout: `{"Service":"postgres","State":"running","Health":"healthy","Publishers":[{"PublishedPort":5432}]}`},
		},
		Default:    exec.Result{ExitCode: 0},
		StreamExit: 0,
	}
}

func TestUp_Success(t *testing.T) {
	f := dockerReadyFake()
	var out bytes.Buffer
	err := runUp(context.Background(), composeFixture, &out, strings.NewReader("y\n"), true, f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := out.String()
	for _, want := range []string{"Plan", "docker compose up -d", "Continue?", "Environment started"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in output:\n%s", want, got)
		}
	}
	// The up command must actually have been issued.
	issued := false
	for _, c := range f.Commands {
		if exec.Key(c) == "docker compose up -d" {
			issued = true
		}
	}
	if !issued {
		t.Error("docker compose up -d was not issued")
	}
}

func TestUp_Cancel(t *testing.T) {
	f := dockerReadyFake()
	var out bytes.Buffer
	err := runUp(context.Background(), composeFixture, &out, strings.NewReader("n\n"), true, f)

	var coded *exit.Error
	if !errors.As(err, &coded) || coded.Code != exit.Canceled {
		t.Fatalf("expected Canceled, got %v", err)
	}
	for _, c := range f.Commands {
		if exec.Key(c) == "docker compose up -d" {
			t.Error("up must not run when the user cancels")
		}
	}
}

func TestUp_RefusesLowConfidence(t *testing.T) {
	f := dockerReadyFake()
	var out bytes.Buffer
	// node-vite is a non-Compose project: up must refuse.
	err := runUp(context.Background(), "../testdata/node-vite", &out, strings.NewReader("y\n"), true, f)

	var coded *exit.Error
	if !errors.As(err, &coded) || coded.Code != exit.Unsupported {
		t.Fatalf("expected Unsupported, got %v", err)
	}
	if len(f.Commands) != 0 {
		t.Errorf("no docker commands should run for a refused project, got %v", f.Commands)
	}
}

func TestUp_MissingDocker(t *testing.T) {
	f := &exec.Fake{Default: exec.Result{ExitCode: 1, Stderr: "Cannot connect to the Docker daemon"}}
	var out bytes.Buffer
	err := runUp(context.Background(), composeFixture, &out, strings.NewReader("y\n"), true, f)

	var coded *exit.Error
	if !errors.As(err, &coded) || coded.Code != exit.MissingPrereq {
		t.Fatalf("expected MissingPrereq, got %v", err)
	}
}

func TestDown_Success(t *testing.T) {
	f := dockerReadyFake()
	var out bytes.Buffer
	err := runDown(context.Background(), composeFixture, &out, strings.NewReader("y\n"), true, f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out.String(), "data was preserved") {
		t.Errorf("down should reassure about data, got:\n%s", out.String())
	}
	for _, c := range f.Commands {
		if strings.Contains(exec.Key(c), "--volumes") {
			t.Error("down must never remove volumes")
		}
	}
}

func TestStatus_Success(t *testing.T) {
	f := dockerReadyFake()
	var out bytes.Buffer
	if err := runStatus(context.Background(), composeFixture, &out, true, f); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "postgres") || !strings.Contains(got, "healthy") {
		t.Errorf("status should show service state, got:\n%s", got)
	}
}
