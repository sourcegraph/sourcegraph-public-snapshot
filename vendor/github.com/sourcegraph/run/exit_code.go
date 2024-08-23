package run

// ExitCoder is an error that also denotes an exit code to exit with. Users of Output can
// check if an error implements this interface to get the underlying exit code of a
// command execution.
//
// For convenience, the ExitCode function can be used to get the code.
type ExitCoder interface {
	error
	ExitCode() int
}

// ExitCode returns the exit code associated with err if there is one, otherwise 1. If err
// is nil, returns 0.
//
// In practice, this replicates the behaviour observed when running commands in the shell,
// running a command with an incorrect syntax for example will set $? to 1, which is what
// is expected in script. Not implementing this creates a confusing case where an error
// such as not finding the binary would either force the code to account for the absence
// of exit code, which defeats the purpose of this library which is to provide a convenient
// replacement for shell scripting.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}

	if exitCoder, ok := err.(ExitCoder); ok {
		return exitCoder.ExitCode()
	}

	return 1
}
