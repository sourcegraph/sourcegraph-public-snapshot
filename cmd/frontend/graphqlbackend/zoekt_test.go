package graphqlbackend

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/zoekt"
	zoektrpc "github.com/google/zoekt/rpc"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/search"
	searchbackend "github.com/sourcegraph/sourcegraph/internal/search/backend"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/schema"
)

func BenchmarkSearchResults(b *testing.B) {
	minimalRepos, _, zoektRepos := generateRepos(5000)
	zoektFileMatches := generateZoektMatches(50)

	z := &searchbackend.Zoekt{
		Client: &fakeSearcher{
			repos:  &zoekt.RepoList{Repos: zoektRepos},
			result: &zoekt.SearchResult{Files: zoektFileMatches},
		},
		DisableCache: true,
	}

	ctx := context.Background()

	mockDecodedViewerFinalSettings = &schema.Settings{}
	defer func() { mockDecodedViewerFinalSettings = nil }()

	db.Mocks.Repos.List = func(_ context.Context, op db.ReposListOptions) ([]*types.Repo, error) {
		return minimalRepos, nil
	}
	defer func() { db.Mocks = db.MockStores{} }()

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		q, err := query.ParseAndCheck(`print index:only count:350`)
		if err != nil {
			b.Fatal(err)
		}
		resolver := &searchResolver{query: q, zoekt: z}
		results, err := resolver.Results(ctx)
		if err != nil {
			b.Fatal("Results:", err)
		}
		if int(results.MatchCount()) != len(zoektFileMatches) {
			b.Fatalf("wrong results length. want=%d, have=%d\n", len(zoektFileMatches), results.MatchCount())
		}
	}
}

func BenchmarkIntegrationSearchResults(b *testing.B) {
	dbtesting.SetupGlobalTestDB(b)

	ctx := context.Background()

	_, repos, zoektRepos := generateRepos(5000)
	zoektFileMatches := generateZoektMatches(50)

	zoektClient, cleanup := zoektRPC(&fakeSearcher{
		repos:  &zoekt.RepoList{Repos: zoektRepos},
		result: &zoekt.SearchResult{Files: zoektFileMatches},
	})
	defer cleanup()
	z := &searchbackend.Zoekt{
		Client:       zoektClient,
		DisableCache: true,
	}

	rows := make([]*sqlf.Query, 0, len(repos))
	for _, r := range repos {
		rows = append(rows, sqlf.Sprintf(
			"(%s, %s, %s, %s, %s, %s, %s)",
			r.Name,
			r.Description,
			r.Fork,
			true,
			r.ExternalRepo.ServiceType,
			r.ExternalRepo.ServiceID,
			r.ExternalRepo.ID,
		))
	}

	q := sqlf.Sprintf(`
		INSERT INTO repo (
			name,
			description,
			fork,
			enabled,
			external_service_type,
			external_service_id,
			external_id
		)
		VALUES %s`,
		sqlf.Join(rows, ","),
	)

	_, err := dbconn.Global.ExecContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		q, err := query.ParseAndCheck(`print index:only count:350`)
		if err != nil {
			b.Fatal(err)
		}
		resolver := &searchResolver{query: q, zoekt: z}
		results, err := resolver.Results(ctx)
		if err != nil {
			b.Fatal("Results:", err)
		}
		if int(results.MatchCount()) != len(zoektFileMatches) {
			b.Fatalf("wrong results length. want=%d, have=%d\n", len(zoektFileMatches), results.MatchCount())
		}
	}
}

func generateRepos(count int) ([]*types.Repo, []*types.Repo, []*zoekt.RepoListEntry) {
	var reposWithIDs []*types.Repo
	var repos []*types.Repo
	var zoektRepos []*zoekt.RepoListEntry

	for i := 1; i <= count; i++ {
		name := fmt.Sprintf("repo-%d", i)

		repoWithIDs := &types.Repo{
			ID:   api.RepoID(i),
			Name: api.RepoName(name),
			ExternalRepo: api.ExternalRepoSpec{
				ID:          name,
				ServiceType: "github",
				ServiceID:   "https://github.com",
			}}

		reposWithIDs = append(reposWithIDs, repoWithIDs)

		repos = append(repos, &types.Repo{

			ID:           repoWithIDs.ID,
			Name:         repoWithIDs.Name,
			ExternalRepo: repoWithIDs.ExternalRepo,

			RepoFields: &types.RepoFields{
				URI:         fmt.Sprintf("https://github.com/foobar/%s", repoWithIDs.Name),
				Description: "this repositoriy contains a side project that I haven't maintained in 2 years",
				Language:    "v-language",
			}})

		zoektRepos = append(zoektRepos, &zoekt.RepoListEntry{
			Repository: zoekt.Repository{
				Name:     name,
				Branches: []zoekt.RepositoryBranch{{Name: "HEAD", Version: "deadbeef"}},
			},
		})
	}
	return reposWithIDs, repos, zoektRepos
}

