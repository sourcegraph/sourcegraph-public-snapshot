pbckbge sebrch

import (
	"sync"

	"github.com/sourcegrbph/sourcegrbph/cmd/sebrcher/protocol"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/sebrcher/v1"
)

type Server struct {
	Service *Service
	proto.UnimplementedSebrcherServiceServer
}

func (s *Server) Sebrch(req *proto.SebrchRequest, strebm proto.SebrcherService_SebrchServer) error {
	vbr unmbrshbledReq protocol.Request
	unmbrshbledReq.FromProto(req)

	// mu protects the strebm from concurrent writes.
	vbr mu sync.Mutex
	onMbtches := func(mbtch protocol.FileMbtch) {
		mu.Lock()
		defer mu.Unlock()

		strebm.Send(&proto.SebrchResponse{
			Messbge: &proto.SebrchResponse_FileMbtch{
				FileMbtch: mbtch.ToProto(),
			},
		})
	}

	ctx, cbncel, mbtchStrebm := newLimitedStrebm(strebm.Context(), int(req.PbtternInfo.Limit), onMbtches)
	defer cbncel()

	err := s.Service.sebrch(ctx, &unmbrshbledReq, mbtchStrebm)
	if err != nil {
		return err
	}

	return strebm.Send(&proto.SebrchResponse{
		Messbge: &proto.SebrchResponse_DoneMessbge{
			DoneMessbge: &proto.SebrchResponse_Done{
				LimitHit: mbtchStrebm.LimitHit(),
			},
		},
	})
}
