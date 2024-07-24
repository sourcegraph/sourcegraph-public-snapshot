package codenav

import (
	"context"
	"testing"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestSearchBasedUsages_ResultWithoutSymbols(t *testing.T) {
	refRange := testRange(1)
	refRange2 := testRange(2)

	mockSearchClient := FakeSearchClient().
		WithFile("path.java", refRange, refRange2).
		Build()

	usages, err := searchBasedUsagesImpl(
		context.Background(), observation.TestTraceLogger(log.NoOp()), mockSearchClient,
		UsagesForSymbolArgs{}, "symbol", "Java", core.None[MappedIndex](),
	)
	require.NoError(t, err)
	expectRanges(t, usages, refRange, refRange2)
}

func TestSearchBasedUsages_ResultWithSymbol(t *testing.T) {
	refRange := testRange(1)
	defRange := testRange(2)
	refRange2 := testRange(3)

	mockSearchClient := FakeSearchClient().
		WithFile("path.java", refRange, refRange2, defRange).
		WithSymbols("path.java", defRange).
		Build()

	usages, err := searchBasedUsagesImpl(
		context.Background(), observation.TestTraceLogger(log.NoOp()), mockSearchClient,
		UsagesForSymbolArgs{}, "symbol", "Java", core.None[MappedIndex](),
	)
	require.NoError(t, err)
	expectRanges(t, usages, refRange, refRange2, defRange)
	expectDefinitionRanges(t, usages, defRange)
}

func TestSearchBasedUsages_SyntacticMatchesGetRemovedFromSearchBasedResults(t *testing.T) {
	commentRange := testRange(1)
	syntacticRange := testRange(2)

	commit := api.CommitID("deadbeef")
	mockSearchClient := FakeSearchClient().WithFile("path.java", commentRange, syntacticRange).Build()
	upload, lsifStore := setupUpload(commit, "", doc("path.java", ref("ref", syntacticRange)))
	fakeMappedIndex := NewMappedIndexFromTranslator(lsifStore, noopTranslator(), upload, commit)

	usages, err := searchBasedUsagesImpl(
		context.Background(), observation.TestTraceLogger(log.NoOp()), mockSearchClient,
		UsagesForSymbolArgs{}, "symbol", "Java", core.Some(fakeMappedIndex),
	)
	require.NoError(t, err)
	expectRanges(t, usages, commentRange)
}

func TestSyntacticUsages(t *testing.T) {
	initialRange := testRange(10)
	refRange := testRange(1)
	defRange := testRange(2)
	commentRange := testRange(3)
	localRange := testRange(4)
	commit := api.CommitID("deadbeef")

	mockSearchClient := FakeSearchClient().
		WithFile("path.java", refRange, defRange, commentRange, localRange).
		WithFile("initial.java", initialRange).
		Build()
	upload, lsifStore := setupUpload(commit, "",
		doc("path.java",
			ref("ref", refRange),
			def("def", defRange),
			local("lcl", localRange)),
		doc("initial.java",
			ref("initial", initialRange)))
	fakeMappedIndex := NewMappedIndexFromTranslator(lsifStore, noopTranslator(), upload, commit)

	syntacticUsages, _, err := syntacticUsagesImpl(
		context.Background(), observation.TestTraceLogger(log.NoOp()),
		mockSearchClient, fakeMappedIndex, UsagesForSymbolArgs{
			Commit:      commit,
			Path:        core.NewRepoRelPathUnchecked("initial.java"),
			SymbolRange: initialRange,
		},
	)
	if err != nil {
		t.Error(t, err)
	}
	// We expect syntactic usages to filter both the comment range that was included in the search result,
	// but not in the index as well as the range referencing the local symbol.
	expectRanges(t, syntacticUsages.Matches, initialRange, refRange, defRange)
	expectDefinitionRanges(t, syntacticUsages.Matches, defRange)
}

func TestSyntacticUsages_DocumentNotInIndex(t *testing.T) {
	initialRange := testRange(1)
	refRange := testRange(2)
	commit := api.CommitID("deadbeef")

	mockSearchClient := FakeSearchClient().WithFile("not-in-index.java", refRange).Build()
	upload, lsifStore := setupUpload(commit, "",
		doc("initial.java",
			ref("initial", initialRange)))
	fakeMappedIndex := NewMappedIndexFromTranslator(lsifStore, noopTranslator(), upload, commit)
	syntacticUsages, _, err := syntacticUsagesImpl(
		context.Background(), observation.TestTraceLogger(log.NoOp()),
		mockSearchClient, fakeMappedIndex, UsagesForSymbolArgs{
			Commit:      commit,
			Path:        core.NewRepoRelPathUnchecked("initial.java"),
			SymbolRange: initialRange,
		},
	)
	if err != nil {
		t.Error(t, err)
	}
	expectRanges(t, syntacticUsages.Matches)
}

func TestSyntacticUsages_IndexCommitTranslated(t *testing.T) {
	initialRange := testRange(10)
	refRange := testRange(1)
	editedRange := testRange(2)
	noMatchRange := testRange(3)
	indexCommit := api.CommitID("deadbeef")
	targetCommit := api.CommitID("beefdead")

	mockSearchClient := FakeSearchClient().WithFile("path.java", refRange, editedRange, noMatchRange).Build()
	upload, lsifStore := setupUpload(indexCommit, "",
		doc("initial.java",
			ref("initial", shiftSCIPRange(initialRange, 2))),
		doc("path.java",
			ref("ref", shiftSCIPRange(refRange, 2)),
			ref("edited", shiftSCIPRange(editedRange, 2)),
			ref("noMatch", noMatchRange)))
	// Ranges in the index are shifted by +2, so the translator needs to shift by -2 to match up with the search results.
	fakeMappedIndex := NewMappedIndexFromTranslator(lsifStore, fakeTranslator(upload.GetCommit(), targetCommit, -2,
		func(_ core.RepoRelPath, r scip.Range) bool {
			// When a line was edited in a diff we invalidate all occurrences on that line.
			return r.CompareStrict(editedRange) == 0
		}), upload, targetCommit)

	syntacticUsages, _, err := syntacticUsagesImpl(
		context.Background(), observation.TestTraceLogger(log.NoOp()),
		mockSearchClient, fakeMappedIndex, UsagesForSymbolArgs{
			Commit:      targetCommit,
			Path:        core.NewRepoRelPathUnchecked("initial.java"),
			SymbolRange: initialRange,
		},
	)
	if err != nil {
		t.Error(t, err)
	}
	expectRanges(t, syntacticUsages.Matches, refRange)
}
