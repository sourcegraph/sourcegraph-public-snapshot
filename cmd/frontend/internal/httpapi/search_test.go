package httpapi

import (
	"context"
	"sort"
	"strings"
	"testing"
	"testing/quick"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/sourcegraph/log/logtest"
	"github.com/sourcegraph/zoekt"
	"google.golang.org/protobuf/testing/protocmp"

	proto "github.com/sourcegraph/zoekt/cmd/zoekt-sourcegraph-indexserver/protos/sourcegraph/zoekt/configuration/v1"

	"github.com/sourcegraph/sourcegraph/internal/api"
	citypes "github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/ctags_config"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestServeConfiguration(t *testing.T) {
	repos := []types.MinimalRepo{{
		ID:    5,
		Name:  "5",
		Stars: 5,
	}, {
		ID:    6,
		Name:  "6",
		Stars: 6,
	}}

	gsClient := gitserver.NewMockClient()
	gsClient.ResolveRevisionFunc.SetDefaultHook(func(_ context.Context, _ api.RepoName, spec string, _ gitserver.ResolveRevisionOptions) (api.CommitID, error) {
		return api.CommitID("!" + spec), nil
	})

	repoStore := &fakeRepoStore{Repos: repos}
	searchContextRepoRevsFunc := func(ctx context.Context, repoIDs []api.RepoID) (map[api.RepoID][]string, error) {
		return map[api.RepoID][]string{6: {"a", "b"}}, nil
	}
	rankingService := &fakeRankingService{}

	// Set up the GRPC server
	grpcServer := searchIndexerGRPCServer{
		server: &searchIndexerServer{
			RepoStore:              repoStore,
			gitserverClient:        gsClient,
			Ranking:                rankingService,
			SearchContextsRepoRevs: searchContextRepoRevsFunc,
		},
	}

	// Setup: create a request for repos 5 and 6, and the non-existent repo 1
	requestedRepoIDs := []int32{1, 5, 6}

	// Execute the first request (no fingerprint)
	var initialRequest proto.SearchConfigurationRequest
	initialRequest.RepoIds = requestedRepoIDs
	initialRequest.Fingerprint = nil

	initialResponse, err := grpcServer.SearchConfiguration(context.Background(), &initialRequest)
	if err != nil {
		t.Fatalf("SearchConfiguration: %s", err)
	}

	// Verify: Check to see that the response contains an error
	// for the non-existent repo 1
	var responseRepo1 *proto.ZoektIndexOptions
	foundRepo1 := false

	var receivedRepositories []*proto.ZoektIndexOptions

	for _, repo := range initialResponse.GetUpdatedOptions() {
		if repo.RepoId == 1 {
			responseRepo1 = repo
			foundRepo1 = true
			continue
		}

		sort.Slice(repo.LanguageMap, func(i, j int) bool {
			return repo.LanguageMap[i].Language > repo.LanguageMap[j].Language
		})
		receivedRepositories = append(receivedRepositories, repo)
	}

	if !foundRepo1 {
		t.Errorf("expected to find repo ID 1 in response: %v", receivedRepositories)
	}

	if foundRepo1 && !strings.Contains(responseRepo1.Error, "repo not found") {
		t.Errorf("expected to find repo not found error in repo 1: %v", responseRepo1)
	}

	languageMap := make([]*proto.LanguageMapping, 0)
	for lang, engine := range ctags_config.DefaultEngines {
		languageMap = append(languageMap, &proto.LanguageMapping{Language: lang, Ctags: proto.CTagsParserType(engine)})
	}

	sort.Slice(languageMap, func(i, j int) bool {
		return languageMap[i].Language > languageMap[j].Language
	})

	// Verify: Check to see that the response the expected repos 5 and 6
	expectedRepo5 := &proto.ZoektIndexOptions{
		RepoId:      5,
		Name:        "5",
		Priority:    5,
		Public:      true,
		Symbols:     true,
		Branches:    []*proto.ZoektRepositoryBranch{{Name: "HEAD", Version: "!HEAD"}},
		LanguageMap: languageMap,
	}

	expectedRepo6 := &proto.ZoektIndexOptions{
		RepoId:   6,
		Name:     "6",
		Priority: 6,
		Public:   true,
		Symbols:  true,
		Branches: []*proto.ZoektRepositoryBranch{
			{Name: "HEAD", Version: "!HEAD"},
			{Name: "a", Version: "!a"},
			{Name: "b", Version: "!b"},
		},
		LanguageMap: languageMap,
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

	if diff := cmp.Diff(expectedRepos, receivedRepositories, protocmp.Transform()); diff != "" {
		t.Fatalf("mismatch in response repositories (-want, +got):\n%s", diff)
	}

	if initialResponse.GetFingerprint() == nil {
		t.Fatalf("expected fingerprint to be set in initial response")
	}

	// Setup: run a second request with the fingerprint from the first response
	// Note: when fingerprint is set we only return a subset. We simulate this by setting RepoStore to only list repo number 5
	grpcServer.server.RepoStore = &fakeRepoStore{Repos: repos[:1]}

	var fingerprintedRequest proto.SearchConfigurationRequest
	fingerprintedRequest.RepoIds = requestedRepoIDs
	fingerprintedRequest.Fingerprint = initialResponse.GetFingerprint()

	// Execute the seconds request
	fingerprintedResponse, err := grpcServer.SearchConfiguration(context.Background(), &fingerprintedRequest)
	if err != nil {
		t.Fatalf("SearchConfiguration: %s", err)
	}

	fingerprintedResponses := fingerprintedResponse.GetUpdatedOptions()

	for _, res := range fingerprintedResponses {
		sort.Slice(res.LanguageMap, func(i, j int) bool {
			return res.LanguageMap[i].Language > res.LanguageMap[j].Language
		})
	}

	// Verify that the response contains the expected repo 5
	if diff := cmp.Diff(fingerprintedResponses, []*proto.ZoektIndexOptions{expectedRepo5}, protocmp.Transform()); diff != "" {
		t.Errorf("mismatch in fingerprinted repositories (-want, +got):\n%s", diff)
	}

	if fingerprintedResponse.GetFingerprint() == nil {
		t.Fatalf("expected fingerprint to be set in fingerprinted response")
	}
}

func TestReposIndex(t *testing.T) {
	allRepos := []types.MinimalRepo{
		{ID: 1, Name: "github.com/popular/foo"},
		{ID: 2, Name: "github.com/popular/bar"},
		{ID: 3, Name: "github.com/alice/foo"},
		{ID: 4, Name: "github.com/alice/bar"},
	}

	indexableRepos := allRepos[:2]

	type parameters struct {
		restBody    string
		grpcRequest *proto.ListRequest
	}

	type testCase struct {
		name       string
		indexable  []types.MinimalRepo
		parameters parameters
		want       []string
	}

	cases := []testCase{{
		name:      "indexers",
		indexable: allRepos,
		parameters: parameters{
			restBody:    `{"Hostname": "foo"}`,
			grpcRequest: &proto.ListRequest{Hostname: "foo"},
		},
		want: []string{"github.com/popular/foo", "github.com/alice/foo"},
	}, {
		name:      "indexedids",
		indexable: allRepos,
		parameters: parameters{
			restBody:    `{"Hostname": "foo", "IndexedIDs": [4]}`,
			grpcRequest: &proto.ListRequest{Hostname: "foo", IndexedIds: []int32{4}},
		},
		want: []string{"github.com/popular/foo", "github.com/alice/foo", "github.com/alice/bar"},
	}, {
		name:      "dot-com indexers",
		indexable: indexableRepos,
		parameters: parameters{
			restBody:    `{"Hostname": "foo"}`,
			grpcRequest: &proto.ListRequest{Hostname: "foo"},
		},
		want: []string{"github.com/popular/foo"},
	}, {
		name:      "dot-com indexedids",
		indexable: indexableRepos,
		parameters: parameters{
			restBody:    `{"Hostname": "foo", "IndexedIDs": [2]}`,
			grpcRequest: &proto.ListRequest{Hostname: "foo", IndexedIds: []int32{2}},
		},
		want: []string{"github.com/popular/foo", "github.com/popular/bar"},
	}, {
		name:      "none",
		indexable: allRepos,
		parameters: parameters{
			restBody:    `{"Hostname": "baz"}`,
			grpcRequest: &proto.ListRequest{Hostname: "baz"},
		},
		want: []string{},
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			grpcServer := &searchIndexerGRPCServer{
				server: &searchIndexerServer{
					ListIndexable: fakeListIndexable(tc.indexable),
					RepoStore: &fakeRepoStore{
						Repos: allRepos,
					},
					Indexers: suffixIndexers(true),
				},
			}

			resp, err := grpcServer.List(context.Background(), tc.parameters.grpcRequest)
			if err != nil {
				t.Fatal(err)
			}

			expectedRepoIDs := make([]api.RepoID, len(tc.want))
			for i, name := range tc.want {
				for _, repo := range allRepos {
					if string(repo.Name) == name {
						expectedRepoIDs[i] = repo.ID
					}
				}
			}

			var receivedRepoIDs []api.RepoID
			for _, id := range resp.GetRepoIds() {
				receivedRepoIDs = append(receivedRepoIDs, api.RepoID(id))
			}

			if d := cmp.Diff(expectedRepoIDs, receivedRepoIDs, cmpopts.EquateEmpty()); d != "" {
				t.Fatalf("ids mismatch (-want +got):\n%s", d)
			}
		})
	}
}

func fakeListIndexable(indexable []types.MinimalRepo) func(context.Context) ([]types.MinimalRepo, error) {
	return func(context.Context) ([]types.MinimalRepo, error) {
		return indexable, nil
	}
}

type fakeRepoStore struct {
	Repos []types.MinimalRepo
}

func (f *fakeRepoStore) List(_ context.Context, opts database.ReposListOptions) ([]*types.Repo, error) {
	var repos []*types.Repo
	for _, r := range f.Repos {
		for _, id := range opts.IDs {
			if id == r.ID {
				repos = append(repos, r.ToRepo())
			}
		}
	}

	return repos, nil
}

func (f *fakeRepoStore) StreamMinimalRepos(ctx context.Context, opt database.ReposListOptions, cb func(*types.MinimalRepo)) error {
	names := make(map[string]bool, len(opt.Names))
	for _, name := range opt.Names {
		names[name] = true
	}

	ids := make(map[api.RepoID]bool, len(opt.IDs))
	for _, id := range opt.IDs {
		ids[id] = true
	}

	for i := range f.Repos {
		r := &f.Repos[i]
		if names[string(r.Name)] || ids[r.ID] {
			cb(&f.Repos[i])
		}
	}

	return nil
}

type fakeRankingService struct{}

func (*fakeRankingService) LastUpdatedAt(ctx context.Context, repoIDs []api.RepoID) (map[api.RepoID]time.Time, error) {
	return map[api.RepoID]time.Time{}, nil
}
func (*fakeRankingService) GetRepoRank(ctx context.Context, repoName api.RepoName) (_ []float64, err error) {
	return nil, nil
}
func (*fakeRankingService) GetDocumentRanks(ctx context.Context, repoName api.RepoName) (_ citypes.RepoPathRanks, err error) {
	return citypes.RepoPathRanks{}, nil
}

// suffixIndexers mocks Indexers. ReposSubset will return all repoNames with
// the suffix of hostname.
type suffixIndexers bool

func (b suffixIndexers) ReposSubset(ctx context.Context, hostname string, indexed zoekt.ReposMap, indexable []types.MinimalRepo) ([]types.MinimalRepo, error) {
	if !b.Enabled() {
		return nil, errors.New("indexers disabled")
	}
	if hostname == "" {
		return nil, errors.New("empty hostname")
	}

	var filter []types.MinimalRepo
	for _, r := range indexable {
		if strings.HasSuffix(string(r.Name), hostname) {
			filter = append(filter, r)
		} else if _, ok := indexed[uint32(r.ID)]; ok {
			filter = append(filter, r)
		}
	}
	return filter, nil
}

func (b suffixIndexers) Enabled() bool {
	return bool(b)
}

func TestRepoRankFromConfig(t *testing.T) {
	cases := []struct {
		name       string
		rankScores map[string]float64
		want       float64
	}{
		{"gh.test/sg/sg", nil, 0},
		{"gh.test/sg/sg", map[string]float64{"gh.test": 100}, 100},
		{"gh.test/sg/sg", map[string]float64{"gh.test": 100, "gh.test/sg": 50}, 150},
		{"gh.test/sg/sg", map[string]float64{"gh.test": 100, "gh.test/sg": 50, "gh.test/sg/sg": -20}, 130},
		{"gh.test/sg/ex", map[string]float64{"gh.test": 100, "gh.test/sg": 50, "gh.test/sg/sg": -20}, 150},
	}
	for _, tc := range cases {
		config := schema.SiteConfiguration{ExperimentalFeatures: &schema.ExperimentalFeatures{
			Ranking: &schema.Ranking{
				RepoScores: tc.rankScores,
			},
		}}
		got := repoRankFromConfig(config, tc.name)
		if got != tc.want {
			t.Errorf("got score %v, want %v, repo %q config %v", got, tc.want, tc.name, tc.rankScores)
		}
	}
}

func TestIndexStatusUpdate(t *testing.T) {
	logger := logtest.Scoped(t)

	wantRepoID := uint32(1234)
	wantBranches := []zoekt.RepositoryBranch{{Name: "main", Version: "f00b4r"}}

	called := false

	zoektReposStore := dbmocks.NewMockZoektReposStore()
	zoektReposStore.UpdateIndexStatusesFunc.SetDefaultHook(func(_ context.Context, indexed zoekt.ReposMap) error {
		entry, ok := indexed[wantRepoID]
		if !ok {
			t.Fatalf("wrong repo ID")
		}
		if d := cmp.Diff(entry.Branches, wantBranches); d != "" {
			t.Fatalf("ids mismatch (-want +got):\n%s", d)
		}
		called = true
		return nil
	})

	db := dbmocks.NewMockDB()
	db.ZoektReposFunc.SetDefaultReturn(zoektReposStore)

	parameters := indexStatusUpdateArgs{
		Repositories: []indexStatusUpdateRepository{
			{RepoID: wantRepoID, Branches: wantBranches},
		},
	}

	srv := &searchIndexerGRPCServer{server: &searchIndexerServer{db: db, logger: logger}}

	_, err := srv.UpdateIndexStatus(context.Background(), parameters.ToProto())
	if err != nil {
		t.Fatal(err)
	}

	if !called {
		t.Fatalf("not called")
	}
}

func TestRepoPathRanks_RoundTrip(t *testing.T) {
	var diff string

	f := func(original citypes.RepoPathRanks) bool {
		converted := repoPathRanksFromProto(repoPathRanksToProto(&original))

		if diff = cmp.Diff(&original, converted); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}
