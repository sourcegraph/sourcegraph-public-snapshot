// +build !windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/unix"
)

// spawn command that specified as proc.
func spawnProc(proc string) bool {
	procObj := procs[proc]
	logger := createLogger(proc, procObj.colorIndex)

	cs := []string{"/bin/sh", "-c", procs[proc].cmdline}
	cmd := exec.Command(cs[0], cs[1:]...)
	cmd.Stdin = nil
	cmd.Stdout = logger
	cmd.Stderr = logger
	cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", procObj.port))
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	fmt.Fprintf(logger, "Starting %s on port %d\n", proc, procObj.port)
	err := cmd.Start()
	if err != nil {
		fmt.Fprintf(logger, "Failed to start %s: %s\n", proc, err)
		return true
	}
	procObj.cmd = cmd
	procObj.quit = true
	procObj.mu.Unlock()
	err = cmd.Wait()
	procObj.mu.Lock()
	procObj.cond.Broadcast()
	procObj.waitErr = err
	procObj.cmd = nil
	fmt.Fprintf(logger, "Terminating %s\n", proc)

	return procObj.quit
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

// killProc kills the proc with pid pid, as well as its children.
func killProc(process *os.Process) error {
	return unix.Kill(-1*process.Pid, syscall.SIGKILL)
}
