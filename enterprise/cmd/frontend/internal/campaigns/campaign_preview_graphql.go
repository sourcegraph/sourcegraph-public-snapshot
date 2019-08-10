package campaigns

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/events"
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
	return nil, errors.New("threads not implemented for CampaignPreview")
}

func (v *gqlCampaignPreview) getThreads(ctx context.Context) ([]graphqlbackend.ToThreadOrThreadPreview, error) {
	connection, err := v.Threads(ctx, &graphqlbackend.ThreadConnectionArgs{})
	if err != nil {
		return nil, err
	}
	return connection.Nodes(ctx)
}

func (v *gqlCampaignPreview) Repositories(ctx context.Context) ([]*graphqlbackend.RepositoryResolver, error) {
	return campaignRepositories(ctx, v)
}

func (v *gqlCampaignPreview) Commits(ctx context.Context) ([]*graphqlbackend.GitCommitResolver, error) {
	return campaignCommits(ctx, v)
}

func (v *gqlCampaignPreview) RepositoryComparisons(ctx context.Context) ([]*graphqlbackend.RepositoryComparisonResolver, error) {
	return campaignRepositoryComparisons(ctx, v)
}

func (v *gqlCampaignPreview) Diagnostics(ctx context.Context, args *graphqlutil.ConnectionArgs) (graphqlbackend.DiagnosticConnection, error) {
	return nil, errors.New("diagnostics not implemented for CampaignPreview")
}

func (v *gqlCampaignPreview) BurndownChart(ctx context.Context) (graphqlbackend.CampaignBurndownChart, error) {
	return campaignBurndownChart(ctx, v)
}

func (v *gqlCampaignPreview) getEvents(ctx context.Context, beforeDate time.Time, eventTypes []events.Type) ([]graphqlbackend.ToEvent, error) {
	return nil, errors.New("getEvents not implemented for CampaignPreview")
}

func (v *gqlCampaignPreview) TimelineItems(ctx context.Context, arg *graphqlbackend.EventConnectionCommonArgs) (graphqlbackend.EventConnection, error) {
	return nil, errors.New("timelineItems not implemented for CampaignPreview")
}
