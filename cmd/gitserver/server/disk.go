package server

import (
	"github.com/ricochet2200/go-disk-usage/du"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

func (s *Server) GetDiskInfo() protocol.DiskInfoResponse {
	usage := du.NewDiskUsage(s.ReposDir)
	usage.Available()
	return protocol.DiskInfoResponse{
		TotalSpace: usage.Size(),
		FreeSpace:  usage.Free(),
	}
}
