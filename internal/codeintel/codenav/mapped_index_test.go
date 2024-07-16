package codenav

import (
	"context"
	"errors"
	"testing"

	"github.com/sourcegraph/scip/bindings/go/scip"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
)

var uploadIDSupply = 0

func newUploadID() int {
	uploadIDSupply += 1
	return uploadIDSupply
}

func testRange(r int) scip.Range {
	return scip.NewRangeUnchecked([]int32{int32(r), int32(r), int32(r)})
}

type fakeOccurrence struct {
	symbol       string
	isDefinition bool
	rg           scip.Range
}

type fakeDocument struct {
	path        core.UploadRelPath
	occurrences []fakeOccurrence
}

func (d fakeDocument) Occurrences() []*scip.Occurrence {
	occs := make([]*scip.Occurrence, 0, len(d.occurrences))
	for _, occ := range d.occurrences {
		var symbolRoles scip.SymbolRole = 0
		if occ.isDefinition {
			symbolRoles = scip.SymbolRole_Definition
		}
		occs = append(occs, &scip.Occurrence{
			Range:       occ.rg.SCIPRange(),
			Symbol:      occ.symbol,
			SymbolRoles: int32(symbolRoles),
		})
	}
	return occs
}

func ref(symbol string, rg int) fakeOccurrence {
	return fakeOccurrence{
		symbol:       symbol,
		isDefinition: false,
		rg:           testRange(rg),
	}
}

func doc(path string, occurrences ...fakeOccurrence) fakeDocument {
	return fakeDocument{
		path:        core.NewUploadRelPathUnchecked(path),
		occurrences: occurrences,
	}
}

// Set up uploads + lsifstore
func setupUpload(commit api.CommitID, root string, documents ...fakeDocument) (uploadsshared.CompletedUpload, lsifstore.LsifStore) {
	id := newUploadID()
	lsifStore := NewMockLsifStore()
	lsifStore.SCIPDocumentFunc.SetDefaultHook(func(ctx context.Context, uploadId int, path core.UploadRelPath) (core.Option[*scip.Document], error) {
		if id != uploadId {
			return core.None[*scip.Document](), errors.New("unknown upload id")
		}
		for _, document := range documents {
			if document.path.Equal(path) {
				return core.Some(&scip.Document{
					RelativePath: document.path.RawValue(),
					Occurrences:  document.Occurrences(),
				}), nil
			}
		}
		return core.None[*scip.Document](), nil
	})

	return uploadsshared.CompletedUpload{
		ID:     id,
		Commit: string(commit),
		Root:   root,
	}, lsifStore
}

func shiftSCIPRange(r scip.Range, numLines int) scip.Range {
	return scip.NewRangeUnchecked([]int32{
		r.Start.Line + int32(numLines),
		r.Start.Character,
		r.End.Line + int32(numLines),
		r.End.Character,
	})
}

func shiftPos(pos shared.Position, numLines int) shared.Position {
	return shared.Position{
		Line:      pos.Line + numLines,
		Character: pos.Character,
	}
}

// A GitTreeTranslator that returns positions and ranges shifted by numLines
// and returns failed translations for path/range pairs if shouldFail returns true
func fakeTranslator(
	targetCommit api.CommitID,
	numLines int,
	shouldFail func(string, shared.Range) bool,
) GitTreeTranslator {
	translator := NewMockGitTreeTranslator()
	translator.GetSourceCommitFunc.SetDefaultReturn(targetCommit)
	translator.GetTargetCommitPositionFromSourcePositionFunc.SetDefaultHook(func(ctx context.Context, commit string, path string, pos shared.Position, reverse bool) (shared.Position, bool, error) {
		numLines := numLines
		if reverse {
			numLines = -numLines
		}
		if shouldFail(path, shared.Range{Start: pos, End: pos}) {
			return shared.Position{}, false, nil
		}
		return shiftPos(pos, numLines), true, nil
	})
	translator.GetTargetCommitRangeFromSourceRangeFunc.SetDefaultHook(func(ctx context.Context, commit string, path string, rg shared.Range, reverse bool) (shared.Range, bool, error) {
		numLines := numLines
		if reverse {
			numLines = -numLines
		}
		if shouldFail(path, rg) {
			return shared.Range{}, false, nil
		}
		return shared.Range{Start: shiftPos(rg.Start, numLines), End: shiftPos(rg.End, numLines)}, true, nil
	})
	return translator
}

