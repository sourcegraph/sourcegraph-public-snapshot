package search

import (
	"context"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	proto "github.com/sourcegraph/sourcegraph/internal/searcher/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	Service *Service
	proto.UnimplementedSearcherServiceServer
}

func (s *Server) Search(req *proto.SearchRequest, stream proto.SearcherService_SearchServer) error {
	var unmarshaledReq protocol.Request
	unmarshaledReq.FromProto(req)

	// mu protects the stream from concurrent writes.
	var mu sync.Mutex
	onMatches := func(match protocol.FileMatch) {
		mu.Lock()
		defer mu.Unlock()

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
		return convertToGRPCError(ctx, err)
	}

	return stream.Send(&proto.SearchResponse{
		Message: &proto.SearchResponse_DoneMessage{
			DoneMessage: &proto.SearchResponse_Done{
				LimitHit: matchStream.LimitHit(),
			},
		},
	})
}

// convertToGRPCError converts an error into a gRPC status error code.
//
// If err is nil, it returns nil.
//
// If err is already a gRPC status error, it is returned as-is.
//
// If the provided context has expired, a grpc codes.Canceled / DeadlineExceeded error is returned.
//
// If the err is a well-known error (such as a process getting killed, etc.),
// it's mapped to the appropriate gRPC status code.
//
// Otherwise, err is converted to an Unknown gRPC error code.
func convertToGRPCError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	// don't convert an existing status error
	if statusErr, ok := status.FromError(err); ok {
		return statusErr.Err()
	}

	// if the context expired, just return that
	if ctxErr := ctx.Err(); ctxErr != nil {
		return status.FromContextError(ctxErr).Err()
	}

	// otherwise convert to a status error
	grpcCode := codes.Unknown
	if strings.Contains(err.Error(), "signal: killed") {
		grpcCode = codes.Aborted
	}

	return status.New(grpcCode, err.Error()).Err()
}
