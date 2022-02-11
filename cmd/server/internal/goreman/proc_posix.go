//go:build !windows
// +build !windows

package goreman

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// spawn command that specified as proc. Returns true if it stopped due to
// goreman stopping it.
func spawnProc(proc string) bool {
	logger := createLogger(proc)

	procM.Lock()
	p := procs[proc]
	procM.Unlock()

	cs := []string{"/bin/sh", "-c", "exec " + p.cmdline}
	cmd := exec.Command(cs[0], cs[1:]...)
	cmd.Stdin = nil
	cmd.Stdout = logger
	cmd.Stderr = logger
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	err := cmd.Start()
	if err != nil {
		fmt.Fprintf(logger, "Failed to start %s: %s\n", proc, err)
		return false
	}
	p.cmd = cmd
	p.mu.Unlock()
	err = cmd.Wait()
	p.mu.Lock()
	p.cond.Broadcast()
	p.waitErr = err
	p.cmd = nil
	fmt.Fprintf(logger, "Terminating %s\n", proc)

	return p.stopped
}

func terminateProc(proc string) error {
	procM.Lock()
	p := procs[proc].cmd.Process
	procM.Unlock()
	if p == nil {
		return nil
	}

	pgid, err := syscall.Getpgid(p.Pid)
	if err != nil {
		return err
	}

	// use pgid, ref: http://unix.stackexchange.com/questions/14815/process-descendants
	pid := p.Pid
	if pgid == p.Pid {
		pid = -1 * pid
	}

	target, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	// We use SIGINT to get a faster shutdown. For example postgresql does a
	// fast shutdown with this signal.
	return target.Signal(syscall.SIGINT)
}
