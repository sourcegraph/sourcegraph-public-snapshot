package graphqlbackend

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
	store "sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
	vcstesting "sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs/testing"
)

func TestSearchCommitsInRepo(t *testing.T) {
	ctx := context.Background()

	calledReposGetByURI := store.Mocks.Repos.MockGetByURI(t, "repo", 1)
	calledReposResolveRev := backend.Mocks.Repos.MockResolveRev_NoCheck(t, "c0")
	var calledVCSRawLogDiffSearch bool
	calledRepoVCSOpen := store.Mocks.RepoVCS.MockOpen(t, 1, vcstesting.MockRepository{
		RawLogDiffSearch_: func(ctx context.Context, opt vcs.RawLogDiffSearchOptions) ([]*vcs.LogCommitSearchResult, bool, error) {
			calledVCSRawLogDiffSearch = true
			if want := "p"; opt.Query.Pattern != want {
				t.Errorf("got %q, want %q", opt.Query.Pattern, want)
			}
			if want := []string{
				"-Sp",
				"--max-count=" + strconv.Itoa(maxGitLogSearchResults+1),
				"--regexp-ignore-case",
				"rev",
			}; !reflect.DeepEqual(opt.Args, want) {
				t.Errorf("got %v, want %v", opt.Args, want)
			}
			return []*vcs.LogCommitSearchResult{
				{
					Commit: vcs.Commit{ID: "c1"},
					Diff:   &vcs.Diff{Raw: "x"},
				},
			}, true, nil
		},
	})

	query, err := resolveQuery("p")
	if err != nil {
		t.Fatal(err)
	}
	repoRevs := repositoryRevisions{repo: "repo", revs: []revspecOrRefGlob{{revspec: "rev"}}}
	results, limitHit, err := searchCommitsInRepo(ctx, repoRevs, &patternInfo{Pattern: "p"}, *query, []string{"-Sp"}, vcs.TextSearchOptions{Pattern: "p"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if want := []*commitSearchResult{
		{
			commit: &commitInfoResolver{
				repository: &repositoryResolver{repo: &sourcegraph.Repo{ID: 1, URI: "repo"}},
				oid:        "c1",
				author:     *toSignatureResolver(&vcs.Signature{}),
			},
			diffPreview: &highlightedString{value: "x", highlights: []*highlightedRange{}},
		},
	}; !reflect.DeepEqual(results, want) {
		t.Errorf("results\ngot  %v\nwant %v", results, want)
	}
	if limitHit {
		t.Error("limitHit")
	}
	if !*calledReposGetByURI {
		t.Error("!calledReposGetByURI")
	}
	if !*calledReposResolveRev {
		t.Error("!calledReposResolveRev")
	}
	if !*calledRepoVCSOpen {
		t.Error("!calledRepoVCSOpen")
	}
	if !calledVCSRawLogDiffSearch {
		t.Error("!calledVCSRawLogDiffSearch")
	}
}

func (c *commitSearchResult) String() string {
	return fmt.Sprintf("{commit: %+v diffPreview: %+v messagePreview: %+v}", c.commit, c.diffPreview, c.messagePreview)
}
