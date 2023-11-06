package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

func (r *GitTreeEntryResolver) Blame(ctx context.Context,
	args *struct {
		StartLine        int32
		EndLine          int32
		IgnoreWhitespace *bool
	}) ([]*hunkResolver, error) {
	hunks, err := r.gitserverClient.BlameFile(ctx, r.commit.repoResolver.RepoName(), r.Path(), &gitserver.BlameOptions{
		NewestCommit:     api.CommitID(r.commit.OID()),
		StartLine:        int(args.StartLine),
		EndLine:          int(args.EndLine),
		IgnoreWhitespace: pointers.Deref(args.IgnoreWhitespace, false),
	})
	if err != nil {
		return nil, err
	}

	hunkResolvers := make([]*hunkResolver, 0, len(hunks))
	for _, hunk := range hunks {
		hunkResolvers = append(hunkResolvers, &hunkResolver{
			db:   r.db,
			repo: r.commit.repoResolver,
			hunk: hunk,
		})
	}

	return hunkResolvers, nil
}
