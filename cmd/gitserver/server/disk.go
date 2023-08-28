package server

import (
	"net/http"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/sourcegraph/sourcegraph/internal/diskusage"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
)

// getDiskInfo returns disk usage info for the gitserver.
//
// It calculates the total and free disk space for the gitserver's repo
// directory using du.DiskUsage. The results are returned as a
// protocol.DiskInfoResponse struct.
func getDiskInfo(dir string) (*protocol.DiskInfoResponse, error) {
	usage, err := diskusage.New(dir)
	if err != nil {
		return nil, err
	}
	return &protocol.DiskInfoResponse{
		TotalSpace: usage.Size(),
		FreeSpace:  usage.Free(),
	}, nil
}

func (s *Server) handleDiskInfo(w http.ResponseWriter, r *http.Request) {
	resp, err := getDiskInfo(s.ReposDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
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
