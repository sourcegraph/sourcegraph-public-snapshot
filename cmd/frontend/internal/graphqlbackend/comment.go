package graphqlbackend

import (
	"context"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
)

func commentByID(ctx context.Context, id graphql.ID) (*commentResolver, error) {
	commentID, err := unmarshalCommentID(id)
	if err != nil {
		return nil, err
	}
	return commentByIDInt32(ctx, commentID)
}

func commentByIDInt32(ctx context.Context, commentID int32) (*commentResolver, error) {
	comment, err := db.Comments.GetByID(ctx, commentID)
	if err != nil {
		return nil, err
	}

	thread, err := db.Threads.Get(ctx, comment.ThreadID)
	if err != nil {
		return nil, err
	}

	repo, err := db.OrgRepos.GetByID(ctx, thread.OrgRepoID)
	if err != nil {
		return nil, err
	}

	org, err := db.Orgs.GetByID(ctx, repo.OrgID)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check that the current user is a member of the org.
	if err := backend.CheckOrgAccess(ctx, org.ID); err != nil {
		return nil, err
	}

	return &commentResolver{
		comment: comment,
		repo:    repo,
		thread:  thread,
		org:     org,
	}, nil
}

func marshalCommentID(id int32) graphql.ID { return relay.MarshalID("Comment", id) }

func unmarshalCommentID(id graphql.ID) (commentID int32, err error) {
	err = relay.UnmarshalSpec(id, &commentID)
	return
}

func (r *commentResolver) ID() graphql.ID { return relay.MarshalID("Comment", r.comment.ID) }

func (r *commentResolver) DatabaseID() int32 { return r.comment.ID }
