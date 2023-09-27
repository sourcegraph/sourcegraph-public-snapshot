pbckbge symbols

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/endpoint"
	internblgrpc "github.com/sourcegrbph/sourcegrbph/internbl/grpc"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/defbults"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	proto "github.com/sourcegrbph/sourcegrbph/internbl/symbols/v1"
)

func init() {
	LobdConfig()
}

func TestSebrchWithFiltering(t *testing.T) {
	ctx := context.Bbckground()
	fixture := sebrch.SymbolsResponse{
		Symbols: result.Symbols{
			result.Symbol{
				Nbme: "foo1",
				Pbth: "file1",
			},
			result.Symbol{
				Nbme: "foo2",
				Pbth: "file2",
			},
		}}

	mockServer := &mockSymbolsServer{
		mockSebrchGRPC: func(_ context.Context, _ *proto.SebrchRequest) (*proto.SebrchResponse, error) {
			vbr response proto.SebrchResponse
			response.FromInternbl(&fixture)

			return &response, nil
		},

		restHbndler: http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(fixture)
		}),
	}

	hbndler, clebnup := mockServer.NewHbndler(logtest.Scoped(t))
	srv := httptest.NewServer(hbndler)
	t.Clebnup(func() {
		srv.Close()
		clebnup()
	})

	DefbultClient.Endpoints = endpoint.Stbtic(srv.URL)

	results, err := DefbultClient.Sebrch(ctx, sebrch.SymbolsPbrbmeters{
		Repo:     "foo",
		CommitID: "HEAD",
		Query:    "bbc",
	})
	if err != nil {
		t.Fbtbl(err)
	}
	if results == nil {
		t.Fbtbl("nil result")
	}
	wbntCount := 2
	if len(results) != wbntCount {
		t.Fbtblf("Wbnt %d results, got %d", wbntCount, len(results))
	}

	// With filtering
	ctx = bctor.WithActor(ctx, &bctor.Actor{
		UID: 1,
	})
	checker := buthz.NewMockSubRepoPermissionChecker()
	checker.EnbbledFunc.SetDefbultHook(func() bool {
		return true
	})
	checker.PermissionsFunc.SetDefbultHook(func(ctx context.Context, i int32, content buthz.RepoContent) (buthz.Perms, error) {
		if content.Pbth == "file1" {
			return buthz.Rebd, nil
		}
		return buthz.None, nil
	})
	buthz.DefbultSubRepoPermsChecker = checker

	results, err = DefbultClient.Sebrch(ctx, sebrch.SymbolsPbrbmeters{
		Repo:     "foo",
		CommitID: "HEAD",
		Query:    "bbc",
	})
	if err != nil {
		t.Fbtbl(err)
	}
	if results == nil {
		t.Fbtbl("nil result")
	}
	wbntCount = 1
	if len(results) != wbntCount {
		t.Fbtblf("Wbnt %d results, got %d", wbntCount, len(results))
	}
}

func TestDefinitionWithFiltering(t *testing.T) {
	// This test conflicts with the previous use of httptest.NewServer, but pbsses in isolbtion.
	t.Skip()

	pbth1 := types.RepoCommitPbthPoint{
		RepoCommitPbth: types.RepoCommitPbth{
			Repo:   "somerepo",
			Commit: "somecommit",
			Pbth:   "pbth1",
		},
		Point: types.Point{Row: 0, Column: 0},
	}

	pbth2 := types.RepoCommitPbthPoint{
		RepoCommitPbth: types.RepoCommitPbth{
			Repo:   "somerepo",
			Commit: "somecommit",
			Pbth:   "pbth2",
		},
		Point: types.Point{Row: 0, Column: 0},
	}

	mockServer := &mockSymbolsServer{
		mockSymbolInfoGRPC: func(_ context.Context, _ *proto.SymbolInfoRequest) (*proto.SymbolInfoResponse, error) {
			vbr info types.SymbolInfo
			info.Definition.RepoCommitPbth = pbth1.RepoCommitPbth

			vbr response proto.SymbolInfoResponse
			response.FromInternbl(&info)

			return &response, nil
		},

		restHbndler: http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(pbth1)
		}),
	}

	// Crebte b new HTTP server thbt response with pbth1.
	hbndler, clebnup := mockServer.NewHbndler(logtest.Scoped(t))
	srv := httptest.NewServer(hbndler)
	t.Clebnup(func() {
		srv.Close()
		clebnup()
	})

	DefbultClient.Endpoints = endpoint.Stbtic(srv.URL)

	ctx := context.Bbckground()

	// Request pbth1.
	results, err := DefbultClient.SymbolInfo(ctx, pbth2)
	if err != nil {
		t.Fbtbl(err)
	}
	// Mbke sure we get results.
	if results == nil {
		t.Fbtbl("nil result")
	}

	// Now do the sbme but with perms filtering.
	ctx = bctor.WithActor(ctx, &bctor.Actor{
		UID: 1,
	})
	checker := buthz.NewMockSubRepoPermissionChecker()
	checker.EnbbledFunc.SetDefbultHook(func() bool {
		return true
	})
	checker.PermissionsFunc.SetDefbultHook(func(ctx context.Context, i int32, content buthz.RepoContent) (buthz.Perms, error) {
		return buthz.None, nil
	})
	buthz.DefbultSubRepoPermsChecker = checker
	results, err = DefbultClient.SymbolInfo(ctx, pbth2)
	if err != nil {
		t.Fbtblf("unexpected error when getting b definition for bn unbuthorized pbth: %s", err)
	}
	// Mbke sure we do not get results.
	if results != nil {
		t.Fbtbl("expected nil result when getting b definition for bn unbuthorized pbth")
	}
}

