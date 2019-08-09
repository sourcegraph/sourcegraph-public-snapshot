package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// CampaignPreview is the interface for the GraphQL type CampaignPreview.
type CampaignPreview interface {
	Name() string
	Author(context.Context) (*Actor, error)
	Body() string
	BodyText() string
	BodyHTML() string
	Threads(context.Context, *ThreadConnectionArgs) (ThreadConnection, error)
	Repositories(context.Context) ([]*RepositoryResolver, error)
	Commits(context.Context) ([]*GitCommitResolver, error)
	RepositoryComparisons(context.Context) ([]*RepositoryComparisonResolver, error)
	Diagnostics(context.Context, *graphqlutil.ConnectionArgs) (DiagnosticConnection, error)
	BurndownChart(context.Context) (CampaignBurndownChart, error)
}
