//go:build cgo

package api

import (
	"context"
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/proto"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/squirrel"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/types"
	internaltypes "github.com/sourcegraph/sourcegraph/internal/types"
)

// addHandlers adds handlers that require cgo.
func addHandlers(
	mux *http.ServeMux,
	searchFunc types.SearchFunc,
	readFileFunc func(context.Context, internaltypes.RepoCommitPath) ([]byte, error),
) {
	mux.HandleFunc("/localCodeIntel", squirrel.LocalCodeIntelHandler(readFileFunc))
	mux.HandleFunc("/debugLocalCodeIntel", squirrel.DebugLocalCodeIntelHandler)
	mux.HandleFunc("/symbolInfo", squirrel.NewSymbolInfoHandler(searchFunc, readFileFunc))
}

// LocalCodeIntel returns local code intelligence for the given file and commit
func (s *grpcServer) LocalCodeIntel(ctx context.Context, request *proto.LocalCodeIntelRequest) (*proto.LocalCodeIntelResponse, error) {
	squirrelService := squirrel.New(s.readFileFunc, nil) // TODO:@ggilmore: why is the symbolsearch field hard-coded to nil?
	defer squirrelService.Close()

	var args internaltypes.RepoCommitPath
	args.FromProto(request.RepoCommitPath)

	payload, err := squirrelService.LocalCodeIntel(ctx, args)
	if err != nil {
		return nil, err
	}

	return payload.ToProto(), nil
}

// SymbolInfo returns information about the symbols specified by the given request.
func (s *grpcServer) SymbolInfo(ctx context.Context, request *proto.SymbolInfoRequest) (*proto.SymbolInfoResponse, error) {
	squirrelService := squirrel.New(s.readFileFunc, s.searchFunc)
	defer squirrelService.Close()

	var args internaltypes.RepoCommitPathPoint
	args.RepoCommitPath.FromProto(request.RepoCommitPath)
	args.Point.FromProto(request.Point)

	info, err := squirrelService.SymbolInfo(ctx, args)
	if err != nil {
		return nil, err
	}

	var response proto.SymbolInfoResponse
	response.Hover = info.Hover
	response.Definition.RepoCommitPath = info.Definition.RepoCommitPath.ToProto()
	response.Definition.Range = info.Definition.Range.ToProto()

	return &response, nil
}
