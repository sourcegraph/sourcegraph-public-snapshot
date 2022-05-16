package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

func (r *GitTreeEntryResolver) Blame(ctx context.Context,
	args *struct {
		StartLine int32
		EndLine   int32
	}) ([]*hunkResolver, error) {
	hunks, err := gitserver.NewClient(r.db).BlameFile(ctx, r.commit.repoResolver.RepoName(), r.Path(), &gitserver.BlameOptions{
		NewestCommit: api.CommitID(r.commit.OID()),
		StartLine:    int(args.StartLine),
		EndLine:      int(args.EndLine),
	}, authz.DefaultSubRepoPermsChecker)
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
