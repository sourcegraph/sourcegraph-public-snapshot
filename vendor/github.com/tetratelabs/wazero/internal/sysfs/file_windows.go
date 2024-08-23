package sysfs

import (
	"syscall"
	"unsafe"

	"github.com/tetratelabs/wazero/internal/platform"
)

const NonBlockingFileIoSupported = true

var kernel32 = syscall.NewLazyDLL("kernel32.dll")

// procPeekNamedPipe is the syscall.LazyProc in kernel32 for PeekNamedPipe
var procPeekNamedPipe = kernel32.NewProc("PeekNamedPipe")

// readFd returns ENOSYS on unsupported platforms.
//
// PeekNamedPipe: https://learn.microsoft.com/en-us/windows/win32/api/namedpipeapi/nf-namedpipeapi-peeknamedpipe
// "GetFileType can assist in determining what device type the handle refers to. A console handle presents as FILE_TYPE_CHAR."
// https://learn.microsoft.com/en-us/windows/console/console-handles
func readFd(fd uintptr, buf []byte) (int, syscall.Errno) {
	handle := syscall.Handle(fd)
	fileType, err := syscall.GetFileType(syscall.Stdin)
	if err != nil {
		return 0, platform.UnwrapOSError(err)
	}
	if fileType&syscall.FILE_TYPE_CHAR == 0 {
		return -1, syscall.ENOSYS
	}
	n, err := peekNamedPipe(handle)
	if err != nil {
		errno := platform.UnwrapOSError(err)
		if errno == syscall.ERROR_BROKEN_PIPE {
			return 0, 0
		}
	}
	if n == 0 {
		return -1, syscall.EAGAIN
	}
	un, err := syscall.Read(handle, buf[0:n])
	return un, platform.UnwrapOSError(err)
}

// peekNamedPipe partially exposes PeekNamedPipe from the Win32 API
// see https://learn.microsoft.com/en-us/windows/win32/api/namedpipeapi/nf-namedpipeapi-peeknamedpipe
func peekNamedPipe(handle syscall.Handle) (uint32, error) {
	var totalBytesAvail uint32
	totalBytesPtr := unsafe.Pointer(&totalBytesAvail)
	_, _, err := procPeekNamedPipe.Call(
		uintptr(handle),        // [in]            HANDLE  hNamedPipe,
		0,                      // [out, optional] LPVOID  lpBuffer,
		0,                      // [in]            DWORD   nBufferSize,
		0,                      // [out, optional] LPDWORD lpBytesRead
		uintptr(totalBytesPtr), // [out, optional] LPDWORD lpTotalBytesAvail,
		0)                      // [out, optional] LPDWORD lpBytesLeftThisMessage
	if err == syscall.Errno(0) {
		return totalBytesAvail, nil
	}
	return totalBytesAvail, err
}
