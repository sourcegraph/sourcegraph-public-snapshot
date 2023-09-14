package aggregation

import (
	"context"
	"testing"
	"time"

	"github.com/hexops/autogold/v2"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	dTypes "github.com/sourcegraph/sourcegraph/internal/types"
	internaltypes "github.com/sourcegraph/sourcegraph/internal/types"
)

func newTestSearchResultsAggregator(ctx context.Context, tabulator AggregationTabulator, countFunc AggregationCountFunc, mode types.SearchAggregationMode, db database.DB) SearchResultsAggregator {
	if db == nil {
		db = dbmocks.NewMockDB()
	}
	return &searchAggregationResults{
		db:        db,
		mode:      mode,
		ctx:       ctx,
		tabulator: tabulator,
		countFunc: countFunc,
	}
}

type testAggregator struct {
	results map[string]int
	errors  []error
}

func (r *testAggregator) AddResult(result *AggregationMatchResult, err error) {
	if err != nil {
		r.errors = append(r.errors, err)
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
		ChunkMatches: matches,
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
		MessagePreview: &result.MatchedString{
			Content: content,
			MatchedRanges: result.Ranges{
				{
					Start: result.Location{Line: 1, Offset: 0, Column: 1},
					End:   result.Location{Line: 1, Offset: 1, Column: 1},
				},
				{
					Start: result.Location{Line: 2, Offset: 0, Column: 1},
					End:   result.Location{Line: 2, Offset: 1, Column: 1},
				}},
		},
	}
}

func diffMatch(repo, author string, repoID int) result.Match {
	return &result.CommitMatch{
		Repo: internaltypes.MinimalRepo{Name: api.RepoName(repo), ID: api.RepoID(repoID)},
		Commit: gitdomain.Commit{
			Author: gitdomain.Signature{Name: author},
		},
		DiffPreview: &result.MatchedString{
			Content: "file3 file4\n@@ -3,4 +1,6 @@\n+needle\n-needle\n",
			MatchedRanges: result.Ranges{{
				Start: result.Location{Offset: 29, Line: 2, Column: 1},
				End:   result.Location{Offset: 35, Line: 2, Column: 7},
			}, {
				Start: result.Location{Offset: 37, Line: 3, Column: 1},
				End:   result.Location{Offset: 43, Line: 3, Column: 7},
			}},
		},
		Diff: []result.DiffFile{{
			OrigName: "file3",
			NewName:  "file4",
			Hunks: []result.Hunk{{
				OldStart: 3,
				NewStart: 1,
				OldCount: 4,
				NewCount: 6,
				Header:   "",
				Lines:    []string{"+needle", "-needle"},
			}},
		}},
	}
}

var sampleDate = time.Date(2022, time.April, 1, 0, 0, 0, 0, time.UTC)

