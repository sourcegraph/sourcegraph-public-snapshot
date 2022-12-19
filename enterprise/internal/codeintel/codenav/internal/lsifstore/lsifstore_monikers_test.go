package lsifstore

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/codenav/shared"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

func TestDatabaseMonikersByPosition(t *testing.T) {
	testCases := []struct {
		name      string
		uploadID  int
		path      string
		line      int
		character int
		expected  [][]precise.MonikerData
	}{
		{
			name:     "lsif",
			uploadID: testLSIFUploadID,
			// `func NewMetaData(id, root string, info ToolInfo) *MetaData {`
			//       ^^^^^^^^^^^
			path: "protocol/protocol.go",
			line: 92, character: 10,
			expected: [][]precise.MonikerData{
				{
					{
						Kind:                 "export",
						Scheme:               "gomod",
						Identifier:           "github.com/sourcegraph/lsif-go/protocol:NewMetaData",
						PackageInformationID: "114",
					},
				},
			},
		},
		{
			name:     "scip",
			uploadID: testSCIPUploadID,
			// `    const enabled = sourcegraph.configuration.get().get('codeIntel.lsif') ?? true`
			//                                  ^^^^^^^^^^^^^
			path: "template/src/lsif/providers.ts",
			line: 25, character: 35,
			expected: [][]precise.MonikerData{
				{
					{
						Kind:                 "import",
						Scheme:               "scip-typescript",
						Identifier:           "scip-typescript npm sourcegraph 25.5.0 src/`sourcegraph.d.ts`/`'sourcegraph'`/configuration.",
						PackageInformationID: "scip:bnBt:c291cmNlZ3JhcGg:MjUuNS4w",
					},
				},
			},
		},
	}

	store := populateTestStore(t)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if actual, err := store.GetMonikersByPosition(context.Background(), testCase.uploadID, testCase.path, testCase.line, testCase.character); err != nil {
				t.Fatalf("unexpected error %s", err)
			} else {
				if diff := cmp.Diff(testCase.expected, actual); diff != "" {
					t.Errorf("unexpected moniker result (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestGetBulkMonikerLocations(t *testing.T) {
	tableName := "references"
	uploadIDs := []int{testLSIFUploadID, testSCIPUploadID}
	monikers := []precise.MonikerData{
		{
			Scheme:     "gomod",
			Identifier: "github.com/sourcegraph/lsif-go/protocol:DefinitionResult.Vertex",
		},
		{
			Scheme:     "scip-typescript",
			Identifier: "scip-typescript npm template 0.0.0-DEVELOPMENT src/util/`helpers.ts`/asArray().",
		},
	}

	store := populateTestStore(t)

	locations, totalCount, err := store.GetBulkMonikerLocations(context.Background(), tableName, uploadIDs, monikers, 100, 0)
	if err != nil {
		t.Fatalf("unexpected error querying bulk moniker locations: %s", err)
	}
	if totalCount != 11 {
		t.Fatalf("unexpected total count: want=%d have=%d\n", 11, totalCount)
	}

	expectedLocations := []shared.Location{
		// LSIF results
		{DumpID: testLSIFUploadID, Path: "protocol/protocol.go", Range: newRange(260, 1, 260, 7)},
		{DumpID: testLSIFUploadID, Path: "protocol/protocol.go", Range: newRange(266, 2, 266, 8)},

		// SCIP results
		{DumpID: testSCIPUploadID, Path: "template/src/providers.ts", Range: newRange(10, 9, 10, 16)},
		{DumpID: testSCIPUploadID, Path: "template/src/providers.ts", Range: newRange(186, 43, 186, 50)},
		{DumpID: testSCIPUploadID, Path: "template/src/providers.ts", Range: newRange(296, 34, 296, 41)},
		{DumpID: testSCIPUploadID, Path: "template/src/providers.ts", Range: newRange(324, 38, 324, 45)},
		{DumpID: testSCIPUploadID, Path: "template/src/providers.ts", Range: newRange(384, 30, 384, 37)},
		{DumpID: testSCIPUploadID, Path: "template/src/providers.ts", Range: newRange(415, 8, 415, 15)},
		{DumpID: testSCIPUploadID, Path: "template/src/providers.ts", Range: newRange(420, 27, 420, 34)},
		{DumpID: testSCIPUploadID, Path: "template/src/search/providers.ts", Range: newRange(9, 9, 9, 16)},
		{DumpID: testSCIPUploadID, Path: "template/src/search/providers.ts", Range: newRange(225, 20, 225, 27)},
	}
	if diff := cmp.Diff(expectedLocations, locations); diff != "" {
		t.Errorf("unexpected locations (-want +got):\n%s", diff)
	}
}
