package search

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/diskcache"
	proto "github.com/sourcegraph/sourcegraph/internal/searcher/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

func (s *Server) CacheState(_ context.Context, req *proto.CacheStateRequest) (*proto.CacheStateResponse, error) {
	repo := api.RepoName(req.Repo)
	commit := api.CommitID(req.CommitOid)

	// TODO hybrid search? We need to correctly represent that state of hybrid
	// search here. It sets path (which we use as nil). We likely won't be
	// able to fully do this, but we should atleast store somewhere so that
	// CacheState returns something sensical (right now it will just always
	// say MISSING)
	var state proto.CacheState
	fromState := s.Service.Store.CacheState(repo, commit, nil)
	switch fromState {
	case diskcache.StateMissing:
		state = proto.CacheState_MISSING
	case diskcache.StateFetching:
		state = proto.CacheState_FETCHING
	case diskcache.StateCached:
		state = proto.CacheState_CACHED
	default:
		return nil, errors.Errorf("unknown diskcache.State %q", fromState)
	}

	return &proto.CacheStateResponse{
		CacheState: state,
	}, nil
}
