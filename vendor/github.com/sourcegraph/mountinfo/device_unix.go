//go:build !windows

// file to hold functions that work on both Linux and Unix operating systems

package mountinfo

import (
	"fmt"

	"golang.org/x/sys/unix"
)

// defined as a variable so that it can be redefined by test routines
var getDeviceNumber = func(filePath string) (string, error) {
	// this is the only explicitely platform-dependent code being used: Stat_t and Stat.
	// (requires a Unix/Linux OS to compile)
	// Other code is implicitly dependent on Linux's sysfs, but will compile on other OSs
	var stat unix.Stat_t
	err := unix.Stat(filePath, &stat)
	if err != nil {
		return "", fmt.Errorf("getDeviceNumber: failed to stat %q: %w", filePath, err)
	}

	//nolint:unconvert // We need the unix.Major/Minor functions to perform the proper bit-shifts
	major, minor := unix.Major(uint64(stat.Dev)), unix.Minor(uint64(stat.Dev))

	// Represent the number in <major>:<minor> format.
	return fmt.Sprintf("%d:%d", major, minor), nil
}
