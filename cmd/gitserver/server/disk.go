package server

import (
	"encoding/json"
	"net/http"

	"github.com/ricochet2200/go-disk-usage/du"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

func (s *Server) getDiskInfo() protocol.DiskInfoResponse {
	usage := du.NewDiskUsage(s.ReposDir)
	return protocol.DiskInfoResponse{
		TotalSpace: usage.Size(),
		FreeSpace:  usage.Free(),
	}
}

func (s *Server) handleDiskInfo(w http.ResponseWriter, r *http.Request) {
	resp := s.getDiskInfo()
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
