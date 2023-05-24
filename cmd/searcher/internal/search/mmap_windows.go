package search

import (
	"io/fs"
	"os"

	mmapgo "github.com/edsrzf/mmap-go"
)

func mmap(path string, f *os.File, fi fs.FileInfo) ([]byte, error) {
	return mmapgo.Map(f, mmapgo.RDONLY, 0)
}

func unmap(data mmapgo.MMap) error {
	return data.Unmap()
}
