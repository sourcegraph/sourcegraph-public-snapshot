package extsvc

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/events"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
	"github.com/sourcegraph/sourcegraph/pkg/repoupdater"
)

const (
	eventTypeCreateThread = "CreateThread"
)

func init() {
	events.Register(eventTypeCreateThread, func(ctx context.Context, common graphqlbackend.EventCommon, data events.EventData, toEvent *graphqlbackend.ToEvent) error {
		thread, err := threadlike.ThreadOrIssueOrChangesetByDBID(ctx, data.Thread)
		if err != nil {
			return err
		}
		toEvent.CreateThreadEvent = &graphqlbackend.CreateThreadEvent{
			EventCommon: common,
			Thread_:     thread,
		}
		return nil
	})
}

var MockImportGitHubThreadEvents func() error // TODO!(sqs)

func ImportGitHubThreadEvents(ctx context.Context, threadID int64, repoID api.RepoID, extRepo api.ExternalRepoSpec, externalThreadURL string) error {
	if MockImportGitHubThreadEvents != nil {
		return MockImportGitHubThreadEvents()
	}

	number, err := parseIssueOrPullRequestNumberFromExternalURL(externalThreadURL)
	if err != nil {
		return err
	}

	gh, err := getClientForRepo(ctx, repoID)
	if err != nil {
		return err
	}

	var data struct {
		Node *struct {
			PullRequest *struct {
				CreatedAt     time.Time
				TimelineItems struct {
					Nodes []struct {
						Typename      string `json:"__typename"`
						ID            graphql.ID
						Actor, Author *struct{ Login string }
						CreatedAt     time.Time
					}
				}
			}
		}
	}
	if err := gh.RequestGraphQL(ctx, "", `
query ImportGitHubThreadEvents($repository: ID!, $pullRequestNumber: Int!) {
	node(id: $repository) {
		... on Repository {
			pullRequest(number: $pullRequestNumber) {
createdAt
				timelineItems(first: 10, itemTypes: [MERGED_EVENT, REOPENED_EVENT, REVIEW_REQUESTED_EVENT, PULL_REQUEST_REVIEW]) {
					nodes {
						__typename
						... on MergedEvent {
							id
							actor { login }
							createdAt
						}
						... on ReopenedEvent {
							id
							actor { login }
							createdAt
						}
						... on ReviewRequestedEvent {
							id
							actor { login }
							createdAt
						}
						... on PullRequestReview {
							id
							author { login }
							createdAt
							bodyText
						}
					}
				}
			}
		}
	}
}
`, map[string]interface{}{
		"repository":        extRepo.ID,
		"pullRequestNumber": number,
	}, &data); err != nil {
		return err
	}
	if data.Node == nil {
		return fmt.Errorf("github repository with ID %q not found", extRepo.ID)
	}
	if data.Node.PullRequest == nil {
		return fmt.Errorf("github repository with ID %q has no pull request #%d", extRepo.ID, number)
	}

	// Add creation event.
	if err := events.CreateEvent(ctx, events.CreationData{
		Type:        eventTypeCreateThread,
		Objects:     events.Objects{Thread: threadID},
		ActorUserID: 0, // TODO!(sqs): determine this, map from github if needed
		CreatedAt:   data.Node.PullRequest.CreatedAt,
	}); err != nil {
		return err
	}

	// Add GitHub timeline events.
	for _, event := range data.Node.PullRequest.TimelineItems.Nodes {
		// TODO!(sqs): map to sourcegraph event types
		if err := events.CreateEvent(ctx, events.CreationData{
			Type:      event.Typename,
			Objects:   events.Objects{Thread: threadID},
			Data:      event,
			CreatedAt: event.CreatedAt,
		}); err != nil {
			return err
		}
	}
	return nil
}

var (
	cliFactory = repos.NewHTTPClientFactory()
)

func getClientForRepo(ctx context.Context, repoID api.RepoID) (*github.Client, error) {
	// 🚨 SECURITY: Only site admins may read external services (they have secrets). TODO!(sqs)
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}

	svcs, err := repoupdater.DefaultClient.RepoExternalServices(ctx, uint32(repoID))
	if err != nil {
		return nil, err
	}
	// TODO!(sqs): how to choose if there are multiple
	if len(svcs) == 0 {
		return nil, fmt.Errorf("no external services exist for repo %d", repoID)
	}
	src, err := repos.NewGithubSource(&repos.ExternalService{
		ID:          svcs[0].ID,
		Kind:        svcs[0].Kind,
		DisplayName: svcs[0].DisplayName,
		Config:      svcs[0].Config,
		CreatedAt:   svcs[0].CreatedAt,
		UpdatedAt:   svcs[0].UpdatedAt,
	}, cliFactory)
	if err != nil {
		return nil, err
	}
	return src.Client(), nil
}

func parseIssueOrPullRequestNumberFromExternalURL(externalURL string) (number int, err error) {
	u, err := url.Parse(externalURL)
	if err != nil {
		return 0, err
	}

	i := strings.LastIndex(u.Path, "/")
	if i == -1 {
		err = errors.New("github url has no number")
		return
	}
	return strconv.Atoi(u.Path[i+1:])
}