// A GitTreeTranslator that returns all positions and ranges shifted by numLines.
func shiftAllTranslator(targetCommit api.CommitID, numLines int) GitTreeTranslator {
	return fakeTranslator(targetCommit, numLines, func(path string, rg shared.Range) bool { return false })
}

// A GitTreeTranslator that returns all positions and ranges unchanged
func noopTranslator(targetCommit api.CommitID) GitTreeTranslator {
	return shiftAllTranslator(targetCommit, 0)
}

func setupSimpleUpload() (api.CommitID, uploadsshared.CompletedUpload, lsifstore.LsifStore) {
	indexCommit := api.CommitID("deadbeef")
	targetCommit := api.CommitID("beefdead")
	upload, lsifStore := setupUpload(indexCommit, "indexRoot/",
		doc("a.go",
			ref("a", 1),
			ref("b", 2),
			ref("c", 3)),
		doc("b.go",
			ref("a", 2)))
	return targetCommit, upload, lsifStore
}

func TestNewMappedIndex(t *testing.T) {
	indexCommit := api.CommitID("deadbeef")
	targetCommit := api.CommitID("beefdead")
	upload, lsifStore := setupUpload(indexCommit, "")
	mappedIndex, err := NewMappedIndex(lsifStore, nil, nil, upload, targetCommit)
	require.NoError(t, err)
	require.Equal(t, indexCommit, mappedIndex.IndexCommit())
	require.Equal(t, targetCommit, mappedIndex.TargetCommit())
}

func TestMappedIndex_GetDocumentNoTranslation(t *testing.T) {
	targetCommit, upload, lsifStore := setupSimpleUpload()
	translator := noopTranslator(targetCommit)
	mappedIndex := NewMappedIndexFromTranslator(lsifStore, translator, upload)

	ctx := context.Background()
	unknownDoc, err := mappedIndex.GetDocument(ctx, core.NewRepoRelPathUnchecked("indexRoot/unknown.go"))
	require.NoError(t, err)
	require.True(t, unknownDoc.IsNone())

	mappedDocumentResult, err := mappedIndex.GetDocument(ctx, core.NewRepoRelPathUnchecked("indexRoot/a.go"))
	require.NoError(t, err)
	mappedDocument := mappedDocumentResult.Unwrap()

	occurrences, err := mappedDocument.GetOccurrencesAtRange(ctx, testRange(1))
	require.NoError(t, err)
	require.Len(t, occurrences, 1)
	require.Equal(t, scip.NewRangeUnchecked(occurrences[0].GetRange()).Start.Line, int32(1))

	noOccurrences, err := mappedDocument.GetOccurrencesAtRange(ctx, testRange(4))
	require.NoError(t, err)
	require.Len(t, noOccurrences, 0)

	allOccurrences, err := mappedDocument.GetOccurrences(ctx)
	require.NoError(t, err)
	require.Len(t, allOccurrences, 3)
}

