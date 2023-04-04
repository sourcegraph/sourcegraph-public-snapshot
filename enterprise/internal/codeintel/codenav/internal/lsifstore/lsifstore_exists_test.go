package lsifstore

import (
	"context"
	"testing"
)

func TestDatabaseExists(t *testing.T) {
	store := populateTestStore(t)

	testCases := []struct {
		uploadID int
		path     string
		expected bool
	}{
		// SCIP
		{testSCIPUploadID, "template/src/lsif/api.ts", true},
		{testSCIPUploadID, "template/src/lsif/util.ts", true},
		{testSCIPUploadID, "missing.ts", false},
	}

	for _, testCase := range testCases {
		if exists, err := store.GetPathExists(context.Background(), testCase.uploadID, testCase.path); err != nil {
			t.Fatalf("unexpected error %s", err)
		} else if exists != testCase.expected {
			t.Errorf("unexpected exists result for %s. want=%v have=%v", testCase.path, testCase.expected, exists)
		}
	}
}
