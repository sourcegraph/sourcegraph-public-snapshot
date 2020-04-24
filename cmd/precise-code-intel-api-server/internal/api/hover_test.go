package api

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-api-server/internal/mocks"
	bundles "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
)

func TestHover(t *testing.T) {
	mockDB := mocks.NewMockDB()
	mockBundleManagerClient := mocks.NewMockBundleManagerClient()
	mockBundleClient := mocks.NewMockBundleClient()

	setMockDBGetDumpByID(t, mockDB, map[int]db.Dump{42: testDump1})
	setMockBundleManagerClientBundleClient(t, mockBundleManagerClient, map[int]bundles.BundleClient{42: mockBundleClient})
	setMockBundleClientHover(t, mockBundleClient, "main.go", 10, 50, "text", testRange1, true)

	api := New(mockDB, mockBundleManagerClient)
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
	mockDB := mocks.NewMockDB()
	mockBundleManagerClient := mocks.NewMockBundleManagerClient()
	setMockDBGetDumpByID(t, mockDB, nil)

	api := New(mockDB, mockBundleManagerClient)
	if _, _, _, err := api.Hover(context.Background(), "sub1/main.go", 10, 50, 42); err != ErrMissingDump {
		t.Fatalf("unexpected error getting hover text. want=%q have=%q", ErrMissingDump, err)
	}
}

func TestHoverRemoteDefinitionHoverText(t *testing.T) {
	mockDB := mocks.NewMockDB()
	mockBundleManagerClient := mocks.NewMockBundleManagerClient()
	mockBundleClient1 := mocks.NewMockBundleClient()
	mockBundleClient2 := mocks.NewMockBundleClient()

	setMockDBGetDumpByID(t, mockDB, map[int]db.Dump{42: testDump1, 50: testDump2})
	setMockBundleManagerClientBundleClient(t, mockBundleManagerClient, map[int]bundles.BundleClient{42: mockBundleClient1, 50: mockBundleClient2})
	setMockBundleClientHover(t, mockBundleClient1, "main.go", 10, 50, "", bundles.Range{}, false)
	setMockBundleClientDefinitions(t, mockBundleClient1, "main.go", 10, 50, nil)
	setMockBundleClientMonikersByPosition(t, mockBundleClient1, "main.go", 10, 50, [][]bundles.MonikerData{{testMoniker1}})
	setMockBundleClientPackageInformation(t, mockBundleClient1, "main.go", "1234", testPackageInformation)
	setMockDBGetPackage(t, mockDB, "gomod", "leftpad", "0.1.0", testDump2, true)
	setMockBundleClientMonikerResults(t, mockBundleClient2, "definitions", "gomod", "pad", 0, 0, []bundles.Location{
		{DumpID: 50, Path: "foo.go", Range: testRange1},
		{DumpID: 50, Path: "bar.go", Range: testRange2},
		{DumpID: 50, Path: "baz.go", Range: testRange3},
	}, 15)
	setMockBundleClientHover(t, mockBundleClient2, "foo.go", 10, 50, "text", testRange4, true)

	api := New(mockDB, mockBundleManagerClient)
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
	mockDB := mocks.NewMockDB()
	mockBundleManagerClient := mocks.NewMockBundleManagerClient()
	mockBundleClient := mocks.NewMockBundleClient()

	setMockDBGetDumpByID(t, mockDB, map[int]db.Dump{42: testDump1})
	setMockBundleManagerClientBundleClient(t, mockBundleManagerClient, map[int]bundles.BundleClient{42: mockBundleClient})
	setMockBundleClientHover(t, mockBundleClient, "main.go", 10, 50, "", bundles.Range{}, false)
	setMockBundleClientDefinitions(t, mockBundleClient, "main.go", 10, 50, nil)
	setMockBundleClientMonikersByPosition(t, mockBundleClient, "main.go", 10, 50, [][]bundles.MonikerData{{testMoniker1}})
	setMockBundleClientPackageInformation(t, mockBundleClient, "main.go", "1234", testPackageInformation)
	setMockDBGetPackage(t, mockDB, "gomod", "leftpad", "0.1.0", db.Dump{}, false)

	api := New(mockDB, mockBundleManagerClient)
	_, _, exists, err := api.Hover(context.Background(), "sub1/main.go", 10, 50, 42)
	if err != nil {
		t.Fatalf("unexpected error getting hover text: %s", err)
	}
	if exists {
		t.Errorf("unexpected hover text")
	}
}
