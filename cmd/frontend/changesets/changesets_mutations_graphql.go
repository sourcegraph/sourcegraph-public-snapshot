package changesets

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (GraphQLResolver) CreateChangeset(ctx context.Context, arg *struct {
	Input graphqlbackend.ChangesetsCreateChangesetInput
}) (graphqlbackend.ChangesetsCreateChangesetPayload, error) {
	panic("TODO!(sqs)")
}
