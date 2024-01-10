package graphqlbackend

import (
	"context"
	"io"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitTreeEntryBlameArgs struct {
	StartLine        int32
	EndLine          int32
	IgnoreWhitespace bool
}

func (r *GitTreeEntryResolver) Blame(ctx context.Context, args *GitTreeEntryBlameArgs) ([]*hunkResolver, error) {
	hr, err := r.gitserverClient.StreamBlameFile(ctx, r.commit.repoResolver.RepoName(), r.Path(), &gitserver.BlameOptions{
		NewestCommit:     api.CommitID(r.commit.OID()),
		StartLine:        int(args.StartLine),
		EndLine:          int(args.EndLine),
		IgnoreWhitespace: args.IgnoreWhitespace,
	})
	if err != nil {
		return nil, err
	}
	defer hr.Close()

	hunkResolvers := []*hunkResolver{}
	for {
		hunk, err := hr.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return hunkResolvers, nil
			}
			return nil, err
		}
		hunkResolvers = append(hunkResolvers, &hunkResolver{
			db:   r.db,
			repo: r.commit.repoResolver,
			hunk: hunk,
		})
	}
}
