//go:build !windows
// +build !windows

package search

import (
	"io/fs"
	"os"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/sourcegraph/log"
)

func mmap(path string, f *os.File, fi fs.FileInfo) ([]byte, error) {
	data, err := unix.Mmap(int(f.Fd()), 0, int(fi.Size()), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}
	if err := unix.Madvise(data, syscall.MADV_SEQUENTIAL); err != nil {
		// best effort at optimization, so only log failures here
		log.Scoped("mmap").Info("failed to madvise", log.String("path", path), log.Error(err))
	}

	return data, nil
}

func unmap(data []byte) error {
	return unix.Munmap(data)
}
