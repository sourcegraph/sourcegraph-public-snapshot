//go:build linux

package memcmd

import (
	"context"
	"io/fs"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/procfs"

	"github.com/sourcegraph/sourcegraph/internal/bytesize"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var defaultSamplingInterval = env.MustGetDuration("MEMORY_OBSERVATION_DEFAULT_SAMPLING_INTERVAL", 1*time.Millisecond, "For memory observers spawned by NewDefaultObserver, the interval at which memory usage is sampled. This environment variable only has an effect on Linux.")

// NewDefaultObserver creates a new Observer that observes the memory usage of a process and its children on Linux.
// This function is a convenience function that uses the default sampling interval specified by the MEMORY_OBSERVATION_DEFAULT_SAMPLING_INTERVAL
// environment variable.
//
// See NewLinuxObserver for more information.
func NewDefaultObserver(ctx context.Context, cmd *exec.Cmd) (Observer, error) {
	return NewLinuxObserver(ctx, cmd, defaultSamplingInterval)
}

// linuxObserver is an Observer that observes the memory usage of a process and its children on Linux.
type linuxObserver struct {
	ctx  context.Context
	proc processInfoProvider

	samplingInterval time.Duration

	startOnce sync.Once
	started   chan struct{}

	stopOnce          sync.Once
	cancelFunc        func()
	explicitlyStopped chan struct{}

	cmd *exec.Cmd

	mu                      sync.RWMutex // mutex ensures that we can read and write the memory usage from different goroutines
	highestMemoryUsageBytes uint64
	errs                    error
}

// NewLinuxObserver creates a new Observer that observes the memory usage of a process and its children on Linux.
//
// The observer will start sampling the memory usage of the process and its children at regular intervals (specified by samplingInterval).
func NewLinuxObserver(ctx context.Context, cmd *exec.Cmd, samplingInterval time.Duration) (Observer, error) {
	if cmd.Process == nil {
		// The process has not been started yet
		return nil, errors.New("process has not been started yet")
	}

	if samplingInterval <= 0 {
		return nil, errors.New("samplingInterval must be greater than 0")
	}

	f, err := procfs.NewDefaultFS()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create procfs")
	}

	proc := &procfsProcessInfoProvider{fs: f}

	if ctx == nil {
		ctx = context.Background()
	}

	ctx, cancel := context.WithCancel(ctx)

	return &linuxObserver{
		ctx:  ctx,
		proc: proc,

		cmd: cmd,

		started:           make(chan struct{}),
		explicitlyStopped: make(chan struct{}),

		cancelFunc: cancel,

		samplingInterval: samplingInterval,
	}, nil
}

func (l *linuxObserver) MaxMemoryUsage() (bytesize.Size, error) {
	select {
	case <-l.started:
	default:
		return 0, errObserverNotStarted
	}

	l.Stop()

	l.mu.RLock()
	defer l.mu.RUnlock()

	return bytesize.Size(l.highestMemoryUsageBytes), l.errs
}

// Start starts the observer.
func (l *linuxObserver) Start() {
	l.startOnce.Do(func() {
		go l.observe()
		close(l.started)
	})
}

func (l *linuxObserver) Stop() {
	l.stopOnce.Do(func() {
		close(l.explicitlyStopped)
		l.cancelFunc()
	})
}

func (l *linuxObserver) observe() {
	// Create a channel to signal when we should collect memory usage

	doCollection := make(chan struct{}, 1)
	doCollection <- struct{}{} // Trigger initial collection
	defer func() {
		for range doCollection {
			// Drain the channel
		}
	}()

	go func() {
		ticker := time.NewTicker(l.samplingInterval)
		defer ticker.Stop()

		defer close(doCollection) // signal that we are done collecting memory usage

		for {
			select {
			case <-l.ctx.Done(): // Shutdown the piping goroutine
				return

			case <-ticker.C: // Trigger memory collection at regular intervals
				doCollection <- struct{}{}
			}
		}
	}()

	for {
		select {
		case <-l.ctx.Done():
			return

		case <-doCollection:
			currentMemoryUsageBytes, err := memoryUsageForPidAndChildren(l.ctx, l.proc, l.cmd.Process.Pid)
			if errMaybeCausedByExplicitStop(err, l.explicitlyStopped) {
				// The error occurred when we were explicitly stopped, so we should skip
				// over this iteration.
				continue
			}

			l.mu.Lock()

			l.errs = errors.Append(l.errs, err)
			if currentMemoryUsageBytes > l.highestMemoryUsageBytes {
				l.highestMemoryUsageBytes = currentMemoryUsageBytes
			}

			l.mu.Unlock()
		}
	}
}

