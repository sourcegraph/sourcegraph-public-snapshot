package api

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestDefinitions(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockGitserverClient := NewMockGitserverClient()

	setMockDBStoreGetDumpByID(t, mockDBStore, map[int]store.Dump{42: testDump1})
	setmockLSIFStoreDefinitions(t, mockLSIFStore, 42, "main.go", 10, 50, []lsifstore.Location{
		{DumpID: 42, Path: "foo.go", Range: testRange1},
		{DumpID: 42, Path: "bar.go", Range: testRange2},
		{DumpID: 42, Path: "baz.go", Range: testRange3},
	})

	api := New(mockDBStore, mockLSIFStore, mockGitserverClient, &observation.TestContext)
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
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockGitserverClient := NewMockGitserverClient()
	setMockDBStoreGetDumpByID(t, mockDBStore, nil)

	api := New(mockDBStore, mockLSIFStore, mockGitserverClient, &observation.TestContext)
	if _, err := api.Definitions(context.Background(), "sub1/main.go", 10, 50, 25); err != ErrMissingDump {
		t.Fatalf("unexpected error getting definitions. want=%q have=%q", ErrMissingDump, err)
	}
}

func TestDefinitionViaSameDumpMoniker(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockGitserverClient := NewMockGitserverClient()

	setMockDBStoreGetDumpByID(t, mockDBStore, map[int]store.Dump{42: testDump1})
	setmockLSIFStoreDefinitions(t, mockLSIFStore, 42, "main.go", 10, 50, nil)
	setmockLSIFStoreMonikersByPosition(t, mockLSIFStore, 42, "main.go", 10, 50, [][]lsifstore.MonikerData{{testMoniker2}})
	setmockLSIFStoreMonikerResults(t, mockLSIFStore, 42, "definitions", "gomod", "pad", 0, 100, []lsifstore.Location{
		{DumpID: 42, Path: "foo.go", Range: testRange1},
		{DumpID: 42, Path: "bar.go", Range: testRange2},
		{DumpID: 42, Path: "baz.go", Range: testRange3},
	}, 3)

	api := New(mockDBStore, mockLSIFStore, mockGitserverClient, &observation.TestContext)
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
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockGitserverClient := NewMockGitserverClient()

	setMockDBStoreGetDumpByID(t, mockDBStore, map[int]store.Dump{42: testDump1, 50: testDump2})
	setmockLSIFStoreDefinitions(t, mockLSIFStore, 42, "main.go", 10, 50, nil)
	setmockLSIFStoreMonikersByPosition(t, mockLSIFStore, 42, "main.go", 10, 50, [][]lsifstore.MonikerData{{testMoniker1}})
	setmockLSIFStorePackageInformation(t, mockLSIFStore, 42, "main.go", "1234", testPackageInformation)
	setMockDBStoreGetPackage(t, mockDBStore, "gomod", "leftpad", "0.1.0", testDump2, true)
	setmockLSIFStoreMonikerResults(t, mockLSIFStore, 50, "definitions", "gomod", "pad", 0, 100, []lsifstore.Location{
		{DumpID: 50, Path: "foo.go", Range: testRange1},
		{DumpID: 50, Path: "bar.go", Range: testRange2},
		{DumpID: 50, Path: "baz.go", Range: testRange3},
	}, 15)

	api := New(mockDBStore, mockLSIFStore, mockGitserverClient, &observation.TestContext)
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
