package lsifstore

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
)

func TestDatabaseFindDocumentIDs(t *testing.T) {
	store := populateTestStore(t)
	const nonExistentUploadID = math.MaxInt32

	testCases := []struct {
		uploadID int
		path     string
		expected bool
	}{
		{testSCIPUploadID, "template/src/lsif/api.ts", true},
		{testSCIPUploadID, "template/src/lsif/util.ts", true},
		{testSCIPUploadID, "missing.ts", false},
		{nonExistentUploadID, "template/src/lsif/util.ts", false},
	}

	for _, tc := range testCases {
		// TODO: Ideally, we would do some bulk tests here, but somewhat blocked on:
		// https://linear.app/sourcegraph/issue/GRAPH-707/make-lsifstore-tests-more-easily-customizable
		input := map[int]core.UploadRelPath{tc.uploadID: core.NewUploadRelPathUnchecked(tc.path)}
		docIDsMap, err := store.FindDocumentIDs(context.Background(), input)
		require.NoError(t, err)
		_, found := docIDsMap[tc.uploadID]
		require.Equalf(t, tc.expected, found, "path: %v", tc.path)
	}
}
