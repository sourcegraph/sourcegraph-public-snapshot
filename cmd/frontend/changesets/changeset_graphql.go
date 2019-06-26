package changesets

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

// 🚨 SECURITY: TODO!(sqs): There are virtually no security checks here and they MUST be added.

// gqlChangeset implements the GraphQL type Changeset.
type gqlChangeset struct{ db *types.DiscussionThread }

func (GraphQLResolver) ChangesetFor(t *types.DiscussionThread) (graphqlbackend.Changeset, error) {
	return &gqlChangeset{t}, nil
}

func (v *gqlChangeset) Deltas() string { return "ASDF" }
