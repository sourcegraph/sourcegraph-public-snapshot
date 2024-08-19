package sysfs

import (
	"io"
	"syscall"

	"github.com/tetratelabs/wazero/internal/fsapi"
	"github.com/tetratelabs/wazero/internal/platform"
)

func adjustReaddirErr(f fsapi.File, isClosed bool, err error) syscall.Errno {
	if err == io.EOF {
		return 0 // e.g. Readdir on darwin returns io.EOF, but linux doesn't.
	} else if errno := platform.UnwrapOSError(err); errno != 0 {
		errno = dirError(f, isClosed, errno)
		// Comply with errors allowed on fsapi.File Readdir
		switch errno {
		case syscall.EINVAL: // os.File Readdir can return this
			return syscall.EBADF
		case syscall.ENOTDIR: // dirError can return this
			return syscall.EBADF
		}
		return errno
	}
	return 0
}
