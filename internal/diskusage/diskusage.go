// Copied from https://sourcegraph.com/github.com/ricochet2200/go-disk-usage
package diskusage

import "syscall"

type DiskUsage interface {
	Free() uint64
	Size() uint64
	PercentUsed() float32
	Available() uint64
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

// Used returns total bytes used in file system
func (du *diskUsage) used() uint64 {
	return du.Size() - du.Free()
}

func (du *diskUsage) usage() float32 {
	return float32(du.used()) / float32(du.Size())
}

// PercentUsed returns percentage of use on the file system
func (du *diskUsage) PercentUsed() float32 {
	return du.usage() * 100
}

// Available return total available bytes on file system to an unprivileged user
func (du *diskUsage) Available() uint64 {
	return du.stat.Bavail * uint64(du.stat.Bsize)
}
