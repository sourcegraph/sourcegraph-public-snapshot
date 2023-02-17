package search

import (
	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	proto "github.com/sourcegraph/sourcegraph/internal/searcher/v1"
)

type Server struct {
	Service *Service
	proto.UnimplementedSearcherServiceServer
}

func (s *Server) Search(req *proto.SearchRequest, stream proto.SearcherService_SearchServer) error {
	var unmarshaledReq protocol.Request
	unmarshaledReq.FromProto(req)

	onMatches := func(match protocol.FileMatch) {
		stream.Send(&proto.SearchResponse{
			Message: &proto.SearchResponse_FileMatch{
				FileMatch: match.ToProto(),
			},
		})
	}

	ctx, cancel, matchStream := newLimitedStream(stream.Context(), int(req.PatternInfo.Limit), onMatches)
	defer cancel()

	err := s.Service.search(ctx, &unmarshaledReq, matchStream)
	if err != nil {
		return err
	}

	return stream.Send(&proto.SearchResponse{
		Message: &proto.SearchResponse_DoneMessage{
			DoneMessage: &proto.SearchResponse_Done{
				LimitHit: matchStream.LimitHit(),
			},
		},
	})
}
