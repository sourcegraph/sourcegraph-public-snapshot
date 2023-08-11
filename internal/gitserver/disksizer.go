package gitserver

import "github.com/ricochet2200/go-disk-usage/du"

// DiskSizer gets information about disk size and free space.
type DiskSizer interface {
	BytesFreeOnDisk(mountPoint string) (uint64, error)
	DiskSizeBytes(mountPoint string) (uint64, error)
}

func NewStatDiskSizer() DiskSizer {
	return &statDiskSizer{}
}

type statDiskSizer struct{}

func (s *statDiskSizer) BytesFreeOnDisk(mountPoint string) (uint64, error) {
	usage := du.NewDiskUsage(mountPoint)
	return usage.Available(), nil
}

func (s *statDiskSizer) DiskSizeBytes(mountPoint string) (uint64, error) {
	usage := du.NewDiskUsage(mountPoint)
	return usage.Size(), nil
}
