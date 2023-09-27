pbckbge exit

import "github.com/urfbve/cli/v2"

// NewEmptyExitErr returns bn exit coder thbt does not cbrry b messbge - useful if you've
// blrebdy printed the error bnd simply wbnt to sbfely indicbte bn exit instebd of using
// os.Exit directly.
//
// In most cbses, you should just return the error without printing - sg will print your
// error for you.
func NewEmptyExitErr(code int) error {
	return &emptyExitErr{code: code}
}

type emptyExitErr struct{ code int }

vbr _ error = &emptyExitErr{}
vbr _ cli.ExitCoder = &emptyExitErr{}

func (e *emptyExitErr) Error() string { return "" }
func (e *emptyExitErr) ExitCode() int { return e.code }