func memoryUsageForPidAndChildren(ctx context.Context, proc processInfoProvider, basePid int) (currentMemoryUsageBytes uint64, err error) {
	select {
	case <-ctx.Done():
		return 0, ctx.Err() // Return early if the context is done
	default:
	}

	var allRSSMemoryBytes uint64
	var errs error

	// This is a depth-first search of the process tree rooted at basePID.
	// For each iteration:
	// 1) we pop the first element from the stack
	// 2) add its memory usage to the total
	// 3) add its children to the stack
	//
	// We continue this process until the stack is empty.
	//
	// This process is best-effort. We might miss some processes if they
	// are created and destroyed between iterations.
	//
	// Some processes' memory information  might also be unavailable to us (e.g. the parent process might have already waited
	// on the child process, and the information is no longer available). In this specific case, we will ignore
	// the error (will be an os.IsNotExist error since we are using procfs) and continue.
	//
	// In the end, we return the sum of all the RSS memory usage of the processes in the tree, and any errors that occurred during the iteration.

	pidStack := []int{basePid}
	for len(pidStack) > 0 {
		select {
		case <-ctx.Done():
			return allRSSMemoryBytes, ctx.Err() // Return early if the context is done
		default:
		}

		currentPid := pidStack[0]
		pidStack = pidStack[1:]

		rss, err := proc.RSS(currentPid)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) { // Ignore no-longer-existent processes
				err = errors.Wrapf(err, "failed to report memory usage for pid %d", currentPid)
				errs = errors.Append(errs, err)
			}

			continue
		}

		allRSSMemoryBytes += rss

		children, err := proc.Children(currentPid)
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) { // Ignore no-longer-existent processes
				err = errors.Wrapf(err, "failed to list all processes")
				errs = errors.Append(errs, err)
			}

			continue
		}

		pidStack = append(pidStack, children...)
	}

	return allRSSMemoryBytes, errs
}

type processInfoProvider interface {
	// Children returns the PIDs of the children of the process with the given PID, or an error.
	// This is a best-effort operation that might miss some children. See the implementation-specific documentation for
	// more information.
	//
	// If the process does not exist, an error that wraps fs.ErrNotExist is returned.
	Children(pid int) (childrenPIDs []int, err error)

	// RSS returns the resident set size (RSS) of the process with the given PID, or an error.
	//
	// If the process does not exist, an error that wraps fs.ErrNotExist is returned.
	RSS(pid int) (rssBytes uint64, err error)
}

type procfsProcessInfoProvider struct {
	fs procfs.FS
}

func (p *procfsProcessInfoProvider) RSS(pid int) (rssBytes uint64, err error) {
	memory, err := func() (uint64, error) {
		proc, err := p.fs.Proc(pid)
		if err != nil {
			return 0, errors.Wrapf(err, "failed to get procfs")
		}

		status, err := proc.NewStatus()
		if err != nil {
			return 0, errors.Wrapf(err, "failed to get status")
		}
		return status.VmRSS, nil
	}()

	if err != nil {
		err = convertESRCH(err) // Ensure that we convert ESRCH errors to fs.ErrNotExist
	}

	return memory, err
}

// Children returns the PIDs of the children of the process with the given PID, or an error.
//
// This is a best-effort operation that might miss some children since it doesn't represent a snapshot of the process tree.
// (e.g. a child process might be created and destroyed between the time we list the processes and the time we list the children).
func (p *procfsProcessInfoProvider) Children(parentPID int) (pids []int, err error) {
	pids, err = func() ([]int, error) {
		procs, err := p.fs.AllProcs()
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list all processes")
		}

		var children []int
		for _, p := range procs {
			stat, err := p.Stat()
			if err != nil {
				if e := convertESRCH(err); !errors.Is(e, fs.ErrNotExist) { // Ignore no-longer-existent processes
					err = errors.Wrapf(err, "failed to stat process %d", p.PID)
					return nil, err
				}

				continue
			}

			if stat.PPID == parentPID {
				children = append(children, p.PID)
			}
		}

		return children, nil
	}()

	if err != nil {
		err = convertESRCH(err) // Ensure that we wrap ESRCH errors with fs.ErrNotExist
	}

	return pids, err
}

func errMaybeCausedByExplicitStop(err error, stopChan chan struct{}) bool {
	if errors.IsContextCanceled(err) {
		select {
		case <-stopChan:
			return true
		default:
		}
	}

	return false
}

// convertESRCH wraps an ESRCH error with fs.ErrNotExist
// to conform to the interface of the processInfoProvider
// (which makes it easier to check for errors).
func convertESRCH(err error) error {
	var e syscall.Errno
	if errors.As(err, &e) {
		// Append fs.ErrNotExist to the error if the error is an ESRCH error (and we haven't already done so)
		if e == syscall.ESRCH && !errors.Is(err, fs.ErrNotExist) {
			return errors.Append(err, fs.ErrNotExist)
		}
	}

	return err
}

var _ processInfoProvider = &procfsProcessInfoProvider{}
