package local

import (
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/store"
)

func (s *builds) GetRepoBuild(ctx context.Context, rev *sourcegraph.RepoRevSpec) (*sourcegraph.Build, error) {
	wasAbs := isAbsCommitID(rev.CommitID) // cache if request was for absolute commit ID

	if err := (&repos{}).resolveRepoRev(ctx, rev); err != nil {
		return nil, err
	}

	build, _, err := store.BuildsFromContext(ctx).GetFirstInCommitOrder(ctx, rev.URI, []string{rev.CommitID}, false)
	if err != nil {
		return nil, err
	}

	if wasAbs {
		veryShortCache(ctx)
	}
	return build, nil
}
