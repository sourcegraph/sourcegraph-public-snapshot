package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"

	"golang.org/x/sys/windows"
)

// spawn command that specified as proc.
func spawnProc(proc string, errCh chan<- error) {
	procObj := procs[proc]
	logger := createLogger(proc, procObj.colorIndex)

	cs := []string{"cmd", "/c", procObj.cmdline}
	cmd := exec.Command(cs[0], cs[1:]...)
	cmd.Stdin = nil
	cmd.Stdout = logger
	cmd.Stderr = logger
	cmd.SysProcAttr = &windows.SysProcAttr{
		CreationFlags: windows.CREATE_UNICODE_ENVIRONMENT | 0x00000200,
	}
	cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", procObj.port))

	fmt.Fprintf(logger, "Starting %s on port %d\n", proc, procObj.port)
	if err := cmd.Start(); err != nil {
		select {
		case errCh <- err:
		default:
		}
		fmt.Fprintf(logger, "Failed to start %s: %s\n", proc, err)
		return
	}
	procObj.cmd = cmd
	procObj.stoppedBySupervisor = false
	procObj.mu.Unlock()
	err := cmd.Wait()
	procObj.mu.Lock()
	procObj.cond.Broadcast()
	if err != nil && procObj.stoppedBySupervisor == false {
		select {
		case errCh <- err:
		default:
		}
	}
	procObj.waitErr = err
	procObj.cmd = nil
	fmt.Fprintf(logger, "Terminating %s\n", proc)
}

func terminateProc(proc string, _ os.Signal) error {
	dll, err := windows.LoadDLL("kernel32.dll")
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
	r1, _, err = f.Call(windows.CTRL_BREAK_EVENT, uintptr(pid))
	if r1 == 0 {
		return err
	}
	return nil
}

func killProc(process *os.Process) error {
	return process.Kill()
}

func notifyCh() <-chan os.Signal {
	sc := make(chan os.Signal, 10)
	signal.Notify(sc, os.Interrupt)
	return sc
}
