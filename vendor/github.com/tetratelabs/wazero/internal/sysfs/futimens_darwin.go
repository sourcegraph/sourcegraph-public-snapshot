package sysfs

import (
	"syscall"
	_ "unsafe" // for go:linkname
)

const (
	_AT_FDCWD               = -0x2
	_AT_SYMLINK_NOFOLLOW    = 0x0020
	_UTIME_NOW              = -1
	_UTIME_OMIT             = -2
	SupportsSymlinkNoFollow = true
)

//go:noescape
//go:linkname utimensat syscall.utimensat
func utimensat(dirfd int, path string, times *[2]syscall.Timespec, flags int) error

func utimens(path string, times *[2]syscall.Timespec, symlinkFollow bool) error {
	var flags int
	if !symlinkFollow {
		flags = _AT_SYMLINK_NOFOLLOW
	}
	return utimensat(_AT_FDCWD, path, times, flags)
}

func futimens(fd uintptr, times *[2]syscall.Timespec) error {
	_p0 := timesToPtr(times)

	// Warning: futimens only exists since High Sierra (10.13).
	_, _, e1 := syscall_syscall6(libc_futimens_trampoline_addr, fd, uintptr(_p0), 0, 0, 0, 0)
	if e1 != 0 {
		return e1
	}
	return nil
}

// libc_futimens_trampoline_addr is the address of the
// `libc_futimens_trampoline` symbol, defined in `futimens_darwin.s`.
//
// We use this to invoke the syscall through syscall_syscall6 imported below.
var libc_futimens_trampoline_addr uintptr

// Imports the futimens symbol from libc as `libc_futimens`.
//
// Note: CGO mechanisms are used in darwin regardless of the CGO_ENABLED value
// or the "cgo" build flag. See /RATIONALE.md for why.
//go:cgo_import_dynamic libc_futimens futimens "/usr/lib/libSystem.B.dylib"
