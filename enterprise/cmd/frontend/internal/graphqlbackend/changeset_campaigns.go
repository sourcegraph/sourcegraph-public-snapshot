package graphqlbackend

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/changesets/campaigns"
)

func init() {
	// Contribute the GraphQL type ChangesetCampaignsMutation.
	graphqlbackend.ChangesetCampaigns = campaigns.GraphQLResolver{}
}
