package codenav

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type FakeFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (fi FakeFileInfo) Name() string {
	return fi.name
}

func (fi FakeFileInfo) Size() int64 {
	return fi.size
}

func (fi FakeFileInfo) Mode() os.FileMode {
	return fi.mode
}

func (fi FakeFileInfo) ModTime() time.Time {
	return fi.modTime
}

func (fi FakeFileInfo) IsDir() bool {
	return fi.isDir
}

func (fi FakeFileInfo) Sys() interface{} {
	return nil
}

func TestSyntacticAndSearchBasedUsages(t *testing.T) {
	mockRepoStore := defaultMockRepoStore()
	mockLsifStore := NewMockLsifStore()
	mockUploadSvc := NewMockUploadService()
	mockGitserverClient := gitserver.NewMockClient()
	mockSearchClient := client.NewMockSearchClient()

	repoId := 20
	repoName := "github.com/syntactic/usages"
	repo := types.Repo{
		ID:   api.RepoID(repoId),
		Name: api.RepoName(repoName),
	}

	pathToRequestSymbol := "path/to/request/symbol.java"
	pathToOccurrence := "path/to/occurrence.java"
	commit := "deadbeef"
	uploadID := 42
	symbolRange := scip.NewRangeUnchecked([]int32{1, 0, 16})
	symbolString := "scip-syntax . . . actualOccurrence#"
	symbolFile := "\nactualOccurrence"
	actualOccurrenceRange := result.Range{
		Start: result.Location{0, 10, 1},
		End:   result.Location{0, 10, 5},
	}
	scipActualOccurrenceRange, _ := scipFromResultRange(actualOccurrenceRange)
	commentRange := result.Range{
		Start: result.Location{0, 0, 0},
		End:   result.Location{0, 0, 10},
	}
	scipCommentRange, _ := scipFromResultRange(commentRange)
	localRange := result.Range{
		Start: result.Location{0, 20, 1},
		End:   result.Location{0, 20, 5},
	}
	scipLocalRange, _ := scipFromResultRange(localRange)

	mockRepoStore.GetReposSetByIDsFunc.SetDefaultReturn(map[api.RepoID]*types.Repo{api.RepoID(repoId): &repo}, nil)

	expectSearchQuery := func(expected string) {
		mockSearchClient.PlanFunc.PushHook(func(_ context.Context, _ string, _ *string, query string, _ search.Mode, _ search.Protocol, _ *int32) (*search.Inputs, error) {
			if query != expected {
				t.Errorf("unexpected query:\nGot: %q\nExp: %q", query, expected)
			}
			return &search.Inputs{}, nil
		})
	}
	mockSearchClient.ExecuteFunc.SetDefaultHook(func(_ context.Context, s streaming.Sender, _ *search.Inputs) (*search.Alert, error) {
		s.Send(streaming.SearchEvent{
			Results: result.Matches{&result.FileMatch{
				File: result.File{Path: pathToOccurrence},
				ChunkMatches: result.ChunkMatches{{
					Ranges: result.Ranges{commentRange, actualOccurrenceRange},
				}, {
					Ranges: result.Ranges{localRange},
				}},
			}},
		})
		return nil, nil
	})

	mockGitserverClient.GetCommitFunc.SetDefaultReturn(&gitdomain.Commit{
		ID: api.CommitID(commit),
	}, nil)

	syntacticUpload := shared.CompletedUpload{
		ID:             uploadID,
		Commit:         commit,
		Root:           "",
		RepositoryID:   repoId,
		RepositoryName: repoName,
		Indexer:        shared.SyntacticIndexer,
		IndexerVersion: "v1.0.0",
	}

	mockUploadSvc.InferClosestUploadsFunc.SetDefaultReturn([]shared.CompletedUpload{syntacticUpload}, nil)
	mockLsifStore.SCIPDocumentFunc.SetDefaultHook(func(_ context.Context, requestedUploadID int, path core.UploadRelPath) (*scip.Document, error) {
		if requestedUploadID == uploadID && path.RawValue() == pathToRequestSymbol {
			return &scip.Document{
				RelativePath: pathToRequestSymbol,
				Occurrences: []*scip.Occurrence{{
					Range:  symbolRange.SCIPRange(),
					Symbol: symbolString,
				}},
			}, nil
		} else if requestedUploadID == uploadID && path.RawValue() == pathToOccurrence {
			return &scip.Document{
				RelativePath: pathToOccurrence,
				Occurrences: []*scip.Occurrence{{
					Range:  scipActualOccurrenceRange.SCIPRange(),
					Symbol: symbolString,
				}, {
					Range:  scipLocalRange.SCIPRange(),
					Symbol: "local 1",
				}},
			}, nil
		}
		return nil, nil
	})
	mockLsifStore.FindDocumentIDsFunc.SetDefaultHook(func(ctx context.Context, paths map[int]core.UploadRelPath) (map[int]int, error) {
		if paths[42].RawValue() == pathToRequestSymbol || paths[42].RawValue() == pathToOccurrence {
			return map[int]int{42: 1}, nil
		}
		return nil, nil
	})

	svc := newService(observation.TestContextTB(t), mockRepoStore, mockLsifStore, mockUploadSvc, mockGitserverClient, mockSearchClient, log.NoOp())

	candidateOccurrenceQuery := "case:yes type:file repo:^github.com/syntactic/usages$ rev:deadbeef language:Java count:500 /\\bactualOccurrence\\b/"
	candidateSymbolQuery := "case:yes type:symbol repo:^github.com/syntactic/usages$ rev:deadbeef language:Java count:50 /\\bactualOccurrence\\b/"

	expectSearchQuery(candidateOccurrenceQuery)
	syntacticUsages, previousSyntacticSearch, err := svc.SyntacticUsages(context.Background(), UsagesForSymbolArgs{
		Repo:        repo,
		Commit:      api.CommitID(commit),
		Path:        core.NewRepoRelPathUnchecked(pathToRequestSymbol),
		SymbolRange: symbolRange,
	})

	if err != nil {
		t.Fatal(err)
	}
	// Check that we return only the actual occurrence, and filter out both the comment and the local occurrence
	if len(syntacticUsages.Matches) != 1 {
		t.Errorf("Expected a single syntactic usage result, but got %+v", syntacticUsages)
	}
	syntacticMatch := syntacticUsages.Matches[0]
	if syntacticMatch.Occurrence.Symbol != symbolString {
		t.Errorf("Expected symbol to be %q, but got %s", symbolString, syntacticUsages.Matches[0].Occurrence.Symbol)
	}
	if syntacticMatch.Range().CompareStrict(scipActualOccurrenceRange) != 0 {
		t.Errorf("Expected syntactic range to be %q, but got %q", scipActualOccurrenceRange.String(), syntacticMatch.Range().String())
	}

	expectSearchQuery(candidateOccurrenceQuery)
	expectSearchQuery(candidateSymbolQuery)
	searchBasedUsagesPrev, searchErrs := svc.SearchBasedUsages(context.Background(), UsagesForSymbolArgs{
		Repo:        repo,
		Commit:      api.CommitID(commit),
		Path:        core.NewRepoRelPathUnchecked(pathToRequestSymbol),
		SymbolRange: symbolRange,
	}, previousSyntacticSearch)
	if searchErrs != nil {
		t.Fatal(err)
	}
	if len(searchBasedUsagesPrev) != 2 {
		t.Errorf("Expected a two search-based usage results, but got %+v", searchBasedUsagesPrev)
	}
	if searchBasedUsagesPrev[0].Range.CompareStrict(scipCommentRange) != 0 {
		t.Errorf("Expected first search-based result to be comment, but got %+v", searchBasedUsagesPrev[0])
	}
	if searchBasedUsagesPrev[1].Range.CompareStrict(scipLocalRange) != 0 {
		t.Errorf("Expected second search-based result to be local, but got %+v", searchBasedUsagesPrev[1])
	}

	// Only mock these calls here to make sure the previous search-based usages call did not invoke them, as
	// it used the `previousSyntacticSearch` to retrieve the symbolName.
	mockGitserverClient.StatFunc.SetDefaultReturn(FakeFileInfo{size: 100}, nil)
	mockGitserverClient.NewFileReaderFunc.SetDefaultReturn(io.NopCloser(bytes.NewReader([]byte(symbolFile))), nil)

	expectSearchQuery(candidateOccurrenceQuery)
	expectSearchQuery(candidateSymbolQuery)
	searchBasedUsages, searchErrs := svc.SearchBasedUsages(context.Background(), UsagesForSymbolArgs{
		Repo:        repo,
		Commit:      api.CommitID(commit),
		Path:        core.NewRepoRelPathUnchecked(pathToRequestSymbol),
		SymbolRange: symbolRange,
	}, nil)
	if searchErrs != nil {
		t.Fatal(err)
	}
	if len(searchBasedUsages) != 2 {
		t.Errorf("Expected a two search-based usage results, but got %+v", searchBasedUsages)
	}
	if searchBasedUsages[0].Range.CompareStrict(scipCommentRange) != 0 {
		t.Errorf("Expected first search-based result to be comment, but got %+v", searchBasedUsages[0])
	}
	if searchBasedUsages[1].Range.CompareStrict(scipLocalRange) != 0 {
		t.Errorf("Expected second search-based result to be local, but got %+v", searchBasedUsages[1])
	}
}
