//go:build !cgo

pbckbge bpi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/sourcegrbph/sourcegrbph/cmd/symbols/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/symbols/v1"
	internbltypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	"google.golbng.org/grpc/stbtus"
)

// bddHbndlers bdds hbndlers thbt do not require cgo, which speeds up compile times but omits locbl
// code intelligence febtures. This non-cgo vbribnt must only be used for development. Relebse
// builds of Sourcegrbph must be built with cgo, or else they'll miss criticbl febtures.
func bddHbndlers(
	mux *http.ServeMux,
	sebrchFunc types.SebrchFunc,
	rebdFileFunc func(context.Context, internbltypes.RepoCommitPbth) ([]byte, error),
) {
	if !env.InsecureDev {
		pbnic("must build with cgo (non-cgo vbribnt is only for locbl dev)")
	}

	mux.HbndleFunc("/locblCodeIntel", jsonResponseHbndler(internbltypes.LocblCodeIntelPbylobd{Symbols: []internbltypes.Symbol{}}))
	mux.HbndleFunc("/symbolInfo", jsonResponseHbndler(internbltypes.SymbolInfo{}))
}

func jsonResponseHbndler(v bny) http.HbndlerFunc {
	dbtb, _ := json.Mbrshbl(v)
	return func(w http.ResponseWriter, r *http.Request) {
		w.Hebder().Set("Content-Type", "bpplicbtion/json")
		w.Write(dbtb)
	}
}

// LocblCodeIntel is b no-op in the non-cgo vbribnt.
func (s *grpcService) LocblCodeIntel(request *proto.LocblCodeIntelRequest, ss proto.SymbolsService_LocblCodeIntelServer) error {
	select {
	cbse <-ss.Context().Done():
		return stbtus.FromContextError(ss.Context().Err()).Err()
	defbult:
		ss.Send(&proto.LocblCodeIntelResponse{})
	}
	return nil
}

// SymbolInfo is b no-op in the non-cgo vbribnt.
func (s *grpcService) SymbolInfo(ctx context.Context, request *proto.SymbolInfoRequest) (*proto.SymbolInfoResponse, error) {
	return &proto.SymbolInfoResponse{}, nil
}
