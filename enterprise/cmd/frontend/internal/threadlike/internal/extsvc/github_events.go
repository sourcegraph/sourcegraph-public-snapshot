package extsvc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/events"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
)

const (
	eventTypeCreateThread    events.Type = "CreateThread"
	eventTypeReview                      = "Review"
	eventTypeReviewRequested             = "ReviewRequested"
	eventTypeMergeChangeset              = "MergeChangeset"
	eventTypeCloseThread                 = "CloseThread"
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
		toEvent.RequestReviewEvent = &graphqlbackend.RequestReviewEvent{
			EventCommon: common,
			Thread_:     thread,
		}
		return nil
	})
	events.Register(eventTypeMergeChangeset, func(ctx context.Context, common graphqlbackend.EventCommon, data events.EventData, toEvent *graphqlbackend.ToEvent) error {
		thread, err := threadlike.ThreadOrIssueOrChangesetByDBID(ctx, data.Thread)
		if err != nil {
			return err
		}
		toEvent.MergeChangesetEvent = &graphqlbackend.MergeChangesetEvent{
			EventCommon: common,
			Changeset_:  thread.Changeset,
		}
		return nil
	})
	events.Register(eventTypeCloseThread, func(ctx context.Context, common graphqlbackend.EventCommon, data events.EventData, toEvent *graphqlbackend.ToEvent) error {
		thread, err := threadlike.ThreadOrIssueOrChangesetByDBID(ctx, data.Thread)
		if err != nil {
			return err
		}
		toEvent.CloseThreadEvent = &graphqlbackend.CloseThreadEvent{
			EventCommon: common,
			Thread_:     thread,
		}
		return nil
	})
}

var MockImportGitHubThreadEvents func() error // TODO!(sqs)

func ImportGitHubThreadEvents(ctx context.Context, threadID, threadExternalServiceID int64, threadExternalID string, repoID api.RepoID) error {
	if MockImportGitHubThreadEvents != nil {
		return MockImportGitHubThreadEvents()
	}

	client, externalServiceID, err := getClientForRepo(ctx, repoID)
	if err != nil {
		return err
	}

	if externalServiceID != threadExternalServiceID {
		// TODO!(sqs): handle this case, not sure when it would happen, also is complicated by when
		// there are multiple external services for a repo.
		return fmt.Errorf("thread %d: external service %d in DB does not match repository external service %d", threadID, threadExternalServiceID, externalServiceID)
	}

	pull, err := getGitHubPullRequestTimelineItems(ctx, client, graphql.ID(threadExternalID))
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
	"MergedEvent":          eventTypeMergeChangeset,
	"ClosedEvent":          eventTypeCloseThread,
}

type githubPullRequestTimelineData struct {
	CreatedAt     time.Time
	TimelineItems struct {
		Nodes []githubEvent
	}
}

type githubEvent struct {
	Typename string     `json:"__typename"`
	ID       graphql.ID `json:"id"`
	Actor    *struct {
		Login string `json:"login"`
	} `json:"actor,omitempty"`
	Author *struct {
		Login string `json:"login"`
	} `json:"author,omitempty"`
	State     string    `json:"state,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

func getGitHubPullRequestTimelineItems(ctx context.Context, client *github.Client, githubPullRequestID graphql.ID) (pull *githubPullRequestTimelineData, err error) {
	var data struct {
		Node *githubPullRequestTimelineData
	}
	if err := client.RequestGraphQL(ctx, "", `
query ImportGitHubThreadEvents($pullRequest: ID!) {
	node(id: $pullRequest) {
		... on PullRequest {
			createdAt
			timelineItems(first: 10, itemTypes: [MERGED_EVENT, CLOSED_EVENT, REOPENED_EVENT, REVIEW_REQUESTED_EVENT, PULL_REQUEST_REVIEW]) {
				nodes {
					__typename
					... on MergedEvent {
						id
						actor { login }
						createdAt
					}
					... on ClosedEvent {
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
`, map[string]interface{}{
		"pullRequest": githubPullRequestID,
	}, &data); err != nil {
		return nil, err
	}
	if data.Node == nil {
		return nil, fmt.Errorf("github pull request with ID %q not found", githubPullRequestID)
	}
	return data.Node, nil
}
