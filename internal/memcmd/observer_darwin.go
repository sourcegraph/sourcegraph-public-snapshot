//go:build darwin

package memcmd

import (
	"os/exec"
	"syscall"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type macObserver struct {
	startOnce sync.Once
	started   chan struct{}

	stopOnce sync.Once
	cmd      *exec.Cmd
}

// NewMacObserver creates a new Observer for a command running on macOS.
// The command must have already been started before calling this function.
func NewMacObserver(cmd *exec.Cmd) (Observer, error) {
	if cmd.Process == nil {
		return nil, errors.New("command has not started")
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

func (o *macObserver) MaxMemoryUsage() (uint64, error) {
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

	return uint64(usage.Maxrss) << 10, nil
}

var _ Observer = &macObserver{}
