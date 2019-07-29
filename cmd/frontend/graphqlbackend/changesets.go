package graphqlbackend

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// Changesets is the implementation of the GraphQL API for changesets queries and mutations. If it
// is not set at runtime, a "not implemented" error is returned to API clients who invoke it.
//
// This is contributed by enterprise.
var Changesets ChangesetsResolver

var errChangesetsNotImplemented = errors.New("changesets is not implemented")

// ChangesetByID is called to look up a Changeset given its GraphQL ID.
func ChangesetByID(ctx context.Context, id graphql.ID) (Changeset, error) {
	if Changesets == nil {
		return nil, errors.New("changesets is not implemented")
	}
	return Changesets.ChangesetByID(ctx, id)
}

// ChangesetInRepository returns a specific changeset in the specified repository.
func ChangesetInRepository(ctx context.Context, repository graphql.ID, number string) (Changeset, error) {
	if Changesets == nil {
		return nil, errChangesetsNotImplemented
	}
	return Changesets.ChangesetInRepository(ctx, repository, number)
}

// ChangesetsForRepository returns an instance of the GraphQL ChangesetConnection type with the list
// of changesets defined in a repository.
func ChangesetsForRepository(ctx context.Context, repository graphql.ID, arg *graphqlutil.ConnectionArgs) (ChangesetConnection, error) {
	if Changesets == nil {
		return nil, errChangesetsNotImplemented
	}
	return Changesets.ChangesetsForRepository(ctx, repository, arg)
}

func (schemaResolver) Changesets(ctx context.Context, arg *graphqlutil.ConnectionArgs) (ChangesetConnection, error) {
	if Changesets == nil {
		return nil, errChangesetsNotImplemented
	}
	return Changesets.Changesets(ctx, arg)
}

func (r schemaResolver) CreateChangeset(ctx context.Context, arg *CreateChangesetArgs) (Changeset, error) {
	if Changesets == nil {
		return nil, errChangesetsNotImplemented
	}
	return Changesets.CreateChangeset(ctx, arg)
}

func (r schemaResolver) UpdateChangeset(ctx context.Context, arg *UpdateChangesetArgs) (Changeset, error) {
	if Changesets == nil {
		return nil, errChangesetsNotImplemented
	}
	return Changesets.UpdateChangeset(ctx, arg)
}

func (r schemaResolver) PublishPreviewChangeset(ctx context.Context, arg *PublishPreviewChangesetArgs) (Changeset, error) {
	if Changesets == nil {
		return nil, errChangesetsNotImplemented
	}
	return Changesets.PublishPreviewChangeset(ctx, arg)
}

func (r schemaResolver) DeleteChangeset(ctx context.Context, arg *DeleteChangesetArgs) (*EmptyResponse, error) {
	if Changesets == nil {
		return nil, errChangesetsNotImplemented
	}
	return Changesets.DeleteChangeset(ctx, arg)
}

// ChangesetsResolver is the interface for the GraphQL changesets queries and mutations.
type ChangesetsResolver interface {
	// Queries
	Changesets(context.Context, *graphqlutil.ConnectionArgs) (ChangesetConnection, error)

	// Mutations
	CreateChangeset(context.Context, *CreateChangesetArgs) (Changeset, error)
	UpdateChangeset(context.Context, *UpdateChangesetArgs) (Changeset, error)
	PublishPreviewChangeset(context.Context, *PublishPreviewChangesetArgs) (Changeset, error)
	DeleteChangeset(context.Context, *DeleteChangesetArgs) (*EmptyResponse, error)

	// ChangesetByID is called by the ChangesetByID func but is not in the GraphQL API.
	ChangesetByID(context.Context, graphql.ID) (Changeset, error)

	// ChangesetInRepository is called by the ChangesetInRepository func but is not in the GraphQL
	// API.
	ChangesetInRepository(ctx context.Context, repository graphql.ID, number string) (Changeset, error)

	// ChangesetsForRepository is called by the ChangesetsForRepository func but is not in the
	// GraphQL API.
	ChangesetsForRepository(ctx context.Context, repository graphql.ID, arg *graphqlutil.ConnectionArgs) (ChangesetConnection, error)
}

type CreateChangesetArgs struct {
	Input struct {
		createThreadCommonInput
		Preview *bool
	}
}

type UpdateChangesetArgs struct {
	Input struct {
		updateThreadCommonInput
	}
}

type PublishPreviewChangesetArgs struct {
	Changeset graphql.ID
}

type DeleteChangesetArgs struct {
	Changeset graphql.ID
}

type ChangesetStatus string

const (
	ChangesetStatusOpen   ChangesetStatus = "OPEN"
	ChangesetStatusMerged                 = "MERGED"
	ChangesetStatusClosed                 = "CLOSED"
)

// Changeset is the interface for the GraphQL type Changeset.
type Changeset interface {
	threadCommon
	Status() ChangesetStatus
	RepositoryComparison(context.Context) (*RepositoryComparisonResolver, error)
}

// ChangesetConnection is the interface for the GraphQL type ChangesetConnection.
type ChangesetConnection interface {
	Nodes(context.Context) ([]Changeset, error)
	TotalCount(context.Context) (int32, error)
	PageInfo(context.Context) (*graphqlutil.PageInfo, error)
}
