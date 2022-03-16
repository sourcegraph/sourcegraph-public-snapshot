package searcher

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/limits"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestRepoShouldBeSearched(t *testing.T) {
	MockSearch = func(ctx context.Context, repo api.RepoName, repoID api.RepoID, commit api.CommitID, p *search.TextPatternInfo, fetchTimeout time.Duration, onMatches func([]*protocol.FileMatch)) (limitHit bool, err error) {
		repoName := repo
		switch repoName {
		case "foo/one":
			onMatches([]*protocol.FileMatch{{Path: "main.go"}})
			return false, nil
		case "foo/no-filematch":
			onMatches([]*protocol.FileMatch{})
			return false, nil
		default:
			return false, errors.New("Unexpected repo")
		}
	}
	defer func() { MockSearch = nil }()
	info := &search.TextPatternInfo{
		FileMatchLimit:               limits.DefaultMaxSearchResults,
		Pattern:                      "foo",
		FilePatternsReposMustInclude: []string{"main"},
	}

	shouldBeSearched, err := repoShouldBeSearched(context.Background(), nil, info, types.MinimalRepo{Name: "foo/one", ID: 1}, "1a2b3c", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if !shouldBeSearched {
		t.Errorf("expected repo to be searched, got shouldn't be searched")
	}

	shouldBeSearched, err = repoShouldBeSearched(context.Background(), nil, info, types.MinimalRepo{Name: "foo/no-filematch", ID: 2}, "1a2b3c", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	if shouldBeSearched {
		t.Errorf("expected repo to not be searched, got should be searched")
	}
}
