package resolvers

import (
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	codeintelresolvers "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type Resolver struct {
	codeIntelResolver func() codeintelresolvers.Resolver
}

func NewResolver(
	db dbutil.DB,
	codeIntelResolver func() codeintelresolvers.Resolver,
	clock func() time.Time,
) graphqlbackend.CodeGraphResolver {
	return &Resolver{
		codeIntelResolver: codeIntelResolver,
	}
}
