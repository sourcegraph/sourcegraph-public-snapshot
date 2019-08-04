package extsvc

import (
	"context"
	"encoding/json"
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
	eventTypeCreateThread    events.Type = "CreateThread"
	eventTypeReview                      = "Review"
	eventTypeReviewRequested             = "ReviewRequested"
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
	events.Register(eventTypeReview, func(ctx context.Context, common graphqlbackend.EventCommon, data events.EventData, toEvent *graphqlbackend.ToEvent) error {
		thread, err := threadlike.ThreadOrIssueOrChangesetByDBID(ctx, data.Thread)
		if err != nil {
			return err
		}
		// TODO!(sqs): validate state
		var o struct {
			State graphqlbackend.ReviewState `json:"state"`
		}
		if err := json.Unmarshal(data.Data, &o); err != nil {
			return err
		}
		toEvent.ReviewEvent = &graphqlbackend.ReviewEvent{
			EventCommon: common,
			Thread_:     thread,
			State_:      o.State,
		}
		return nil
	})
	events.Register(eventTypeReviewRequested, func(ctx context.Context, common graphqlbackend.EventCommon, data events.EventData, toEvent *graphqlbackend.ToEvent) error {
		thread, err := threadlike.ThreadOrIssueOrChangesetByDBID(ctx, data.Thread)
		if err != nil {
			return err
		}
		toEvent.ReviewRequestedEvent = &graphqlbackend.ReviewRequestedEvent{
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

	pull, externalServiceID, err := getGitHubPullRequest(ctx, repoID, extRepo, externalThreadURL)
	if err != nil {
		return err
	}

	objects := events.Objects{Thread: threadID}

	toImport := make([]events.CreationData, 0, len(pull.TimelineItems.Nodes)+1)

	// Creation event.
	toImport = append(toImport, events.CreationData{
		Type:        eventTypeCreateThread,
		Objects:     objects,
		ActorUserID: 0, // TODO!(sqs): determine this, map from github if needed
		CreatedAt:   pull.CreatedAt,
	})

	// GitHub timeline events.
	for _, ghEvent := range pull.TimelineItems.Nodes {
		if eventType, ok := githubEventTypes[ghEvent.Typename]; ok {
			toImport = append(toImport, events.CreationData{
				Type:      eventType,
				Objects:   objects,
				Data:      ghEvent,
				CreatedAt: ghEvent.CreatedAt,
			})
		}
	}

	return events.ImportExternalEvents(ctx, externalServiceID, objects, toImport)
}

var githubEventTypes = map[string]events.Type{
	"PullRequestReview":    eventTypeReview,
	"ReviewRequestedEvent": eventTypeReviewRequested,
}

var (
	cliFactory = repos.NewHTTPClientFactory()
)

func getClientForRepo(ctx context.Context, repoID api.RepoID) (client *github.Client, externalServiceID int64, err error) {
	// 🚨 SECURITY: Only site admins may read external services (they have secrets). TODO!(sqs)
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, 0, err
	}

	svcs, err := repoupdater.DefaultClient.RepoExternalServices(ctx, uint32(repoID))
	if err != nil {
		return nil, 0, err
	}
	// TODO!(sqs): how to choose if there are multiple
	if len(svcs) == 0 {
		return nil, 0, fmt.Errorf("no external services exist for repo %d", repoID)
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
		return nil, 0, err
	}
	return src.Client(), svcs[0].ID, nil
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

type githubPullRequest struct {
	CreatedAt     time.Time
	TimelineItems struct {
		Nodes []githubEvent
	}
}

type githubEvent struct {
	Typename      string `json:"__typename"`
	ID            graphql.ID
	Actor, Author *struct{ Login string }
	CreatedAt     time.Time
}

func getGitHubPullRequest(ctx context.Context, repoID api.RepoID, extRepo api.ExternalRepoSpec, externalThreadURL string) (pull *githubPullRequest, externalServiceID int64, err error) {
	number, err := parseIssueOrPullRequestNumberFromExternalURL(externalThreadURL)
	if err != nil {
		return nil, 0, err
	}

	gh, externalServiceID, err := getClientForRepo(ctx, repoID)
	if err != nil {
		return nil, 0, err
	}

	var data struct {
		Node *struct {
			PullRequest *githubPullRequest
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
							state
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
		return nil, 0, err
	}
	if data.Node == nil {
		return nil, 0, fmt.Errorf("github repository with ID %q not found", extRepo.ID)
	}
	if data.Node.PullRequest == nil {
		return nil, 0, fmt.Errorf("github repository with ID %q has no pull request #%d", extRepo.ID, number)
	}
	return data.Node.PullRequest, externalServiceID, nil
}
