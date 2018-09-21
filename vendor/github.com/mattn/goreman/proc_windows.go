package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

// spawn command that specified as proc.
func spawnProc(proc string) bool {
	procObj := procs[proc]
	logger := createLogger(proc, procObj.colorIndex)

	cs := []string{"cmd", "/c", procObj.cmdline}
	cmd := exec.Command(cs[0], cs[1:]...)
	cmd.Stdin = nil
	cmd.Stdout = logger
	cmd.Stderr = logger
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_UNICODE_ENVIRONMENT | 0x00000200,
	}
	cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", procObj.port))

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

	return procs[proc].quit
}

func terminateProc(proc string, signal os.Signal) error {
	dll, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return err
	}
	defer dll.Release()

	pid := procs[proc].cmd.Process.Pid

	f, err := dll.FindProc("SetConsoleCtrlHandler")
	if err != nil {
		return err
	}
	r1, _, err := f.Call(0, 1)
	if r1 == 0 {
		return err
	}
	f, err = dll.FindProc("GenerateConsoleCtrlEvent")
	if err != nil {
		return err
	}
	r1, _, err = f.Call(syscall.CTRL_BREAK_EVENT, uintptr(pid))
	if r1 == 0 {
		return err
	}
	return nil
}

func killProc(process *os.Process) error {
	return process.Kill()
}
