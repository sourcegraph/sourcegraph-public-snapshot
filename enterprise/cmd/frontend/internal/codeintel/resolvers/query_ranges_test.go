package resolvers

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestRanges(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockGitserverClient := NewMockGitserverClient()
	mockPositionAdjuster := noopPositionAdjuster()

	testLocation1 := lsifstore.Location{DumpID: 50, Path: "a.go", Range: testRange1}
	testLocation2 := lsifstore.Location{DumpID: 51, Path: "b.go", Range: testRange2}
	testLocation3 := lsifstore.Location{DumpID: 51, Path: "c.go", Range: testRange1}
	testLocation4 := lsifstore.Location{DumpID: 51, Path: "d.go", Range: testRange2}
	testLocation5 := lsifstore.Location{DumpID: 51, Path: "e.go", Range: testRange1}
	testLocation6 := lsifstore.Location{DumpID: 51, Path: "a.go", Range: testRange2}
	testLocation7 := lsifstore.Location{DumpID: 51, Path: "a.go", Range: testRange3}
	testLocation8 := lsifstore.Location{DumpID: 52, Path: "a.go", Range: testRange4}

	ranges := []lsifstore.CodeIntelligenceRange{
		{Range: testRange1, HoverText: "text1", Definitions: nil, References: []lsifstore.Location{testLocation1}},
		{Range: testRange2, HoverText: "text2", Definitions: []lsifstore.Location{testLocation2}, References: []lsifstore.Location{testLocation3}},
		{Range: testRange3, HoverText: "text3", Definitions: []lsifstore.Location{testLocation4}, References: []lsifstore.Location{testLocation5}},
		{Range: testRange4, HoverText: "text4", Definitions: []lsifstore.Location{testLocation6}, References: []lsifstore.Location{testLocation7}},
		{Range: testRange5, HoverText: "text5", Definitions: []lsifstore.Location{testLocation8}, References: nil},
	}

	mockLSIFStore.RangesFunc.PushReturn(ranges[0:1], nil)
	mockLSIFStore.RangesFunc.PushReturn(ranges[1:4], nil)
	mockLSIFStore.RangesFunc.PushReturn(ranges[4:], nil)

	uploads := []dbstore.Dump{
		{ID: 50, Commit: "deadbeef", Root: "sub1/"},
		{ID: 51, Commit: "deadbeef", Root: "sub2/"},
		{ID: 52, Commit: "deadbeef", Root: "sub3/"},
		{ID: 53, Commit: "deadbeef", Root: "sub4/"},
	}
	resolver := newQueryResolver(
		mockDBStore,
		mockLSIFStore,
		newCachedCommitChecker(mockGitserverClient),
		mockPositionAdjuster,
		42,
		"deadbeef",
		"s1/main.go",
		uploads,
		newOperations(&observation.TestContext),
	)
	adjustedRanges, err := resolver.Ranges(context.Background(), 10, 20)
	if err != nil {
		t.Fatalf("unexpected error querying ranges: %s", err)
	}

	adjustedLocation1 := AdjustedLocation{Dump: uploads[0], Path: "sub1/a.go", AdjustedCommit: "deadbeef", AdjustedRange: testRange1}
	adjustedLocation2 := AdjustedLocation{Dump: uploads[1], Path: "sub2/b.go", AdjustedCommit: "deadbeef", AdjustedRange: testRange2}
	adjustedLocation3 := AdjustedLocation{Dump: uploads[1], Path: "sub2/c.go", AdjustedCommit: "deadbeef", AdjustedRange: testRange1}
	adjustedLocation4 := AdjustedLocation{Dump: uploads[1], Path: "sub2/d.go", AdjustedCommit: "deadbeef", AdjustedRange: testRange2}
	adjustedLocation5 := AdjustedLocation{Dump: uploads[1], Path: "sub2/e.go", AdjustedCommit: "deadbeef", AdjustedRange: testRange1}
	adjustedLocation6 := AdjustedLocation{Dump: uploads[1], Path: "sub2/a.go", AdjustedCommit: "deadbeef", AdjustedRange: testRange2}
	adjustedLocation7 := AdjustedLocation{Dump: uploads[1], Path: "sub2/a.go", AdjustedCommit: "deadbeef", AdjustedRange: testRange3}
	adjustedLocation8 := AdjustedLocation{Dump: uploads[2], Path: "sub3/a.go", AdjustedCommit: "deadbeef", AdjustedRange: testRange4}

	expectedRanges := []AdjustedCodeIntelligenceRange{
		{Range: testRange1, HoverText: "text1", Definitions: []AdjustedLocation{}, References: []AdjustedLocation{adjustedLocation1}},
		{Range: testRange2, HoverText: "text2", Definitions: []AdjustedLocation{adjustedLocation2}, References: []AdjustedLocation{adjustedLocation3}},
		{Range: testRange3, HoverText: "text3", Definitions: []AdjustedLocation{adjustedLocation4}, References: []AdjustedLocation{adjustedLocation5}},
		{Range: testRange4, HoverText: "text4", Definitions: []AdjustedLocation{adjustedLocation6}, References: []AdjustedLocation{adjustedLocation7}},
		{Range: testRange5, HoverText: "text5", Definitions: []AdjustedLocation{adjustedLocation8}, References: []AdjustedLocation{}},
	}
	if diff := cmp.Diff(expectedRanges, adjustedRanges); diff != "" {
		t.Errorf("unexpected ranges (-want +got):\n%s", diff)
	}
}
