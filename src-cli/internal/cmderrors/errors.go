package cmderrors

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// UsageError is an error type that subcommands can return in order to signal
// that a usage error has occurred.
type UsageError struct {
	error
}

func Usage(msg string) *UsageError {
	return &UsageError{errors.New(msg)}
}

func Usagef(f string, args ...interface{}) *UsageError {
	return &UsageError{errors.Newf(f, args...)}
}

func ExitCode(code int, err error) *ExitCodeError {
	return &ExitCodeError{error: err, exitCode: code}
}

// ExitCodeError is an error type that subcommands can return in order to
// specify the exact exit code.
type ExitCodeError struct {
	error
	exitCode int
}

func (e *ExitCodeError) HasError() bool { return e.error != nil }
func (e *ExitCodeError) Code() int      { return e.exitCode }

func (e *ExitCodeError) Error() string {
	if e.error != nil {
		return fmt.Sprintf("%s (exit code: %d)", e.error, e.exitCode)
	}
	return fmt.Sprintf("exit code: %d", e.exitCode)
}

const (
	GraphqlErrorsExitCode = 2
)

var ExitCode1 = &ExitCodeError{exitCode: 1}
