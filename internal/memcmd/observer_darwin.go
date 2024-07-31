//go:build darwin

package memcmd

import (
	"context"
	"os/exec"
	"sync"
	"syscall"

	"github.com/sourcegraph/sourcegraph/internal/bytesize"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type macObserver struct {
	startOnce sync.Once
	started   chan struct{}

	stopOnce sync.Once
	cmd      *exec.Cmd
}

// NewDefaultObserver creates a new Observer for a command running on macOS.
// The command must have already been started before calling this function.
// The command must have also been started with its own process group ID (cmd.SysProcAttr.Setpgid == true).
func NewDefaultObserver(_ context.Context, cmd *exec.Cmd) (Observer, error) {
	return NewMacObserver(cmd)
}

// NewMacObserver creates a new Observer for a command running on macOS.
// The command must have already been started before calling this function.
// The command must have also been started with its own process group ID (cmd.SysProcAttr.Setpgid == true).
func NewMacObserver(cmd *exec.Cmd) (Observer, error) {
	if cmd.Process == nil {
		return nil, errors.New("command has not started")
	}

	attr := cmd.SysProcAttr
	if !(attr != nil && attr.Setpgid) {
		return nil, errProcessNotWithinOwnProcessGroup
	}

	return &macObserver{
		started: make(chan struct{}),
		cmd:     cmd,
	}, nil
}

func (o *macObserver) Start() {
	o.startOnce.Do(func() {
		close(o.started)
	})
}

func (o *macObserver) Stop() {
	o.stopOnce.Do(func() {})
}

func (o *macObserver) MaxMemoryUsage() (bytesize.Size, error) {
	select {
	case <-o.started:
	default:
		return 0, errObserverNotStarted
	}

	o.Stop()

	state := o.cmd.ProcessState
	if state == nil {
		return 0, errProcessNotStopped
	}

	usage, ok := state.SysUsage().(*syscall.Rusage)
	if !ok {
		return 0, errors.New("failed to get rusage")
	}

	// On macOS, MAXRSS is the maximum resident set size used (in bytes, not kilobytes).
	// See getrusage(2) for more information.
	return bytesize.Size(usage.Maxrss), nil
}

var _ Observer = &macObserver{}

var errProcessNotWithinOwnProcessGroup = errors.New("command must be started with its own process group ID (cmd.SysProcAttr.Setpgid = true)")
