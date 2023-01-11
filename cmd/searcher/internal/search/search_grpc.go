package search

import (
	"github.com/sourcegraph/sourcegraph/cmd/searcher/proto"
	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
)

type Server struct {
	Service *Service
	proto.UnimplementedSearcherServer
}

func (s *Server) Search(req *proto.SearchRequest, stream proto.Searcher_SearchServer) error {
	var unmarshaledReq protocol.Request
	unmarshaledReq.FromProto(req)

	onMatches := func(match protocol.FileMatch) {
		stream.Send(&proto.SearchResponse{
			FileMatch:   match.ToProto(),
			LimitHit:    false,
			DeadlineHit: false,
		})
	}

	ctx, cancel, matchStream := newLimitedStream(stream.Context(), int(req.PatternInfo.Limit), onMatches)
	defer cancel()

	return s.Service.search(ctx, &unmarshaledReq, matchStream)
}
