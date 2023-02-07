package search

import (
	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/searcher/v1"
)

type Server struct {
	Service *Service
	v1.UnimplementedSearcherServiceServer
}

func (s *Server) Search(req *v1.SearchRequest, stream v1.SearcherService_SearchServer) error {
	var unmarshaledReq protocol.Request
	unmarshaledReq.FromProto(req)

	onMatches := func(match protocol.FileMatch) {
		stream.Send(&v1.SearchResponse{
			Message: &v1.SearchResponse_FileMatch{
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

	return stream.Send(&v1.SearchResponse{
		Message: &v1.SearchResponse_DoneMessage{
			DoneMessage: &v1.SearchResponse_Done{
				LimitHit: matchStream.LimitHit(),
			},
		},
	})
}
