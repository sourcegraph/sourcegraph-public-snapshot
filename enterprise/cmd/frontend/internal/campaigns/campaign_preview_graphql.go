package campaigns

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments"
)

func (GraphQLResolver) CampaignPreview(ctx context.Context, arg *graphqlbackend.CampaignPreviewArgs) (graphqlbackend.CampaignPreview, error) {
	return &gqlCampaignPreview{input: arg.Input}, nil
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

func (v *gqlCampaignPreview) Threads(context.Context, *graphqlbackend.ThreadConnectionArgs) (graphqlbackend.ThreadOrThreadPreviewConnection, error) {
	panic("TODO!(sqs)")
}

func (v *gqlCampaignPreview) getThreads(ctx context.Context) ([]graphqlbackend.ToThreadOrThreadPreview, error) {
	connection, err := v.Threads(ctx, &graphqlbackend.ThreadConnectionArgs{})
	if err != nil {
		return nil, err
	}
	return connection.Nodes(ctx)
}

func (v *gqlCampaignPreview) Repositories(ctx context.Context) ([]*graphqlbackend.RepositoryResolver, error) {
	threads, err := v.getThreads(ctx)
	if err != nil {
		return nil, err
	}
	return campaignRepositories(ctx, threads)
}

func (v *gqlCampaignPreview) Commits(ctx context.Context) ([]*graphqlbackend.GitCommitResolver, error) {
	threads, err := v.getThreads(ctx)
	if err != nil {
		return nil, err
	}
	return campaignCommits(ctx, threads)
}

func (v *gqlCampaignPreview) RepositoryComparisons(ctx context.Context) ([]*graphqlbackend.RepositoryComparisonResolver, error) {
	threads, err := v.getThreads(ctx)
	if err != nil {
		return nil, err
	}
	return campaignRepositoryComparisons(ctx, threads)
}

func (v *gqlCampaignPreview) Diagnostics(context.Context, *graphqlutil.ConnectionArgs) (graphqlbackend.DiagnosticConnection, error) {
	panic("TODO!(sqs)")
}

func (v *gqlCampaignPreview) BurndownChart(context.Context) (graphqlbackend.CampaignBurndownChart, error) {
	panic("TODO!(sqs)")
}
