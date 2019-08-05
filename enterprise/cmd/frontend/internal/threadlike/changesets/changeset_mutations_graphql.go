package changesets

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/internal"
)

func (GraphQLResolver) CreateChangeset(ctx context.Context, arg *graphqlbackend.CreateChangesetArgs) (graphqlbackend.Changeset, error) {
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

	changeset, err := internal.DBThreads{}.Create(ctx, nil, &internal.DBThread{
		Type:         internal.DBThreadTypeChangeset,
		RepositoryID: repo.DBID(),
		Title:        arg.Input.Title,
		State:        string(graphqlbackend.ChangesetStateOpen),
		IsPreview:    arg.Input.Preview != nil && *arg.Input.Preview,
		BaseRef:      arg.Input.BaseRef,
		HeadRef:      arg.Input.HeadRef,
	}, comment)
	if err != nil {
		return nil, err
	}
	return newGQLChangeset(changeset), nil
}

func (GraphQLResolver) UpdateChangeset(ctx context.Context, arg *graphqlbackend.UpdateChangesetArgs) (graphqlbackend.Changeset, error) {
	l, err := changesetByID(ctx, arg.Input.ID)
	if err != nil {
		return nil, err
	}
	changeset, err := internal.DBThreads{}.Update(ctx, l.db.ID, internal.DBThreadUpdate{
		Title: arg.Input.Title,
		// TODO!(sqs): handle body update
		BaseRef: arg.Input.BaseRef,
		HeadRef: arg.Input.HeadRef,
	})
	if err != nil {
		return nil, err
	}
	return newGQLChangeset(changeset), nil
}

func (GraphQLResolver) PublishPreviewChangeset(ctx context.Context, arg *graphqlbackend.PublishPreviewChangesetArgs) (graphqlbackend.Changeset, error) {
	l, err := changesetByID(ctx, arg.Changeset)
	if err != nil {
		return nil, err
	}

	if !l.IsPreview() {
		return nil, errors.New("changeset has already been published (and is not in preview)")
	}

	v := false
	changeset, err := internal.DBThreads{}.Update(ctx, l.db.ID, internal.DBThreadUpdate{
		IsPreview: &v,
	})
	if err != nil {
		return nil, err
	}
	return newGQLChangeset(changeset), nil
}

func (GraphQLResolver) DeleteChangeset(ctx context.Context, arg *graphqlbackend.DeleteChangesetArgs) (*graphqlbackend.EmptyResponse, error) {
	gqlChangeset, err := changesetByID(ctx, arg.Changeset)
	if err != nil {
		return nil, err
	}
	return nil, internal.DBThreads{}.DeleteByID(ctx, gqlChangeset.db.ID)
}
