package resolvers

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type Resolver struct {
}

func NewResolver(db dbutil.DB, clock func() time.Time) graphqlbackend.CodeGraphResolver {
	return &Resolver{}
}

func (r *Resolver) UserCodeGraph(ctx context.Context, user *graphqlbackend.UserResolver) (graphqlbackend.CodeGraphPersonNodeResolver, error) {
	return &CodeGraphPersonNodeResolver{
		user: user,
	}, nil
}

type CodeGraphPersonNodeResolver struct {
	user *graphqlbackend.UserResolver
}

func (CodeGraphPersonNodeResolver) Dependencies() []string {
	return []string{"mydependency1", "mydependency2", "mydependency3"}
}

func (CodeGraphPersonNodeResolver) Dependents() []string {
	return []string{"mydependent1", "mydependent2"}
}
