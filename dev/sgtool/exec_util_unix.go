// +build !windows

package main

import (
	"os/exec"
	"syscall"
)

func pgrep(program string) (found bool, err error) {
	err = exec.Command("pgrep", "rego").Run()
	if err != nil {
		if _, ok := err.(*exec.Error); ok {
			return false, err
		}
		if status, _ := exitStatus(err); status == 1 {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func exitStatus(err error) (uint32, error) {
	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			// There is no platform independent way to retrieve
			// the exit code, but the following will work on Unix
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				return uint32(status.ExitStatus()), nil
			}
		}
		return 0, err
	}
	return 0, nil
}
