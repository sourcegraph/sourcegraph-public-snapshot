package exit

import "github.com/urfave/cli/v2"

// NewEmptyExitErr returns an exit coder that does not carry a message - useful if you've
// already printed the error and simply want to safely indicate an exit instead of using
// os.Exit directly.
//
// In most cases, you should just return the error without printing - sg will print your
// error for you.
func NewEmptyExitErr(code int) error {
	return &emptyExitErr{code: code}
}

type emptyExitErr struct{ code int }

var _ error = &emptyExitErr{}
var _ cli.ExitCoder = &emptyExitErr{}

func (e *emptyExitErr) Error() string { return "" }
func (e *emptyExitErr) ExitCode() int { return e.code }
