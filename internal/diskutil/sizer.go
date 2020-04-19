package diskutil

import (
	"syscall"

	"github.com/pkg/errors"
)

// DiskSizer gets information about disk size and free space.
type DiskSizer interface {
	// MountPoint returns the root of the device.
	MountPoint() string

	// Size returns the total and available size of a disk in bytes.
	Size() (diskSizeBytes uint64, freeBytes uint64, err error)
}

type statDiskSizer struct {
	mountPoint string
}

// NewDiskSizer creates a DiskSizer for the mount point containing
// directory d.
func NewDiskSizer(d string) (DiskSizer, error) {
	mountPoint, err := findMountPoint(d)
	if err != nil {
		return nil, err
	}

	return statDiskSizer{
		mountPoint: mountPoint,
	}, nil
}

func (s statDiskSizer) MountPoint() string {
	return s.mountPoint
}

func (s statDiskSizer) Size() (uint64, uint64, error) {
	var fs syscall.Statfs_t
	if err := syscall.Statfs(s.mountPoint, &fs); err != nil {
		return 0, 0, errors.Wrap(err, "statting")
	}

	return fs.Blocks * uint64(fs.Bsize), fs.Bavail * uint64(fs.Bsize), nil
}
