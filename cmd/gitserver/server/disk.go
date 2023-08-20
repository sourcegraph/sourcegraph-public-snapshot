package server

import (
	"net/http"

	"github.com/ricochet2200/go-disk-usage/du"
	"google.golang.org/protobuf/encoding/protojson"

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

	protoResponse := resp.ToProto()

	jsonBytes, err := protojson.Marshal(protoResponse)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}
