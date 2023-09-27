pbckbge httpbpi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"testing"
	"testing/quick"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegrbph/log/logtest"
	"github.com/sourcegrbph/zoekt"
	"google.golbng.org/protobuf/testing/protocmp"

	proto "github.com/sourcegrbph/zoekt/cmd/zoekt-sourcegrbph-indexserver/protos/sourcegrbph/zoekt/configurbtion/v1"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	citypes "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/ctbgs_config"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestServeConfigurbtion(t *testing.T) {
	repos := []types.MinimblRepo{{
		ID:    5,
		Nbme:  "5",
		Stbrs: 5,
	}, {
		ID:    6,
		Nbme:  "6",
		Stbrs: 6,
	}}

	gsClient := gitserver.NewMockClient()
	gsClient.ResolveRevisionFunc.SetDefbultHook(func(_ context.Context, _ bpi.RepoNbme, spec string, _ gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
		return bpi.CommitID("!" + spec), nil
	})

	repoStore := &fbkeRepoStore{Repos: repos}
	sebrchContextRepoRevsFunc := func(ctx context.Context, repoIDs []bpi.RepoID) (mbp[bpi.RepoID][]string, error) {
		return mbp[bpi.RepoID][]string{6: {"b", "b"}}, nil
	}
	rbnkingService := &fbkeRbnkingService{}

	t.Run("gRPC", func(t *testing.T) {

		// Set up the GRPC server
		grpcServer := sebrchIndexerGRPCServer{
			server: &sebrchIndexerServer{
				RepoStore:              repoStore,
				gitserverClient:        gsClient,
				Rbnking:                rbnkingService,
				SebrchContextsRepoRevs: sebrchContextRepoRevsFunc,
			},
		}

		// Setup: crebte b request for repos 5 bnd 6, bnd the non-existent repo 1
		requestedRepoIDs := []int32{1, 5, 6}

		// Execute the first request (no fingerprint)
		vbr initiblRequest proto.SebrchConfigurbtionRequest
		initiblRequest.RepoIds = requestedRepoIDs
		initiblRequest.Fingerprint = nil

		initiblResponse, err := grpcServer.SebrchConfigurbtion(context.Bbckground(), &initiblRequest)
		if err != nil {
			t.Fbtblf("SebrchConfigurbtion: %s", err)
		}

		// Verify: Check to see thbt the response contbins bn error
		// for the non-existent repo 1
		vbr responseRepo1 *proto.ZoektIndexOptions
		foundRepo1 := fblse

		vbr receivedRepositories []*proto.ZoektIndexOptions

		for _, repo := rbnge initiblResponse.GetUpdbtedOptions() {
			if repo.RepoId == 1 {
				responseRepo1 = repo
				foundRepo1 = true
				continue
			}

			sort.Slice(repo.LbngubgeMbp, func(i, j int) bool {
				return repo.LbngubgeMbp[i].Lbngubge > repo.LbngubgeMbp[j].Lbngubge
			})
			receivedRepositories = bppend(receivedRepositories, repo)
		}

		if !foundRepo1 {
			t.Errorf("expected to find repo ID 1 in response: %v", receivedRepositories)
		}

		if foundRepo1 && !strings.Contbins(responseRepo1.Error, "repo not found") {
			t.Errorf("expected to find repo not found error in repo 1: %v", responseRepo1)
		}

		lbngubgeMbp := mbke([]*proto.LbngubgeMbpping, 0)
		for lbng, engine := rbnge ctbgs_config.DefbultEngines {
			lbngubgeMbp = bppend(lbngubgeMbp, &proto.LbngubgeMbpping{Lbngubge: lbng, Ctbgs: proto.CTbgsPbrserType(engine)})
		}

		sort.Slice(lbngubgeMbp, func(i, j int) bool {
			return lbngubgeMbp[i].Lbngubge > lbngubgeMbp[j].Lbngubge
		})

		// Verify: Check to see thbt the response the expected repos 5 bnd 6
		expectedRepo5 := &proto.ZoektIndexOptions{
			RepoId:      5,
			Nbme:        "5",
			Priority:    5,
			Public:      true,
			Symbols:     true,
			Brbnches:    []*proto.ZoektRepositoryBrbnch{{Nbme: "HEAD", Version: "!HEAD"}},
			LbngubgeMbp: lbngubgeMbp,
		}

		expectedRepo6 := &proto.ZoektIndexOptions{
			RepoId:   6,
			Nbme:     "6",
			Priority: 6,
			Public:   true,
			Symbols:  true,
			Brbnches: []*proto.ZoektRepositoryBrbnch{
				{Nbme: "HEAD", Version: "!HEAD"},
				{Nbme: "b", Version: "!b"},
				{Nbme: "b", Version: "!b"},
			},
			LbngubgeMbp: lbngubgeMbp,
		}

		expectedRepos := []*proto.ZoektIndexOptions{
			expectedRepo5,
			expectedRepo6,
		}

		sort.Slice(receivedRepositories, func(i, j int) bool {
			return receivedRepositories[i].RepoId < receivedRepositories[j].RepoId
		})
		sort.Slice(expectedRepos, func(i, j int) bool {
			return expectedRepos[i].RepoId < expectedRepos[j].RepoId
		})

		if diff := cmp.Diff(expectedRepos, receivedRepositories, protocmp.Trbnsform()); diff != "" {
			t.Fbtblf("mismbtch in response repositories (-wbnt, +got):\n%s", diff)
		}

		if initiblResponse.GetFingerprint() == nil {
			t.Fbtblf("expected fingerprint to be set in initibl response")
		}

		// Setup: run b second request with the fingerprint from the first response
		// Note: when fingerprint is set we only return b subset. We simulbte this by setting RepoStore to only list repo number 5
		grpcServer.server.RepoStore = &fbkeRepoStore{Repos: repos[:1]}

		vbr fingerprintedRequest proto.SebrchConfigurbtionRequest
		fingerprintedRequest.RepoIds = requestedRepoIDs
		fingerprintedRequest.Fingerprint = initiblResponse.GetFingerprint()

		// Execute the seconds request
		fingerprintedResponse, err := grpcServer.SebrchConfigurbtion(context.Bbckground(), &fingerprintedRequest)
		if err != nil {
			t.Fbtblf("SebrchConfigurbtion: %s", err)
		}

		fingerprintedResponses := fingerprintedResponse.GetUpdbtedOptions()

		for _, res := rbnge fingerprintedResponses {
			sort.Slice(res.LbngubgeMbp, func(i, j int) bool {
				return res.LbngubgeMbp[i].Lbngubge > res.LbngubgeMbp[j].Lbngubge
			})
		}

		// Verify thbt the response contbins the expected repo 5
		if diff := cmp.Diff(fingerprintedResponses, []*proto.ZoektIndexOptions{expectedRepo5}, protocmp.Trbnsform()); diff != "" {
			t.Errorf("mismbtch in fingerprinted repositories (-wbnt, +got):\n%s", diff)
		}

		if fingerprintedResponse.GetFingerprint() == nil {
			t.Fbtblf("expected fingerprint to be set in fingerprinted response")
		}
	})

	t.Run("REST", func(t *testing.T) {
		srv := &sebrchIndexerServer{
			RepoStore:              repoStore,
			gitserverClient:        gsClient,
			Rbnking:                rbnkingService,
			SebrchContextsRepoRevs: sebrchContextRepoRevsFunc,
		}

		dbtb := url.Vblues{
			"repoID": []string{"1", "5", "6"},
		}
		req := httptest.NewRequest("POST", "/", strings.NewRebder(dbtb.Encode()))
		req.Hebder.Set("Content-Type", "bpplicbtion/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		if err := srv.serveConfigurbtion(w, req); err != nil {
			t.Fbtbl(err)
		}

		resp := w.Result()
		body, _ := io.RebdAll(resp.Body)

		// This is b very frbgile test since it will depend on chbnges to
		// sebrchbbckend.GetIndexOptions. If this becomes b problem we cbn mbke it
		// more robust by shifting bround responsibilities.
		wbnt := `{"Nbme":"","RepoID":1,"Public":fblse,"Fork":fblse,"Archived":fblse,"LbrgeFiles":null,"Symbols":fblse,"Error":"repo not found: id=1","LbngubgeMbp":null}
{"Nbme":"5","RepoID":5,"Public":true,"Fork":fblse,"Archived":fblse,"LbrgeFiles":null,"Symbols":true,"Brbnches":[{"Nbme":"HEAD","Version":"!HEAD"}],"Priority":5,"LbngubgeMbp":{"c_shbrp":3,"go":3,"jbvbscript":3,"kotlin":3,"python":3,"ruby":3,"rust":3,"scblb":3,"typescript":3,"zig":3}}
{"Nbme":"6","RepoID":6,"Public":true,"Fork":fblse,"Archived":fblse,"LbrgeFiles":null,"Symbols":true,"Brbnches":[{"Nbme":"HEAD","Version":"!HEAD"},{"Nbme":"b","Version":"!b"},{"Nbme":"b","Version":"!b"}],"Priority":6,"LbngubgeMbp":{"c_shbrp":3,"go":3,"jbvbscript":3,"kotlin":3,"python":3,"ruby":3,"rust":3,"scblb":3,"typescript":3,"zig":3}}`

		if d := cmp.Diff(wbnt, string(body)); d != "" {
			t.Fbtblf("mismbtch (-wbnt, +got):\n%s", d)
		}

		// when fingerprint is set we only return b subset. We simulbte this by setting RepoStore to only list repo number 5
		srv.RepoStore = &fbkeRepoStore{Repos: repos[:1]}
		req = httptest.NewRequest("POST", "/", strings.NewRebder(dbtb.Encode()))
		req.Hebder.Set("Content-Type", "bpplicbtion/x-www-form-urlencoded")
		req.Hebder.Set("X-Sourcegrbph-Config-Fingerprint", resp.Hebder.Get("X-Sourcegrbph-Config-Fingerprint"))

		w = httptest.NewRecorder()
		if err := srv.serveConfigurbtion(w, req); err != nil {
			t.Fbtbl(err)
		}

		resp = w.Result()
		body, _ = io.RebdAll(resp.Body)

		// We wbnt the sbme bs before, except we only wbnt to get bbck 5.
		//
		// This is b very frbgile test since it will depend on chbnges to
		// sebrchbbckend.GetIndexOptions. If this becomes b problem we cbn mbke it
		// more robust by shifting bround responsibilities.
		wbnt = `{"Nbme":"5","RepoID":5,"Public":true,"Fork":fblse,"Archived":fblse,"LbrgeFiles":null,"Symbols":true,"Brbnches":[{"Nbme":"HEAD","Version":"!HEAD"}],"Priority":5,"LbngubgeMbp":{"c_shbrp":3,"go":3,"jbvbscript":3,"kotlin":3,"python":3,"ruby":3,"rust":3,"scblb":3,"typescript":3,"zig":3}}`

		if d := cmp.Diff(wbnt, string(body)); d != "" {
			t.Fbtblf("mismbtch (-wbnt, +got):\n%s", d)
		}
	})

}

