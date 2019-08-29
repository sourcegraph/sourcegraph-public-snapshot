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
	IsDraft() bool
	StartDate() *DateTime
	DueDate() *DateTime
	Threads(context.Context, *ThreadConnectionArgs) (ThreadOrThreadPreviewConnection, error)
	Repositories(context.Context) ([]*RepositoryResolver, error)
	Commits(context.Context) ([]*GitCommitResolver, error)
	RepositoryComparisons(context.Context) ([]RepositoryComparison, error)
	Diagnostics(context.Context, *graphqlutil.ConnectionArgs) (DiagnosticConnection, error)
	BurndownChart(context.Context) (CampaignBurndownChart, error)
	hasParticipants
}
