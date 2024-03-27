//go:build !cgo

package api

import (
	"context"
	"encoding/json"
	"net/http"

	"google.golang.org/grpc/status"

	proto "github.com/sourcegraph/sourcegraph/internal/symbols/v1"
)

func jsonResponseHandler(v any) http.HandlerFunc {
	data, _ := json.Marshal(v)
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}
}

// LocalCodeIntel is a no-op in the non-cgo variant.
func (s *grpcService) LocalCodeIntel(request *proto.LocalCodeIntelRequest, ss proto.SymbolsService_LocalCodeIntelServer) error {
	select {
	case <-ss.Context().Done():
		return status.FromContextError(ss.Context().Err()).Err()
	default:
		ss.Send(&proto.LocalCodeIntelResponse{})
	}
	return nil
}

// SymbolInfo is a no-op in the non-cgo variant.
func (s *grpcService) SymbolInfo(ctx context.Context, request *proto.SymbolInfoRequest) (*proto.SymbolInfoResponse, error) {
	return &proto.SymbolInfoResponse{}, nil
}