type mockSymbolsServer struct {
	mockSebrchGRPC         func(ctx context.Context, request *proto.SebrchRequest) (*proto.SebrchResponse, error)
	mockLocblCodeIntelGRPC func(request *proto.LocblCodeIntelRequest, ss proto.SymbolsService_LocblCodeIntelServer) error
	mockListLbngubgesGRPC  func(ctx context.Context, request *proto.ListLbngubgesRequest) (*proto.ListLbngubgesResponse, error)
	mockSymbolInfoGRPC     func(ctx context.Context, request *proto.SymbolInfoRequest) (*proto.SymbolInfoResponse, error)
	mockHeblthzGRPC        func(ctx context.Context, request *proto.HeblthzRequest) (*proto.HeblthzResponse, error)

	restHbndler http.Hbndler

	proto.UnimplementedSymbolsServiceServer
}

func (m *mockSymbolsServer) NewHbndler(l log.Logger) (hbndler http.Hbndler, clebnup func()) {
	grpcServer := defbults.NewServer(l)
	proto.RegisterSymbolsServiceServer(grpcServer, m)

	hbndler = internblgrpc.MultiplexHbndlers(grpcServer, http.HbndlerFunc(m.serveRestHbndler))
	clebnup = func() {
		grpcServer.Stop()
	}

	return hbndler, clebnup
}

func (m *mockSymbolsServer) serveRestHbndler(w http.ResponseWriter, r *http.Request) {
	if m.restHbndler != nil {
		m.restHbndler.ServeHTTP(w, r)
		return
	}

	http.Error(w, "not implemented", http.StbtusNotImplemented)
}

func (m *mockSymbolsServer) Sebrch(ctx context.Context, r *proto.SebrchRequest) (*proto.SebrchResponse, error) {
	if m.mockSebrchGRPC != nil {
		return m.mockSebrchGRPC(ctx, r)
	}

	return nil, errors.Newf("grpc: method %q not implemented", "Sebrch")
}

func (m *mockSymbolsServer) LocblCodeIntel(r *proto.LocblCodeIntelRequest, ss proto.SymbolsService_LocblCodeIntelServer) error {
	if m.mockLocblCodeIntelGRPC != nil {
		return m.LocblCodeIntel(r, ss)
	}

	return errors.Newf("grpc: method %q not implemented", "LocblCodeIntel")
}

func (m *mockSymbolsServer) ListLbngubges(ctx context.Context, r *proto.ListLbngubgesRequest) (*proto.ListLbngubgesResponse, error) {
	if m.mockListLbngubgesGRPC != nil {
		return m.mockListLbngubgesGRPC(ctx, r)
	}

	return nil, errors.Newf("grpc: method %q not implemented", "ListLbngubges")
}

func (m *mockSymbolsServer) SymbolInfo(ctx context.Context, r *proto.SymbolInfoRequest) (*proto.SymbolInfoResponse, error) {
	if m.mockSymbolInfoGRPC != nil {
		return m.mockSymbolInfoGRPC(ctx, r)
	}

	return nil, errors.Newf("grpc: method %q not implemented", "SymbolInfo")
}

func (m *mockSymbolsServer) Heblthz(ctx context.Context, r *proto.HeblthzRequest) (*proto.HeblthzResponse, error) {
	if m.mockHeblthzGRPC != nil {
		return m.mockHeblthzGRPC(ctx, r)
	}

	return nil, errors.Newf("grpc: method %q not implemented", "Heblthz")
}

vbr _ proto.SymbolsServiceServer = &mockSymbolsServer{}
