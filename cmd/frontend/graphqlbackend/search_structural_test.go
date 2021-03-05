package graphqlbackend

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/zoekt"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/search"
	searchbackend "github.com/sourcegraph/sourcegraph/internal/search/backend"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

// Tests that indexed repos are filtered in structural search
func TestStructuralSearchRepoFilter(t *testing.T) {
	db := new(dbtesting.MockDB)

	repoName := "indexed/one"
	indexedFileName := "indexed.go"

	indexedRepo := &types.RepoName{Name: api.RepoName(repoName)}

	unindexedRepo := &types.RepoName{Name: api.RepoName("unindexed/one")}

	database.Mocks.Repos.ListRepoNames = func(_ context.Context, op database.ReposListOptions) ([]*types.RepoName, error) {
		return []*types.RepoName{indexedRepo, unindexedRepo}, nil
	}
	defer func() { database.Mocks = database.MockStores{} }()

	mockSearchFilesInRepo = func(
		ctx context.Context,
		repo *types.RepoName,
		gitserverRepo api.RepoName,
		rev string,
		info *search.TextPatternInfo,
		fetchTimeout time.Duration,
	) (
		matches []*FileMatchResolver,
		limitHit bool,
		err error,
	) {
		repoName := repo.Name
		switch repoName {
		case "indexed/one":
			return []*FileMatchResolver{mkFileMatch(db, nil, indexedFileName)}, false, nil
		case "unindexed/one":
			return []*FileMatchResolver{mkFileMatch(db, nil, "unindexed.go")}, false, nil
		default:
			return nil, false, errors.New("Unexpected repo")
		}
	}
	database.Mocks.Repos.Count = mockCount
	defer func() { mockSearchFilesInRepo = nil }()

	zoektRepo := &zoekt.RepoListEntry{
		Repository: zoekt.Repository{
			Name:     string(indexedRepo.Name),
			Branches: []zoekt.RepositoryBranch{{Name: "HEAD", Version: "deadbeef"}},
		},
	}

	zoektFileMatches := []zoekt.FileMatch{{
		FileName:   indexedFileName,
		Repository: string(indexedRepo.Name),
		LineMatches: []zoekt.LineMatch{
			{
				Line: nil,
			},
		},
	}}

	z := &searchbackend.Zoekt{
		Client: &searchbackend.FakeSearcher{
			Repos:  []*zoekt.RepoListEntry{zoektRepo},
			Result: &zoekt.SearchResult{Files: zoektFileMatches},
		},
		DisableCache: true,
	}

	ctx := context.Background()

	q, err := query.ParseLiteral(`patterntype:structural index:only foo`)
	if err != nil {
		t.Fatal(err)
	}
	resolver := &searchResolver{
		db: db,
		SearchInputs: &SearchInputs{
			Query:        q,
			PatternType:  query.SearchTypeStructural,
			UserSettings: &schema.Settings{},
		},
		zoekt:        z,
		searcherURLs: endpoint.Static("test"),
		reposMu:      &sync.Mutex{},
		resolved:     &repos.Resolved{},
	}
	results, err := resolver.Results(ctx)
	if err != nil {
		t.Fatal("Results:", err)
	}

	fm, _ := results.Results()[0].ToFileMatch()
	if got, want := fm.Path, indexedFileName; got != want {
		t.Fatalf("wrong indexed filename. want=%s, have=%s\n", want, got)
	}
}

func TestBuildQuery(t *testing.T) {
	pattern := ":[x~*]"
	want := "error parsing regexp: missing argument to repetition operator: `*`"
	t.Run("build query", func(t *testing.T) {
		_, err := buildQuery(
			&search.TextParameters{
				PatternInfo: &search.TextPatternInfo{Pattern: pattern},
			},
			nil,
			nil,
			false,
		)
		if diff := cmp.Diff(err.Error(), want); diff != "" {
			t.Error(diff)
		}
	})
}
