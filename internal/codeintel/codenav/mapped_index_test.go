package codenav

import (
	"context"
	"testing"

	"github.com/sourcegraph/scip/bindings/go/scip"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/internal/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
)

func setupSimpleUpload() (api.CommitID, uploadsshared.CompletedUpload, lsifstore.LsifStore) {
	indexCommit := api.CommitID("deadbeef")
	targetCommit := api.CommitID("beefdead")
	upload, lsifStore := setupUpload(indexCommit, "indexRoot/",
		doc("a.go",
			ref("a", testRange(1)),
			ref("b", testRange(2)),
			ref("c", testRange(3))),
		doc("b.go",
			ref("a", testRange(2))))
	return targetCommit, upload, lsifStore
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
	// The translator passed to NewMappedIndex uses the targetCommit as its base.
	// This -2 thus means going from index -> target shifts by +2 lines.
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
