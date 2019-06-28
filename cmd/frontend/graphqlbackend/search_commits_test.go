package graphqlbackend

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/kylelemons/godebug/pretty"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

func TestSearchCommitsInRepo(t *testing.T) {
	ctx := context.Background()

	var calledVCSRawLogDiffSearch bool
	gitSignatureWithDate := git.Signature{Date: time.Now().UTC().AddDate(0, 0, -1)}
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
				Commit: git.Commit{ID: "c1", Author: gitSignatureWithDate},
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
		Repo: &db.MinimalRepo{ID: 1, Name: "repo"},
		Revs: []search.RevisionSpecifier{{RevSpec: "rev"}},
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

	wantCommit := gitCommitResolver{
		repo:   &repositoryResolver{repo: &db.MinimalRepo{ID: 1, Name: "repo"}},
		oid:    "c1",
		author: *toSignatureResolver(&gitSignatureWithDate),
	}

	if want := []*commitSearchResultResolver{
		{
			commit:      &wantCommit,
			diffPreview: &highlightedString{value: "x", highlights: []*highlightedRange{}},
			icon:        "data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0iVVRGLTgiPz48IURPQ1RZUEUgc3ZnIFBVQkxJQyAiLS8vVzNDLy9EVEQgU1ZHIDEuMS8vRU4iICJodHRwOi8vd3d3LnczLm9yZy9HcmFwaGljcy9TVkcvMS4xL0RURC9zdmcxMS5kdGQiPjxzdmcgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIiB4bWxuczp4bGluaz0iaHR0cDovL3d3dy53My5vcmcvMTk5OS94bGluayIgdmVyc2lvbj0iMS4xIiB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHZpZXdCb3g9IjAgMCAyNCAyNCI+PHBhdGggZD0iTTE3LDEyQzE3LDE0LjQyIDE1LjI4LDE2LjQ0IDEzLDE2LjlWMjFIMTFWMTYuOUM4LjcyLDE2LjQ0IDcsMTQuNDIgNywxMkM3LDkuNTggOC43Miw3LjU2IDExLDcuMVYzSDEzVjcuMUMxNS4yOCw3LjU2IDE3LDkuNTggMTcsMTJNMTIsOUEzLDMgMCAwLDAgOSwxMkEzLDMgMCAwLDAgMTIsMTVBMywzIDAgMCwwIDE1LDEyQTMsMyAwIDAsMCAxMiw5WiIgLz48L3N2Zz4=",
			label:       "[repo](/repo) › [](/repo/-/commit/c1): [](/repo/-/commit/c1)",
			url:         "/repo/-/commit/c1",
			detail:      "[`c1` one day ago](/repo/-/commit/c1)",
			matches:     []*searchResultMatchResolver{{url: "/repo/-/commit/c1", body: "```diff\nx```", highlights: []*highlightedRange{}}},
		},
	}; !reflect.DeepEqual(results, want) {
		t.Errorf("results\ngot  %v\nwant %v\ndiff: %v", results, want, pretty.Compare(results, want))
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

func (r *commitSearchResultResolver) String() string {
	return fmt.Sprintf("{commit: %+v diffPreview: %+v messagePreview: %+v}", r.commit, r.diffPreview, r.messagePreview)
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
