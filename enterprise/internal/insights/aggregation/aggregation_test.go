package aggregation

import (
	"testing"
	"time"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	internaltypes "github.com/sourcegraph/sourcegraph/internal/types"
)

type testAggregator struct {
	results map[string]int
}

func (r *testAggregator) AddResult(result *AggregationMatchResult, err error) {
	if err != nil {
		return
	}
	current, _ := r.results[result.Key.Group]
	r.results[result.Key.Group] = result.Count + current
}

func contentMatch(repo, path string, repoID int32, chunks ...string) result.Match {
	matches := make([]result.ChunkMatch, 0, len(chunks))
	for _, content := range chunks {
		matches = append(matches, result.ChunkMatch{
			Content:      content,
			ContentStart: result.Location{Offset: 0, Line: 1, Column: 0},
			Ranges: result.Ranges{{
				Start: result.Location{Offset: 0, Line: 1, Column: 0},
				End:   result.Location{Offset: len(content), Line: 1, Column: len(content)},
			}},
		})
	}

	return &result.FileMatch{
		File: result.File{
			Repo: internaltypes.MinimalRepo{Name: api.RepoName(repo), ID: api.RepoID(repoID)},
			Path: path,
		},
		ChunkMatches: result.ChunkMatches(matches),
	}
}

func repoMatch(repo string, repoID int32) result.Match {
	return &result.RepoMatch{
		Name: api.RepoName(repo),
		ID:   api.RepoID(repoID),
	}
}

func pathMatch(repo, path string, repoID int32) result.Match {
	return &result.FileMatch{
		File: result.File{
			Repo: internaltypes.MinimalRepo{Name: api.RepoName(repo), ID: api.RepoID(repoID)},
			Path: path,
		},
	}
}

func symbolMatch(repo, path string, repoID int32, symbols ...string) result.Match {
	symbolMatches := make([]*result.SymbolMatch, 0, len(symbols))
	for _, s := range symbols {
		symbolMatches = append(symbolMatches, &result.SymbolMatch{Symbol: result.Symbol{Name: s}})
	}

	return &result.FileMatch{
		File: result.File{
			Repo: internaltypes.MinimalRepo{Name: api.RepoName(repo), ID: api.RepoID(repoID)},
			Path: path,
		},
		Symbols: symbolMatches,
	}
}

func commitMatch(repo, author string, date time.Time, repoID, numRanges int32, content string) result.Match {

	return &result.CommitMatch{
		Commit: gitdomain.Commit{
			Author:    gitdomain.Signature{Name: author},
			Committer: &gitdomain.Signature{},
			Message:   gitdomain.Message(content),
		},
		Repo: internaltypes.MinimalRepo{Name: api.RepoName(repo), ID: api.RepoID(repoID)},
	}
}

var sampleDate time.Time = time.Date(2022, time.April, 1, 0, 0, 0, 0, time.UTC)

func TestRepoAggregation(t *testing.T) {
	testCases := []struct {
		mode        types.SearchAggregationMode
		searchEvent streaming.SearchEvent
		want        autogold.Value
	}{
		{types.REPO_AGGREGATION_MODE,
			streaming.SearchEvent{Results: []result.Match{}},
			autogold.Want("No results", map[string]int{})},
		{
			types.REPO_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{contentMatch("myRepo", "file.go", 1, "a", "b")},
			},
			autogold.Want("Single file match multiple results", map[string]int{"myRepo": 2}),
		},
		{
			types.REPO_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					contentMatch("myRepo", "file.go", 1, "a", "b"),
					contentMatch("myRepo", "file2.go", 1, "d", "e"),
				}},
			autogold.Want("Multiple file match multiple results", map[string]int{"myRepo": 4}),
		},
		{
			types.REPO_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					contentMatch("myRepo", "file.go", 1, "a", "b"),
					contentMatch("myRepo2", "file2.go", 2, "a", "b"),
				}},
			autogold.Want("Multiple repos multiple match", map[string]int{"myRepo": 2, "myRepo2": 2}),
		},
		{
			types.REPO_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					commitMatch("myRepo", "Author A", sampleDate, 1, 2, "a"),
					commitMatch("myRepo", "Author B", sampleDate, 1, 2, "b"),
				}},
			autogold.Want("Count repos on commit matches", map[string]int{"myRepo": 2}),
		},
		{
			types.REPO_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					repoMatch("myRepo", 1),
					repoMatch("myRepo2", 2),
				}},
			autogold.Want("Count repos on repo match", map[string]int{"myRepo": 1, "myRepo2": 1}),
		},
		{
			types.REPO_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					pathMatch("myRepo", "file1.go", 1),
					pathMatch("myRepo", "file2.go", 1),
					pathMatch("myRepoB", "file3.go", 2),
				}},
			autogold.Want("Count repos on path matches", map[string]int{"myRepo": 2, "myRepoB": 1}),
		},
		{
			types.REPO_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					symbolMatch("myRepo", "file1.go", 1, "a", "b"),
					symbolMatch("myRepo", "file2.go", 1, "c", "d"),
				}},
			autogold.Want("Count repos on symbol matches", map[string]int{"myRepo": 4}),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			aggregator := testAggregator{results: make(map[string]int)}
			countFunc, _ := GetCountFuncForMode("", "", tc.mode)
			sra := NewSearchResultsAggregator(aggregator.AddResult, countFunc)
			sra.Send(tc.searchEvent)
			tc.want.Equal(t, aggregator.results)
		})
	}
}

