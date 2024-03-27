package internal

import (
	"github.com/sourcegraph/sourcegraph/internal/diskusage"
	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
)

// getDiskInfo returns disk usage info for the gitserver.
//
// It calculates the total and free disk space for the gitserver's repo
// directory using du.DiskUsage. The results are returned as a
// protocol.DiskInfoResponse struct.
func getDiskInfo(dir string) (*proto.DiskInfoResponse, error) {
	usage, err := diskusage.New(dir)
	if err != nil {
		return nil, err
	}
	return &proto.DiskInfoResponse{
		TotalSpace:  usage.Size(),
		FreeSpace:   usage.Free(),
		PercentUsed: usage.PercentUsed(),
	}, nil
}