func TestRepoAggregation(t *testing.T) {
	testCases := []struct {
		name        string
		mode        types.SearchAggregationMode
		searchEvent streaming.SearchEvent
		want        autogold.Value
	}{
		{
			"No results",
			types.REPO_AGGREGATION_MODE,
			streaming.SearchEvent{Results: []result.Match{}},
			autogold.Expect(map[string]int{})},
		{
			"Single file match multiple results",
			types.REPO_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{contentMatch("myRepo", "file.go", 1, "a", "b")},
			},
			autogold.Expect(map[string]int{"myRepo": 2}),
		},
		{
			"Multiple file match multiple results",
			types.REPO_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					contentMatch("myRepo", "file.go", 1, "a", "b"),
					contentMatch("myRepo", "file2.go", 1, "d", "e"),
				}},
			autogold.Expect(map[string]int{"myRepo": 4}),
		},
		{
			"Multiple repo multiple match",
			types.REPO_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					contentMatch("myRepo", "file.go", 1, "a", "b"),
					contentMatch("myRepo2", "file2.go", 2, "a", "b"),
				}},
			autogold.Expect(map[string]int{"myRepo": 2, "myRepo2": 2}),
		},
		{
			"Count repos on commit matches",
			types.REPO_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					commitMatch("myRepo", "Author A", sampleDate, 1, 2, "a"),
					commitMatch("myRepo", "Author B", sampleDate, 1, 2, "b"),
				}},
			autogold.Expect(map[string]int{"myRepo": 4}),
		},
		{
			"Count repos on repo match",
			types.REPO_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					repoMatch("myRepo", 1),
					repoMatch("myRepo2", 2),
				}},
			autogold.Expect(map[string]int{"myRepo": 1, "myRepo2": 1}),
		},
		{
			"Count repos on path matches",
			types.REPO_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					pathMatch("myRepo", "file1.go", 1),
					pathMatch("myRepo", "file2.go", 1),
					pathMatch("myRepoB", "file3.go", 2),
				}},
			autogold.Expect(map[string]int{"myRepo": 2, "myRepoB": 1}),
		},
		{
			"Count repos on symbol matches",
			types.REPO_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					symbolMatch("myRepo", "file1.go", 1, "a", "b"),
					symbolMatch("myRepo", "file2.go", 1, "c", "d"),
				}},
			autogold.Expect(map[string]int{"myRepo": 4}),
		},
		{
			"Count repos on diff matches",
			types.REPO_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					diffMatch("myRepo", "author-a", 1),
					diffMatch("myRepo", "author-b", 1),
				}},
			autogold.Expect(map[string]int{"myRepo": 4}),
		},
		{
			"Count multiple repos on diff matches",
			types.REPO_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					diffMatch("myRepo", "author-a", 1),
					diffMatch("myRepo2", "author-b", 2),
				}},
			autogold.Expect(map[string]int{"myRepo": 2, "myRepo2": 2}),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			aggregator := testAggregator{results: make(map[string]int)}
			countFunc, _ := GetCountFuncForMode("", "", tc.mode)
			sra := newTestSearchResultsAggregator(context.Background(), aggregator.AddResult, countFunc, tc.mode, nil)
			sra.Send(tc.searchEvent)
			tc.want.Equal(t, aggregator.results)
		})
	}
}

func TestAuthorAggregation(t *testing.T) {
	testCases := []struct {
		name        string
		mode        types.SearchAggregationMode
		searchEvent streaming.SearchEvent
		want        autogold.Value
	}{
		{
			"No results",
			types.AUTHOR_AGGREGATION_MODE, streaming.SearchEvent{}, autogold.Expect(map[string]int{})},
		{
			"No author for content match",
			types.AUTHOR_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{contentMatch("myRepo", "file.go", 1, "a", "b")},
			},
			autogold.Expect(map[string]int{}),
		},
		{
			"No author for symbol match",
			types.AUTHOR_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{symbolMatch("myRepo", "file.go", 1, "a", "b")},
			},
			autogold.Expect(map[string]int{}),
		},
		{
			"No author for path match",
			types.AUTHOR_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{pathMatch("myRepo", "file.go", 1)},
			},
			autogold.Expect(map[string]int{}),
		},
		{
			"counts by author",
			types.AUTHOR_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					commitMatch("repoA", "Author A", sampleDate, 1, 2, "a"),
					commitMatch("repoA", "Author B", sampleDate, 1, 2, "a"),
					commitMatch("repoB", "Author B", sampleDate, 2, 2, "a"),
					commitMatch("repoB", "Author C", sampleDate, 2, 2, "a"),
				},
			},
			autogold.Expect(map[string]int{"Author A": 2, "Author B": 4, "Author C": 2}),
		},
		{
			"Count authors on diff matches",
			types.AUTHOR_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					diffMatch("myRepo", "author-a", 1),
					diffMatch("myRepo2", "author-a", 2),
					diffMatch("myRepo2", "author-b", 2),
				}},
			autogold.Expect(map[string]int{"author-a": 4, "author-b": 2}),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			aggregator := testAggregator{results: make(map[string]int)}
			countFunc, _ := GetCountFuncForMode("", "", tc.mode)
			sra := newTestSearchResultsAggregator(context.Background(), aggregator.AddResult, countFunc, tc.mode, nil)
			sra.Send(tc.searchEvent)
			tc.want.Equal(t, aggregator.results)
		})
	}
}

