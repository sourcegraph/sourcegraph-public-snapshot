package campaigns

import (
	"context"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/events"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/repos/git"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/diagnostics"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
	"github.com/sourcegraph/sourcegraph/pkg/gituri"
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

func (v *gqlCampaignPreview) getThreads(ctx context.Context) ([]graphqlbackend.ToThreadOrThreadPreview, error) {
	connection, err := v.Threads(ctx, &graphqlbackend.ThreadConnectionArgs{})
	if err != nil {
		return nil, err
	}
	return connection.Nodes(ctx)
}

func (v *gqlCampaignPreview) Threads(ctx context.Context, args *graphqlbackend.ThreadConnectionArgs) (graphqlbackend.ThreadOrThreadPreviewConnection, error) {
	var allThreads []graphqlbackend.ToThreadOrThreadPreview

	getRepoFromURI := func(uriStr string) (*types.Repo, error) {
		uri, err := gituri.Parse(uriStr)
		if err != nil {
			return nil, err
		}
		return backend.Repos.GetByName(ctx, uri.Repo())
	}

	// Include issues for each diagnostic.
	diagnostics, err := v.getDiagnostics(ctx)
	if err != nil {
		return nil, err
	}
	for _, d := range diagnostics {
		var dd struct {
			Resource string
			Message  string
		}
		if err := json.Unmarshal([]byte(d.Data().Value.(json.RawMessage)), &dd); err != nil {
			return nil, err
		}
		repo, err := getRepoFromURI(dd.Resource)
		if err != nil {
			return nil, err
		}
		allThreads = append(allThreads, graphqlbackend.ToThreadOrThreadPreview{
			ThreadPreview: threads.NewGQLThreadPreview(graphqlbackend.CreateThreadInput{
				Repository: graphqlbackend.NewRepositoryResolver(repo).ID(),
				Title:      dd.Message,
			}, nil),
		})
	}

	// Include changesets for each diff.
	type fileDiff struct {
		parsed *diff.FileDiff
		repo   *types.Repo
	}
	diffs := make([]fileDiff, len(v.input.RawFileDiffs))
	for i, diffStr := range v.input.RawFileDiffs {
		var err error
		diffs[i].parsed, err = diff.ParseFileDiff([]byte(diffStr))
		if err != nil {
			return nil, err
		}
		diffs[i].repo, err = getRepoFromURI(diffs[i].parsed.NewName)
		if err != nil {
			return nil, err
		}
	}
	for _, d := range diffs {
		repoComparison := &git.GQLRepositoryComparisonPreview{
			BaseRepository_: graphqlbackend.NewRepositoryResolver(d.repo),
			HeadRepository_: graphqlbackend.NewRepositoryResolver(d.repo),
			FileDiffs_:      []*diff.FileDiff{d.parsed},
		}
		allThreads = append(allThreads, graphqlbackend.ToThreadOrThreadPreview{
			ThreadPreview: threads.NewGQLThreadPreview(graphqlbackend.CreateThreadInput{
				Repository: graphqlbackend.NewRepositoryResolver(d.repo).ID(),
				Title:      "DIFF TODO!(sqs)",
			}, repoComparison),
		})
	}
	// TODO!(sqs): include existing issues/threads matched by rules

	return threads.ConstThreadOrThreadPreviewConnection(allThreads), nil
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
	ds := make([]graphqlbackend.Diagnostic, len(v.input.RawDiagnostics))
	for i, diagnosticStr := range v.input.RawDiagnostics {
		var d diagnostics.GQLDiagnostic
		if err := json.Unmarshal([]byte(diagnosticStr), &d); err != nil {
			return nil, err
		}
		ds[i] = d
	}
	return ds, nil
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
