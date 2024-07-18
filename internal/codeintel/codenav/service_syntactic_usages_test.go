package codenav

import (
	"context"
	"testing"

	genslices "github.com/life4/genesis/slices"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func expectSearchRanges(t *testing.T, matches []SearchBasedMatch, ranges ...scip.Range) {
	t.Helper()
	for _, match := range matches {
		_, err := genslices.Find(ranges, func(r scip.Range) bool {
			return match.Range.CompareStrict(r) == 0
		})
		if err != nil {
			t.Errorf("Did not expect match at %q", match.Range.String())
		}
	}

	for _, r := range ranges {
		_, err := genslices.Find(matches, func(match SearchBasedMatch) bool {
			return match.Range.CompareStrict(r) == 0
		})
		if err != nil {
			t.Errorf("Expected match at %q", r.String())
		}
	}
}

func expectSyntacticRanges(t *testing.T, matches []SyntacticMatch, ranges ...scip.Range) {
	t.Helper()
	for _, match := range matches {
		_, err := genslices.Find(ranges, func(r scip.Range) bool {
			return match.Range.CompareStrict(r) == 0
		})
		if err != nil {
			t.Errorf("Did not expect match at %q", match.Range.String())
		}
	}

	for _, r := range ranges {
		_, err := genslices.Find(matches, func(match SyntacticMatch) bool {
			return match.Range.CompareStrict(r) == 0
		})
		if err != nil {
			t.Errorf("Expected match at %q", r.String())
		}
	}
}

func expectDefinitionRanges(t *testing.T, matches []SearchBasedMatch, ranges ...scip.Range) {
	t.Helper()
	for _, match := range matches {
		_, err := genslices.Find(ranges, func(r scip.Range) bool {
			return match.Range.CompareStrict(r) == 0
		})
		if match.IsDefinition && err != nil {
			t.Errorf("Did not expect match at %q to be a definition", match.Range.String())
			return
		} else if !match.IsDefinition && err == nil {
			t.Errorf("Expected match at %q to be a definition", match.Range.String())
			return
		}
	}
}

func expectSyntacticDefinitionRanges(t *testing.T, matches []SyntacticMatch, ranges ...scip.Range) {
	t.Helper()
	for _, match := range matches {
		_, err := genslices.Find(ranges, func(r scip.Range) bool {
			return match.Range.CompareStrict(r) == 0
		})
		if match.IsDefinition && err != nil {
			t.Errorf("Did not expect match at %q to be a definition", match.Range.String())
			return
		} else if !match.IsDefinition && err == nil {
			t.Errorf("Expected match at %q to be a definition", match.Range.String())
			return
		}
	}
}

func TestSearchBasedUsages_ResultWithoutSymbols(t *testing.T) {
	refRange := testRange(1)
	refRange2 := testRange(2)

	mockSearchClient := FakeSearchClient().
		WithFile("path.java", refRange, refRange2).
		Build()

	usages, searchErr := searchBasedUsagesImpl(
		context.Background(), observation.TestTraceLogger(log.NoOp()), mockSearchClient,
		UsagesForSymbolArgs{}, "symbol", "Java", core.None[MappedIndex](),
	)
	require.NoError(t, searchErr)
	expectSearchRanges(t, usages, refRange, refRange2)
}

func TestSearchBasedUsages_ResultWithSymbol(t *testing.T) {
	refRange := testRange(1)
	defRange := testRange(2)
	refRange2 := testRange(3)

	mockSearchClient := FakeSearchClient().
		WithFile("path.java", refRange, refRange2, defRange).
		WithSymbols("path.java", defRange).
		Build()

	usages, searchErr := searchBasedUsagesImpl(
		context.Background(), observation.TestTraceLogger(log.NoOp()), mockSearchClient,
		UsagesForSymbolArgs{}, "symbol", "Java", core.None[MappedIndex](),
	)
	require.NoError(t, searchErr)
	expectSearchRanges(t, usages, refRange, refRange2, defRange)
	expectDefinitionRanges(t, usages, defRange)
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
	fakeMappedIndex := NewMappedIndexFromTranslator(lsifStore, noopTranslator(commit), upload)

	syntacticUsages, _, err := syntacticUsagesImpl(
		context.Background(), observation.TestTraceLogger(log.NoOp()),
		mockSearchClient, fakeMappedIndex, UsagesForSymbolArgs{
			Commit:      commit,
			Path:        core.NewRepoRelPathUnchecked("initial.java"),
			SymbolRange: initialRange,
		},
	)

	if err != nil {
		t.Error(err)
	}

	// We expect syntactic usages to filter both the comment range that was included in the search result,
	// but not in the index as well as the range referencing the local symbol.
	expectSyntacticRanges(t, syntacticUsages.Matches, initialRange, refRange, defRange)
	expectSyntacticDefinitionRanges(t, syntacticUsages.Matches, defRange)
}
