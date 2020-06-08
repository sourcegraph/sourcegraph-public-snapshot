package api

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
	bundles "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
	bundlemocks "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client/mocks"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	dbmocks "github.com/sourcegraph/sourcegraph/internal/codeintel/db/mocks"
	gitservermocks "github.com/sourcegraph/sourcegraph/internal/codeintel/gitserver/mocks"
)

func TestDiagnostics(t *testing.T) {
	mockDB := dbmocks.NewMockDB()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()
	mockBundleClient := bundlemocks.NewMockBundleClient()
	mockGitserverClient := gitservermocks.NewMockClient()

	expectedDiagnostics := []client.Diagnostic{
		{
			Path:           "internal/foo.go",
			Severity:       1,
			Code:           "c1",
			Message:        "m1",
			Source:         "s1",
			StartLine:      11,
			StartCharacter: 12,
			EndLine:        13,
			EndCharacter:   14,
		},
		{
			Path:           "internal/bar.go",
			Severity:       2,
			Code:           "c2",
			Message:        "m2",
			Source:         "s2",
			StartLine:      21,
			StartCharacter: 22,
			EndLine:        23,
			EndCharacter:   24,
		},
		{
			Path:           "internal/baz.go",
			Severity:       3,
			Code:           "c3",
			Message:        "m3",
			Source:         "s3",
			StartLine:      31,
			StartCharacter: 32,
			EndLine:        33,
			EndCharacter:   34,
		},
	}

	setMockDBGetDumpByID(t, mockDB, map[int]db.Dump{42: testDump1})
	setMockBundleManagerClientBundleClient(t, mockBundleManagerClient, map[int]bundles.BundleClient{42: mockBundleClient})
	setMockBundleClientDiagnostics(t, mockBundleClient, "sub1", expectedDiagnostics)

	api := testAPI(mockDB, mockBundleManagerClient, mockGitserverClient)
	diagnostics, err := api.Diagnostics(context.Background(), "sub1", 42)
	if err != nil {
		t.Fatalf("expected error getting diagnostics: %s", err)
	}

	if diff := cmp.Diff(expectedDiagnostics, diagnostics); diff != "" {
		t.Errorf("unexpected range (-want +got):\n%s", diff)
	}
}

func TestDiagnosticsUnknownDump(t *testing.T) {
	mockDB := dbmocks.NewMockDB()
	mockBundleManagerClient := bundlemocks.NewMockBundleManagerClient()
	mockGitserverClient := gitservermocks.NewMockClient()
	setMockDBGetDumpByID(t, mockDB, nil)

	api := testAPI(mockDB, mockBundleManagerClient, mockGitserverClient)
	if _, err := api.Diagnostics(context.Background(), "sub1", 42); err != ErrMissingDump {
		t.Fatalf("unexpected error getting diagnostics. want=%q have=%q", ErrMissingDump, err)
	}
}
