package sysfs

import (
	"io/fs"
	"syscall"

	"github.com/tetratelabs/wazero/internal/platform"
	"github.com/tetratelabs/wazero/sys"
)

func defaultStatFile(f fs.File) (sys.Stat_t, syscall.Errno) {
	if info, err := f.Stat(); err != nil {
		return sys.Stat_t{}, platform.UnwrapOSError(err)
	} else {
		return sys.NewStat_t(info), 0
	}
}
