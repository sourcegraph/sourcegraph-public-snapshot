package graphqlbackend

import (
	"context"
	"io/fs"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type HistoryArgs struct {
	graphqlutil.ConnectionArgs
	After *string

	// TODO(@camdencheek): implement follow. Right now, we wouldn't have
	// a way to get an updated filename for commits that traverse renames.
}

func (r *GitTreeEntryResolver) History(ctx context.Context, args HistoryArgs) *fileHistoryConnection {
	return &fileHistoryConnection{
		stat: r.stat,
		commits: r.commit.Ancestors(ctx, &AncestorsArgs{
			ConnectionArgs: args.ConnectionArgs,
			AfterCursor:    args.After,
			Path:           pointers.Ptr(r.Path()),
		}),
	}
}

type fileHistoryConnection struct {
	stat    fs.FileInfo
	commits *gitCommitConnectionResolver
}

func (r *fileHistoryConnection) Nodes(ctx context.Context) ([]*GitTreeEntryResolver, error) {
	commits, err := r.commits.Nodes(ctx)
	if err != nil {
		return nil, err
	}

	treeEntries := make([]*GitTreeEntryResolver, len(commits))
	for i, commit := range commits {
		treeEntries[i] = NewGitTreeEntryResolver(commit.db, commit.gitserverClient, GitTreeEntryResolverOpts{
			Commit: commit,
			Stat:   r.stat,
		})
	}
	return treeEntries, nil
}

func (r *fileHistoryConnection) TotalCount(ctx context.Context) (*int32, error) {
	return r.commits.TotalCount(ctx)
}

func (r *fileHistoryConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return r.commits.PageInfo(ctx)
}
