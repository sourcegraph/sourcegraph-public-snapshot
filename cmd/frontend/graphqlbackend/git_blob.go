package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func (r *GitTreeEntryResolver) Blame(ctx context.Context,
	args *struct {
		StartLine int32
		EndLine   int32
	}) ([]*hunkResolver, error) {
	hunks, err := git.BlameFile(ctx, r.commit.repoResolver.name, r.Path(), &git.BlameOptions{
		NewestCommit: api.CommitID(r.commit.OID()),
		StartLine:    int(args.StartLine),
		EndLine:      int(args.EndLine),
	})
	if err != nil {
		return nil, err
	}

	var hunksResolver []*hunkResolver
	for _, hunk := range hunks {
		hunksResolver = append(hunksResolver, &hunkResolver{
			db:   r.db,
			repo: r.commit.repoResolver,
			hunk: hunk,
		})
	}

	return hunksResolver, nil
}
