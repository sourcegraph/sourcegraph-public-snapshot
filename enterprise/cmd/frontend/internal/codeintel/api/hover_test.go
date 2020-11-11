package api

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func TestHover(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockGitserverClient := NewMockGitserverClient()

	setMockDBStoreGetDumpByID(t, mockDBStore, map[int]store.Dump{42: testDump1})
	setmockLSIFStoreHover(t, mockLSIFStore, 42, "main.go", 10, 50, "text", testRange1, true)

	api := New(mockDBStore, mockLSIFStore, mockGitserverClient, &observation.TestContext)
	text, r, exists, err := api.Hover(context.Background(), "sub1/main.go", 10, 50, 42)
	if err != nil {
		t.Fatalf("expected error getting hover text: %s", err)
	}
	if !exists {
		t.Fatalf("expected hover text to exist.")
	}

	if text != "text" {
		t.Errorf("unexpected text. want=%s have=%s", "text", text)
	}
	if diff := cmp.Diff(testRange1, r); diff != "" {
		t.Errorf("unexpected range (-want +got):\n%s", diff)
	}
}

func TestHoverUnknownDump(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockGitserverClient := NewMockGitserverClient()
	setMockDBStoreGetDumpByID(t, mockDBStore, nil)

	api := New(mockDBStore, mockLSIFStore, mockGitserverClient, &observation.TestContext)
	if _, _, _, err := api.Hover(context.Background(), "sub1/main.go", 10, 50, 42); err != ErrMissingDump {
		t.Fatalf("unexpected error getting hover text. want=%q have=%q", ErrMissingDump, err)
	}
}

func TestHoverRemoteDefinitionHoverText(t *testing.T) {
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
	setMultimockLSIFStoreHover(
		t,
		mockLSIFStore,
		hoverSpec{42, "main.go", 10, 50, "", lsifstore.Range{}, false},
		hoverSpec{50, "foo.go", 10, 50, "text", testRange4, true},
	)

	api := New(mockDBStore, mockLSIFStore, mockGitserverClient, &observation.TestContext)
	text, r, exists, err := api.Hover(context.Background(), "sub1/main.go", 10, 50, 42)
	if err != nil {
		t.Fatalf("expected error getting hover text: %s", err)
	}
	if !exists {
		t.Fatalf("expected hover text to exist.")
	}

	if text != "text" {
		t.Errorf("unexpected text. want=%s have=%s", "text", text)
	}
	if diff := cmp.Diff(testRange4, r); diff != "" {
		t.Errorf("unexpected range (-want +got):\n%s", diff)
	}
}

func TestHoverUnknownDefinition(t *testing.T) {
	mockDBStore := NewMockDBStore()
	mockLSIFStore := NewMockLSIFStore()
	mockGitserverClient := NewMockGitserverClient()

	setMockDBStoreGetDumpByID(t, mockDBStore, map[int]store.Dump{42: testDump1})
	setmockLSIFStoreHover(t, mockLSIFStore, 42, "main.go", 10, 50, "", lsifstore.Range{}, false)
	setmockLSIFStoreDefinitions(t, mockLSIFStore, 42, "main.go", 10, 50, nil)
	setmockLSIFStoreMonikersByPosition(t, mockLSIFStore, 42, "main.go", 10, 50, [][]lsifstore.MonikerData{{testMoniker1}})
	setmockLSIFStorePackageInformation(t, mockLSIFStore, 42, "main.go", "1234", testPackageInformation)
	setMockDBStoreGetPackage(t, mockDBStore, "gomod", "leftpad", "0.1.0", store.Dump{}, false)

	api := New(mockDBStore, mockLSIFStore, mockGitserverClient, &observation.TestContext)
	_, _, exists, err := api.Hover(context.Background(), "sub1/main.go", 10, 50, 42)
	if err != nil {
		t.Fatalf("unexpected error getting hover text: %s", err)
	}
	if exists {
		t.Errorf("unexpected hover text")
	}
}
