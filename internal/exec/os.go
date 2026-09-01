package exec

import (
	"bytes"
	"context"
	"errors"
	osexec "os/exec"
)

// OS is the real executor backed by os/exec.
type OS struct{}

func (OS) Capture(ctx context.Context, c Command) (Result, error) {
	cmd := osexec.CommandContext(ctx, c.Name, c.Args...)
	cmd.Dir = c.Dir
	cmd.Env = c.Env

	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf

	err := cmd.Run()
	res := Result{Stdout: out.String(), Stderr: errBuf.String()}

	if err != nil {
		var ee *osexec.ExitError
		if errors.As(err, &ee) {
			res.ExitCode = ee.ExitCode()
			return res, nil // ran, but non-zero: reported via ExitCode, not error
		}
		return res, err // could not start
	}
	return res, nil
}

func (OS) Stream(ctx context.Context, c Command, sio Stdio) (int, error) {
	cmd := osexec.CommandContext(ctx, c.Name, c.Args...)
	cmd.Dir = c.Dir
	cmd.Env = c.Env
	cmd.Stdin = sio.In
	cmd.Stdout = sio.Out
	cmd.Stderr = sio.Err

	err := cmd.Run()
	if err != nil {
		var ee *osexec.ExitError
		if errors.As(err, &ee) {
			return ee.ExitCode(), nil
		}
		return -1, err
	}
	return 0, nil
}
