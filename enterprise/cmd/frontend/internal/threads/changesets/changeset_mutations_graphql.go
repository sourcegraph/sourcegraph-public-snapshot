package changesets

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (GraphQLResolver) CreateChangeset(ctx context.Context, arg *graphqlbackend.CreateChangesetArgs) (graphqlbackend.Changeset, error) {
	repo, err := graphqlbackend.RepositoryByID(ctx, arg.Input.Repository)
	if err != nil {
		return nil, err
	}

	db := &dbChangeset{
		RepositoryID: repo.DBID(),
		Title:        arg.Input.Title,
		ExternalURL:  arg.Input.ExternalURL,
		Type:         arg.Input.Type,
	}
	// Apply default status.
	if arg.Input.Status != nil {
		db.Status = *arg.Input.Status
	} else {
		db.Status = graphqlbackend.ChangesetStatusOpen
	}

	// Validate.
	if !graphqlbackend.IsValidChangesetStatus(string(db.Status)) {
		return nil, errors.New("invalid changeset status")
	}
	if !graphqlbackend.IsValidChangesetType(string(db.Type)) {
		return nil, errors.New("invalid changeset type")
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
