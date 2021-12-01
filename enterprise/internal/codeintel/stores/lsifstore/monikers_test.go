package lsifstore

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

func TestDatabaseMonikersByPosition(t *testing.T) {
	store := populateTestStore(t)

	// `func NewMetaData(id, root string, info ToolInfo) *MetaData {`
	//       ^^^^^^^^^^^

	if actual, err := store.MonikersByPosition(context.Background(), testBundleID, "protocol/protocol.go", 92, 10); err != nil {
		t.Fatalf("unexpected error %s", err)
	} else {
		expected := [][]precise.MonikerData{
			{
				{
					Kind:                 "export",
					Scheme:               "gomod",
					Identifier:           "github.com/sourcegraph/lsif-go/protocol:NewMetaData",
					PackageInformationID: "114",
				},
			},
		}

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("unexpected moniker result (-want +got):\n%s", diff)
		}
	}
}

func TestDatabaseBulkMonikerResults(t *testing.T) {
	store := populateTestStore(t)

	edgeDefinitionLocations := []Location{
		{DumpID: testBundleID, Path: "protocol/protocol.go", Range: newRange(410, 5, 410, 9)},
	}

	edgeReferenceLocations := []Location{
		{DumpID: testBundleID, Path: "protocol/protocol.go", Range: newRange(410, 5, 410, 9)},
		{DumpID: testBundleID, Path: "protocol/protocol.go", Range: newRange(440, 1, 440, 5)},
		{DumpID: testBundleID, Path: "protocol/protocol.go", Range: newRange(448, 8, 448, 12)},
		{DumpID: testBundleID, Path: "protocol/protocol.go", Range: newRange(462, 1, 462, 5)},
		{DumpID: testBundleID, Path: "protocol/protocol.go", Range: newRange(470, 8, 470, 12)},
	}

	markdownReferenceLocations := []Location{
		{DumpID: testBundleID, Path: "internal/index/helper.go", Range: newRange(78, 6, 78, 16)},
	}

	combinedReferences := append(markdownReferenceLocations, edgeReferenceLocations...)
	edgeMoniker := precise.MonikerData{Scheme: "gomod", Identifier: "github.com/sourcegraph/lsif-go/protocol:Edge"}
	markdownMoniker := precise.MonikerData{Scheme: "gomod", Identifier: "github.com/slimsag/godocmd:ToMarkdown"}

	testCases := []struct {
		tableName          string
		uploadIDs          []int
		monikers           []precise.MonikerData
		limit              int
		offset             int
		expectedLocations  []Location
		expectedTotalCount int
	}{
		// empty cases
		{"definitions", []int{}, []precise.MonikerData{edgeMoniker}, 5, 0, nil, 0},
		{"definitions", []int{testBundleID}, []precise.MonikerData{}, 5, 0, nil, 0},

		// single definitions
		{"definitions", []int{testBundleID}, []precise.MonikerData{edgeMoniker}, 5, 0, edgeDefinitionLocations, 1},
		{"definitions", []int{testBundleID}, []precise.MonikerData{edgeMoniker}, 1, 0, edgeDefinitionLocations[:1], 1},
		{"definitions", []int{testBundleID}, []precise.MonikerData{edgeMoniker}, 5, 1, edgeDefinitionLocations[1:], 1},

		// single references
		{"references", []int{testBundleID}, []precise.MonikerData{edgeMoniker}, 5, 0, edgeReferenceLocations[:5], 19},
		{"references", []int{testBundleID}, []precise.MonikerData{edgeMoniker}, 2, 2, edgeReferenceLocations[2:4], 19},
		{"references", []int{testBundleID}, []precise.MonikerData{markdownMoniker}, 5, 0, markdownReferenceLocations, 1},

		// multiple monikers
		{"references", []int{testBundleID}, []precise.MonikerData{edgeMoniker, markdownMoniker}, 5, 0, combinedReferences[:5], 20},
		{"references", []int{testBundleID}, []precise.MonikerData{edgeMoniker, markdownMoniker}, 5, 1, combinedReferences[1:6], 20},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("i=%d", i), func(t *testing.T) {
			if actual, totalCount, err := store.BulkMonikerResults(
				context.Background(),
				testCase.tableName,
				testCase.uploadIDs,
				testCase.monikers,
				testCase.limit,
				testCase.offset,
			); err != nil {
				t.Fatalf("unexpected error for test case #%d: %s", i, err)
			} else {
				if totalCount != testCase.expectedTotalCount {
					t.Errorf("unexpected moniker result total count for test case #%d. want=%d have=%d", i, testCase.expectedTotalCount, totalCount)
				}

				if diff := cmp.Diff(testCase.expectedLocations, actual); diff != "" {
					t.Errorf("unexpected moniker result locations for test case #%d (-want +got):\n%s", i, diff)
				}
			}
		})
	}
}
