package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
)

// NewAuthzResolver will be set by enterprise
var NewAuthzResolver func() AuthzResolver

type AuthzResolver interface {
	SetRepositoryPermissionsForUsers(ctx context.Context, args *RepoPermsArgs) (*EmptyResponse, error)
	AuthorizedUserRepositories(ctx context.Context, args *AuthorizedRepoArgs) (RepositoryConnectionResolver, error)
	UsersWithPendingPermissions(ctx context.Context) ([]string, error)
	AuthorizedUsers(ctx context.Context, args *RepoAuthorizedUserArgs) (UserConnectionResolver, error)
}

type RepoPermsArgs struct {
	Repository graphql.ID
	BindIDs    []string
	Perm       string
}

var authzInEnterprise = errors.New("authorization mutations and queries are only available in enterprise")

func (*schemaResolver) SetRepositoryPermissionsForUsers(ctx context.Context, args *RepoPermsArgs) (*EmptyResponse, error) {
	if EnterpriseResolvers.authzResolver == nil {
		return nil, authzInEnterprise
	}
	return EnterpriseResolvers.authzResolver.SetRepositoryPermissionsForUsers(ctx, args)
}

type AuthorizedRepoArgs struct {
	Username *string
	Email    *string
	Perm     string
	First    int32
	After    *string
}

func (*schemaResolver) AuthorizedUserRepositories(ctx context.Context, args *AuthorizedRepoArgs) (RepositoryConnectionResolver, error) {
	if EnterpriseResolvers.authzResolver == nil {
		return nil, authzInEnterprise
	}
	return EnterpriseResolvers.authzResolver.AuthorizedUserRepositories(ctx, args)
}

func (*schemaResolver) UsersWithPendingPermissions(ctx context.Context) ([]string, error) {
	if EnterpriseResolvers.authzResolver == nil {
		return nil, authzInEnterprise
	}
	return EnterpriseResolvers.authzResolver.UsersWithPendingPermissions(ctx)
}
