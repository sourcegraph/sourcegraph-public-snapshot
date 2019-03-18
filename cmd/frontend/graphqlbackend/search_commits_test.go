package graphqlbackend

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

func TestSearchCommitsInRepo(t *testing.T) {
	ctx := context.Background()

	var calledVCSRawLogDiffSearch bool
	git.Mocks.RawLogDiffSearch = func(opt git.RawLogDiffSearchOptions) ([]*git.LogCommitSearchResult, bool, error) {
		calledVCSRawLogDiffSearch = true
		if want := "p"; opt.Query.Pattern != want {
			t.Errorf("got %q, want %q", opt.Query.Pattern, want)
		}
		if want := []string{
			"--no-prefix",
			"--max-count=" + strconv.Itoa(defaultMaxSearchResults+1),
			"--unified=0",
			"--regexp-ignore-case",
			"rev",
		}; !reflect.DeepEqual(opt.Args, want) {
			t.Errorf("got %v, want %v", opt.Args, want)
		}
		return []*git.LogCommitSearchResult{
			{
				Commit: git.Commit{ID: "c1"},
				Diff:   &git.Diff{Raw: "x"},
			},
		}, true, nil
	}
	defer git.ResetMocks()

	query, err := query.ParseAndCheck("p")
	if err != nil {
		t.Fatal(err)
	}
	repoRevs := search.RepositoryRevisions{
		Repo:          &types.Repo{ID: 1, URI: "repo"},
		GitserverRepo: gitserver.Repo{Name: "repo", URL: "u"},
		Revs:          []search.RevisionSpecifier{{RevSpec: "rev"}},
	}
	results, limitHit, timedOut, err := searchCommitsInRepo(ctx, commitSearchOp{
		repoRevs:          repoRevs,
		info:              &search.PatternInfo{Pattern: "p", FileMatchLimit: int32(defaultMaxSearchResults)},
		query:             query,
		diff:              true,
		textSearchOptions: git.TextSearchOptions{Pattern: "p"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if want := []*commitSearchResultResolver{
		{
			commit: &gitCommitResolver{
				repo:   &repositoryResolver{repo: &types.Repo{ID: 1, URI: "repo"}},
				oid:    "c1",
				author: *toSignatureResolver(&git.Signature{}),
			},
			diffPreview: &highlightedString{value: "x", highlights: []*highlightedRange{}},
		},
	}; !reflect.DeepEqual(results, want) {
		t.Errorf("results\ngot  %v\nwant %v", results, want)
	}
	if limitHit {
		t.Error("limitHit")
	}
	if timedOut {
		t.Error("timedOut")
	}
	if !calledVCSRawLogDiffSearch {
		t.Error("!calledVCSRawLogDiffSearch")
	}
}

func (c *commitSearchResultResolver) String() string {
	return fmt.Sprintf("{commit: %+v diffPreview: %+v messagePreview: %+v}", c.commit, c.diffPreview, c.messagePreview)
}

func TestExpandUsernamesToEmails(t *testing.T) {
	resetMocks()
	db.Mocks.Users.GetByUsername = func(ctx context.Context, username string) (*types.User, error) {
		if want := "alice"; username != want {
			t.Errorf("got %q, want %q", username, want)
		}
		return &types.User{ID: 123}, nil
	}
	db.Mocks.UserEmails.ListByUser = func(id int32) ([]*db.UserEmail, error) {
		if want := int32(123); id != want {
			t.Errorf("got %v, want %v", id, want)
		}
		t := time.Now()
		return []*db.UserEmail{
			{Email: "alice@example.com", VerifiedAt: &t},
			{Email: "alice@example.org", VerifiedAt: &t},
		}, nil
	}

	x, err := expandUsernamesToEmails(context.Background(), []string{"foo", "@alice"})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"foo", `alice@example\.com`, `alice@example\.org`}; !reflect.DeepEqual(x, want) {
		t.Errorf("got %q, want %q", x, want)
	}
}