func generateZoektMatches(count int) []zoekt.FileMatch {
	var zoektFileMatches []zoekt.FileMatch
	for i := 1; i <= count; i++ {
		repoName := fmt.Sprintf("repo-%d", i)
		fileName := fmt.Sprintf("foobar-%d.go", i)

		zoektFileMatches = append(zoektFileMatches, zoekt.FileMatch{
			Score:      5.0,
			FileName:   fileName,
			Repository: repoName, // Important: this needs to match a name in `repos`
			Branches:   []string{"master"},
			LineMatches: []zoekt.LineMatch{
				{
					Line: nil,
				},
			},
			Checksum: []byte{0, 1, 2},
		})
	}
	return zoektFileMatches
}

// zoektRPC starts zoekts rpc interface and returns a client to
// searcher. Useful for capturing CPU/memory usage when benchmarking the zoekt
// client.
func zoektRPC(s zoekt.Searcher) (zoekt.Searcher, func()) {
	mux := http.NewServeMux()
	mux.Handle(zoektrpc.DefaultRPCPath, zoektrpc.Server(s))
	ts := httptest.NewServer(mux)
	cl := zoektrpc.Client(strings.TrimPrefix(ts.URL, "http://"))
	return cl, func() {
		cl.Close()
		ts.Close()
	}
}

func TestZoektSingleIndexedRepo(t *testing.T) {
	repoRev := func(revSpec string) *search.RepositoryRevisions {
		return &search.RepositoryRevisions{
			Repo: &types.Repo{ID: api.RepoID(0), Name: "test/repo"},
			Revs: []search.RevisionSpecifier{
				{RevSpec: revSpec},
			},
		}
	}
	zoektRepos := []*zoekt.RepoListEntry{{
		Repository: zoekt.Repository{
			Name: "test/repo",
			Branches: []zoekt.RepositoryBranch{
				{
					Name:    "HEAD",
					Version: "df3f4e499698e48152b39cd655d8901eaf583fa5",
				},
				{
					Name:    "NOT-HEAD",
					Version: "8ec975423738fe7851676083ebf660a062ed1578",
				},
			},
		},
	}}
	z := &searchbackend.Zoekt{
		Client: &fakeSearcher{
			repos: &zoekt.RepoList{Repos: zoektRepos},
		},
		DisableCache: true,
	}
	cases := []struct {
		rev           string
		wantIndexed   []*search.RepositoryRevisions
		wantUnindexed []*search.RepositoryRevisions
	}{
		{
			rev:           "",
			wantIndexed:   []*search.RepositoryRevisions{repoRev("")},
			wantUnindexed: []*search.RepositoryRevisions{},
		},
		{
			rev:           "HEAD",
			wantIndexed:   []*search.RepositoryRevisions{repoRev("HEAD")},
			wantUnindexed: []*search.RepositoryRevisions{},
		},
		{
			rev:           "df3f4e499698e48152b39cd655d8901eaf583fa5",
			wantIndexed:   []*search.RepositoryRevisions{repoRev("df3f4e499698e48152b39cd655d8901eaf583fa5")},
			wantUnindexed: []*search.RepositoryRevisions{},
		},
		{
			rev:           "df3f4e",
			wantIndexed:   []*search.RepositoryRevisions{repoRev("df3f4e")},
			wantUnindexed: []*search.RepositoryRevisions{},
		},
		{
			rev:           "d",
			wantIndexed:   []*search.RepositoryRevisions{},
			wantUnindexed: []*search.RepositoryRevisions{repoRev("d")},
		},
		{
			rev:           "HEAD^1",
			wantIndexed:   []*search.RepositoryRevisions{},
			wantUnindexed: []*search.RepositoryRevisions{repoRev("HEAD^1")},
		},
		{
			rev:           "8ec975423738fe7851676083ebf660a062ed1578",
			wantUnindexed: []*search.RepositoryRevisions{},
			wantIndexed:   []*search.RepositoryRevisions{repoRev("8ec975423738fe7851676083ebf660a062ed1578")},
		},
	}

	type ret struct {
		Indexed, Unindexed []*search.RepositoryRevisions
	}

	for _, tt := range cases {
		filter := func(*zoekt.Repository) bool { return true }
		indexed, unindexed, err := zoektSingleIndexedRepo(context.Background(), z, repoRev(tt.rev), filter)
		if err != nil {
			t.Fatal(err)
		}
		got := ret{
			Indexed:   indexed,
			Unindexed: unindexed,
		}
		want := ret{
			Indexed:   tt.wantIndexed,
			Unindexed: tt.wantUnindexed,
		}
		if !cmp.Equal(want, got) {
			t.Errorf("%s mismatch (-want +got):\n%s", tt.rev, cmp.Diff(want, got))
		}
	}
}
