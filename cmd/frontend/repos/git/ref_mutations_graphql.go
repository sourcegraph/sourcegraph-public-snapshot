package git

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (GraphQLResolver) CreateRefFromPatch(ctx context.Context, arg *struct {
	Input graphqlbackend.GitCreateRefFromPatchInput
}) (graphqlbackend.GitCreateRefFromPatchPayload, error) {
	panic("TODO!(sqs)")
}
