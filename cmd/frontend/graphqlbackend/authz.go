package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// NewAuthzResolver will be set by enterprise
var NewAuthzResolver func() AuthzResolver

type AuthzResolver interface {
	SetRepositoryPermissionsForUsers(ctx context.Context, args *SetRepoPermsArgs) (*EmptyResponse, error)
	AuthorizedUserRepositories(ctx context.Context, args *AuthorizedRepoArgs) (RepositoryConnectionResolver, error)
	UsersWithPendingPermissions(ctx context.Context) ([]string, error)
	AuthorizedUsers(ctx context.Context, args *RepoAuthorizedUserArgs) (UserConnectionResolver, error)
	RepositoriesPermissions(ctx context.Context, args *ReposPermsArgs) (RepositoryPermissionsConnectionResolver, error)
}

var authzInEnterprise = errors.New("authorization mutations and queries are only available in enterprise")

type defaultAuthzResolver struct{}

func (defaultAuthzResolver) SetRepositoryPermissionsForUsers(ctx context.Context, args *SetRepoPermsArgs) (*EmptyResponse, error) {
	return nil, authzInEnterprise
}

func (defaultAuthzResolver) AuthorizedUserRepositories(ctx context.Context, args *AuthorizedRepoArgs) (RepositoryConnectionResolver, error) {
	return nil, authzInEnterprise
}

func (defaultAuthzResolver) UsersWithPendingPermissions(ctx context.Context) ([]string, error) {
	return nil, authzInEnterprise
}

func (defaultAuthzResolver) AuthorizedUsers(ctx context.Context, args *RepoAuthorizedUserArgs) (UserConnectionResolver, error) {
	return nil, authzInEnterprise
}

func (defaultAuthzResolver) RepositoriesPermissions(ctx context.Context, args *ReposPermsArgs) (RepositoryPermissionsConnectionResolver, error) {
	return nil, authzInEnterprise
}

type ReposPermsArgs struct {
	graphqlutil.ConnectionArgs
	Query      *string
	OrderBy    string
	Descending bool
}

type SetRepoPermsArgs struct {
	Repository graphql.ID
	BindIDs    []string
	Perm       string
}

type AuthorizedRepoArgs struct {
	Username *string
	Email    *string
	Perm     string
	First    int32
	After    *string
}

type RepositoryPermissionsConnectionResolver interface {
	Nodes(ctx context.Context) ([]RepositoryPermissionsResolver, error)
	TotalCount(ctx context.Context) (int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

type RepositoryPermissionsResolver interface {
	Repository() *RepositoryResolver
	Permissions() RepositoryPermissionsInfoResolver
}

type RepositoryPermissionsInfoResolver interface {
	Perm() string
	UpdatedAt() DateTime
}
