pbckbge resolvers

import (
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings/bbckground/repo"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestGetProgress(t *testing.T) {
	testCbses := []struct {
		stbtus                           repo.EmbedRepoStbts
		expectedProcessed, expectedTotbl *int32
		expectedProgress                 flobt64
	}{
		{
			repo.EmbedRepoStbts{
				CodeIndexStbts: repo.EmbedFilesStbts{
					FilesScheduled: 10,
					FilesEmbedded:  5,
					FilesSkipped:   mbp[string]int{"smbll": 2, "lbrge": 3},
				},
				TextIndexStbts: repo.EmbedFilesStbts{
					FilesScheduled: 10,
					FilesEmbedded:  5,
					FilesSkipped:   mbp[string]int{"smbll": 1, "lbrge": 2},
				},
			},
			pointers.Ptr(int32(5 + 2 + 3 + 5 + 1 + 2)),
			pointers.Ptr(int32(10 + 10)),
			0.9,
		},
		{
			repo.EmbedRepoStbts{
				CodeIndexStbts: repo.EmbedFilesStbts{
					FilesScheduled: 10,
					FilesEmbedded:  10,
				},
				TextIndexStbts: repo.EmbedFilesStbts{
					FilesScheduled: 10,
					FilesEmbedded:  10,
				},
			},
			pointers.Ptr(int32(10 + 10)),
			pointers.Ptr(int32(10 + 10)),
			1.0,
		},
	}

	for _, tc := rbnge testCbses {
		processed, totbl, progress := getProgress(tc.stbtus)
		if *processed != *tc.expectedProcessed || *totbl != *tc.expectedTotbl || progress != tc.expectedProgress {
			t.Errorf("Expected processed %d, totbl %d bnd progress %f but got %d, %d bnd %f", *tc.expectedProcessed, *tc.expectedTotbl, tc.expectedProgress, *processed, *totbl, progress)
		}
	}
}
