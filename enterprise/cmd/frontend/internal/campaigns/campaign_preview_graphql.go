package campaigns

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/events"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/diagnostics"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
)

func (GraphQLResolver) CampaignPreview(ctx context.Context, arg *graphqlbackend.CampaignPreviewArgs) (graphqlbackend.CampaignPreview, error) {
	return &gqlCampaignPreview{input: arg.Input}, nil
}

type gqlCampaignPreview struct {
	input graphqlbackend.CampaignPreviewInput
}

func (v *gqlCampaignPreview) Name() string { return v.input.Campaign.Name }

func (v *gqlCampaignPreview) Author(ctx context.Context) (*graphqlbackend.Actor, error) {
	user, err := graphqlbackend.CurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	return &graphqlbackend.Actor{User: user}, nil
}

func (v *gqlCampaignPreview) Body() string {
	if v.input.Campaign.Body == nil {
		return ""
	}
	return *v.input.Campaign.Body
}

func (v *gqlCampaignPreview) BodyText() string { return comments.ToBodyText(v.Body()) }

func (v *gqlCampaignPreview) BodyHTML() string { return comments.ToBodyHTML(v.Body()) }

func (v *gqlCampaignPreview) IsDraft() bool {
	return v.input.Campaign.Draft != nil && *v.input.Campaign.Draft
}

func (v *gqlCampaignPreview) StartDate() *graphqlbackend.DateTime { return v.input.Campaign.StartDate }

func (v *gqlCampaignPreview) DueDate() *graphqlbackend.DateTime { return v.input.Campaign.DueDate }

func (v *gqlCampaignPreview) getThreads(ctx context.Context) ([]graphqlbackend.ToThreadOrThreadPreview, error) {
	connection, err := v.Threads(ctx, &graphqlbackend.ThreadConnectionArgs{})
	if err != nil {
		return nil, err
	}
	return connection.Nodes(ctx)
}

func (v *gqlCampaignPreview) Threads(ctx context.Context, args *graphqlbackend.ThreadConnectionArgs) (graphqlbackend.ThreadOrThreadPreviewConnection, error) {
	// TODO!(sqs): dont ignore args
	allThreads, err := (&rulesExecutor{
		campaign: ruleExecutorCampaignInfo{
			name:    v.input.Campaign.Name,
			body:    stringOrEmpty(v.input.Campaign.Body),
			isDraft: v.input.Campaign.Draft != nil && *v.input.Campaign.Draft,
		},
		extensionData: &v.input.Campaign.ExtensionData,
	}).planThreads(ctx)
	if err != nil {
		return nil, err
	}
	return threads.ConstThreadOrThreadPreviewConnection(threads.ToThreadOrThreadPreviews(nil, allThreads)), nil
}

func (v *gqlCampaignPreview) Repositories(ctx context.Context) ([]*graphqlbackend.RepositoryResolver, error) {
	return campaignRepositories(ctx, v)
}

func (v *gqlCampaignPreview) Commits(ctx context.Context) ([]*graphqlbackend.GitCommitResolver, error) {
	return campaignCommits(ctx, v)
}

func (v *gqlCampaignPreview) RepositoryComparisons(ctx context.Context) ([]graphqlbackend.RepositoryComparison, error) {
	return campaignRepositoryComparisons(ctx, v)
}

func (v *gqlCampaignPreview) getDiagnostics(ctx context.Context) ([]graphqlbackend.Diagnostic, error) {
	return extdata{}.parseDiagnostics(&v.input.Campaign.ExtensionData)
}

func (v *gqlCampaignPreview) Diagnostics(ctx context.Context, args *graphqlutil.ConnectionArgs) (graphqlbackend.DiagnosticConnection, error) {
	ds, err := v.getDiagnostics(ctx)
	if err != nil {
		return nil, err
	}
	return diagnostics.ConstConnection(ds), nil
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

func (v *gqlCampaignPreview) Participants(ctx context.Context, arg *graphqlbackend.ParticipantConnectionArgs) (graphqlbackend.ParticipantConnection, error) {
	return campaignParticipants(ctx, v)
}