func TestAuthorAggregation(t *testing.T) {
	testCases := []struct {
		mode        types.SearchAggregationMode
		searchEvent streaming.SearchEvent
		want        autogold.Value
	}{
		{types.AUTHOR_AGGREGATION_MODE, streaming.SearchEvent{}, autogold.Want("No results", map[string]int{})},
		{
			types.AUTHOR_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{contentMatch("myRepo", "file.go", 1, "a", "b")},
			},
			autogold.Want("No author for content match", map[string]int{}),
		},
		{
			types.AUTHOR_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{symbolMatch("myRepo", "file.go", 1, "a", "b")},
			},
			autogold.Want("No author for symbol match", map[string]int{}),
		},
		{
			types.AUTHOR_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{pathMatch("myRepo", "file.go", 1)},
			},
			autogold.Want("No author for path match", map[string]int{}),
		},
		{
			types.AUTHOR_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					commitMatch("repoA", "Author A", sampleDate, 1, 2, "a"),
					commitMatch("repoA", "Author B", sampleDate, 1, 2, "a"),
					commitMatch("repoB", "Author B", sampleDate, 2, 2, "a"),
					commitMatch("repoB", "Author C", sampleDate, 2, 2, "a"),
				},
			},
			autogold.Want("counts by author", map[string]int{"Author A": 1, "Author B": 2, "Author C": 1}),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			aggregator := testAggregator{results: make(map[string]int)}
			countFunc, _ := GetCountFuncForMode("", "", tc.mode)
			sra := NewSearchResultsAggregator(aggregator.AddResult, countFunc)
			sra.Send(tc.searchEvent)
			tc.want.Equal(t, aggregator.results)
		})
	}
}

