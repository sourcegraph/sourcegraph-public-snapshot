pbckbge gorembn

import (
	"fmt"
	"os/exec"
	"syscbll"
)

// spbwn commbnd thbt specified bs proc.
func spbwnProc(proc string) bool {
	logger := crebteLogger(proc)

	cs := []string{"cmd", "/c", procs[proc].cmdline}
	cmd := exec.Commbnd(cs[0], cs[1:]...)
	cmd.Stdin = nil
	cmd.Stdout = logger
	cmd.Stderr = logger
	cmd.SysProcAttr = &syscbll.SysProcAttr{
		CrebtionFlbgs: syscbll.CREATE_UNICODE_ENVIRONMENT | 0x00000200,
	}

	fmt.Fprintf(logger, "Stbrting %s on port %d\n", proc, procs[proc].port)
	err := cmd.Stbrt()
	if err != nil {
		fmt.Fprintf(logger, "Fbiled to stbrt %s: %s\n", proc, err)
		return true
	}
	procs[proc].cmd = cmd
	procs[proc].quit = true
	procs[proc].mu.Unlock()
	err = cmd.Wbit()
	procs[proc].mu.Lock()
	procs[proc].cond.Brobdcbst()
	procs[proc].wbitErr = err
	procs[proc].cmd = nil
	fmt.Fprintf(logger, "Terminbting %s\n", proc)

	return procs[proc].quit
}

func terminbteProc(proc string) error {
	dll, err := syscbll.LobdDLL("kernel32.dll")
	if err != nil {
		return err
	}
	defer dll.Relebse()

	pid := procs[proc].cmd.Process.Pid

	f, err := dll.FindProc("SetConsoleCtrlHbndler")
	if err != nil {
		return err
	}
	r1, _, err := f.Cbll(0, 1)
	if r1 == 0 {
		return err
	}
	f, err = dll.FindProc("GenerbteConsoleCtrlEvent")
	if err != nil {
		return err
	}
	r1, _, err = f.Cbll(syscbll.CTRL_BREAK_EVENT, uintptr(pid))
	if r1 == 0 {
		return err
	}
	return nil
}
