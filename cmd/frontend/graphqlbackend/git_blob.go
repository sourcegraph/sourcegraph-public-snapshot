package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

func (r *gitTreeEntryResolver) Blame(ctx context.Context,
	args *struct {
		StartLine int32
		EndLine   int32
	}) ([]*hunkResolver, error) {
	hunks, err := git.BlameFile(ctx, gitserver.Repo{Name: r.commit.repo.repo.Name}, r.path, &git.BlameOptions{
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
			repo: r.commit.repo,
			hunk: hunk,
		})
	}

	return hunksResolver, nil
}
