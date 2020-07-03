package api

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	bundlemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client/mocks"
	gitservermocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver/mocks"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	storemocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store/mocks"
)

func TestWindow(t *testing.T) {
	mockStore := storemocks.NewMockStore()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()
	mockBundleClient := bundlemocks.NewMockBundleClient()
	mockGitserverClient := gitservermocks.NewMockClient()

	sourceRanges := []bundles.AggregateCodeIntelligence{
		{
			Range:       bundles.Range{Start: bundles.Position{1, 2}, End: bundles.Position{3, 4}},
			Definitions: []bundles.Location{},
			References:  []bundles.Location{},
			HoverText:   "",
		},
		{
			Range:       bundles.Range{Start: bundles.Position{2, 3}, End: bundles.Position{4, 5}},
			Definitions: []bundles.Location{{Path: "foo.go", Range: bundles.Range{Start: bundles.Position{10, 20}, End: bundles.Position{30, 40}}}},
			References:  []bundles.Location{{Path: "bar.go", Range: bundles.Range{Start: bundles.Position{100, 200}, End: bundles.Position{300, 400}}}},
			HoverText:   "ht2",
		},
		{
			Range:       bundles.Range{Start: bundles.Position{3, 4}, End: bundles.Position{5, 6}},
			Definitions: []bundles.Location{{Path: "bar.go", Range: bundles.Range{Start: bundles.Position{11, 21}, End: bundles.Position{31, 41}}}},
			References:  []bundles.Location{{Path: "foo.go", Range: bundles.Range{Start: bundles.Position{101, 201}, End: bundles.Position{301, 401}}}},
			HoverText:   "ht3",
		},
	}

	setMockStoreGetDumpByID(t, mockStore, map[int]store.Dump{42: testDump1})
	setMockBundleManagerClientBundleClient(t, mockBundleManagerClient, map[int]bundles.BundleClient{42: mockBundleClient})
	setMockBundleClientWindow(t, mockBundleClient, "main.go", 10, 20, sourceRanges)

	api := testAPI(mockStore, mockBundleManagerClient, mockGitserverClient)
	ranges, err := api.Window(context.Background(), "sub1/main.go", 10, 20, 42)
	if err != nil {
		t.Fatalf("expected error getting window: %s", err)
	}

	expectedRanges := []ResolvedAggregateCodeIntelligence{
		{
			Range:       bundles.Range{Start: bundles.Position{1, 2}, End: bundles.Position{3, 4}},
			Definitions: nil,
			References:  nil,
			HoverText:   "",
		},
		{
			Range:       bundles.Range{Start: bundles.Position{2, 3}, End: bundles.Position{4, 5}},
			Definitions: []ResolvedLocation{{Dump: testDump1, Path: "sub1/foo.go", Range: bundles.Range{Start: bundles.Position{10, 20}, End: bundles.Position{30, 40}}}},
			References:  []ResolvedLocation{{Dump: testDump1, Path: "sub1/bar.go", Range: bundles.Range{Start: bundles.Position{100, 200}, End: bundles.Position{300, 400}}}},
			HoverText:   "ht2",
		},
		{
			Range:       bundles.Range{Start: bundles.Position{3, 4}, End: bundles.Position{5, 6}},
			Definitions: []ResolvedLocation{{Dump: testDump1, Path: "sub1/bar.go", Range: bundles.Range{Start: bundles.Position{11, 21}, End: bundles.Position{31, 41}}}},
			References:  []ResolvedLocation{{Dump: testDump1, Path: "sub1/foo.go", Range: bundles.Range{Start: bundles.Position{101, 201}, End: bundles.Position{301, 401}}}},
			HoverText:   "ht3",
		},
	}
	if diff := cmp.Diff(expectedRanges, ranges); diff != "" {
		t.Errorf("unexpected range (-want +got):\n%s", diff)
	}
}

func TestWindowUnknownDump(t *testing.T) {
	mockStore := storemocks.NewMockStore()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()
	mockGitserverClient := gitservermocks.NewMockClient()
	setMockStoreGetDumpByID(t, mockStore, nil)

	api := testAPI(mockStore, mockBundleManagerClient, mockGitserverClient)
	if _, err := api.Window(context.Background(), "sub1", 42, 0, 10); err != ErrMissingDump {
		t.Fatalf("unexpected error getting window. want=%q have=%q", ErrMissingDump, err)
	}
}
