package resolvers

import (
	"context"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/codenav/shared"
	store "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore"
)

// filterUploadsWithCommits removes the uploads for commits which are unknown to gitserver from the given
// slice. The slice is filtered in-place and returned (to update the slice length).
func filterUploadsWithCommits(ctx context.Context, cachedCommitChecker *cachedCommitChecker, uploads []store.Dump) ([]store.Dump, error) {
	rcs := make([]gitserver.RepositoryCommit, 0, len(uploads))
	for _, upload := range uploads {
		rcs = append(rcs, gitserver.RepositoryCommit{
			RepositoryID: upload.RepositoryID,
			Commit:       upload.Commit,
		})
	}
	exists, err := cachedCommitChecker.existsBatch(ctx, rcs)
	if err != nil {
		return nil, err
	}

	filtered := uploads[:0]
	for i, upload := range uploads {
		if exists[i] {
			filtered = append(filtered, upload)
		}
	}

	return filtered, nil
}

func uploadIDsToString(vs []store.Dump) string {
	ids := make([]string, 0, len(vs))
	for _, v := range vs {
		ids = append(ids, strconv.Itoa(v.ID))
	}

	return strings.Join(ids, ", ")
}

func sharedRangeTolsifstoreRange(r shared.Range) lsifstore.Range {
	return lsifstore.Range{
		Start: lsifstore.Position(r.Start),
		End:   lsifstore.Position(r.End),
	}
}
