// +build !windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// spawn command that specified as proc.
func spawnProc(proc string) bool {
	logger := createLogger(proc)

	cs := []string{"/bin/sh", "-c", procs[proc].cmdline}
	cmd := exec.Command(cs[0], cs[1:]...)
	cmd.Stdin = nil
	cmd.Stdout = logger
	cmd.Stderr = logger
	cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", procs[proc].port))
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	fmt.Fprintf(logger, "Starting %s on port %d\n", proc, procs[proc].port)
	err := cmd.Start()
	if err != nil {
		fmt.Fprintf(logger, "Failed to start %s: %s\n", proc, err)
		return true
	}
	procs[proc].cmd = cmd
	procs[proc].quit = true
	procs[proc].mu.Unlock()
	err = cmd.Wait()
	procs[proc].mu.Lock()
	procs[proc].cond.Broadcast()
	procs[proc].waitErr = err
	procs[proc].cmd = nil
	fmt.Fprintf(logger, "Terminating %s\n", proc)

	return procs[proc].quit
}

func terminateProc(proc string, signal os.Signal) error {
	p := procs[proc].cmd.Process
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
	return target.Signal(signal)
}
