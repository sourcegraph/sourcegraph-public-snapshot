package executil

import (
	"errors"
	"os/exec"
	"time"
)

var ErrCmdTimeout = errors.New("command timed out")

// CmdWaitWithTimeout runs cmd.Wait() with the specified timeout. If
// the timeout elapses before cmd.Wait() returns, ErrCmdTimeout is
// returned.
func CmdWaitWithTimeout(timeout time.Duration, cmd *exec.Cmd) error {
	errc := make(chan error, 1)
	go func() {
		errc <- cmd.Wait()
	}()
	var err error
	select {
	case <-time.After(timeout):
		cmd.Process.Kill()
		return ErrCmdTimeout
	case err = <-errc:
	}
	return err
}
