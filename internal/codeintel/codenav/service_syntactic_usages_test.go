package codenav

import (
	"context"
	"testing"

	genslices "github.com/life4/genesis/slices"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
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

func TestSearchBasedUsages_ResultWithoutSymbols(t *testing.T) {
	refRange := scipRange(1)
	refRange2 := scipRange(2)

	mockSearchClient := FakeSearchClient().
		WithFile("path.java", refRange, refRange2).
		Build()

	svc := newService(
		observation.TestContextTB(t), defaultMockRepoStore(),
		NewMockLsifStore(), NewMockUploadService(), gitserver.NewMockClient(),
		mockSearchClient, log.NoOp(),
	)

	usages, searchErr := svc.searchBasedUsagesInner(
		context.Background(), observation.TestTraceLogger(log.NoOp()), UsagesForSymbolArgs{},
		"symbol", "Java", core.None[MappedIndex](),
	)
	require.NoError(t, searchErr)
	expectSearchRanges(t, usages, refRange, refRange2)
}

func TestSearchBasedUsages_ResultWithSymbol(t *testing.T) {
	refRange := scipRange(1)
	defRange := scipRange(2)
	refRange2 := scipRange(3)

	mockSearchClient := FakeSearchClient().
		WithFile("path.java", refRange, refRange2, defRange).
		WithSymbols("path.java", defRange).
		Build()

	svc := newService(
		observation.TestContextTB(t), defaultMockRepoStore(),
		NewMockLsifStore(), NewMockUploadService(), gitserver.NewMockClient(),
		mockSearchClient, log.NoOp(),
	)

	usages, searchErr := svc.searchBasedUsagesInner(
		context.Background(), observation.TestTraceLogger(log.NoOp()), UsagesForSymbolArgs{},
		"symbol", "Java", core.None[MappedIndex](),
	)
	require.NoError(t, searchErr)
	expectSearchRanges(t, usages, refRange, refRange2, defRange)
	expectDefinitionRanges(t, usages, defRange)
}
