package changesets

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/internal"
)

func (GraphQLResolver) CreateChangeset(ctx context.Context, arg *graphqlbackend.CreateChangesetArgs) (graphqlbackend.Changeset, error) {
	repo, err := graphqlbackend.RepositoryByID(ctx, arg.Input.Repository)
	if err != nil {
		return nil, err
	}

	changeset, err := internal.DBThreads{}.Create(ctx, &internal.DBThread{
		Type:         graphqlbackend.ThreadlikeTypeChangeset,
		RepositoryID: repo.DBID(),
		Title:        arg.Input.Title,
		ExternalURL:  arg.Input.ExternalURL,
		Status:       string(graphqlbackend.ChangesetStatusOpen),
		IsPreview:    arg.Input.Preview != nil && *arg.Input.Preview,
		BaseRef:      arg.Input.BaseRef,
		HeadRef:      arg.Input.HeadRef,
	})
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
		Title:       arg.Input.Title,
		ExternalURL: arg.Input.ExternalURL,
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
	panic("TODO!(sqs)")
}

func (GraphQLResolver) DeleteChangeset(ctx context.Context, arg *graphqlbackend.DeleteChangesetArgs) (*graphqlbackend.EmptyResponse, error) {
	gqlChangeset, err := changesetByID(ctx, arg.Changeset)
	if err != nil {
		return nil, err
	}
	return nil, internal.DBThreads{}.DeleteByID(ctx, gqlChangeset.db.ID)
}
