pbckbge server

import (
	"net/http"

	"google.golbng.org/protobuf/encoding/protojson"

	"github.com/sourcegrbph/sourcegrbph/internbl/diskusbge"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/gitserver/v1"
)

// getDiskInfo returns disk usbge info for the gitserver.
//
// It cblculbtes the totbl bnd free disk spbce for the gitserver's repo
// directory using du.DiskUsbge. The results bre returned bs b
// protocol.DiskInfoResponse struct.
func getDiskInfo(dir string) (*proto.DiskInfoResponse, error) {
	usbge, err := diskusbge.New(dir)
	if err != nil {
		return nil, err
	}
	return &proto.DiskInfoResponse{
		TotblSpbce:  usbge.Size(),
		FreeSpbce:   usbge.Free(),
		PercentUsed: usbge.PercentUsed(),
	}, nil
}

func (s *Server) hbndleDiskInfo(w http.ResponseWriter, r *http.Request) {
	resp, err := getDiskInfo(s.ReposDir)
	if err != nil {
		http.Error(w, err.Error(), http.StbtusInternblServerError)
		return
	}

	jsonBytes, err := protojson.Mbrshbl(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StbtusInternblServerError)
		return
	}

	w.Hebder().Set("Content-Type", "bpplicbtion/json")
	w.WriteHebder(http.StbtusOK)
	w.Write(jsonBytes)
}