func TestPathAggregation(t *testing.T) {
	testCases := []struct {
		name        string
		mode        types.SearchAggregationMode
		searchEvent streaming.SearchEvent
		want        autogold.Value
	}{
		{
			"No results",
			types.PATH_AGGREGATION_MODE, streaming.SearchEvent{}, autogold.Expect(map[string]int{})},
		{
			"no path for commit",
			types.PATH_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					commitMatch("repoA", "Author A", sampleDate, 1, 2, "a"),
				},
			},
			autogold.Expect(map[string]int{}),
		},
		{
			"no path on repo match",
			types.PATH_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					repoMatch("myRepo", 1),
				},
			},
			autogold.Expect(map[string]int{}),
		},
		{
			"Single file match multiple results",
			types.PATH_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{contentMatch("myRepo", "file.go", 1, "a", "b")},
			},
			autogold.Expect(map[string]int{"file.go": 2}),
		},
		{
			"Multiple file match multiple results",
			types.PATH_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					contentMatch("myRepo", "file.go", 1, "a", "b"),
					contentMatch("myRepo", "file2.go", 1, "d", "e"),
				},
			},
			autogold.Expect(map[string]int{"file.go": 2, "file2.go": 2}),
		},
		{
			"Multiple repos same file",
			types.PATH_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					contentMatch("myRepo", "file.go", 1, "a", "b"),
					contentMatch("myRepo2", "file.go", 2, "a", "b"),
				},
			},
			autogold.Expect(map[string]int{"file.go": 4}),
		},
		{
			"Count paths on path matches",
			types.PATH_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					pathMatch("myRepo", "file1.go", 1),
					pathMatch("myRepo", "file2.go", 1),
					pathMatch("myRepoB", "file3.go", 2),
				},
			},
			autogold.Expect(map[string]int{"file1.go": 1, "file2.go": 1, "file3.go": 1}),
		},
		{
			"Count paths on symbol matches",
			types.PATH_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					symbolMatch("myRepo", "file1.go", 1, "a", "b"),
					symbolMatch("myRepo", "file2.go", 1, "c", "d"),
				},
			},
			autogold.Expect(map[string]int{"file1.go": 2, "file2.go": 2}),
		},
		{
			"Count paths on multiple matche types",
			types.PATH_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					repoMatch("myRepo", 1),
					pathMatch("myRepo", "file1.go", 1),
					symbolMatch("myRepo", "file1.go", 1, "c", "d"),
					contentMatch("myRepo", "file.go", 1, "a", "b"),
				},
			},
			autogold.Expect(map[string]int{"file.go": 2, "file1.go": 3}),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			aggregator := testAggregator{results: make(map[string]int)}
			countFunc, _ := GetCountFuncForMode("", "", tc.mode)
			sra := newTestSearchResultsAggregator(context.Background(), aggregator.AddResult, countFunc, tc.mode, nil)
			sra.Send(tc.searchEvent)
			tc.want.Equal(t, aggregator.results)
		})
	}
}

