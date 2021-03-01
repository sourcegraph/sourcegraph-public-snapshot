package graphqlbackend

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"testing"
	"testing/quick"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func TestSearchCommitsInRepo(t *testing.T) {
	ctx := context.Background()
	db := new(dbtesting.MockDB)

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
				Diff:   &git.RawDiff{Raw: "x"},
			},
		}, true, nil
	}
	defer git.ResetMocks()

	q, err := query.ParseLiteral("p")
	if err != nil {
		t.Fatal(err)
	}
	repoRevs := &search.RepositoryRevisions{
		Repo: &types.RepoName{ID: 1, Name: "repo"},
		Revs: []search.RevisionSpecifier{{RevSpec: "rev"}},
	}
	results, limitHit, timedOut, err := searchCommitsInRepo(ctx, db, search.CommitParameters{
		RepoRevs:    repoRevs,
		PatternInfo: &search.CommitPatternInfo{Pattern: "p", FileMatchLimit: int32(defaultMaxSearchResults)},
		Query:       q,
		Diff:        true,
	})
	if err != nil {
		t.Fatal(err)
	}

	want := []*CommitSearchResultResolver{{
		db: db,
		CommitSearchResult: CommitSearchResult{
			commit:      git.Commit{ID: "c1", Author: gitSignatureWithDate},
			repoName:    types.RepoName{ID: 1, Name: "repo"},
			diffPreview: &highlightedString{value: "x", highlights: []*highlightedRange{}},
			body:        "```diff\nx```",
			highlights:  []*highlightedRange{},
		},
	}}

	if !reflect.DeepEqual(results, want) {
		t.Errorf("results\ngot  %v\nwant %v", results, want)
	}

	wantDetail := Markdown("[`c1` one day ago](/repo/-/commit/c1)")
	if gotDetail := want[0].Detail(); gotDetail != wantDetail {
		t.Errorf("detail\ngot  %v\nwant %v", gotDetail, wantDetail)
	}

	wantLabel := Markdown("[repo](/repo) › [](/repo/-/commit/c1): [](/repo/-/commit/c1)")
	if gotLabel := want[0].Label(); gotLabel != wantLabel {
		t.Errorf("label\ngot  %v\nwant %v", gotLabel, wantLabel)
	}

	wantURL := "/repo/-/commit/c1"
	if gotURL := want[0].URL(); gotURL != wantURL {
		t.Errorf("url\ngot  %v\nwant %v", gotURL, wantURL)
	}

	wantMatches := []*searchResultMatchResolver{{url: "/repo/-/commit/c1", body: "```diff\nx```", highlights: []*highlightedRange{}}}
	if gotMatches := want[0].Matches(); !reflect.DeepEqual(gotMatches, wantMatches) {
		t.Errorf("matches\ngot  %v\nwant %v", gotMatches, wantMatches)
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

func (r *CommitSearchResultResolver) String() string {
	return fmt.Sprintf("{commit: %+v diffPreview: %+v messagePreview: %+v}", r.Commit(), r.diffPreview, r.messagePreview)
}

func TestExpandUsernamesToEmails(t *testing.T) {
	resetMocks()
	database.Mocks.Users.GetByUsername = func(ctx context.Context, username string) (*types.User, error) {
		if want := "alice"; username != want {
			t.Errorf("got %q, want %q", username, want)
		}
		return &types.User{ID: 123}, nil
	}
	database.Mocks.UserEmails.ListByUser = func(_ context.Context, opt database.UserEmailsListOptions) ([]*database.UserEmail, error) {
		if want := int32(123); opt.UserID != want {
			t.Errorf("got %v, want %v", opt.UserID, want)
		}
		t := time.Now()
		return []*database.UserEmail{
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

func TestHighlightMatches(t *testing.T) {
	type args struct {
		pattern *regexp.Regexp
		data    []byte
	}
	tests := []struct {
		name string
		args args
		want *highlightedString
	}{
		{
			// https://github.com/sourcegraph/sourcegraph/issues/4512
			name: "match at end",
			args: args{
				pattern: regexp.MustCompile(`白`),
				data:    []byte(`加一行空白`),
			},
			want: &highlightedString{
				value: "加一行空白",
				highlights: []*highlightedRange{
					{
						line:      1,
						character: 4,
						length:    1,
					},
				},
			},
		},
		{
			// https://github.com/sourcegraph/sourcegraph/issues/4512
			name: "two character match in middle",
			args: args{
				pattern: regexp.MustCompile(`行空`),
				data:    []byte(`加一行空白`),
			},
			want: &highlightedString{
				value: "加一行空白",
				highlights: []*highlightedRange{
					{
						line:      1,
						character: 2,
						length:    2,
					},
				},
			},
		},
		{
			// https://github.com/sourcegraph/sourcegraph/issues/4512
			name: "match at beginning",
			args: args{
				pattern: regexp.MustCompile(`加`),
				data:    []byte(`加一行空白`),
			},
			want: &highlightedString{
				value: "加一行空白",
				highlights: []*highlightedRange{
					{
						line:      1,
						character: 0,
						length:    1,
					},
				},
			},
		},

		{
			name: "invalid utf-8 ",
			args: args{
				pattern: regexp.MustCompile(`.`),
				data:    []byte("a\xc5z"),
			},
			want: &highlightedString{
				value: "a\xc5z",
				highlights: []*highlightedRange{
					{
						line:      1,
						character: 0,
						length:    1,
					},
					{
						line:      1,
						character: 1,
						length:    1,
					},
					{
						line:      1,
						character: 2,
						length:    1,
					},
				},
			},
		},

		{
			name: "multiline",
			args: args{
				pattern: regexp.MustCompile(`行`),
				data:    []byte("加一行空白\n加一空行白"),
			},
			want: &highlightedString{
				value: "加一行空白\n加一空行白",
				highlights: []*highlightedRange{
					{
						line:      1,
						character: 2,
						length:    1,
					},
					{
						line:      2,
						character: 3,
						length:    1,
					},
				},
			},
		},

		// https://github.com/sourcegraph/sourcegraph/issues/4791
		{
			name: "unicode search that would be broken by tolower",
			args: args{
				pattern: regexp.MustCompile(`İ`),
				data:    []byte(`İi`),
			},
			want: &highlightedString{
				value: "İi",
				highlights: []*highlightedRange{
					{
						line:      1,
						character: 0,
						length:    1,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := highlightMatches(tt.args.pattern, tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("highlightMatches() = %v, want %v", spew.Sdump(got), spew.Sdump(tt.want))
			}
		})
	}
}

func Benchmark_highlightMatches(b *testing.B) {
	as := bytes.Repeat([]byte{'a'}, 5000)
	lines := append(as, byte('\n'))
	lines = append(lines, as...)
	rx := regexp.MustCompile(`a`)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = highlightMatches(rx, lines)
	}
}

// searchCommitsInRepo is a blocking version of searchCommitsInRepoStream.
func searchCommitsInRepo(ctx context.Context, db dbutil.DB, op search.CommitParameters) (results []*CommitSearchResultResolver, limitHit, timedOut bool, err error) {
	for event := range searchCommitsInRepoStream(ctx, db, op) {
		results = append(results, event.Results...)
		limitHit = event.LimitHit
		timedOut = event.TimedOut
		err = event.Error
	}
	return results, limitHit, timedOut, err
}

func TestCommitSearchResult_Limit(t *testing.T) {
	f := func(nHighlights []int, limitInput uint32) bool {
		cr := &CommitSearchResult{
			highlights: make([]*highlightedRange, len(nHighlights)),
		}

		// It isn't interesting to test limit > ResultCount, so we bound it to
		// [1, ResultCount]
		count := cr.ResultCount()
		limit := (int(limitInput) % count) + 1

		after := cr.Limit(limit)
		newCount := cr.ResultCount()

		if after == 0 && newCount == limit {
			return true
		}

		t.Logf("failed limit=%d count=%d => after=%d newCount=%d", limit, count, after, newCount)
		return false
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error("quick check failed")
	}

	for nSymbols := 0; nSymbols <= 3; nSymbols++ {
		for limit := 0; limit <= nSymbols; limit++ {
			if !f(make([]int, nSymbols), uint32(limit)) {
				t.Error("small exhaustive check failed")
			}
		}
	}
}
