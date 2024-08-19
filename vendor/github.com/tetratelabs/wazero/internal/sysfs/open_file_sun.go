//go:build illumos || solaris

package sysfs

import (
	"io/fs"
	"os"
	"syscall"

	"github.com/tetratelabs/wazero/internal/platform"
)

func openFile(path string, flag int, perm fs.FileMode) (*os.File, syscall.Errno) {
	f, err := os.OpenFile(path, flag, perm)
	return f, platform.UnwrapOSError(err)
}
