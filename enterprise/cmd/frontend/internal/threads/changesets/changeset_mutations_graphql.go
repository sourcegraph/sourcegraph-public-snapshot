package changesets

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
)

func (GraphQLResolver) CreateChangeset(ctx context.Context, arg *graphqlbackend.CreateChangesetArgs) (graphqlbackend.Changeset, error) {
	repo, err := graphqlbackend.RepositoryByID(ctx, arg.Input.Repository)
	if err != nil {
		return nil, err
	}

	db := &dbChangeset{
		DBThreadCommon: threads.DBThreadCommon{
			RepositoryID: repo.DBID(),
			Title:        arg.Input.Title,
			ExternalURL:  arg.Input.ExternalURL,
		},
		Status:    graphqlbackend.ChangesetStatusOpen,
		IsPreview: arg.Input.Preview == nil || *arg.Input.Preview,
		BaseRef:   arg.Input.BaseRef,
		HeadRef:   arg.Input.HeadRef,
	}

	changeset, err := dbChangesets{}.Create(ctx, db)
	if err != nil {
		return nil, err
	}
	return &gqlChangeset{db: changeset}, nil
}

func (GraphQLResolver) UpdateChangeset(ctx context.Context, arg *graphqlbackend.UpdateChangesetArgs) (graphqlbackend.Changeset, error) {
	update := dbChangesetUpdate{
		Title:       arg.Input.Title,
		ExternalURL: arg.Input.ExternalURL,
		Status:      arg.Input.Status,
	}
	if update.Status != nil && !graphqlbackend.IsValidChangesetStatus(string(*update.Status)) {
		return nil, errors.New("invalid changeset status")
	}

	l, err := changesetByID(ctx, arg.Input.ID)
	if err != nil {
		return nil, err
	}
	changeset, err := dbChangesets{}.Update(ctx, l.db.ID, update)
	if err != nil {
		return nil, err
	}
	return &gqlChangeset{db: changeset}, nil
}

func (GraphQLResolver) DeleteChangeset(ctx context.Context, arg *graphqlbackend.DeleteChangesetArgs) (*graphqlbackend.EmptyResponse, error) {
	gqlChangeset, err := changesetByID(ctx, arg.Changeset)
	if err != nil {
		return nil, err
	}
	return nil, dbChangesets{}.DeleteByID(ctx, gqlChangeset.db.ID)
}