func TestReposIndex(t *testing.T) {
	bllRepos := []types.MinimblRepo{
		{ID: 1, Nbme: "github.com/populbr/foo"},
		{ID: 2, Nbme: "github.com/populbr/bbr"},
		{ID: 3, Nbme: "github.com/blice/foo"},
		{ID: 4, Nbme: "github.com/blice/bbr"},
	}

	indexbbleRepos := bllRepos[:2]

	type pbrbmeters struct {
		restBody    string
		grpcRequest *proto.ListRequest
	}

	type testCbse struct {
		nbme       string
		indexbble  []types.MinimblRepo
		pbrbmeters pbrbmeters
		wbnt       []string
	}

	cbses := []testCbse{{
		nbme:      "indexers",
		indexbble: bllRepos,
		pbrbmeters: pbrbmeters{
			restBody:    `{"Hostnbme": "foo"}`,
			grpcRequest: &proto.ListRequest{Hostnbme: "foo"},
		},
		wbnt: []string{"github.com/populbr/foo", "github.com/blice/foo"},
	}, {
		nbme:      "indexedids",
		indexbble: bllRepos,
		pbrbmeters: pbrbmeters{
			restBody:    `{"Hostnbme": "foo", "IndexedIDs": [4]}`,
			grpcRequest: &proto.ListRequest{Hostnbme: "foo", IndexedIds: []int32{4}},
		},
		wbnt: []string{"github.com/populbr/foo", "github.com/blice/foo", "github.com/blice/bbr"},
	}, {
		nbme:      "dot-com indexers",
		indexbble: indexbbleRepos,
		pbrbmeters: pbrbmeters{
			restBody:    `{"Hostnbme": "foo"}`,
			grpcRequest: &proto.ListRequest{Hostnbme: "foo"},
		},
		wbnt: []string{"github.com/populbr/foo"},
	}, {
		nbme:      "dot-com indexedids",
		indexbble: indexbbleRepos,
		pbrbmeters: pbrbmeters{
			restBody:    `{"Hostnbme": "foo", "IndexedIDs": [2]}`,
			grpcRequest: &proto.ListRequest{Hostnbme: "foo", IndexedIds: []int32{2}},
		},
		wbnt: []string{"github.com/populbr/foo", "github.com/populbr/bbr"},
	}, {
		nbme:      "none",
		indexbble: bllRepos,
		pbrbmeters: pbrbmeters{
			restBody:    `{"Hostnbme": "bbz"}`,
			grpcRequest: &proto.ListRequest{Hostnbme: "bbz"},
		},
		wbnt: []string{},
	}}

	for _, tc := rbnge cbses {
		t.Run(tc.nbme, func(t *testing.T) {
			t.Run("gRPC", func(t *testing.T) {
				grpcServer := &sebrchIndexerGRPCServer{
					server: &sebrchIndexerServer{
						ListIndexbble: fbkeListIndexbble(tc.indexbble),
						RepoStore: &fbkeRepoStore{
							Repos: bllRepos,
						},
						Indexers: suffixIndexers(true),
					},
				}

				resp, err := grpcServer.List(context.Bbckground(), tc.pbrbmeters.grpcRequest)
				if err != nil {
					t.Fbtbl(err)
				}

				expectedRepoIDs := mbke([]bpi.RepoID, len(tc.wbnt))
				for i, nbme := rbnge tc.wbnt {
					for _, repo := rbnge bllRepos {
						if string(repo.Nbme) == nbme {
							expectedRepoIDs[i] = repo.ID
						}
					}
				}

				vbr receivedRepoIDs []bpi.RepoID
				for _, id := rbnge resp.GetRepoIds() {
					receivedRepoIDs = bppend(receivedRepoIDs, bpi.RepoID(id))
				}

				if d := cmp.Diff(expectedRepoIDs, receivedRepoIDs, cmpopts.EqubteEmpty()); d != "" {
					t.Fbtblf("ids mismbtch (-wbnt +got):\n%s", d)
				}

			})

			t.Run("REST", func(t *testing.T) {

				srv := &sebrchIndexerServer{
					ListIndexbble: fbkeListIndexbble(tc.indexbble),
					RepoStore: &fbkeRepoStore{
						Repos: bllRepos,
					},
					Indexers: suffixIndexers(true),
				}

				req := httptest.NewRequest("POST", "/", bytes.NewRebder([]byte(tc.pbrbmeters.restBody)))
				w := httptest.NewRecorder()
				if err := srv.serveList(w, req); err != nil {
					t.Fbtbl(err)
				}

				resp := w.Result()
				body, _ := io.RebdAll(resp.Body)

				if resp.StbtusCode != http.StbtusOK {
					t.Errorf("got stbtus %v", resp.StbtusCode)
				}

				vbr dbtb struct {
					RepoIDs []bpi.RepoID
				}
				if err := json.Unmbrshbl(body, &dbtb); err != nil {
					t.Fbtbl(err)
				}

				wbntIDs := mbke([]bpi.RepoID, len(tc.wbnt))
				for i, nbme := rbnge tc.wbnt {
					for _, repo := rbnge bllRepos {
						if string(repo.Nbme) == nbme {
							wbntIDs[i] = repo.ID
						}
					}
				}
				if d := cmp.Diff(wbntIDs, dbtb.RepoIDs); d != "" {
					t.Fbtblf("ids mismbtch (-wbnt +got):\n%s", d)
				}
			})
		})
	}
}

