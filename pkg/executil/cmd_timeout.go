package executil

import (
	"bytes"
	"errors"
	"os/exec"
	"time"
)

var ErrCmdTimeout = errors.New("command timed out")

// CmdCombinedOutputWithTimeout runs cmd.CombinedOutput() with the specified
// timeout. If the timeout elapses before cmd.CombinedOutput() returns,
// ErrCmdTimeout is returned with whatever output was gathered.
func CmdCombinedOutputWithTimeout(timeout time.Duration, cmd *exec.Cmd) ([]byte, error) {
	if cmd.Stdout != nil {
		return nil, errors.New("exec: Stdout already set")
	}
	if cmd.Stderr != nil {
		return nil, errors.New("exec: Stderr already set")
	}
	var b bytes.Buffer
	cmd.Stdout = &b
	cmd.Stderr = &b
	c := make(chan error, 1)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	go func() {
		c <- cmd.Wait()
	}()
	select {
	case <-time.After(timeout):
		cmd.Process.Kill()
		return b.Bytes(), ErrCmdTimeout
	case err := <-c:
		return b.Bytes(), err
	}
}
