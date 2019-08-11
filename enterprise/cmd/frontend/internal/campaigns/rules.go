package campaigns

import (
	"context"
	"fmt"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/repos/git"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
	"github.com/sourcegraph/sourcegraph/pkg/gituri"
)

type rulesExecutor struct {
	extensionData graphqlbackend.CampaignExtensionData
}

func (x *rulesExecutor) planThreads(ctx context.Context) ([]graphqlbackend.ThreadPreview, error) {
	planIssues := func(ctx context.Context) (issues []graphqlbackend.ThreadPreview, err error) {
		// Include issues for each diagnostic.
		diagnostics, err := extdata{}.parseDiagnosticInfos(x.extensionData)
		if err != nil {
			return nil, err
		}
		for _, d := range diagnostics {
			repo, err := backend.Repos.GetByName(ctx, d.ResourceURI.Repo())
			if err != nil {
				return nil, err
			}
			issues = append(issues, threads.NewGQLThreadPreview(graphqlbackend.CreateThreadInput{
				Repository: graphqlbackend.NewRepositoryResolver(repo).ID(),
				Title:      d.Message,
			}, nil))
		}
		return issues, nil
	}

	planChangesets := func(ctx context.Context) (changesets []graphqlbackend.ThreadPreview, err error) {
		// Include changesets for each diff.
		diffs, err := extdata{}.parseRawFileDiffs(x.extensionData)
		if err != nil {
			return nil, err
		}
		for _, d := range diffs {
			newNameURI, err := gituri.Parse(d.NewName)
			if err != nil {
				return nil, err
			}
			repo, err := backend.Repos.GetByName(ctx, newNameURI.Repo())
			if err != nil {
				return nil, err
			}
			repoComparison := &git.GQLRepositoryComparisonPreview{
				BaseRepository_: graphqlbackend.NewRepositoryResolver(repo),
				HeadRepository_: graphqlbackend.NewRepositoryResolver(repo),
				FileDiffs_:      []*diff.FileDiff{d},
			}

			// TODO!(sqs) hack get title from diagnostic threads
			threadTitle := "Fix problems" // TODO!(sqs)

			changesets = append(changesets, threads.NewGQLThreadPreview(graphqlbackend.CreateThreadInput{
				Repository: graphqlbackend.NewRepositoryResolver(repo).ID(),
				Title:      fmt.Sprintf("Fix: %s", threadTitle),
			}, repoComparison))
		}
		return changesets, nil
	}
	// TODO!(sqs): include existing issues/threads matched by rules

	issues, err := planIssues(ctx)
	if err != nil {
		return nil, err
	}
	changesets, err := planChangesets(ctx)
	if err != nil {
		return nil, err
	}
	return append(issues, changesets...), nil
}

func (x *rulesExecutor) syncThreads(ctx context.Context, campaignID int64) error {
	allThreads, err := x.planThreads(ctx)
	if err != nil {
		return err
	}

	// TODO!(sqs): sync issues too - right now we only sync changesets because they are easier to
	// sync because they have a base/head that uniquely identifies them.
	for _, thread := range allThreads {
		kind, err := thread.Kind(ctx)
		if err != nil {
			return err
		}
		if kind != graphqlbackend.ThreadKindChangeset {
			continue
		}
		repo, err := thread.Repository(ctx)
		if err != nil {
			return err
		}
		threadID, err := threads.CreateOrGetExistingGitHubPullRequest(ctx, repo.DBID(), repo.DBExternalRepo(), threads.CreateChangesetData{
			BaseRefName: "master",          // TODO!(sqs): hack
			HeadRefName: "sourcegraph-a8n", // TODO!(sqs): hack
			Title:       thread.Title(),
		})
		if err != nil {
			return err
		}
		if err := (dbCampaignsThreads{}).AddThreadsToCampaign(ctx, campaignID, []int64{threadID}); err != nil {
			return err
		}
	}

	return nil
}

func (x *rulesExecutor) executeRules(ctx context.Context, campaignID int64) error {
	return x.syncThreads(ctx, campaignID)
}