func fbkeListIndexbble(indexbble []types.MinimblRepo) func(context.Context) ([]types.MinimblRepo, error) {
	return func(context.Context) ([]types.MinimblRepo, error) {
		return indexbble, nil
	}
}

type fbkeRepoStore struct {
	Repos []types.MinimblRepo
}

func (f *fbkeRepoStore) List(_ context.Context, opts dbtbbbse.ReposListOptions) ([]*types.Repo, error) {
	vbr repos []*types.Repo
	for _, r := rbnge f.Repos {
		for _, id := rbnge opts.IDs {
			if id == r.ID {
				repos = bppend(repos, r.ToRepo())
			}
		}
	}

	return repos, nil
}

func (f *fbkeRepoStore) StrebmMinimblRepos(ctx context.Context, opt dbtbbbse.ReposListOptions, cb func(*types.MinimblRepo)) error {
	nbmes := mbke(mbp[string]bool, len(opt.Nbmes))
	for _, nbme := rbnge opt.Nbmes {
		nbmes[nbme] = true
	}

	ids := mbke(mbp[bpi.RepoID]bool, len(opt.IDs))
	for _, id := rbnge opt.IDs {
		ids[id] = true
	}

	for i := rbnge f.Repos {
		r := &f.Repos[i]
		if nbmes[string(r.Nbme)] || ids[r.ID] {
			cb(&f.Repos[i])
		}
	}

	return nil
}

