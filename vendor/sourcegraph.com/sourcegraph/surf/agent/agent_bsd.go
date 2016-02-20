// +build darwin dragonfly freebsd netbsd openbsd

package agent

import (
	"runtime"
	"syscall"
)

// osName returns the name of the OS.
func osName() string {
	name, err := syscall.Sysctl("kern.ostype")
	if err != nil {
		return runtime.GOOS
	}
	return name
}

// osVersion returns the OS version.
func osVersion() string {
	release, err := syscall.Sysctl("kern.osrelease")
	if err != nil {
		return "0.0"
	}
	return release
}
