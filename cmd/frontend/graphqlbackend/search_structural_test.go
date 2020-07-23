package graphqlbackend

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/zoekt"
	zoektquery "github.com/google/zoekt/query"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search"
	searchbackend "github.com/sourcegraph/sourcegraph/internal/search/backend"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/schema"
)

// Tests that indexed repos are filtered in structural search
func TestStructuralSearchRepoFilter(t *testing.T) {
	repoName := "indexed/one"
	indexedFileName := "indexed.go"

	indexedRepo := &types.Repo{Name: api.RepoName(repoName)}

	unindexedRepo := &types.Repo{Name: api.RepoName("unindexed/one")}

	mockDecodedViewerFinalSettings = &schema.Settings{}
	defer func() { mockDecodedViewerFinalSettings = nil }()

	db.Mocks.Repos.List = func(_ context.Context, op db.ReposListOptions) ([]*types.Repo, error) {
		return []*types.Repo{indexedRepo, unindexedRepo}, nil
	}
	defer func() { db.Mocks = db.MockStores{} }()

	mockSearchFilesInRepo = func(
		ctx context.Context,
		repo *types.Repo,
		gitserverRepo gitserver.Repo,
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
			return []*FileMatchResolver{{JPath: indexedFileName}}, false, nil
		case "unindexed/one":
			return []*FileMatchResolver{{JPath: "unindexed.go"}}, false, nil
		default:
			return nil, false, errors.New("Unexpected repo")
		}
	}
	db.Mocks.Repos.Count = mockCount
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
		Client: &fakeSearcher{
			repos:  []*zoekt.RepoListEntry{zoektRepo},
			result: &zoekt.SearchResult{Files: zoektFileMatches},
		},
		DisableCache: true,
	}

	ctx := context.Background()

	q, err := query.ParseAndCheck(`patterntype:structural index:only foo`)
	if err != nil {
		t.Fatal(err)
	}
	resolver := &searchResolver{
		query:        q,
		patternType:  query.SearchTypeStructural,
		zoekt:        z,
		searcherURLs: endpoint.Static("test"),
	}
	results, err := resolver.Results(ctx)
	if err != nil {
		t.Fatal("Results:", err)
	}

	fm, _ := results.Results()[0].ToFileMatch()
	if fm.JPath != indexedFileName {
		t.Fatalf("wrong indexed filename. want=%s, have=%s\n", indexedFileName, fm.JPath)
	}
}

func TestStructuralPatToRegexpQuery(t *testing.T) {
	cases := []struct {
		Name     string
		Pattern  string
		Function func(string, bool) (zoektquery.Q, error)
		Want     string
	}{
		{
			Name:     "Just a hole",
			Pattern:  ":[1]",
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"()((?s:.))*?()")`,
		},
		{
			Name:     "Adjacent holes",
			Pattern:  ":[1]:[2]:[3]",
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"()((?s:.))*?()((?s:.))*?()((?s:.))*?()")`,
		},
		{
			Name:     "Substring between holes",
			Pattern:  ":[1] substring :[2]",
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"()((?s:.))*?([\\t-\\n\\f-\\r ]+substring[\\t-\\n\\f-\\r ]+)((?s:.))*?()")`,
		},
		{
			Name:     "Substring before and after different hole kinds",
			Pattern:  "prefix :[[1]] :[2.] suffix",
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"(prefix[\\t-\\n\\f-\\r ]+)((?s:.))*?([\\t-\\n\\f-\\r ]+)((?s:.))*?([\\t-\\n\\f-\\r ]+suffix)")`,
		},
		{
			Name:     "Substrings covering all hole kinds.",
			Pattern:  `1. :[1] 2. :[[2]] 3. :[3.] 4. :[4\n] 5. :[ ] 6. :[ 6] done.`,
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"(1\\.[\\t-\\n\\f-\\r ]+)((?s:.))*?([\\t-\\n\\f-\\r ]+2\\.[\\t-\\n\\f-\\r ]+)((?s:.))*?([\\t-\\n\\f-\\r ]+3\\.[\\t-\\n\\f-\\r ]+)((?s:.))*?([\\t-\\n\\f-\\r ]+4\\.[\\t-\\n\\f-\\r ]+)((?s:.))*?([\\t-\\n\\f-\\r ]+5\\.[\\t-\\n\\f-\\r ]+)((?s:.))*?([\\t-\\n\\f-\\r ]+6\\.[\\t-\\n\\f-\\r ]+)((?s:.))*?([\\t-\\n\\f-\\r ]+done\\.)")`,
		},
		{
			Name:     "Substrings across multiple lines.",
			Pattern:  ``,
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"()")`,
		},
		{
			Name:     "Allow alphanumeric identifiers in holes",
			Pattern:  "sub :[alphanum_ident_123] string",
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"(sub[\\t-\\n\\f-\\r ]+)((?s:.))*?([\\t-\\n\\f-\\r ]+string)")`,
		},

		{
			Name:     "Whitespace separated holes",
			Pattern:  ":[1] :[2]",
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"()((?s:.))*?([\\t-\\n\\f-\\r ]+)((?s:.))*?()")`,
		},
		{
			Name:     "Expect newline separated pattern",
			Pattern:  "ParseInt(:[stuff], :[x]) if err ",
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"(ParseInt\\()((?s:.))*?(,[\\t-\\n\\f-\\r ]+)((?s:.))*?(\\)[\\t-\\n\\f-\\r ]+if[\\t-\\n\\f-\\r ]+err[\\t-\\n\\f-\\r ]+)")`,
		},
		{
			Name: "Contiguous whitespace is replaced by regex",
			Pattern: `ParseInt(:[stuff],    :[x])
             if err `,
			Function: StructuralPatToRegexpQuery,
			Want:     `(and case_regex:"(ParseInt\\()((?s:.))*?(,[\\t-\\n\\f-\\r ]+)((?s:.))*?(\\)[\\t-\\n\\f-\\r ]+if[\\t-\\n\\f-\\r ]+err[\\t-\\n\\f-\\r ]+)")`,
		},
	}
	for _, tt := range cases {
		t.Run(tt.Name, func(t *testing.T) {
			got, _ := tt.Function(tt.Pattern, false)
			if got.String() != tt.Want {
				t.Fatalf("mismatched queries\ngot  %s\nwant %s", got.String(), tt.Want)
			}
		})
	}
}
