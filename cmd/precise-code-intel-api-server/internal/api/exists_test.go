package api

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	bundles "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-api-server/internal/mocks"
)

func TestFindClosestDatabase(t *testing.T) {
	mockDB := mocks.NewMockDB()
	mockBundleManagerClient := mocks.NewMockBundleManagerClient()
	mockBundleClient1 := mocks.NewMockBundleClient()
	mockBundleClient2 := mocks.NewMockBundleClient()
	mockBundleClient3 := mocks.NewMockBundleClient()
	mockBundleClient4 := mocks.NewMockBundleClient()

	setMockDBFindClosestDumps(t, mockDB, 42, testCommit, "s1/main.go", []db.Dump{
		{ID: 50, Root: "s1/"},
		{ID: 51, Root: "s1/"},
		{ID: 52, Root: "s1/"},
		{ID: 53, Root: "s2/"},
	})
	setMockBundleManagerClientBundleClient(t, mockBundleManagerClient, map[int]bundles.BundleClient{
		50: mockBundleClient1,
		51: mockBundleClient2,
		52: mockBundleClient3,
		53: mockBundleClient4,
	})
	setMockBundleClientExists(t, mockBundleClient1, "main.go", true)
	setMockBundleClientExists(t, mockBundleClient2, "main.go", false)
	setMockBundleClientExists(t, mockBundleClient3, "main.go", true)
	setMockBundleClientExists(t, mockBundleClient4, "s1/main.go", false)

	api := New(mockDB, mockBundleManagerClient)
	dumps, err := api.FindClosestDumps(context.Background(), 42, testCommit, "s1/main.go")
	if err != nil {
		t.Fatalf("unexpected error finding closest database: %s", err)
	}

	expected := []db.Dump{
		{ID: 50, Root: "s1/"},
		{ID: 52, Root: "s1/"},
	}
	if diff := cmp.Diff(expected, dumps); diff != "" {
		t.Errorf("unexpected dumps (-want +got):\n%s", diff)
	}
}
