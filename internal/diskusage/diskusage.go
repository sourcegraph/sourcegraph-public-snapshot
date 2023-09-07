// Copied from https://sourcegraph.com/github.com/ricochet2200/go-disk-usage
package diskusage

import "syscall"

type DiskUsage interface {
	Free() uint64
	Size() uint64
}

// DiskUsage contains usage data and provides user-friendly access methods
type diskUsage struct {
	stat *syscall.Statfs_t
}

// New returns an object holding the disk usage of volumePath
// or nil in case of error (invalid path, etc)
func New(volumePath string) (DiskUsage, error) {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(volumePath, &stat); err != nil {
		return nil, err
	}
	return &diskUsage{stat: &stat}, nil
}

// Free returns total free bytes on file system
func (du *diskUsage) Free() uint64 {
	return du.stat.Bfree * uint64(du.stat.Bsize)
}

// Size returns total size of the file system
func (du *diskUsage) Size() uint64 {
	return uint64(du.stat.Blocks) * uint64(du.stat.Bsize)
}
