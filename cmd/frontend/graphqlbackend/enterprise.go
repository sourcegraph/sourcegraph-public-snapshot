package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

type EnterpriseResolver interface {
	OrgRepositories(ctx context.Context, args *ListOrgRepositoriesArgs, org *types.Org) (RepositoryConnectionResolver, error)
}
