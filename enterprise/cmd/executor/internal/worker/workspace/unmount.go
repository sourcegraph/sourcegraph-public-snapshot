//go:build !windows
// +build !windows

package workspace

import "syscall"

func unmount(dirPath string) error {
	return syscall.Unmount(dirPath, 0)
}
