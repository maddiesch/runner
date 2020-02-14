// Package runner provides a wrapper around exec.Cmd that supports contexts
// with a timeout.
package runner

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"path/filepath"
	"time"
)

// Runner is the main structure that handles the execution of of a command and
// the metadata associated with that execution.
type Runner struct {
	// The sub-command
	Cmd *exec.Cmd

	outBuf bytes.Buffer
	errBuf bytes.Buffer

	out io.Writer
	err io.Writer
}

// Command returns a new Runner command
func Command(cmd string, args ...string) *Runner {
	r := &Runner{
		Cmd: exec.Command(cmd, args...),
	}

	r.out = &r.outBuf
	r.err = &r.errBuf

	return r
}

// Cd sets the command's working directory
func (r *Runner) Cd(path string) *Runner {
	r.Cmd.Dir = filepath.Clean(path)

	return r
}

// AddStdout appends a new stdout writer to the output stack
func (r *Runner) AddStdout(w io.Writer) *Runner {
	r.out = io.MultiWriter(r.out, w)

	return r
}

// AddStderr appends a new stderr writer to the output stack
func (r *Runner) AddStderr(w io.Writer) *Runner {
	r.err = io.MultiWriter(r.err, w)

	return r
}

// Run executes the command
func (r *Runner) Run(ctx context.Context) (*Result, error) {
	deadline, ok := ctx.Deadline()
	duration := 300 * time.Second
	if ok {
		duration = time.Until(deadline)
	}

	tCtx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()

	r.Cmd.Stdout = r.out
	r.Cmd.Stderr = r.err

	if err := r.Cmd.Start(); err != nil {
		return nil, err
	}

	done := make(chan error)

	go func() {
		err := r.Cmd.Wait()

		done <- err

		close(done)
	}()

	var err error

	select {
	case <-tCtx.Done():
		_ = r.Cmd.Process.Kill()
		err = tCtx.Err()
	case e := <-done:
		err = e
	}

	return &Result{Stdout: r.outBuf.Bytes(), Stderr: r.errBuf.Bytes()}, err
}

// Result contains the result data from a command
type Result struct {
	// Data written to stdout by the command
	Stdout []byte

	// Data written to stderr by the command
	Stderr []byte
}
