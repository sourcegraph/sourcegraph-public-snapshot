// +build windows

package agent

import (
	"fmt"
	"runtime"
	"syscall"
)

// osName returns the name of the OS.
func osName() string {
	return runtime.GOOS
}

// osVersion returns the OS version.
func osVersion() string {
	v, err := syscall.GetVersion()
	if err != nil {
		return "0.0"
	}
	major := uint8(v)
	minor := uint8(v >> 8)
	return fmt.Sprintf("%d.%d", major, minor)
}
