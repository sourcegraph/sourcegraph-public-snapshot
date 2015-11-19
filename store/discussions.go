package store

import (
	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

type Discussions interface {
	Create(ctx context.Context, d *sourcegraph.Discussion) error
	Get(ctx context.Context, repo sourcegraph.RepoSpec, id int64) (*sourcegraph.Discussion, error)
	List(ctx context.Context, in *sourcegraph.DiscussionListOp) (*sourcegraph.DiscussionList, error)
	CreateComment(ctx context.Context, discussionID int64, comment *sourcegraph.DiscussionComment) error
}
