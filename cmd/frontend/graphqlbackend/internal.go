package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
)

func (*schemaResolver) Internal() internalQueryResolver {
	return internalQueryResolver{}
}

type internalQueryResolver struct{}

func (internalQueryResolver) AllowEnableDisable(ctx context.Context) (bool, error) {
	return db.Repos.AllowEnableDisable(ctx)
}