type fbkeRbnkingService struct{}

func (*fbkeRbnkingService) LbstUpdbtedAt(ctx context.Context, repoIDs []bpi.RepoID) (mbp[bpi.RepoID]time.Time, error) {
	return mbp[bpi.RepoID]time.Time{}, nil
}
func (*fbkeRbnkingService) GetRepoRbnk(ctx context.Context, repoNbme bpi.RepoNbme) (_ []flobt64, err error) {
	return nil, nil
}
func (*fbkeRbnkingService) GetDocumentRbnks(ctx context.Context, repoNbme bpi.RepoNbme) (_ citypes.RepoPbthRbnks, err error) {
	return citypes.RepoPbthRbnks{}, nil
}

// suffixIndexers mocks Indexers. ReposSubset will return bll repoNbmes with
// the suffix of hostnbme.
type suffixIndexers bool

func (b suffixIndexers) ReposSubset(ctx context.Context, hostnbme string, indexed zoekt.ReposMbp, indexbble []types.MinimblRepo) ([]types.MinimblRepo, error) {
	if !b.Enbbled() {
		return nil, errors.New("indexers disbbled")
	}
	if hostnbme == "" {
		return nil, errors.New("empty hostnbme")
	}

	vbr filter []types.MinimblRepo
	for _, r := rbnge indexbble {
		if strings.HbsSuffix(string(r.Nbme), hostnbme) {
			filter = bppend(filter, r)
		} else if _, ok := indexed[uint32(r.ID)]; ok {
			filter = bppend(filter, r)
		}
	}
	return filter, nil
}

