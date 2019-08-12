package threads

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/events"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
)

const (
	eventTypeCreateThread    events.Type = "CreateThread"
	eventTypeReview                      = "Review"
	eventTypeReviewRequested             = "ReviewRequested"
	eventTypeMergeThread                 = "MergeThread"
	eventTypeCloseThread                 = "CloseThread"
	eventTypeReopenThread                = "ReopenThread"
)

func init() {
	events.Register(eventTypeCreateThread, func(ctx context.Context, common graphqlbackend.EventCommon, data events.EventData, toEvent *graphqlbackend.ToEvent) error {
		thread, err := threadByDBID(ctx, data.Thread)
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
		thread, err := threadByDBID(ctx, data.Thread)
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
		thread, err := threadByDBID(ctx, data.Thread)
		if err != nil {
			return err
		}
		toEvent.RequestReviewEvent = &graphqlbackend.RequestReviewEvent{
			EventCommon: common,
			Thread_:     thread,
		}
		return nil
	})
	events.Register(eventTypeMergeThread, func(ctx context.Context, common graphqlbackend.EventCommon, data events.EventData, toEvent *graphqlbackend.ToEvent) error {
		thread, err := threadByDBID(ctx, data.Thread)
		if err != nil {
			return err
		}
		toEvent.MergeThreadEvent = &graphqlbackend.MergeThreadEvent{
			EventCommon: common,
			Thread_:     thread,
		}
		return nil
	})
	events.Register(eventTypeCloseThread, func(ctx context.Context, common graphqlbackend.EventCommon, data events.EventData, toEvent *graphqlbackend.ToEvent) error {
		thread, err := threadByDBID(ctx, data.Thread)
		if err != nil {
			return err
		}
		toEvent.CloseThreadEvent = &graphqlbackend.CloseThreadEvent{
			EventCommon: common,
			Thread_:     thread,
		}
		return nil
	})
	events.Register(eventTypeReopenThread, func(ctx context.Context, common graphqlbackend.EventCommon, data events.EventData, toEvent *graphqlbackend.ToEvent) error {
		thread, err := threadByDBID(ctx, data.Thread)
		if err != nil {
			return err
		}
		toEvent.ReopenThreadEvent = &graphqlbackend.ReopenThreadEvent{
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
		// there are multiple external services for a repo.  TODO!(sqs): also make this look up the
		// external service using the externalServiceID directly when repo-updater exposes an API to
		// do that.
		return fmt.Errorf("thread %d: external service %d in DB does not match repository external service %d", threadID, threadExternalServiceID, externalServiceID)
	}

	result, err := getGitHubIssueOrPullRequestTimelineItems(ctx, client, graphql.ID(threadExternalID))
	if err != nil {
		return err
	}

	objects := events.Objects{Thread: threadID}

	toImport := make([]events.CreationData, 0, len(result.TimelineItems.Nodes)+1)

	// Creation event.
	toImport = append(toImport, events.CreationData{
		Type:                  eventTypeCreateThread,
		Objects:               objects,
		ActorUserID:           0, // TODO!(sqs): determine this, map from github if needed
		ExternalActorUsername: result.Author.Login,
		ExternalActorURL:      result.Author.URL,
		CreatedAt:             result.CreatedAt,
	})

	// GitHub timeline events.
	for _, ghEvent := range result.TimelineItems.Nodes {
		if eventType, ok := githubEventTypes[ghEvent.Typename]; ok {
			data := events.CreationData{
				Type:      eventType,
				Objects:   objects,
				Data:      ghEvent,
				CreatedAt: ghEvent.CreatedAt,
			}

			var actor *githubActor
			if ghEvent.Author != nil {
				actor = ghEvent.Author
			} else if ghEvent.Actor != nil {
				actor = ghEvent.Actor
			}
			if actor != nil {
				data.ExternalActorUsername = actor.Login
				data.ExternalActorURL = actor.URL
			}

			toImport = append(toImport, data)
		}
	}

	return events.ImportExternalEvents(ctx, externalServiceID, objects, toImport)
}

var githubEventTypes = map[string]events.Type{
	"PullRequestReview":    eventTypeReview,
	"ReviewRequestedEvent": eventTypeReviewRequested,
	"MergedEvent":          eventTypeMergeThread,
	"ClosedEvent":          eventTypeCloseThread,
	"ReopenedEvent":        eventTypeReopenThread,
}

type githubPullRequestOrIssueTimelineData struct {
	CreatedAt     time.Time    `json:"createdAt"`
	Author        *githubActor `json:"author"`
	TimelineItems struct {
		Nodes []githubEvent
	}
}

type githubEvent struct {
	Typename  string       `json:"__typename"`
	ID        graphql.ID   `json:"id"`
	Actor     *githubActor `json:"actor,omitempty"`
	Author    *githubActor `json:"author,omitempty"`
	State     string       `json:"state,omitempty"`
	CreatedAt time.Time    `json:"createdAt"`
}

const (
	githubIssueOrPullRequestTimelineItemsCommonQuery = `
			author { ...ActorFields }
			createdAt
`
	githubIssueOrPullRequestEventCommonTimelineItemTypes = `CLOSED_EVENT, REOPENED_EVENT`
	githubIssueOrPullRequestEventCommonQuery             = `
... on ClosedEvent {
	id
	actor { ... ActorFields }
	createdAt
}
... on ReopenedEvent {
	id
	actor { ... ActorFields }
	createdAt
}
`
)

func getGitHubIssueOrPullRequestTimelineItems(ctx context.Context, client *github.Client, githubIssueOrPullRequestID graphql.ID) (pull *githubPullRequestOrIssueTimelineData, err error) {
	var data struct {
		Node *githubPullRequestOrIssueTimelineData
	}
	if err := client.RequestGraphQL(ctx, "", `
query ImportGitHubThreadEvents($issueOrPullRequest: ID!) {
	node(id: $issueOrPullRequest) {
		... on Issue {
		`+githubIssueOrPullRequestTimelineItemsCommonQuery+`
					timelineItems(first: 10, itemTypes: [`+githubIssueOrPullRequestEventCommonTimelineItemTypes+`]) {
						nodes {
							__typename
		`+githubIssueOrPullRequestEventCommonQuery+`
						}
					}
				}
		... on PullRequest {
`+githubIssueOrPullRequestTimelineItemsCommonQuery+`
			timelineItems(first: 10, itemTypes: [MERGED_EVENT, REVIEW_REQUESTED_EVENT, PULL_REQUEST_REVIEW, `+githubIssueOrPullRequestEventCommonTimelineItemTypes+`]) {
				nodes {
					__typename
`+githubIssueOrPullRequestEventCommonQuery+`
					... on MergedEvent {
						id
						actor { ...ActorFields }
						createdAt
					}
					... on ReviewRequestedEvent {
						id
						actor { ... ActorFields }
						createdAt
					}
					... on PullRequestReview {
						id
						author { ... ActorFields }
						createdAt
						state
					}
				}
			}
		}
	}
}
`+githubActorFieldsFragment, map[string]interface{}{
		"issueOrPullRequest": githubIssueOrPullRequestID,
	}, &data); err != nil {
		return nil, err
	}
	if data.Node == nil {
		return nil, fmt.Errorf("github issue or pull request with ID %q not found", githubIssueOrPullRequestID)
	}
	return data.Node, nil
}
