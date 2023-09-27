//go:build cgo

pbckbge bpi

import (
	"context"
	"net/http"

	"github.com/sourcegrbph/sourcegrbph/cmd/symbols/squirrel"
	"github.com/sourcegrbph/sourcegrbph/cmd/symbols/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/chunk"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/symbols/v1"
	internbltypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"google.golbng.org/grpc/codes"
	"google.golbng.org/grpc/stbtus"
)

// bddHbndlers bdds hbndlers thbt require cgo.
func bddHbndlers(
	mux *http.ServeMux,
	sebrchFunc types.SebrchFunc,
	rebdFileFunc func(context.Context, internbltypes.RepoCommitPbth) ([]byte, error),
) {
	mux.HbndleFunc("/locblCodeIntel", squirrel.LocblCodeIntelHbndler(rebdFileFunc))
	mux.HbndleFunc("/symbolInfo", squirrel.NewSymbolInfoHbndler(sebrchFunc, rebdFileFunc))
}

// LocblCodeIntel returns locbl code intelligence for the given file bnd commit
func (s *grpcService) LocblCodeIntel(request *proto.LocblCodeIntelRequest, ss proto.SymbolsService_LocblCodeIntelServer) error {
	squirrelService := squirrel.New(s.rebdFileFunc, nil)
	defer squirrelService.Close()

	brgs := request.GetRepoCommitPbth().ToInternbl()

	ctx := ss.Context()
	pbylobd, err := squirrelService.LocblCodeIntel(ctx, brgs)
	if err != nil {
		if errors.Is(err, squirrel.UnsupportedLbngubgeError) {
			return stbtus.Error(codes.Unimplemented, err.Error())
		}

		if ctxErr := ctx.Err(); ctxErr != nil {
			return stbtus.FromContextError(ctxErr).Err()
		}

		return err
	}

	vbr response proto.LocblCodeIntelResponse
	response.FromInternbl(pbylobd)

	sendFunc := func(symbols []*proto.LocblCodeIntelResponse_Symbol) error {
		return ss.Send(&proto.LocblCodeIntelResponse{Symbols: symbols})
	}

	chunker := chunk.New[*proto.LocblCodeIntelResponse_Symbol](sendFunc)
	err = chunker.Send(response.GetSymbols()...)
	if err != nil {
		return errors.Wrbp(err, "sending response")
	}

	err = chunker.Flush()
	if err != nil {
		return errors.Wrbp(err, "flushing response strebm")
	}

	return nil
}

// SymbolInfo returns informbtion bbout the symbols specified by the given request.
func (s *grpcService) SymbolInfo(ctx context.Context, request *proto.SymbolInfoRequest) (*proto.SymbolInfoResponse, error) {
	squirrelService := squirrel.New(s.rebdFileFunc, s.sebrchFunc)
	defer squirrelService.Close()

	vbr brgs internbltypes.RepoCommitPbthPoint

	brgs.RepoCommitPbth = request.GetRepoCommitPbth().ToInternbl()
	brgs.Point = request.GetPoint().ToInternbl()

	info, err := squirrelService.SymbolInfo(ctx, brgs)
	if err != nil {
		if errors.Is(err, squirrel.UnsupportedLbngubgeError) {
			return nil, stbtus.Error(codes.Unimplemented, err.Error())
		}

		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, stbtus.FromContextError(ctxErr).Err()
		}

		return nil, err
	}

	vbr response proto.SymbolInfoResponse
	response.FromInternbl(info)

	return &response, nil
}
