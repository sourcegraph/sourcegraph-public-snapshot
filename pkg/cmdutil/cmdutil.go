// Package cmdutil is like the exec package except it provides nice errors.
//
// This package simply invokes an existing *exec.Cmd in the standard execution
// modes like cmd.Output or cmd.Run, except in the event of command failure the
// error will include the command which was ran and the stderr output:
//
// 	exit status 128 (running 'rm foobar'): stderr: "rm: foobar: No such file or directory"
//
// Instead of what you would usually get by invoking the *exec.Cmd method
// yourself:
//
//  exit status 128
//
package cmdutil

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// ExitError is a wrapper around exec.ExitError, but also records the command
// run.
type ExitError struct {
	*exec.ExitError
	Args []string
}

func (e *ExitError) Error() string {
	return fmt.Sprintf("%v (running '%s'): stderr: %q", e.ExitError, strings.Join(e.Args, " "), string(e.ExitError.Stderr))
}

// Output is functionally the same as invoking c.Output except it attaches
// stderr to a buffer and returns nice errors (see package description) in the
// event of an error.
func Output(c *exec.Cmd) ([]byte, error) {
	output, err := c.Output()
	if err != nil {
		if e, ok := err.(*exec.ExitError); ok {
			return nil, &ExitError{
				ExitError: e,
				Args:      c.Args,
			}
		}
		return nil, fmt.Errorf("%v (running '%s')", err, strings.Join(c.Args, " "))
	}
	return output, nil
}

// Run is functionally the same as invoking c.Run except it attaches stderr to
// a buffer and returns nice errors (see package description) in the event of
// an error.
func Run(c *exec.Cmd) error {
	if c.Stderr != nil {
		return fmt.Errorf("cmdutil (running '%s'): c.Stderr != nil", strings.Join(c.Args, " "))
	}
	var stderr bytes.Buffer
	c.Stderr = &stderr
	err := c.Run()
	if err != nil {
		return fmt.Errorf("%v (running '%s'): stderr: %q", err, strings.Join(c.Args, " "), stderr.String())
	}
	return nil
}
