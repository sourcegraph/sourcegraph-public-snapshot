package server

import (
	"encoding/json"
	"net/http"

	"github.com/ricochet2200/go-disk-usage/du"

	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

// getDiskInfo returns disk usage info for the gitserver.
//
// It calculates the total and free disk space for the gitserver's repo
// directory using du.DiskUsage. The results are returned as a
// protocol.DiskInfoResponse struct.
func (s *Server) getDiskInfo() protocol.DiskInfoResponse {
	usage := du.NewDiskUsage(s.ReposDir)
	return protocol.DiskInfoResponse{
		TotalSpace: usage.Size(),
		FreeSpace:  usage.Free(),
	}
}

func (s *Server) handleDiskInfo(w http.ResponseWriter, r *http.Request) {
	resp := s.getDiskInfo()
	// Encode and write the disk info response.
	//
	// Attempts to JSON encode the disk info response object and write it to the
	// HTTP response writer. If encoding fails, writes an internal server error
	// HTTP status code and error message to the response writer instead.
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
