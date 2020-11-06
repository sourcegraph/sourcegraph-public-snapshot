package api

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	storemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore/mocks"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	bundlemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore/mocks"
)

func TestRanges(t *testing.T) {
	mockStore := storemocks.NewMockStore()
	mockBundleStore := bundlemocks.NewMockStore()
	mockGitserverClient := NewMockGitserverClient()

	sourceRanges := []lsifstore.CodeIntelligenceRange{
		{
			Range:       lsifstore.Range{Start: lsifstore.Position{1, 2}, End: lsifstore.Position{3, 4}},
			Definitions: []lsifstore.Location{},
			References:  []lsifstore.Location{},
			HoverText:   "",
		},
		{
			Range:       lsifstore.Range{Start: lsifstore.Position{2, 3}, End: lsifstore.Position{4, 5}},
			Definitions: []lsifstore.Location{{Path: "foo.go", Range: lsifstore.Range{Start: lsifstore.Position{10, 20}, End: lsifstore.Position{30, 40}}}},
			References:  []lsifstore.Location{{Path: "bar.go", Range: lsifstore.Range{Start: lsifstore.Position{100, 200}, End: lsifstore.Position{300, 400}}}},
			HoverText:   "ht2",
		},
		{
			Range:       lsifstore.Range{Start: lsifstore.Position{3, 4}, End: lsifstore.Position{5, 6}},
			Definitions: []lsifstore.Location{{Path: "bar.go", Range: lsifstore.Range{Start: lsifstore.Position{11, 21}, End: lsifstore.Position{31, 41}}}},
			References:  []lsifstore.Location{{Path: "foo.go", Range: lsifstore.Range{Start: lsifstore.Position{101, 201}, End: lsifstore.Position{301, 401}}}},
			HoverText:   "ht3",
		},
	}

	setMockStoreGetDumpByID(t, mockStore, map[int]store.Dump{42: testDump1})
	setMockBundleStoreRanges(t, mockBundleStore, 42, "main.go", 10, 20, sourceRanges)

	api := testAPI(mockStore, mockBundleStore, mockGitserverClient)
	ranges, err := api.Ranges(context.Background(), "sub1/main.go", 10, 20, 42)
	if err != nil {
		t.Fatalf("expected error getting ranges: %s", err)
	}

	expectedRanges := []ResolvedCodeIntelligenceRange{
		{
			Range:       lsifstore.Range{Start: lsifstore.Position{1, 2}, End: lsifstore.Position{3, 4}},
			Definitions: nil,
			References:  nil,
			HoverText:   "",
		},
		{
			Range:       lsifstore.Range{Start: lsifstore.Position{2, 3}, End: lsifstore.Position{4, 5}},
			Definitions: []ResolvedLocation{{Dump: testDump1, Path: "sub1/foo.go", Range: lsifstore.Range{Start: lsifstore.Position{10, 20}, End: lsifstore.Position{30, 40}}}},
			References:  []ResolvedLocation{{Dump: testDump1, Path: "sub1/bar.go", Range: lsifstore.Range{Start: lsifstore.Position{100, 200}, End: lsifstore.Position{300, 400}}}},
			HoverText:   "ht2",
		},
		{
			Range:       lsifstore.Range{Start: lsifstore.Position{3, 4}, End: lsifstore.Position{5, 6}},
			Definitions: []ResolvedLocation{{Dump: testDump1, Path: "sub1/bar.go", Range: lsifstore.Range{Start: lsifstore.Position{11, 21}, End: lsifstore.Position{31, 41}}}},
			References:  []ResolvedLocation{{Dump: testDump1, Path: "sub1/foo.go", Range: lsifstore.Range{Start: lsifstore.Position{101, 201}, End: lsifstore.Position{301, 401}}}},
			HoverText:   "ht3",
		},
	}
	if diff := cmp.Diff(expectedRanges, ranges); diff != "" {
		t.Errorf("unexpected range (-want +got):\n%s", diff)
	}
}

func TestRangesUnknownDump(t *testing.T) {
	mockStore := storemocks.NewMockStore()
	mockBundleStore := bundlemocks.NewMockStore()
	mockGitserverClient := NewMockGitserverClient()
	setMockStoreGetDumpByID(t, mockStore, nil)

	api := testAPI(mockStore, mockBundleStore, mockGitserverClient)
	if _, err := api.Ranges(context.Background(), "sub1", 42, 0, 10); err != ErrMissingDump {
		t.Fatalf("unexpected error getting ranges. want=%q have=%q", ErrMissingDump, err)
	}
}
