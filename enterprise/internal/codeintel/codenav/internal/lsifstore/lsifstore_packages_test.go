package lsifstore

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
)

func TestGetPackageInformation(t *testing.T) {
	testCases := []struct {
		name                 string
		uploadID             int
		path                 string
		packageInformationID string
		expectedData         precise.PackageInformationData
	}{
		{
			name:                 "scip",
			uploadID:             testSCIPUploadID,
			path:                 "protocol/protocol.go",
			packageInformationID: "scip:dGVzdC1tYW5hZ2Vy:dGVzdC1uYW1l:dGVzdC12ZXJzaW9u",
			expectedData: precise.PackageInformationData{
				Manager: "test-manager",
				Name:    "test-name",
				Version: "test-version",
			},
		},
	}

	store := populateTestStore(t)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if actual, exists, err := store.GetPackageInformation(context.Background(), testCase.uploadID, testCase.path, testCase.packageInformationID); err != nil {
				t.Fatalf("unexpected error %s", err)
			} else if !exists {
				t.Errorf("no package information")
			} else {
				if diff := cmp.Diff(testCase.expectedData, actual); diff != "" {
					t.Errorf("unexpected package information (-want +got):\n%s", diff)
				}
			}
		})
	}
}