func TestCaptureGroupAggregation(t *testing.T) {
	longCaptureGroup := "111111111|222222222|333333333|444444444|555555555|666666666|777777777|888888888|999999999|000000000|"
	testCases := []struct {
		name        string
		mode        types.SearchAggregationMode
		searchEvent streaming.SearchEvent
		query       string
		want        autogold.Value
	}{
		{
			"no results",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			streaming.SearchEvent{},
			"TEST",
			autogold.Expect(map[string]int{})},
		{
			"two keys from 1 chunk",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{contentMatch("myRepo", "file.go", 1, "python2.7 python3.9")},
			},
			`python([0-9]\.[0-9])`,
			autogold.Expect(map[string]int{"2.7": 1, "3.9": 1}),
		},
		{
			"count 2 from 1 chunk",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{contentMatch("myRepo", "file.go", 1, "python2.7 python2.7")},
			},
			`python([0-9]\.[0-9])`,
			autogold.Expect(map[string]int{"2.7": 2}),
		},
		{
			"count multiple results",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					contentMatch("myRepo", "file.go", 1, "python2.7 python3.9"),
					contentMatch("myRepo2", "file2.go", 2, "python2.7 python3.9"),
				},
			},
			`python([0-9]\.[0-9])`,
			autogold.Expect(map[string]int{"2.7": 2, "3.9": 2}),
		},
		{
			"skips non capturing group",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					contentMatch("myRepo", "file.go", 1, "python2.7 python3.9"),
				},
			},
			`python(?:[0-9])\.([0-9])`,
			autogold.Expect(map[string]int{"7": 1, "9": 1}),
		},
		{
			"capture match respects case:no",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{contentMatch("myRepo", "file.go", 1, "Python.7 PyThoN2.7")},
			},
			`repo:^github\.com/sourcegraph/sourcegraph python([0-9]\.[0-9]) case:no`,
			autogold.Expect(map[string]int{"2.7": 1}),
		},
		{
			"capture match respects case:yes",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{contentMatch("myRepo", "file.go", 1, "Python.7 PyThoN2.7")},
			},
			`repo:^github\.com/sourcegraph/sourcegraph python([0-9]\.[0-9]) case:yes`,
			autogold.Expect(map[string]int{}),
		},
		{
			"only get values from first capture group",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					contentMatch("myRepo", "file.go", 1, "python2.7 python2.7"),
					contentMatch("myRepo", "file2.go", 1, "python2.8 python2.9"),
				},
			},
			`python([0-9])\.([0-9])`,
			autogold.Expect(map[string]int{"2": 4}),
		},
		{
			"whole match only",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					contentMatch("myRepo", "file.go", 1, "2.7"),
					contentMatch("myRepo", "file2.go", 1, "2.9"),
				},
			},
			`([0-9]\.[0-9])`,
			autogold.Expect(map[string]int{"2.7": 1, "2.9": 1}),
		},
		{
			"no more than 100 characters",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					contentMatch("myRepo", "file.go", 1, "z"+longCaptureGroup+"extraz"),
					contentMatch("myRepo", "file2.go", 1, "zsmallMatchz"),
				},
			},
			`z(.*)z`,
			autogold.Expect(map[string]int{longCaptureGroup: 1, "smallMatch": 1}),
		},
		{
			"accepts exactly 100 characters",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					contentMatch("myRepo", "file.go", 1, "z"+longCaptureGroup+"z"),
				},
			},
			`z(.*)z`,
			autogold.Expect(map[string]int{longCaptureGroup: 1}),
		},
		{
			"capture groups against whole file matches",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					pathMatch("myRepo", "dir1/file1.go", 1),
					pathMatch("myRepo", "dir2/file2.go", 1),
					pathMatch("myRepo", "dir2/file3.go", 1),
				},
			},
			`(.*?)\/`,
			autogold.Expect(map[string]int{"dir1": 1, "dir2": 2}),
		},
		{
			"capture groups against repo matches",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					repoMatch("myRepo-a", 1),
					repoMatch("myRepo-a", 1),
					repoMatch("myRepo-b", 2),
				},
			},
			`myrepo-(.*)`,
			autogold.Expect(map[string]int{"a": 2, "b": 1}),
		},
		{
			"capture groups against commit matches",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					commitMatch("myRepo", "Author A", sampleDate, 1, 2, "python2.7 python2.7"),
					commitMatch("myRepo", "Author B", sampleDate, 1, 2, "python2.7 python2.8"),
				},
			},
			`python([0-9]\.[0-9])`,
			autogold.Expect(map[string]int{"2.7": 3, "2.8": 1}),
		},
		{
			"capture groups against commit matches case sensitive",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					commitMatch("myRepo", "Author A", sampleDate, 1, 2, "Python2.7 Python2.7"),
					commitMatch("myRepo", "Author B", sampleDate, 1, 2, "python2.7 Python2.8"),
				},
			},
			`python([0-9]\.[0-9]) case:yes`,
			autogold.Expect(map[string]int{"2.7": 1}),
		},
		{
			"capture groups against multiple match types",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					repoMatch("sourcegraph-repo1", 1),
					repoMatch("sourcegraph-repo2", 2),
					pathMatch("sourcegraph-repo1", "/dir/sourcegraph-test/file1.go", 1),
					pathMatch("sourcegraph-repo1", "/dir/sourcegraph-client/file1.go", 1),
					contentMatch("sourcegraph-repo1", "/dir/sourcegraph-client/app.css", 1, ".sourcegraph-notifications {", ".sourcegraph-alerts {"),
					contentMatch("sourcegraph-repo1", "/dir/sourcegraph-client-legacy/app.css", 1, ".sourcegraph-notifications {"),
				},
			},
			`/sourcegraph-(\\w+)/ patterntype:standard`,
			autogold.Expect(map[string]int{"repo1": 1, "repo2": 1, "test": 1, "client": 1, "notifications": 2, "alerts": 1}),
		},
		{
			"capture groups ignores diff types",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					diffMatch("sourcegraph-repo1", "author-a", 1),
				},
			},
			`/need(.)/ patterntype:standard`,
			autogold.Expect(map[string]int{}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			aggregator := testAggregator{results: make(map[string]int)}
			countFunc, err := GetCountFuncForMode(tc.query, "regexp", tc.mode)
			if err != nil {
				t.Errorf("expected test not to error, got %v", err)
				t.FailNow()
			}
			sra := newTestSearchResultsAggregator(context.Background(), aggregator.AddResult, countFunc, tc.mode, nil)
			sra.Send(tc.searchEvent)
			tc.want.Equal(t, aggregator.results)
		})
	}
}

