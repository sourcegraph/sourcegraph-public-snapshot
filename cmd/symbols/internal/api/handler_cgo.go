//go:build cgo

package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/squirrel"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/types"
	"github.com/sourcegraph/sourcegraph/internal/grpc/chunk"
	proto "github.com/sourcegraph/sourcegraph/internal/symbols/v1"
	internaltypes "github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
func (s *grpcService) LocalCodeIntel(request *proto.LocalCodeIntelRequest, ss proto.SymbolsService_LocalCodeIntelServer) error {
	squirrelService := squirrel.New(s.readFileFunc, nil)
	defer squirrelService.Close()

	args := request.GetRepoCommitPath().ToInternal()

	ctx := ss.Context()
	payload, err := squirrelService.LocalCodeIntel(ctx, args)
	if err != nil {
		if errors.Is(err, squirrel.UnsupportedLanguageError) {
			return status.Error(codes.Unimplemented, err.Error())
		}

		if ctxErr := ctx.Err(); ctxErr != nil {
			return status.FromContextError(ctxErr).Err()
		}

		return err
	}

	var response proto.LocalCodeIntelResponse
	response.FromInternal(payload)

	sendFunc := func(symbols []*proto.LocalCodeIntelResponse_Symbol) error {
		return ss.Send(&proto.LocalCodeIntelResponse{Symbols: symbols})
	}

	chunker := chunk.New[*proto.LocalCodeIntelResponse_Symbol](sendFunc)
	err = chunker.Send(response.GetSymbols()...)
	if err != nil {
		return errors.Wrap(err, "sending response")
	}

	err = chunker.Flush()
	if err != nil {
		return errors.Wrap(err, "flushing response stream")
	}

	return nil
}

// SymbolInfo returns information about the symbols specified by the given request.
func (s *grpcService) SymbolInfo(ctx context.Context, request *proto.SymbolInfoRequest) (*proto.SymbolInfoResponse, error) {
	squirrelService := squirrel.New(s.readFileFunc, s.searchFunc)
	defer squirrelService.Close()

	var args internaltypes.RepoCommitPathPoint

	args.RepoCommitPath = request.GetRepoCommitPath().ToInternal()
	args.Point = request.GetPoint().ToInternal()

	panic("GRPC panic!!!!")

	fmt.Printf("grpc: Called squirrel.SymbolInfo(%v)\n", args.RepoCommitPath)

	info, err := squirrelService.SymbolInfo(ctx, args)
	if err != nil {
		fmt.Printf("grpc: Got squirrel error, %v (UFEE=%v)\n", err.Error(),
			errors.Is(err, squirrel.UnrecognizedFileExtensionError))
		if errors.Is(err, squirrel.UnrecognizedFileExtensionError) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if errors.Is(err, squirrel.UnsupportedLanguageError) {
			return nil, status.Error(codes.Unimplemented, err.Error())
		}

		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, status.FromContextError(ctxErr).Err()
		}

		return nil, err
	}

	var response proto.SymbolInfoResponse
	response.FromInternal(info)

	return &response, nil
}
