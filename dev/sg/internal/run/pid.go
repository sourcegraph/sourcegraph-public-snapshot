package run

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/sourcegraph/sourcegraph/dev/sg/interrupt"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type pidFile struct {
	Args []string `json:"args"`
	Pid  int      `json:"pid"`
}

func PidExistsWithArgs(args []string) (int, bool, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return 0, false, errors.Wrap(err, "could not check pidfiles")
	}

	pattern := filepath.Join(homeDir, ".sourcegraph", "sg.pid.*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return 0, false, errors.Wrap(err, "could not list pidfiles")
	}

	for _, match := range matches {
		f, err := os.Open(match)
		if err != nil {
			return 0, false, errors.Wrapf(err, "could not check pidfile %q", match)
		}
		defer f.Close()

		var content pidFile
		if err := json.NewDecoder(f).Decode(&content); err != nil {
			return 0, false, errors.Wrapf(err, "could not check pidfile %q", match)
		}

		if argsEqual(content.Args, args) {
			// If a pid already exists, let's check if it's still running.
			alive, err := isPidAlive(int32(content.Pid))
			if err != nil {
				return 0, true, errors.Wrapf(err, "could not check if pid %d is alive", content.Pid)
			}
			if alive {
				return content.Pid, true, nil
			}

			// Trash the pidfile and return that it's not running.
			_ = os.Remove(match)
			return 0, false, nil
		}
	}

	return 0, false, nil
}

func argsEqual(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i, n := range s1 {
		if n != s2[i] {
			return false
		}
	}
	return true
}

// On Unix systems, os.FindProcess always returns a os.Proc, regardless if the process is running or not. Therefore,
// we need more work to check if it's alive or not.
// Reference: https://github.com/shirou/gopsutil/blob/c141152a7b8f59b63e060fa8450f5cd5e7196dfb/process/process_posix.go#L73
func isPidAlive(pid int32) (bool, error) {
	if pid <= 0 {
		return false, errors.Newf("invalid pid %v", pid)
	}
	proc, err := os.FindProcess(int(pid))
	if err != nil {
		return false, err
	}
	err = proc.Signal(syscall.Signal(0))
	if err == nil {
		return true, nil
	}
	if err.Error() == "os: process already finished" {
		return false, nil
	}
	errno, ok := err.(syscall.Errno)
	if !ok {
		return false, err
	}
	switch errno {
	case syscall.ESRCH:
		return false, nil
	case syscall.EPERM:
		return true, nil
	}
	return false, err
}

func writePid() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	pidFileName := fmt.Sprintf("sg.pid.%d.json", os.Getpid())
	pidFilePath := filepath.Join(homeDir, ".sourcegraph", pidFileName)

	content := pidFile{
		Args: os.Args[1:],
		Pid:  os.Getpid(),
	}

	b, err := json.Marshal(content)
	if err != nil {
		return err
	}

	if err := os.WriteFile(pidFilePath, b, 0644); err != nil {
		return err
	}

	interrupt.Register(func() { os.Remove(pidFilePath) })

	return nil
}
