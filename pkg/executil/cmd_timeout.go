package executil

import (
	"bytes"
	"errors"
	"os/exec"
	"sync"
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
	var b syncBuffer
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

// syncBuffer is like bytes.Buffer (it implements Writer and has a
// Bytes method) but Write and Bytes calls are synchronized using a
// mutex. This avoids race conditions when used as the stdout/stderr
// destination in CmdCombinedOutputWithTimeout.
type syncBuffer struct {
	mu sync.RWMutex // guards b
	b  bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	n, err = b.b.Write(p)
	b.mu.Unlock()
	return
}

func (b *syncBuffer) Bytes() []byte {
	b.mu.RLock()
	p := b.b.Bytes()
	b.mu.RUnlock()
	return p
}
