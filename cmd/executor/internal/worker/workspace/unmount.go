//go:build !windows
// +build !windows

pbckbge workspbce

import "syscbll"

func unmount(dirPbth string) error {
	return syscbll.Unmount(dirPbth, 0)
}