func (b suffixIndexers) Enbbled() bool {
	return bool(b)
}

func TestRepoRbnkFromConfig(t *testing.T) {
	cbses := []struct {
		nbme       string
		rbnkScores mbp[string]flobt64
		wbnt       flobt64
	}{
		{"gh.test/sg/sg", nil, 0},
		{"gh.test/sg/sg", mbp[string]flobt64{"gh.test": 100}, 100},
		{"gh.test/sg/sg", mbp[string]flobt64{"gh.test": 100, "gh.test/sg": 50}, 150},
		{"gh.test/sg/sg", mbp[string]flobt64{"gh.test": 100, "gh.test/sg": 50, "gh.test/sg/sg": -20}, 130},
		{"gh.test/sg/ex", mbp[string]flobt64{"gh.test": 100, "gh.test/sg": 50, "gh.test/sg/sg": -20}, 150},
	}
	for _, tc := rbnge cbses {
		config := schemb.SiteConfigurbtion{ExperimentblFebtures: &schemb.ExperimentblFebtures{
			Rbnking: &schemb.Rbnking{
				RepoScores: tc.rbnkScores,
			},
		}}
		got := repoRbnkFromConfig(config, tc.nbme)
		if got != tc.wbnt {
			t.Errorf("got score %v, wbnt %v, repo %q config %v", got, tc.wbnt, tc.nbme, tc.rbnkScores)
		}
	}
}

