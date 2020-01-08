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
