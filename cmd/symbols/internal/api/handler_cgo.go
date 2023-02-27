//go:build cgo

package api

import (
	"context"
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/squirrel"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/types"
	proto "github.com/sourcegraph/sourcegraph/internal/symbols/v1"
	internaltypes "github.com/sourcegraph/sourcegraph/internal/types"
)

// addHandlers adds handlers that require cgo.
func addHandlers(
	mux *http.ServeMux,
	searchFunc types.SearchFunc,
	readFileFunc func(context.Context, internaltypes.RepoCommitPath) ([]byte, error),
) {
	mux.HandleFunc("/localCodeIntel", squirrel.LocalCodeIntelHandler(readFileFunc))
	mux.HandleFunc("/symbolInfo", squirrel.NewSymbolInfoHandler(searchFunc, readFileFunc))
}

// LocalCodeIntel returns local code intelligence for the given file and commit
func (s *grpcService) LocalCodeIntel(ctx context.Context, request *proto.LocalCodeIntelRequest) (*proto.LocalCodeIntelResponse, error) {
	squirrelService := squirrel.New(s.readFileFunc, nil)
	defer squirrelService.Close()

	args := request.GetRepoCommitPath().ToInternal()

	payload, err := squirrelService.LocalCodeIntel(ctx, args)
	if err != nil {
		// TODO(camdencheek): This ignores errors from LocalCodeIntel to match the behavior found here:
		// https://sourcegraph.com/github.com/sourcegraph/sourcegraph@a1631d58604815917096acc3356447c55baebf22/-/blob/cmd/symbols/squirrel/http_handlers.go?L57-57
		//
		// This is weird, and maybe not intentional, but things break if we return an error.
		return nil, nil
	}

	var response proto.LocalCodeIntelResponse
	response.FromInternal(payload)

	return &response, nil
}

// SymbolInfo returns information about the symbols specified by the given request.
func (s *grpcService) SymbolInfo(ctx context.Context, request *proto.SymbolInfoRequest) (*proto.SymbolInfoResponse, error) {
	squirrelService := squirrel.New(s.readFileFunc, s.searchFunc)
	defer squirrelService.Close()

	var args internaltypes.RepoCommitPathPoint

	args.RepoCommitPath = request.GetRepoCommitPath().ToInternal()
	args.Point = request.GetPoint().ToInternal()

	info, err := squirrelService.SymbolInfo(ctx, args)
	if err != nil {
		return nil, err
	}

	var response proto.SymbolInfoResponse
	response.FromInternal(info)

	return &response, nil
}
