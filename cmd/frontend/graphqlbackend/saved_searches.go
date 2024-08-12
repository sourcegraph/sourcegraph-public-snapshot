package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

func (s *schemaResolver) SavedSearches(ctx context.Context, args SavedSearchesArgs) (*SavedSearchConnectionResolver, error) {
	return s.SavedSearchesResolver.SavedSearches(ctx, args)
}

type SavedSearchesResolver interface {
	// Query
	SavedSearches(ctx context.Context, args SavedSearchesArgs) (*SavedSearchConnectionResolver, error)
	SavedSearchByID(ctx context.Context, id graphql.ID) (SavedSearchResolver, error)

	// Mutations
	CreateSavedSearch(ctx context.Context, args *CreateSavedSearchArgs) (SavedSearchResolver, error)
	UpdateSavedSearch(ctx context.Context, args *UpdateSavedSearchArgs) (SavedSearchResolver, error)
	DeleteSavedSearch(ctx context.Context, args *DeleteSavedSearchArgs) (*EmptyResponse, error)
	TransferSavedSearchOwnership(ctx context.Context, args *TransferSavedSearchOwnershipArgs) (SavedSearchResolver, error)
	ChangeSavedSearchVisibility(ctx context.Context, args *ChangeSavedSearchVisibilityArgs) (SavedSearchResolver, error)

	NodeResolvers() map[string]NodeByIDFunc
}

type SavedSearchesOrderBy string

const (
	SavedSearchesOrderByDescription SavedSearchesOrderBy = "SAVED_SEARCH_DESCRIPTION"
	SavedSearchesOrderByUpdatedAt   SavedSearchesOrderBy = "SAVED_SEARCH_UPDATED_AT"
)

type SavedSearchVisibility string

const (
	SavedSearchVisibilityPublic SavedSearchVisibility = "PUBLIC"
	SavedSearchVisibilitySecret SavedSearchVisibility = "SECRET"
)

func (v SavedSearchVisibility) IsSecret() bool {
	return v != SavedSearchVisibilityPublic
}

type SavedSearchConnectionResolver = gqlutil.ConnectionResolver[SavedSearchResolver]

type SavedSearchResolver interface {
	ID() graphql.ID
	Description() string
	Query() string
	Draft() bool
	Owner(context.Context) (*NamespaceResolver, error)
	Visibility() SavedSearchVisibility
	CreatedBy(context.Context) (*UserResolver, error)
	CreatedAt() gqlutil.DateTime
	UpdatedBy(context.Context) (*UserResolver, error)
	UpdatedAt() gqlutil.DateTime
	URL() string
	ViewerCanAdminister(context.Context) bool
}

type SavedSearchesArgs struct {
	gqlutil.ConnectionResolverArgs
	Query              *string
	Owner              *graphql.ID
	ViewerIsAffiliated *bool
	IncludeDrafts      bool
	OrderBy            SavedSearchesOrderBy
}

type CreateSavedSearchArgs struct {
	Input SavedSearchInput
}

type SavedSearchInput struct {
	Owner       graphql.ID
	Description string
	Query       string
	Draft       bool
	Visibility  SavedSearchVisibility
}

type UpdateSavedSearchArgs struct {
	ID    graphql.ID
	Input SavedSearchUpdateInput
}

type SavedSearchUpdateInput struct {
	Description string
	Query       string
	Draft       bool
}

type DeleteSavedSearchArgs struct {
	ID graphql.ID
}

type TransferSavedSearchOwnershipArgs struct {
	ID       graphql.ID
	NewOwner graphql.ID
}

type ChangeSavedSearchVisibilityArgs struct {
	ID            graphql.ID
	NewVisibility SavedSearchVisibility
}
