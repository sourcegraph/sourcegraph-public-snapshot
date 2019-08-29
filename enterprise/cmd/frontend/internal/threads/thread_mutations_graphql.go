package threads

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
)

func (GraphQLResolver) CreateThread(ctx context.Context, arg *graphqlbackend.CreateThreadArgs) (graphqlbackend.Thread, error) {
	repo, err := graphqlbackend.RepositoryByID(ctx, arg.Input.Repository)
	if err != nil {
		return nil, err
	}

	author, err := comments.CommentActorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	comment := commentobjectdb.DBObjectCommentFields{Author: author}
	if arg.Input.Body != nil {
		comment.Body = *arg.Input.Body
	}

	data := &DBThread{
		RepositoryID: repo.DBID(),
		Title:        arg.Input.Title,
		IsDraft:      arg.Input.Draft != nil && *arg.Input.Draft,
		State:        string(graphqlbackend.ThreadStateOpen),
	}
	if arg.Input.BaseRef != nil {
		data.BaseRef = *arg.Input.BaseRef
	}
	if arg.Input.HeadRef != nil {
		data.HeadRef = *arg.Input.HeadRef
	}
	thread, err := dbThreads{}.Create(ctx, nil, data, comment)
	if err != nil {
		return nil, err
	}
	gqlThread := newGQLThread(thread)

	if arg.Input.RawDiagnostics != nil {
		if _, err := graphqlbackend.ThreadDiagnostics.AddDiagnosticsToThread(ctx, &graphqlbackend.AddDiagnosticsToThreadArgs{Thread: gqlThread.ID(), RawDiagnostics: *arg.Input.RawDiagnostics}); err != nil {
			return nil, err
		}
	}

	return gqlThread, nil
}

func (GraphQLResolver) UpdateThread(ctx context.Context, arg *graphqlbackend.UpdateThreadArgs) (graphqlbackend.Thread, error) {
	l, err := threadByID(ctx, arg.Input.ID)
	if err != nil {
		return nil, err
	}
	thread, err := dbThreads{}.Update(ctx, l.db.ID, dbThreadUpdate{
		Title: arg.Input.Title,
		// TODO!(sqs): handle body update
		BaseRef: arg.Input.BaseRef,
		HeadRef: arg.Input.HeadRef,
	})
	if err != nil {
		return nil, err
	}
	return newGQLThread(thread), nil
}

func (GraphQLResolver) PublishDraftThread(ctx context.Context, arg *graphqlbackend.PublishDraftThreadArgs) (graphqlbackend.Thread, error) {
	l, err := threadByID(ctx, arg.Thread)
	if err != nil {
		return nil, err
	}
	if !l.db.IsDraft {
		return nil, errors.New("thread is not a draft thread")
	}
	tmp := false
	thread, err := dbThreads{}.Update(ctx, l.db.ID, dbThreadUpdate{
		IsDraft: &tmp,
	})
	if err != nil {
		return nil, err
	}
	return newGQLThread(thread), nil
}

func (GraphQLResolver) DeleteThread(ctx context.Context, arg *graphqlbackend.DeleteThreadArgs) (*graphqlbackend.EmptyResponse, error) {
	gqlThread, err := threadByID(ctx, arg.Thread)
	if err != nil {
		return nil, err
	}
	return nil, dbThreads{}.DeleteByID(ctx, gqlThread.db.ID)
}
