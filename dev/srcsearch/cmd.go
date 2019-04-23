package main

// usageError is an error type that subcommands can return in order to signal
// that a usage error has occurred.
type usageError struct {
	error
}

// exitCodeError is an error type that subcommands can return in order to
// specify the exact exit code.
type exitCodeError struct {
	error
	exitCode int
}

const (
	graphqlErrorsExitCode = 2
)
