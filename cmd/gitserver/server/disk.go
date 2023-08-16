package server

import "github.com/ricochet2200/go-disk-usage/du"

type GetDiskInfoResult struct {
	TotalSpace uint64
	FreeSpace  uint64
}

func (s *Server) GetDiskInfo() GetDiskInfoResult {
	usage := du.NewDiskUsage(s.ReposDir)
	usage.Available()
	return GetDiskInfoResult{
		TotalSpace: usage.Size(),
		FreeSpace:  usage.Free(),
	}
}
