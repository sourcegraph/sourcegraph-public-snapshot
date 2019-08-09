package campaigns

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments"
)

func (GraphQLResolver) CampaignPreview(ctx context.Context, arg *graphqlbackend.CampaignPreviewArgs) (graphqlbackend.CampaignPreview, error) {
	return &gqlCampaignPreview{input: arg.Input}
}

type gqlCampaignPreview struct {
	input graphqlbackend.CreateCampaignInput
}

func (v *gqlCampaignPreview) Name() string { return v.input.Name }

func (v *gqlCampaignPreview) Author(ctx context.Context) (*graphqlbackend.Actor, error) {
	user, err := graphqlbackend.CurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	return &graphqlbackend.Actor{User: user}, nil
}

func (v *gqlCampaignPreview) Body() string {
	if v.input.Body == nil {
		return ""
	}
	return *v.input.Body
}

func (v *gqlCampaignPreview) BodyText() string { return comments.ToBodyText(v.Body()) }

func (v *gqlCampaignPreview) BodyHTML() string { return comments.ToBodyHTML(v.Body()) }

func (v *gqlCampaignPreview) Threads(context.Context, *graphqlbackend.ThreadConnectionArgs) (graphqlbackend.ThreadConnection, error) {

}

func (v *gqlCampaignPreview) Repositories(context.Context) ([]*graphqlbackend.RepositoryResolver, error) {
}

func (v *gqlCampaignPreview) Commits(context.Context) ([]*graphqlbackend.GitCommitResolver, error) {}

func (v *gqlCampaignPreview) RepositoryComparisons(context.Context) ([]*graphqlbackend.RepositoryComparisonResolver, error) {

}

func (v *gqlCampaignPreview) Diagnostics(context.Context, *graphqlutil.ConnectionArgs) (graphqlbackend.DiagnosticConnection, error) {

}

func (v *gqlCampaignPreview) BurndownChart(context.Context) (graphqlbackend.CampaignBurndownChart, error) {
}
