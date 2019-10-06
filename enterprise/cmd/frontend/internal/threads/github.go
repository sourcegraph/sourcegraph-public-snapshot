package threads

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/actor"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/events"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments"
	commentobjectdb "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func newGitHubExternalThread(result *githubIssueOrPullRequest, repoID api.RepoID, externalServiceID int64) externalThread {
	thread, threadComment := githubIssueOrPullRequestToThread(result)
	thread.RepositoryID = repoID
	thread.ExternalServiceID = externalServiceID

	replyComments := make([]comments.ExternalComment, len(result.Comments.Nodes))
	for i, c := range result.Comments.Nodes {
		replyComments[i] = githubIssueCommentToExternalComment(c)
	}
	return externalThread{
		thread:        thread,
		threadComment: threadComment,
		comments:      replyComments,
	}
}

func githubIssueOrPullRequestToThread(v *githubIssueOrPullRequest) (*DBThread, commentobjectdb.DBObjectCommentFields) {
	thread := &DBThread{
		Title:      v.Title,
		State:      v.State,
		CreatedAt:  v.CreatedAt,
		UpdatedAt:  v.UpdatedAt,
		BaseRef:    v.BaseRefName,
		BaseRefOID: v.BaseRefOid,
		// TODO!(sqs): fill in headrepository
		HeadRef:    v.HeadRefName,
		HeadRefOID: v.HeadRefOid,
		ExternalThreadData: ExternalThreadData{
			ExternalID: string(v.ID),
		},
	}
	if len(v.Assignees.Nodes) >= 1 {
		// TODO!(sqs): support multiple assignees
		thread.Assignee = actor.DBColumns{
			ExternalActorUsername: v.Assignees.Nodes[0].Login,
			ExternalActorURL:      v.Assignees.Nodes[0].URL,
		}
	}
	var err error
	thread.ExternalMetadata, err = json.Marshal(v)
	if err != nil {
		panic(err)
	}

	comment := commentobjectdb.DBObjectCommentFields{
		Body:      v.Body,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}
	githubActorSetDBObjectCommentFields(v.Author, &comment)
	return thread, comment
}

func githubActorSetDBObjectCommentFields(actor *githubActor, f *commentobjectdb.DBObjectCommentFields) {
	// TODO!(sqs): map to sourcegraph user if possible
	f.Author.ExternalActorUsername = actor.Login
	f.Author.ExternalActorURL = actor.URL
}

func githubIssueCommentToExternalComment(v *githubIssueComment) comments.ExternalComment {
	comment := comments.ExternalComment{}
	githubActorSetDBObjectCommentFields(v.Author, &comment.DBObjectCommentFields)
	comment.CreatedAt = v.CreatedAt
	comment.UpdatedAt = v.UpdatedAt
	comment.Body = v.Body
	return comment
}

type githubIssueOrPullRequest struct {
	Typename   string     `json:"__typename"`
	ID         graphql.ID `json:"id"`
	Repository struct {
		ID            graphql.ID `json:"id"`
		NameWithOwner string     `json:"nameWithOwner"`
	} `json:"repository"`
	Number            int          `json:"number"`
	Title             string       `json:"title"`
	Body              string       `json:"body"`
	CreatedAt         time.Time    `json:"createdAt"`
	UpdatedAt         time.Time    `json:"updatedAt"`
	BaseRefName       string       `json:"baseRefName"`
	BaseRefOid        string       `json:"baseRefOid"`
	HeadRefName       string       `json:"headRefName"`
	HeadRefOid        string       `json:"headRefOid"`
	IsCrossRepository bool         `json:"isCrossRepository"`
	URL               string       `json:"url"`
	State             string       `json:"state"`
	Author            *githubActor `json:"author"`
	Assignees         struct {
		Nodes []*githubActor `json:"nodes"`
	} `json:"assignees"`
	Comments struct {
		Nodes []*githubIssueComment `json:"nodes"`
	} `json:"comments"`
}

type githubRef struct {
	Name   string `json:"name"`
	Prefix string `json:"prefix"`
}

