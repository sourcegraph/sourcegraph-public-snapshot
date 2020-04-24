package api

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	bundles "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-api-server/internal/mocks"
)

func TestLookupMoniker(t *testing.T) {
	mockDB := mocks.NewMockDB()
	mockBundleManagerClient := mocks.NewMockBundleManagerClient()
	mockBundleClient1 := mocks.NewMockBundleClient()
	mockBundleClient2 := mocks.NewMockBundleClient()

	setMockBundleManagerClientBundleClient(t, mockBundleManagerClient, map[int]bundles.BundleClient{42: mockBundleClient1, 50: mockBundleClient2})
	setMockBundleClientPackageInformation(t, mockBundleClient1, "sub2/main.go", "1234", testPackageInformation)
	setMockDBGetPackage(t, mockDB, "gomod", "leftpad", "0.1.0", testDump2, true)
	setMockBundleClientMonikerResults(t, mockBundleClient2, "definitions", "gomod", "pad", 10, 5, []bundles.Location{
		{DumpID: 42, Path: "foo.go", Range: testRange1},
		{DumpID: 42, Path: "bar.go", Range: testRange2},
		{DumpID: 42, Path: "baz.go", Range: testRange3},
		{DumpID: 42, Path: "bar.go", Range: testRange4},
		{DumpID: 42, Path: "baz.go", Range: testRange5},
	}, 15)

	locations, totalCount, err := lookupMoniker(mockDB, mockBundleManagerClient, 42, "sub2/main.go", "definitions", testMoniker2, 10, 5)
	if err != nil {
		t.Fatalf("unexpected error querying moniker: %s", err)
	}
	if totalCount != 15 {
		t.Errorf("unexpected total count. want=%d have=%d", 5, totalCount)
	}

	expectedLocations := []ResolvedLocation{
		{Dump: testDump2, Path: "sub2/foo.go", Range: testRange1},
		{Dump: testDump2, Path: "sub2/bar.go", Range: testRange2},
		{Dump: testDump2, Path: "sub2/baz.go", Range: testRange3},
		{Dump: testDump2, Path: "sub2/bar.go", Range: testRange4},
		{Dump: testDump2, Path: "sub2/baz.go", Range: testRange5},
	}
	if diff := cmp.Diff(expectedLocations, locations); diff != "" {
		t.Errorf("unexpected definitions (-want +got):\n%s", diff)
	}
}

func TestLookupMonikerNoPackageInformationID(t *testing.T) {
	mockDB := mocks.NewMockDB()
	mockBundleManagerClient := mocks.NewMockBundleManagerClient()

	_, totalCount, err := lookupMoniker(mockDB, mockBundleManagerClient, 42, "sub/main.go", "definitions", testMoniker3, 10, 5)
	if err != nil {
		t.Fatalf("unexpected error querying moniker: %s", err)
	}
	if totalCount != 0 {
		t.Errorf("unexpected total count. want=%d have=%d", 0, totalCount)
	}
}

func TestLookupMonikerNoPackage(t *testing.T) {
	mockDB := mocks.NewMockDB()
	mockBundleManagerClient := mocks.NewMockBundleManagerClient()
	mockBundleClient := mocks.NewMockBundleClient()

	setMockBundleManagerClientBundleClient(t, mockBundleManagerClient, map[int]bundles.BundleClient{42: mockBundleClient})
	setMockBundleClientPackageInformation(t, mockBundleClient, "main.go", "1234", testPackageInformation)
	setMockDBGetPackage(t, mockDB, "gomod", "leftpad", "0.1.0", db.Dump{}, false)

	_, totalCount, err := lookupMoniker(mockDB, mockBundleManagerClient, 42, "main.go", "definitions", testMoniker1, 10, 5)
	if err != nil {
		t.Fatalf("unexpected error querying moniker: %s", err)
	}
	if totalCount != 0 {
		t.Errorf("unexpected total count. want=%d have=%d", 0, totalCount)
	}
}
