package shared

import "github.com/ricochet2200/go-disk-usage/du"

type StatDiskSizer struct{}

func (s *StatDiskSizer) BytesFreeOnDisk(mountPoint string) (uint64, error) {
	usage := du.NewDiskUsage(mountPoint)
	return usage.Available(), nil
}

func (s *StatDiskSizer) DiskSizeBytes(mountPoint string) (uint64, error) {
	usage := du.NewDiskUsage(mountPoint)
	return usage.Size(), nil
}
