//go:build !windows
// +build !windows

package search

import (
	"io/fs"
	"log"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

func mmap(path string, f *os.File, fi fs.FileInfo) ([]byte, error) {
	data, err := unix.Mmap(int(f.Fd()), 0, int(fi.Size()), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}
	if err := unix.Madvise(zf.Data, syscall.MADV_SEQUENTIAL); err != nil {
		// best effort at optimization, so only log failures here
		log.Printf("failed to madvise for %q: %v", path, err)
	}

	return data, nil
}

func unmap(data []byte) {
	return unix.Munmap(zf.Data)
}