func TestRepoMetadataAggregation(t *testing.T) {
	testCases := []struct {
		name        string
		mode        types.SearchAggregationMode
		searchEvent streaming.SearchEvent
		want        autogold.Value
	}{
		{
			"No results",
			types.REPO_METADATA_AGGREGATION_MODE,
			streaming.SearchEvent{Results: []result.Match{}},
			autogold.Expect(map[string]int{}),
		},
		{
			"Single repo match no metadata",
			types.REPO_METADATA_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{repoMatch("myRepo2", 1)},
			},
			autogold.Expect(map[string]int{"No metadata": 1}),
		},
		{
			"Single repo match multiple metadata",
			types.REPO_METADATA_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{repoMatch("myRepo", 1), repoMatch("myRepo2", 2), repoMatch("myRepo3", 3)},
			},
			autogold.Expect(map[string]int{"open-source": 1, "No metadata": 1, "team:sourcegraph": 1}),
		},
	}
	db := dbmocks.NewMockDB()
	repos := dbmocks.NewMockRepoStore()
	sgString := "sourcegraph"
	repos.ListFunc.SetDefaultReturn([]*dTypes.Repo{
		{Name: "myRepo", ID: 1},
		{Name: "myRepo2", ID: 2, KeyValuePairs: map[string]*string{"open-source": nil}},
		{Name: "myRepo3", ID: 3, KeyValuePairs: map[string]*string{"team": &sgString}},
	}, nil)
	db.ReposFunc.SetDefaultReturn(repos)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			aggregator := testAggregator{results: make(map[string]int)}
			countFunc, _ := GetCountFuncForMode("", "", tc.mode)
			sra := newTestSearchResultsAggregator(context.Background(), aggregator.AddResult, countFunc, tc.mode, db)
			sra.Send(tc.searchEvent)
			tc.want.Equal(t, aggregator.results)
		})
	}
}

func TestAggregationCancelation(t *testing.T) {
	testCases := []struct {
		name        string
		mode        types.SearchAggregationMode
		searchEvent streaming.SearchEvent
		query       string
		want        autogold.Value
	}{
		{
			"aggregator stops counting if context canceled",
			types.CAPTURE_GROUP_AGGREGATION_MODE,
			streaming.SearchEvent{
				Results: []result.Match{
					contentMatch("myRepo", "file.go", 1, "python2.7 python3.9"),
					contentMatch("myRepo2", "file2.go", 2, "python2.7 python3.9"),
				},
			},
			`python([0-9]\.[0-9])`,
			autogold.Expect(map[string]int{"2.7": 2, "3.9": 2}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			aggregator := testAggregator{results: make(map[string]int)}
			countFunc, err := GetCountFuncForMode(tc.query, "regexp", tc.mode)
			if err != nil {
				t.Errorf("expected test not to error, got %v", err)
				t.FailNow()
			}
			ctx, cancel := context.WithCancel(context.Background())
			sra := newTestSearchResultsAggregator(ctx, aggregator.AddResult, countFunc, tc.mode, nil)
			sra.Send(tc.searchEvent)
			cancel()
			sra.Send(tc.searchEvent)
			tc.want.Equal(t, aggregator.results)
			if len(aggregator.errors) != 1 {
				t.Errorf("context cancel should be captured as an error")
			}
		})
	}
}
