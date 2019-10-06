package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// Assignable implements the Assignable GraphQL interface.
type Assignable interface {
	Assignees(context.Context, *graphqlutil.ConnectionArgs) (ActorConnection, error)
}