type githubIssueComment struct {
	ID        graphql.ID   `json:"id"`
	Author    *githubActor `json:"author"`
	Body      string       `json:"body"`
	URL       string       `json:"url"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
}

const (
	githubIssueOrPullRequestCommonQuery = `
__typename
id
repository {
	id
	nameWithOwner
}
number
title
body
createdAt
updatedAt
url
state
author {
	...ActorFields
}
assignees(first: 1) {
	nodes {
		login
		url
	}
}
comments(first: 10) {
	nodes {
		id
		author { ... ActorFields  }
		body
		url
		createdAt
		updatedAt
	}
}
`

	githubPullRequestQuery = `
baseRefName
baseRefOid
headRefName
headRefOid
isCrossRepository
`
)

type githubActor struct {
	Login string `json:"login"`
	URL   string `json:"url"`
}

const githubActorFieldsFragment = `
fragment ActorFields on Actor {
	... on User {
		login
		url
	}
	... on Bot {
		login
		url
	}
}
`

type githubExternalServiceClient struct {
	src *repos.GithubSource
}

func (c *githubExternalServiceClient) CreateOrUpdateThread(ctx context.Context, repoName api.RepoName, repoID api.RepoID, extRepo api.ExternalRepoSpec, data CreateChangesetData) (threadID int64, err error) {
	githubRepositoryID := graphql.ID(extRepo.ID)
	thread, err := c.createGitHubPullRequest(ctx, githubRepositoryID, data)
	if err != nil && strings.Contains(err.Error(), "A pull request already exists") {
		thread, err = c.getExistingGitHubPullRequest(ctx, githubRepositoryID, data)
	}
	if err != nil {
		return 0, err
	}
	// TODO!(sqs): doesnt actually update title/body/etc.
	return ensureExternalThreadIsPersisted(ctx, newGitHubExternalThread(thread, repoID, c.src.ExternalServices()[0].ID), 0)
}

func (c *githubExternalServiceClient) createGitHubPullRequest(ctx context.Context, githubRepositoryID graphql.ID, data CreateChangesetData) (*githubIssueOrPullRequest, error) {
	var resp struct {
		CreatePullRequest struct {
			PullRequest githubIssueOrPullRequest
		}
	}
	if err := c.src.Client().RequestGraphQL(ctx, "", `
mutation CreatePullRequest($input: CreatePullRequestInput!) {
    createPullRequest(input: $input) {
        pullRequest {
`+githubIssueOrPullRequestCommonQuery+`
`+githubPullRequestQuery+`
        }
    }
}
`+githubActorFieldsFragment, map[string]interface{}{
		"input": map[string]interface{}{
			"repositoryId": githubRepositoryID,
			"baseRefName":  data.BaseRefName,
			"headRefName":  data.HeadRefName,
			"title":        data.Title,
			"body":         data.Body,
		},
	}, &resp); err != nil {
		return nil, err
	}
	return &resp.CreatePullRequest.PullRequest, nil
}

func (c *githubExternalServiceClient) getExistingGitHubPullRequest(ctx context.Context, githubRepositoryID graphql.ID, data CreateChangesetData) (*githubIssueOrPullRequest, error) {
	var resp struct {
		Node *struct {
			PullRequests struct {
				Nodes []*githubIssueOrPullRequest
			}
		}
	}
	if err := c.src.Client().RequestGraphQL(ctx, "", `
query GetPullRequest($repositoryId: ID!, $headRefName: String!) {
    node(id: $repositoryId) {
        ... on Repository {
            pullRequests(first: 1, headRefName: $headRefName)  {
                nodes {
`+githubIssueOrPullRequestCommonQuery+`
`+githubPullRequestQuery+`
                }
            }
        }
    }
}
`+githubActorFieldsFragment, map[string]interface{}{
		"repositoryId": githubRepositoryID,
		"headRefName":  data.HeadRefName,
	}, &resp); err != nil {
		return nil, err
	}
	if resp.Node == nil {
		return nil, fmt.Errorf("github repository with ID %q not found", githubRepositoryID)
	}
	if len(resp.Node.PullRequests.Nodes) == 0 {
		return nil, fmt.Errorf("no github pull requests in repository %q with head ref %q", githubRepositoryID, data.HeadRefName)
	}
	return resp.Node.PullRequests.Nodes[0], nil
}

func (c *githubExternalServiceClient) RefreshThreadMetadata(ctx context.Context, threadID, threadExternalServiceID int64, externalID string, repoID api.RepoID) error {
	externalServiceID := c.src.ExternalServices()[0].ID

	if externalServiceID != threadExternalServiceID {
		// TODO!(sqs): handle this case, not sure when it would happen, also is complicated by when
		// there are multiple external services for a repo.  TODO!(sqs): also make this look up the
		// external service using the externalServiceID directly when repo-updater exposes an API to
		// do that.
		return fmt.Errorf("thread %d: external service %d in DB does not match repository external service %d", threadID, threadExternalServiceID, externalServiceID)
	}

	var data struct {
		Node *githubIssueOrPullRequest
	}
	if err := c.src.Client().RequestGraphQL(ctx, "", `
query($id: ID!) {
	node(id: $id) {
		... on Issue {
`+githubIssueOrPullRequestCommonQuery+`
		}
		... on PullRequest {
`+githubIssueOrPullRequestCommonQuery+`
`+githubPullRequestQuery+`
		}
	}
}
`+githubActorFieldsFragment, map[string]interface{}{
		"id": externalID,
	}, &data); err != nil {
		return err
	}
	if data.Node == nil {
		return fmt.Errorf("github issue or pull request with ID %q not found", externalID)
	}

	externalThread := newGitHubExternalThread(data.Node, repoID, externalServiceID)
	return dbUpdateExternalThread(ctx, threadID, externalThread)
}

func (c *githubExternalServiceClient) GetThreadTimelineItems(ctx context.Context, threadExternalID string) ([]events.CreationData, error) {
	result, err := c.getGitHubIssueOrPullRequestTimelineItems(ctx, graphql.ID(threadExternalID))
	if err != nil {
		return nil, err
	}

	items := make([]events.CreationData, 0, len(result.TimelineItems.Nodes)+1)

	// Creation event.
	items = append(items, events.CreationData{
		Type:                  eventTypeCreateThread,
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

			items = append(items, data)
		}
	}
	return items, nil
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

	toImport, err := client.GetThreadTimelineItems(ctx, threadExternalID)
	if err != nil {
		return err
	}
	for i := range toImport {
		toImport[i].Objects.Thread = threadID
	}
	return events.ImportExternalEvents(ctx, externalServiceID, events.Objects{Thread: threadID}, toImport)
}

var githubEventTypes = map[string]events.Type{
	"IssueComment":         comments.EventTypeComment,
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
	githubIssueOrPullRequestEventCommonTimelineItemTypes = `ISSUE_COMMENT, CLOSED_EVENT, REOPENED_EVENT`
	githubIssueOrPullRequestEventCommonQuery             = `
... on IssueComment {
	id
	actor: author { ... ActorFields }
	createdAt
}
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

func (c *githubExternalServiceClient) getGitHubIssueOrPullRequestTimelineItems(ctx context.Context, githubIssueOrPullRequestID graphql.ID) (pull *githubPullRequestOrIssueTimelineData, err error) {
	var data struct {
		Node *githubPullRequestOrIssueTimelineData
	}
	if err := c.src.Client().RequestGraphQL(ctx, "", `
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
