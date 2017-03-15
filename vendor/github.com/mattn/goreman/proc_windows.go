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

	cs := []string{"cmd", "/c", procs[proc].cmdline}
	cmd := exec.Command(cs[0], cs[1:]...)
	cmd.Stdin = nil
	cmd.Stdout = logger
	cmd.Stderr = logger
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_UNICODE_ENVIRONMENT | 0x00000200,
	}
	cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", procs[proc].port))

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

func terminateProc(proc string) error {
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
