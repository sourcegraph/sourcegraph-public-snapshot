package graphqlbackend

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

func TestSearchPagination_unmarshalSearchCursor(t *testing.T) {
	got, err := unmarshalSearchCursor(nil)
	if got != nil || err != nil {
		t.Fatal("expected got == nil && err == nil for nil input")
	}

	want := &searchCursor{
		RepositoryOffset: 1,
		ResultOffset:     2,
	}
	enc := marshalSearchCursor(want)
	if enc == "" {
		t.Fatal("expected encoded string")
	}
	got, err = unmarshalSearchCursor(&enc)
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatal("expected got == want")
	}
}

func TestSearchPagination_sliceSearchResults(t *testing.T) {
	repo := func(name string) *types.Repo {
		return &types.Repo{Name: api.RepoName(name)}
	}
	result := func(repo *types.Repo, path string) *fileMatchResolver {
		return &fileMatchResolver{JPath: path, repo: repo}
	}
	format := func(r slicedSearchResults) string {
		var b bytes.Buffer
		fmt.Fprintf(&b, "results:\n")
		for i, result := range r.results {
			fm, _ := result.ToFileMatch()
			fmt.Fprintf(&b, "	[%d] %s %s\n", i, fm.repo.Name, fm.JPath)
		}
		fmt.Fprintf(&b, "common.repos:\n")
		for i, r := range r.common.repos {
			fmt.Fprintf(&b, "	[%d] %s\n", i, r.Name)
		}
		fmt.Fprintf(&b, "common.resultCount: %v\n", r.common.resultCount)
		fmt.Fprintf(&b, "resultOffset: %d\n", r.resultOffset)
		if r.lastRepoConsumed == nil {
			fmt.Fprintf(&b, "lastRepoConsumed: nil\n")
		} else {
			fmt.Fprintf(&b, "lastRepoConsumed: %s\n", r.lastRepoConsumed.Name)
		}
		fmt.Fprintf(&b, "lastRepoConsumedPartially: %v\n", r.lastRepoConsumedPartially)
		fmt.Fprintf(&b, "limitHit: %v\n", r.limitHit)
		return b.String()
	}
	sharedResult := []searchResultResolver{
		result(repo("org/repo1"), "a.go"),
		result(repo("org/repo1"), "b.go"),
		result(repo("org/repo1"), "c.go"),
		result(repo("org/repo2"), "a.go"),
		result(repo("org/repo2"), "b.go"),
		result(repo("org/repo3"), "a.go"),
		result(repo("org/repo4"), "a.go"),
		result(repo("org/repo4"), "b.go"),
		result(repo("org/repo4"), "c.go"),
		result(repo("org/repo5"), "a.go"),
		result(repo("org/repo5"), "b.go"),
		result(repo("org/repo5"), "c.go"),
		result(repo("org/repo5"), "d.go"),
		result(repo("org/repo5"), "e.go"),
	}
	sharedCommon := &searchResultsCommon{
		repos: []*types.Repo{repo("org/repo1"), repo("org/repo2"), repo("org/repo3")},
	}
	tests := []struct {
		name          string
		results       []searchResultResolver
		common        *searchResultsCommon
		offset, limit int
		want          slicedSearchResults
	}{
		{
			name:    "empty result set",
			results: []searchResultResolver{},
			common:  &searchResultsCommon{},
			offset:  0,
			limit:   3,
			want: slicedSearchResults{
				results: []searchResultResolver{},
				common: &searchResultsCommon{
					resultCount: 0,
					repos:       nil,
					partial:     nil,
				},
				resultOffset:              0,
				lastRepoConsumed:          nil,
				lastRepoConsumedPartially: false,
				limitHit:                  false,
			},
		},
		{
			name:    "limit repo boundary",
			results: sharedResult,
			common:  sharedCommon,
			offset:  0,
			limit:   3,
			want: slicedSearchResults{
				results: []searchResultResolver{
					result(repo("org/repo1"), "a.go"),
					result(repo("org/repo1"), "b.go"),
					result(repo("org/repo1"), "c.go"),
				},
				common: &searchResultsCommon{
					resultCount: 3,
					repos:       []*types.Repo{repo("org/repo1")},
					partial:     make(map[api.RepoName]struct{}),
				},
				resultOffset:              0,
				lastRepoConsumed:          repo("org/repo1"),
				lastRepoConsumedPartially: false,
				limitHit:                  true,
			},
		},
		{
			name:    "limit non repo boundary",
			results: sharedResult,
			common:  sharedCommon,
			offset:  0,
			limit:   2,
			want: slicedSearchResults{
				results: []searchResultResolver{
					result(repo("org/repo1"), "a.go"),
					result(repo("org/repo1"), "b.go"),
				},
				common: &searchResultsCommon{
					resultCount: 2,
					repos:       []*types.Repo{repo("org/repo1")},
					partial:     make(map[api.RepoName]struct{}),
				},
				resultOffset:              2,
				lastRepoConsumed:          repo("org/repo1"),
				lastRepoConsumedPartially: true,
				limitHit:                  true,
			},
		},
		{
			name:    "offset repo boundary",
			results: sharedResult,
			common:  sharedCommon,
			offset:  3,
			limit:   3,
			want: slicedSearchResults{
				results: []searchResultResolver{
					result(repo("org/repo2"), "a.go"),
					result(repo("org/repo2"), "b.go"),
					result(repo("org/repo3"), "a.go"),
				},
				common: &searchResultsCommon{
					resultCount: 3,
					repos:       []*types.Repo{repo("org/repo2"), repo("org/repo3")},
					partial:     make(map[api.RepoName]struct{}),
				},
				resultOffset:              0,
				lastRepoConsumed:          repo("org/repo3"),
				lastRepoConsumedPartially: false,
				limitHit:                  true,
			},
		},
		{
			name:    "offset non-repo boundary",
			results: sharedResult,
			common:  sharedCommon,
			offset:  2,
			limit:   3,
			want: slicedSearchResults{
				results: []searchResultResolver{
					result(repo("org/repo1"), "c.go"),
					result(repo("org/repo2"), "a.go"),
					result(repo("org/repo2"), "b.go"),
				},
				common: &searchResultsCommon{
					resultCount: 3,
					repos:       []*types.Repo{repo("org/repo1"), repo("org/repo2")},
					partial:     make(map[api.RepoName]struct{}),
				},
				resultOffset:              0,
				lastRepoConsumed:          repo("org/repo2"),
				lastRepoConsumedPartially: false,
				limitHit:                  true,
			},
		},
		{
			name: "offset repo boundary fully consumed",
			results: []searchResultResolver{
				result(repo("org/repo1"), "a.go"),
				result(repo("org/repo1"), "b.go"),
				result(repo("org/repo1"), "c.go"),
				result(repo("org/repo2"), "a.go"),
				result(repo("org/repo2"), "b.go"),
				result(repo("org/repo2"), "c.go"),
			},
			common: &searchResultsCommon{
				repos:       []*types.Repo{repo("org/repo1"), repo("org/repo2")},
				resultCount: 3,
			},
			offset: 3,
			limit:  3,
			want: slicedSearchResults{
				results: []searchResultResolver{
					result(repo("org/repo2"), "a.go"),
					result(repo("org/repo2"), "b.go"),
					result(repo("org/repo2"), "c.go"),
				},
				common: &searchResultsCommon{
					resultCount: 3,
					repos:       []*types.Repo{repo("org/repo1"), repo("org/repo2")},
					partial:     nil,
				},
				resultOffset:              0,
				lastRepoConsumed:          repo("org/repo2"),
				lastRepoConsumedPartially: false,
				limitHit:                  false,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := sliceSearchResults(test.results, test.common, test.offset, test.limit)
			if !reflect.DeepEqual(got, test.want) {
				t.Logf("got != want")
				gotFormatted := format(got)
				wantFormatted := format(test.want)
				t.Logf("got:\n%s\n", gotFormatted)
				t.Logf("want:\n%s\n", wantFormatted)
				dmp := diffmatchpatch.New()
				t.Error("diff(got, want):\n", dmp.DiffPrettyText(dmp.DiffMain(wantFormatted, gotFormatted, true)))

				if wantFormatted == gotFormatted {
					dmp = diffmatchpatch.New()
					t.Error("diff(got, want):\n", dmp.DiffPrettyText(dmp.DiffMain(spew.Sdump(test.want), spew.Sdump(got), true)))
				}
			}
		})
	}
}
