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

func TestDefinitions(t *testing.T) {
	mockStore := storemocks.NewMockStore()
	mockBundleStore := bundlemocks.NewMockStore()
	mockGitserverClient := NewMockGitserverClient()

	setMockStoreGetDumpByID(t, mockStore, map[int]store.Dump{42: testDump1})
	setMockBundleStoreDefinitions(t, mockBundleStore, 42, "main.go", 10, 50, []lsifstore.Location{
		{DumpID: 42, Path: "foo.go", Range: testRange1},
		{DumpID: 42, Path: "bar.go", Range: testRange2},
		{DumpID: 42, Path: "baz.go", Range: testRange3},
	})

	api := testAPI(mockStore, mockBundleStore, mockGitserverClient)
	definitions, err := api.Definitions(context.Background(), "sub1/main.go", 10, 50, 42)
	if err != nil {
		t.Fatalf("expected error getting definitions: %s", err)
	}

	expectedDefinitions := []ResolvedLocation{
		{Dump: testDump1, Path: "sub1/foo.go", Range: testRange1},
		{Dump: testDump1, Path: "sub1/bar.go", Range: testRange2},
		{Dump: testDump1, Path: "sub1/baz.go", Range: testRange3},
	}
	if diff := cmp.Diff(expectedDefinitions, definitions); diff != "" {
		t.Errorf("unexpected definitions (-want +got):\n%s", diff)
	}
}

func TestDefinitionsUnknownDump(t *testing.T) {
	mockStore := storemocks.NewMockStore()
	mockBundleStore := bundlemocks.NewMockStore()
	mockGitserverClient := NewMockGitserverClient()
	setMockStoreGetDumpByID(t, mockStore, nil)

	api := testAPI(mockStore, mockBundleStore, mockGitserverClient)
	if _, err := api.Definitions(context.Background(), "sub1/main.go", 10, 50, 25); err != ErrMissingDump {
		t.Fatalf("unexpected error getting definitions. want=%q have=%q", ErrMissingDump, err)
	}
}

func TestDefinitionViaSameDumpMoniker(t *testing.T) {
	mockStore := storemocks.NewMockStore()
	mockBundleStore := bundlemocks.NewMockStore()
	mockGitserverClient := NewMockGitserverClient()

	setMockStoreGetDumpByID(t, mockStore, map[int]store.Dump{42: testDump1})
	setMockBundleStoreDefinitions(t, mockBundleStore, 42, "main.go", 10, 50, nil)
	setMockBundleStoreMonikersByPosition(t, mockBundleStore, 42, "main.go", 10, 50, [][]lsifstore.MonikerData{{testMoniker2}})
	setMockBundleStoreMonikerResults(t, mockBundleStore, 42, "definitions", "gomod", "pad", 0, 100, []lsifstore.Location{
		{DumpID: 42, Path: "foo.go", Range: testRange1},
		{DumpID: 42, Path: "bar.go", Range: testRange2},
		{DumpID: 42, Path: "baz.go", Range: testRange3},
	}, 3)

	api := testAPI(mockStore, mockBundleStore, mockGitserverClient)
	definitions, err := api.Definitions(context.Background(), "sub1/main.go", 10, 50, 42)
	if err != nil {
		t.Fatalf("expected error getting definitions: %s", err)
	}

	expectedDefinitions := []ResolvedLocation{
		{Dump: testDump1, Path: "sub1/foo.go", Range: testRange1},
		{Dump: testDump1, Path: "sub1/bar.go", Range: testRange2},
		{Dump: testDump1, Path: "sub1/baz.go", Range: testRange3},
	}
	if diff := cmp.Diff(expectedDefinitions, definitions); diff != "" {
		t.Errorf("unexpected definitions (-want +got):\n%s", diff)
	}
}

func TestDefinitionViaRemoteDumpMoniker(t *testing.T) {
	mockStore := storemocks.NewMockStore()
	mockBundleStore := bundlemocks.NewMockStore()
	mockGitserverClient := NewMockGitserverClient()

	setMockStoreGetDumpByID(t, mockStore, map[int]store.Dump{42: testDump1, 50: testDump2})
	setMockBundleStoreDefinitions(t, mockBundleStore, 42, "main.go", 10, 50, nil)
	setMockBundleStoreMonikersByPosition(t, mockBundleStore, 42, "main.go", 10, 50, [][]lsifstore.MonikerData{{testMoniker1}})
	setMockBundleStorePackageInformation(t, mockBundleStore, 42, "main.go", "1234", testPackageInformation)
	setMockStoreGetPackage(t, mockStore, "gomod", "leftpad", "0.1.0", testDump2, true)
	setMockBundleStoreMonikerResults(t, mockBundleStore, 50, "definitions", "gomod", "pad", 0, 100, []lsifstore.Location{
		{DumpID: 50, Path: "foo.go", Range: testRange1},
		{DumpID: 50, Path: "bar.go", Range: testRange2},
		{DumpID: 50, Path: "baz.go", Range: testRange3},
	}, 15)

	api := testAPI(mockStore, mockBundleStore, mockGitserverClient)
	definitions, err := api.Definitions(context.Background(), "sub1/main.go", 10, 50, 42)
	if err != nil {
		t.Fatalf("expected error getting definitions: %s", err)
	}

	expectedDefinitions := []ResolvedLocation{
		{Dump: testDump2, Path: "sub2/foo.go", Range: testRange1},
		{Dump: testDump2, Path: "sub2/bar.go", Range: testRange2},
		{Dump: testDump2, Path: "sub2/baz.go", Range: testRange3},
	}
	if diff := cmp.Diff(expectedDefinitions, definitions); diff != "" {
		t.Errorf("unexpected definitions (-want +got):\n%s", diff)
	}
}
