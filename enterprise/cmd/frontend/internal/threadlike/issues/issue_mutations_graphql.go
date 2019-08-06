package issues

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/internal"
)

func (GraphQLResolver) CreateIssue(ctx context.Context, arg *graphqlbackend.CreateIssueArgs) (graphqlbackend.Issue, error) {
	repo, err := graphqlbackend.RepositoryByID(ctx, arg.Input.Repository)
	if err != nil {
		return nil, err
	}

	authorUserID, err := comments.CommentActorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	comment := commentobjectdb.DBObjectCommentFields{AuthorUserID: authorUserID}
	if arg.Input.Body != nil {
		comment.Body = *arg.Input.Body
	}

	issue, err := internal.DBThreads{}.Create(ctx, nil, &internal.DBThread{
		Type:         internal.DBThreadTypeIssue,
		RepositoryID: repo.DBID(),
		Title:        arg.Input.Title,
		State:        string(graphqlbackend.ThreadStateOpen),
		// TODO!(sqs): set diagnostics data
	}, comment)
	if err != nil {
		return nil, err
	}
	return newGQLIssue(issue), nil
}

func (GraphQLResolver) UpdateIssue(ctx context.Context, arg *graphqlbackend.UpdateIssueArgs) (graphqlbackend.Issue, error) {
	l, err := issueByID(ctx, arg.Input.ID)
	if err != nil {
		return nil, err
	}
	issue, err := internal.DBThreads{}.Update(ctx, l.db.ID, internal.DBThreadUpdate{
		Title: arg.Input.Title,
		// TODO!(sqs): handle body update
	})
	if err != nil {
		return nil, err
	}
	return newGQLIssue(issue), nil
}

func (GraphQLResolver) DeleteIssue(ctx context.Context, arg *graphqlbackend.DeleteIssueArgs) (*graphqlbackend.EmptyResponse, error) {
	gqlIssue, err := issueByID(ctx, arg.Issue)
	if err != nil {
		return nil, err
	}
	return nil, internal.DBThreads{}.DeleteByID(ctx, gqlIssue.db.ID)
}
