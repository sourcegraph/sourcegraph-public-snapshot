package run

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
)

// runError wraps exec.ExitError such that it always includes the embedded stderr.
type runError struct{ execErr *exec.ExitError }

var _ ExitCoder = &runError{}

// newError creats a new *Error, and can be provided a nil error and/or nil stdErr
func newError(err error, stdErr io.Reader) error {
	if err == nil {
		return nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if stdErr != nil {
			// Not assigned by default using cmd.Start(), so we consume our copy of stderr
			// and set it here. If an error occurs we just don't do anything with stderr.
			if b, err := io.ReadAll(stdErr); err == nil {
				exitErr.Stderr = bytes.TrimSpace(b)
			}
		}
		return &runError{execErr: exitErr}
	}

	return err
}

func (e *runError) Error() string {
	if len(e.execErr.Stderr) == 0 {
		return e.execErr.String()
	}
	return fmt.Sprintf("%s: %s", e.execErr.String(), string(e.execErr.Stderr))
}

func (e *runError) ExitCode() int {
	return e.execErr.ExitCode()
}
