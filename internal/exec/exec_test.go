package exec

import (
	"context"
	"runtime"
	"testing"
)

func TestOSCapture(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a POSIX shell")
	}
	res, err := OS{}.Capture(context.Background(), Command{Name: "sh", Args: []string{"-c", "printf hello"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ExitCode != 0 || res.Stdout != "hello" {
		t.Errorf("got exit=%d stdout=%q, want exit=0 stdout=hello", res.ExitCode, res.Stdout)
	}
}

func TestOSCaptureNonZero(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a POSIX shell")
	}
	res, err := OS{}.Capture(context.Background(), Command{Name: "sh", Args: []string{"-c", "exit 3"}})
	if err != nil {
		t.Fatalf("a non-zero exit must not be a Go error, got: %v", err)
	}
	if res.ExitCode != 3 {
		t.Errorf("exit code = %d, want 3", res.ExitCode)
	}
}

func TestOSCaptureCannotStart(t *testing.T) {
	_, err := OS{}.Capture(context.Background(), Command{Name: "this-binary-does-not-exist-omadev"})
	if err == nil {
		t.Error("expected an error when the command cannot start")
	}
}

func TestFakeRecordsAndScripts(t *testing.T) {
	f := &Fake{
		Results: map[string]Result{
			"docker compose ps": {ExitCode: 0, Stdout: "running"},
		},
		Default: Result{ExitCode: 0},
	}

	res, _ := f.Capture(context.Background(), Command{Name: "docker", Args: []string{"compose", "ps"}})
	if res.Stdout != "running" {
		t.Errorf("scripted result not returned: %+v", res)
	}

	_, _ = f.Capture(context.Background(), Command{Name: "docker", Args: []string{"compose", "up", "-d"}})
	last, ok := f.Last()
	if !ok || Key(last) != "docker compose up -d" {
		t.Errorf("last command = %q, want 'docker compose up -d'", Key(last))
	}
	if len(f.Commands) != 2 {
		t.Errorf("recorded %d commands, want 2", len(f.Commands))
	}
}