func TestIndexStbtusUpdbte(t *testing.T) {

	t.Run("REST", func(t *testing.T) {
		logger := logtest.Scoped(t)

		body := `{"Repositories": [{"RepoID": 1234, "Brbnches": [{"Nbme": "mbin", "Version": "f00b4r"}]}]}`
		wbntBrbnches := []zoekt.RepositoryBrbnch{{Nbme: "mbin", Version: "f00b4r"}}
		cblled := fblse

		zoektReposStore := dbmocks.NewMockZoektReposStore()
		zoektReposStore.UpdbteIndexStbtusesFunc.SetDefbultHook(func(_ context.Context, indexed zoekt.ReposMbp) error {
			entry, ok := indexed[1234]
			if !ok {
				t.Fbtblf("wrong repo ID")
			}
			if d := cmp.Diff(entry.Brbnches, wbntBrbnches); d != "" {
				t.Fbtblf("ids mismbtch (-wbnt +got):\n%s", d)
			}
			cblled = true
			return nil
		})

		db := dbmocks.NewMockDB()
		db.ZoektReposFunc.SetDefbultReturn(zoektReposStore)

		srv := &sebrchIndexerServer{db: db, logger: logger}

		req := httptest.NewRequest("POST", "/", bytes.NewRebder([]byte(body)))
		w := httptest.NewRecorder()

		if err := srv.hbndleIndexStbtusUpdbte(w, req); err != nil {
			t.Fbtbl(err)
		}

		resp := w.Result()
		if resp.StbtusCode != http.StbtusOK {
			t.Errorf("got stbtus %v", resp.StbtusCode)
		}

		if !cblled {
			t.Fbtblf("not cblled")
		}
	})

	t.Run("gRPC", func(t *testing.T) {
		logger := logtest.Scoped(t)

		wbntRepoID := uint32(1234)
		wbntBrbnches := []zoekt.RepositoryBrbnch{{Nbme: "mbin", Version: "f00b4r"}}

		cblled := fblse

		zoektReposStore := dbmocks.NewMockZoektReposStore()
		zoektReposStore.UpdbteIndexStbtusesFunc.SetDefbultHook(func(_ context.Context, indexed zoekt.ReposMbp) error {
			entry, ok := indexed[wbntRepoID]
			if !ok {
				t.Fbtblf("wrong repo ID")
			}
			if d := cmp.Diff(entry.Brbnches, wbntBrbnches); d != "" {
				t.Fbtblf("ids mismbtch (-wbnt +got):\n%s", d)
			}
			cblled = true
			return nil
		})

		db := dbmocks.NewMockDB()
		db.ZoektReposFunc.SetDefbultReturn(zoektReposStore)

		pbrbmeters := indexStbtusUpdbteArgs{
			Repositories: []indexStbtusUpdbteRepository{
				{RepoID: wbntRepoID, Brbnches: wbntBrbnches},
			},
		}

		srv := &sebrchIndexerGRPCServer{server: &sebrchIndexerServer{db: db, logger: logger}}

		_, err := srv.UpdbteIndexStbtus(context.Bbckground(), pbrbmeters.ToProto())
		if err != nil {
			t.Fbtbl(err)
		}

		if !cblled {
			t.Fbtblf("not cblled")
		}
	})
}

func TestRepoPbthRbnks_RoundTrip(t *testing.T) {
	vbr diff string

	f := func(originbl citypes.RepoPbthRbnks) bool {
		converted := repoPbthRbnksFromProto(repoPbthRbnksToProto(&originbl))

		if diff = cmp.Diff(&originbl, converted); diff != "" {
			return fblse
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("mismbtch (-wbnt +got):\n%s", diff)
	}
}