func TestPathAggregation(t *testing.T) {
	testCases := []struct {
		mode        types.SearchAggregationMode
		searchEvent streaming.SearchEvent
		want        autogold.Value
	}{
		{types.PATH_AGGREGATION_MODE, streaming.SearchEvent{}, autogold.Want("No results", map[string]int{})},
		{
			types.PATH_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					commitMatch("repoA", "Author A", sampleDate, 1, 2, "a"),
				},
			},
			autogold.Want("no path for commit", map[string]int{}),
		},
		{
			types.PATH_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					repoMatch("myRepo", 1),
				},
			},
			autogold.Want("no path on repo match", map[string]int{}),
		},
		{
			types.PATH_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{contentMatch("myRepo", "file.go", 1, "a", "b")},
			},
			autogold.Want("Single file match multiple results", map[string]int{"file.go": 2}),
		},
		{
			types.PATH_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					contentMatch("myRepo", "file.go", 1, "a", "b"),
					contentMatch("myRepo", "file2.go", 1, "d", "e"),
				},
			},
			autogold.Want("Multiple file match multiple results", map[string]int{"file.go": 2, "file2.go": 2}),
		},
		{
			types.PATH_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					contentMatch("myRepo", "file.go", 1, "a", "b"),
					contentMatch("myRepo2", "file.go", 2, "a", "b"),
				},
			},
			autogold.Want("Multiple repos same file", map[string]int{"file.go": 4}),
		},
		{
			types.PATH_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					pathMatch("myRepo", "file1.go", 1),
					pathMatch("myRepo", "file2.go", 1),
					pathMatch("myRepoB", "file3.go", 2),
				},
			},
			autogold.Want("Count paths on path matches", map[string]int{"file1.go": 1, "file2.go": 1, "file3.go": 1}),
		},
		{
			types.PATH_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					symbolMatch("myRepo", "file1.go", 1, "a", "b"),
					symbolMatch("myRepo", "file2.go", 1, "c", "d"),
				},
			},
			autogold.Want("Count paths on symbol matches", map[string]int{"file1.go": 2, "file2.go": 2}),
		},
		{
			types.PATH_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					repoMatch("myRepo", 1),
					pathMatch("myRepo", "file1.go", 1),
					symbolMatch("myRepo", "file1.go", 1, "c", "d"),
					contentMatch("myRepo", "file.go", 1, "a", "b"),
				},
			},
			autogold.Want("Count paths on multiple matche types", map[string]int{"file.go": 2, "file1.go": 3}),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			aggregator := testAggregator{results: make(map[string]int)}
			countFunc, _ := GetCountFuncForMode("", "", tc.mode)
			sra := NewSearchResultsAggregator(aggregator.AddResult, countFunc)
			sra.Send(tc.searchEvent)
			tc.want.Equal(t, aggregator.results)
		})
	}
}

func TestCaptureGroupAggregation(t *testing.T) {
	testCases := []struct {
		mode        types.SearchAggregationMode
		searchEvent streaming.SearchEvent
		query       string
		want        autogold.Value
	}{
		{types.CAPTURE_GROUP_AGGREGATION_MODE, streaming.SearchEvent{}, "TEST", autogold.Want("no results", map[string]int{})},
		{
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{contentMatch("myRepo", "file.go", 1, "python2.7 python3.9")},
			},
			`python([0-9]\.[0-9])`,
			autogold.Want("two keys from 1 chunk", map[string]int{"2.7": 1, "3.9": 1}),
		},
		{
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{contentMatch("myRepo", "file.go", 1, "python2.7 python2.7")},
			},
			`python([0-9]\.[0-9])`,
			autogold.Want("count 2 from 1 chunk", map[string]int{"2.7": 2}),
		},
		{
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					contentMatch("myRepo", "file.go", 1, "python2.7 python3.9"),
					contentMatch("myRepo2", "file2.go", 2, "python2.7 python3.9"),
				},
			},
			`python([0-9]\.[0-9])`,
			autogold.Want("count multiple results", map[string]int{"2.7": 2, "3.9": 2}),
		},
		{
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					contentMatch("myRepo", "file.go", 1, "python2.7 python3.9"),
				},
			},
			`python(?:[0-9])\.([0-9])`,
			autogold.Want("skips non capturing group", map[string]int{"7": 1, "9": 1}),
		},
		{
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{contentMatch("myRepo", "file.go", 1, "Python.7 PyThoN2.7")},
			},
			`repo:^github\.com/sourcegraph/sourcegraph python([0-9]\.[0-9]) case:no`,
			autogold.Want("capture match respects case:no", map[string]int{"2.7": 1}),
		},
		{
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{contentMatch("myRepo", "file.go", 1, "Python.7 PyThoN2.7")},
			},
			`repo:^github\.com/sourcegraph/sourcegraph python([0-9]\.[0-9]) case:yes`,
			autogold.Want("capture match respects case:yes", map[string]int{}),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			aggregator := testAggregator{results: make(map[string]int)}
			countFunc, err := GetCountFuncForMode(tc.query, "regexp", tc.mode)
			if err != nil {
				t.Errorf("expected test not to error, got %v", err)
				t.FailNow()
			}
			sra := NewSearchResultsAggregator(aggregator.AddResult, countFunc)
			sra.Send(tc.searchEvent)
			tc.want.Equal(t, aggregator.results)
		})
	}
}