func TestMappedIndex_GetDocumentWithTranslation(t *testing.T) {
	targetCommit, upload, lsifStore := setupSimpleUpload()
	translator := shiftAllTranslator(targetCommit, -2)
	mappedIndex := NewMappedIndexFromTranslator(lsifStore, translator, upload)

	ctx := context.Background()
	mappedDocumentOption, err := mappedIndex.GetDocument(ctx, core.NewRepoRelPathUnchecked("indexRoot/a.go"))
	require.NoError(t, err)
	mappedDocument := mappedDocumentOption.Unwrap()

	noOccurrences, err := mappedDocument.GetOccurrencesAtRange(ctx, testRange(1))
	require.NoError(t, err)
	require.Len(t, noOccurrences, 0)

	occurrences, err := mappedDocument.GetOccurrencesAtRange(ctx, shiftSCIPRange(testRange(1), 2))
	require.NoError(t, err)
	require.Len(t, occurrences, 1)
	require.Equal(t, scip.NewRangeUnchecked(occurrences[0].GetRange()).Start.Line, int32(3))

	allOccurrences, err := mappedDocument.GetOccurrences(ctx)
	require.NoError(t, err)
	require.Len(t, allOccurrences, 3)
}

// This test is here to check MappedDocument 's internals, by getting all occurrences first,
// we're testing that the `isMapped` logic does not change the results of GetOccurrencesAtRange
func TestMappedIndex_GetOccurrencesAtRangeAfterGetOccurrences(t *testing.T) {
	targetCommit, upload, lsifStore := setupSimpleUpload()
	translator := shiftAllTranslator(targetCommit, -2)
	mappedIndex := NewMappedIndexFromTranslator(lsifStore, translator, upload)

	ctx := context.Background()
	mappedDocumentOption, err := mappedIndex.GetDocument(ctx, core.NewRepoRelPathUnchecked("indexRoot/a.go"))
	require.NoError(t, err)
	mappedDocument := mappedDocumentOption.Unwrap()

	allOccurrences, err := mappedDocument.GetOccurrences(ctx)
	require.NoError(t, err)
	require.Len(t, allOccurrences, 3)

	noOccurrences, err := mappedDocument.GetOccurrencesAtRange(ctx, testRange(1))
	require.NoError(t, err)
	require.Len(t, noOccurrences, 0)

	occurrences, err := mappedDocument.GetOccurrencesAtRange(ctx, shiftSCIPRange(testRange(1), 2))
	require.NoError(t, err)
	require.Len(t, occurrences, 1)
	require.Equal(t, scip.NewRangeUnchecked(occurrences[0].GetRange()).Start.Line, int32(3))
}

func TestMappedIndex_GetDocumentsFiltersFailedTranslation(t *testing.T) {
	targetCommit, upload, lsifStore := setupSimpleUpload()
	translator := fakeTranslator(targetCommit, 0, func(path string, rg shared.Range) bool {
		return rg.ToSCIPRange().CompareStrict(testRange(1)) == 0
	})
	mappedIndex := NewMappedIndexFromTranslator(lsifStore, translator, upload)

	ctx := context.Background()
	mappedDocumentOption, err := mappedIndex.GetDocument(ctx, core.NewRepoRelPathUnchecked("indexRoot/a.go"))
	require.NoError(t, err)
	mappedDocument := mappedDocumentOption.Unwrap()
	allOccurrences, err := mappedDocument.GetOccurrences(ctx)
	require.NoError(t, err)
	require.Len(t, allOccurrences, 2)
}

func TestMappedIndex_GetDocumentFailedTranslation(t *testing.T) {
	targetCommit, upload, lsifStore := setupSimpleUpload()
	translator := fakeTranslator(targetCommit, 0, func(path string, rg shared.Range) bool {
		return path == "indexRoot/b.go" || rg.ToSCIPRange().CompareStrict(testRange(1)) == 0
	})
	mappedIndex := NewMappedIndexFromTranslator(lsifStore, translator, upload)

	ctx := context.Background()
	mappedDocumentOption, err := mappedIndex.GetDocument(ctx, core.NewRepoRelPathUnchecked("indexRoot/b.go"))
	require.NoError(t, err)
	mappedDocument := mappedDocumentOption.Unwrap()

	occurrences, err := mappedDocument.GetOccurrencesAtRange(ctx, testRange(2))
	require.NoError(t, err)
	require.Len(t, occurrences, 0)
}
