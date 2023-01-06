package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

type RoleResolver interface {
	ID() graphql.ID
	Name() string
	Readonly() bool
	CreatedAt() gqlutil.DateTime
	DeletedAt() gqlutil.DateTime
	Permissions() (PermissionConnectionResolver, error)
}

type RoleConnectionResolver interface {
	Nodes(ctx context.Context) ([]RoleResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type PermissionResolver interface {
	ID() graphql.ID
	Namespace() string
	Action() string
	CreatedAt() gqlutil.DateTime
}

type PermissionConnectionResolver interface {
	Nodes(ctx context.Context) ([]PermissionResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type RBACResolver interface {
	// MUTATIONS

	// QUERIES
	Role(ctx context.Context, args *RoleArgs) (RoleResolver, error)
	Roles(ctx context.Context, args *ListRoleArgs) (RoleConnectionResolver, error)

	Permissions(ctx context.Context, args *ListPermissionArgs) (PermissionConnectionResolver, error)

	NodeResolvers() map[string]NodeByIDFunc
}

type RoleArgs struct {
	ID graphql.ID
}

type ListRoleArgs struct {
	User  *graphql.ID
	First int32
	After *string
}

type ListPermissionArgs struct {
	Role  *graphql.ID
	User  *graphql.ID
	First int32
	After *string
}
