package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

func (s *schemaResolver) Prompts(ctx context.Context, args PromptsArgs) (*PromptConnectionResolver, error) {
	return s.PromptsResolver.Prompts(ctx, args)
}

type PromptsResolver interface {
	// Query
	Prompts(ctx context.Context, args PromptsArgs) (*PromptConnectionResolver, error)
	PromptByID(ctx context.Context, id graphql.ID) (PromptResolver, error)

	// Mutations
	CreatePrompt(ctx context.Context, args *CreatePromptArgs) (PromptResolver, error)
	UpdatePrompt(ctx context.Context, args *UpdatePromptArgs) (PromptResolver, error)
	DeletePrompt(ctx context.Context, args *DeletePromptArgs) (*EmptyResponse, error)
	TransferPromptOwnership(ctx context.Context, args *TransferPromptOwnershipArgs) (PromptResolver, error)
	ChangePromptVisibility(ctx context.Context, args *ChangePromptVisibilityArgs) (PromptResolver, error)

	NodeResolvers() map[string]NodeByIDFunc
}

type PromptsOrderBy string

const (
	PromptsOrderByNameWithOwner PromptsOrderBy = "PROMPT_NAME_WITH_OWNER"
	PromptsOrderByUpdatedAt     PromptsOrderBy = "PROMPT_UPDATED_AT"
)

type PromptVisibility string

const (
	PromptVisibilityPublic PromptVisibility = "PUBLIC"
	PromptVisibilitySecret PromptVisibility = "SECRET"
)

func (v PromptVisibility) IsSecret() bool {
	return v != PromptVisibilityPublic
}

type PromptConnectionResolver = gqlutil.ConnectionResolver[PromptResolver]

type PromptResolver interface {
	ID() graphql.ID
	Name() string
	Description() string
	Definition() PromptDefinitionResolver
	Draft() bool
	Owner(context.Context) (*NamespaceResolver, error)
	Visibility() PromptVisibility
	NameWithOwner(context.Context) (string, error)
	CreatedBy(context.Context) (*UserResolver, error)
	CreatedAt() gqlutil.DateTime
	UpdatedBy(context.Context) (*UserResolver, error)
	UpdatedAt() gqlutil.DateTime
	URL() string
	ViewerCanAdminister(context.Context) bool
}

type PromptDefinitionResolver struct {
	Text_ string
}

func (r PromptDefinitionResolver) Text() string {
	return r.Text_
}

type PromptsArgs struct {
	gqlutil.ConnectionResolverArgs
	Query              *string
	Owner              *graphql.ID
	ViewerIsAffiliated *bool
	IncludeDrafts      bool
	OrderBy            PromptsOrderBy
}

type CreatePromptArgs struct {
	Input PromptInput
}

type PromptInput struct {
	Owner          graphql.ID
	Name           string
	Description    string
	DefinitionText string
	Draft          bool
	Visibility     PromptVisibility
}

type UpdatePromptArgs struct {
	ID    graphql.ID
	Input PromptUpdateInput
}

type PromptUpdateInput struct {
	Name           string
	Description    string
	DefinitionText string
	Draft          bool
}

type DeletePromptArgs struct {
	ID graphql.ID
}

type TransferPromptOwnershipArgs struct {
	ID       graphql.ID
	NewOwner graphql.ID
}

type ChangePromptVisibilityArgs struct {
	ID            graphql.ID
	NewVisibility PromptVisibility
}
