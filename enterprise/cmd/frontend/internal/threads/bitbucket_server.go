package threads

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/events"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments"
	commentobjectdb "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func newBitbucketServerExternalThread(result *bitbucketServerPullRequest, repoID api.RepoID, externalServiceID int64) externalThread {
	thread, threadComment := bitbucketServerPullRequestToThread(result)
	thread.RepositoryID = repoID
	thread.ExternalServiceID = externalServiceID

	// replyComments := make([]comments.ExternalComment, len(result.Comments.Nodes))
	// for i, c := range result.Comments.Nodes {
	// 	replyComments[i] = bitbucketServerPullRequestCommentToExternalComment(c)
	// }
	return externalThread{
		thread:        thread,
		threadComment: threadComment,
		comments:      nil, // TODO!!(sqs) support comments
	}
}

func bitbucketServerPullRequestToThread(v *bitbucketServerPullRequest) (*DBThread, commentobjectdb.DBObjectCommentFields) {
	thread := &DBThread{
		Title:      v.Title,
		State:      v.State,
		CreatedAt:  time.Unix(0, v.CreatedDate*int64(time.Millisecond)),
		UpdatedAt:  time.Unix(0, v.UpdatedDate*int64(time.Millisecond)),
		BaseRef:    v.ToRef.ID,
		BaseRefOID: v.ToRef.LatestCommit,
		// TODO!(sqs): fill in headrepository
		HeadRef:            v.FromRef.ID,
		HeadRefOID:         v.FromRef.LatestCommit,
		ExternalThreadData: ExternalThreadData{ExternalID: strconv.Itoa(v.ID)},
	}
	// if len(v.Assignees.Nodes) >= 1 {
	// 	// TODO!(sqs): support multiple assignees
	// 	thread.Assignee = actor.DBColumns{
	// 		ExternalActorUsername: v.Assignees.Nodes[0].Login,
	// 		ExternalActorURL:      v.Assignees.Nodes[0].URL,
	// 	}
	// }
	var err error
	thread.ExternalMetadata, err = json.Marshal(v)
	if err != nil {
		panic(err)
	}

	comment := commentobjectdb.DBObjectCommentFields{
		Body:      v.Description,
		CreatedAt: time.Unix(0, v.CreatedDate*int64(time.Millisecond)),
		UpdatedAt: time.Unix(0, v.UpdatedDate*int64(time.Millisecond)),
	}
	bitbucketServerUserSetDBObjectCommentFields(v.Author.User, &comment)
	return thread, comment
}

func bitbucketServerUserSetDBObjectCommentFields(user *bitbucketServerUser, f *commentobjectdb.DBObjectCommentFields) {
	// TODO!(sqs): map to sourcegraph user if possible
	f.Author.ExternalActorUsername = user.Name
	f.Author.ExternalActorURL = user.Links.Self[0].Href
}

func bitbucketServerPullRequestCommentToExternalComment(v *bitbucketServerPullRequestComment) comments.ExternalComment {
	comment := comments.ExternalComment{}
	bitbucketServerUserSetDBObjectCommentFields(v.Author, &comment.DBObjectCommentFields)
	comment.CreatedAt = time.Unix(0, v.CreatedDate*int64(time.Millisecond))
	comment.UpdatedAt = time.Unix(0, v.UpdatedDate*int64(time.Millisecond))
	comment.Body = v.Text
	return comment
}

type bitbucketServerPullRequest struct {
	Typename    string                                 `json:"__typename"`
	ID          int                                    `json:"id"`
	Title       string                                 `json:"title"`
	Description string                                 `json:"description"`
	CreatedDate int64                                  `json:"createdDate"`
	UpdatedDate int64                                  `json:"updatedDate"`
	FromRef     bitbucketServerRef                     `json:"fromRef"`
	ToRef       bitbucketServerRef                     `json:"toRef"`
	Links       bitbucketServerSelfLink                `json:"links"`
	State       string                                 `json:"state"`
	Author      *bitbucketServerPullRequestParticipant `json:"author,omitempty"`
}

type bitbucketServerSelfLink struct {
	Self [1]struct {
		Href string `json:"href"`
	} `json:"self"`
}

type bitbucketServerRef struct {
	ID           string                     `json:"id"`
	LatestCommit string                     `json:"latestCommit"`
	Repository   *bitbucketServerRepository `json:"repository"`
	// TODO!(sqs): support cross-repo PRs
}

type bitbucketServerRepository struct {
	ID      int                    `json:"id"`
	Name    string                 `json:"name"`
	Slug    string                 `json:"slug"`
	Project bitbucketServerProject `json:"project"`
}

type bitbucketServerProject struct {
	ID   int    `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

type bitbucketServerPullRequestParticipant struct {
	User *bitbucketServerUser `json:"user"`
	Role string               `json:"role"`
}

type bitbucketServerPullRequestComment struct {
	ID          int `json:"id"`
	Text        string
	Author      *bitbucketServerUser `json:"author"`
	CreatedDate int64                `json:"createdDate"`
	UpdatedDate int64                `json:"updatedDate"`
}

const (
	bitbucketServerPullRequestCommonQuery = `
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

	bitbucketServerPullRequestQuery = `
baseRefName
baseRefOid
headRefName
headRefOid
isCrossRepository
`
)

type bitbucketServerUser struct {
	ID           int                     `json:"id"`
	Name         string                  `json:"name"`
	DisplayName  string                  `json:"displayName"`
	EmailAddress string                  `json:"emailAddress"`
	Links        bitbucketServerSelfLink `json:"links"`
}

const bitbucketServerUserFieldsFragment = `
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

type bitbucketServerExternalServiceClient struct {
	src *repos.BitbucketServerSource
}

func getBitbucketServerRepositoryInput(repo api.RepoName) bitbucketServerRepository {
	// TODO!!(sqs) validate this assumption, can be violated by repositoryPathPattern
	parts := strings.SplitN(string(repo), "/", 2)
	return bitbucketServerRepository{
		Slug: parts[1],
		Project: bitbucketServerProject{
			Key: parts[0],
		},
	}
}

func (c *bitbucketServerExternalServiceClient) CreateOrUpdateThread(ctx context.Context, repoName api.RepoName, repoID api.RepoID, extRepo api.ExternalRepoSpec, data CreateChangesetData) (threadID int64, err error) {
	bitbucketServerRepository := getBitbucketServerRepositoryInput(repoName)
	thread, err := c.createBitbucketServerPullRequest(ctx, bitbucketServerRepository, data)
	if err != nil && strings.Contains(err.Error(), "Only one pull request may be open for a given source and target") {
		thread, err = c.getExistingBitbucketServerPullRequest(ctx, bitbucketServerRepository, data)
	}
	if err != nil {
		return 0, err
	}
	// TODO!(sqs): doesnt actually update title/body/etc.
	return ensureExternalThreadIsPersisted(ctx, newBitbucketServerExternalThread(thread, repoID, c.src.ExternalServices()[0].ID), 0)
}

func (c *bitbucketServerExternalServiceClient) createBitbucketServerPullRequest(ctx context.Context, repo bitbucketServerRepository, data CreateChangesetData) (*bitbucketServerPullRequest, error) {
	var result bitbucketServerPullRequest
	if err := c.src.Client().Send(ctx, "POST", fmt.Sprintf("rest/api/1.0/projects/%s/repos/%s/pull-requests", repo.Project.Key, repo.Slug), nil, bitbucketServerPullRequest{
		Title:       data.Title,
		Description: data.Body,
		ToRef: bitbucketServerRef{
			ID:         data.BaseRefName,
			Repository: &repo,
		},
		FromRef: bitbucketServerRef{
			ID:         data.HeadRefName,
			Repository: &repo,
		},
	}, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *bitbucketServerExternalServiceClient) getExistingBitbucketServerPullRequest(ctx context.Context, repo bitbucketServerRepository, data CreateChangesetData) (*bitbucketServerPullRequest, error) {
	params := url.Values{}
	params.Set("at", data.HeadRefName)
	params.Set("withAttributes", "true")
	params.Set("withProperties", "true")
	params.Set("state", "OPEN")

	// TODO!(sqs) support pagination
	var pulls []bitbucketServerPullRequest
	if _, err := c.src.Client().Page(ctx, fmt.Sprintf("rest/api/1.0/projects/%s/repos/%s/pull-requests", repo.Project.Key, repo.Slug), params, nil, &pulls); err != nil {
		return nil, err
	}
	for _, pull := range pulls {
		if pull.FromRef.ID == data.HeadRefName {
			return &pull, nil
		}
	}
	return nil, fmt.Errorf("no bitbucketServer pull requests in repository %+v with head ref %q", repo, data.HeadRefName)
}

func (c *bitbucketServerExternalServiceClient) RefreshThreadMetadata(ctx context.Context, threadID, threadExternalServiceID int64, externalID string, repoID api.RepoID) error {
	repoObj, err := backend.Repos.Get(ctx, repoID)
	if err != nil {
		return err
	}
	repo := getBitbucketServerRepositoryInput(repoObj.Name)

	pullRequestID, err := strconv.Atoi(externalID)
	if err != nil {
		return err
	}
	var pull bitbucketServerPullRequest
	if err := c.src.Client().Send(ctx, "GET", fmt.Sprintf("rest/api/1.0/projects/%s/repos/%s/pull-requests/%d", repo.Project.Key, repo.Slug, pullRequestID), nil, nil, &pull); err != nil {
		return err
	}
	externalThread := newBitbucketServerExternalThread(&pull, repoID, threadExternalServiceID)
	return dbUpdateExternalThread(ctx, threadID, externalThread)
}

func (c *bitbucketServerExternalServiceClient) GetThreadTimelineItems(ctx context.Context, threadExternalID string) ([]events.CreationData, error) {
	panic("TODO!(sqs)")
	// result, err := c.getBitbucketServerPullRequestOrPullRequestTimelineItems(ctx, graphql.ID(threadExternalID))
	// if err != nil {
	// 	return nil, err
	// }

	// items := make([]events.CreationData, 0, len(result.TimelineItems.Nodes)+1)

	// // Creation event.
	// items = append(items, events.CreationData{
	// 	Type:                  eventTypeCreateThread,
	// 	ActorUserID:           0, // TODO!(sqs): determine this, map from bitbucketServer if needed
	// 	ExternalActorUsername: result.Author.Login,
	// 	ExternalActorURL:      result.Author.URL,
	// 	CreatedAt:             result.CreatedAt,
	// })

	// // BitbucketServer timeline events.
	// for _, ghEvent := range result.TimelineItems.Nodes {
	// 	if eventType, ok := bitbucketServerEventTypes[ghEvent.Typename]; ok {
	// 		data := events.CreationData{
	// 			Type:      eventType,
	// 			Data:      ghEvent,
	// 			CreatedAt: ghEvent.CreatedAt,
	// 		}

	// 		var actor *bitbucketServerUser
	// 		if ghEvent.Author != nil {
	// 			actor = ghEvent.Author
	// 		} else if ghEvent.Actor != nil {
	// 			actor = ghEvent.Actor
	// 		}
	// 		if actor != nil {
	// 			data.ExternalActorUsername = actor.Login
	// 			data.ExternalActorURL = actor.URL
	// 		}

	// 		items = append(items, data)
	// 	}
	// }
	// return items, nil
}

var MockImportBitbucketServerThreadEvents func() error // TODO!(sqs)

func ImportBitbucketServerThreadEvents(ctx context.Context, threadID, threadExternalServiceID int64, threadExternalID string, repoID api.RepoID) error {
	if MockImportBitbucketServerThreadEvents != nil {
		return MockImportBitbucketServerThreadEvents()
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

var bitbucketServerEventTypes = map[string]events.Type{
	"IssueComment":         comments.EventTypeComment,
	"PullRequestReview":    eventTypeReview,
	"ReviewRequestedEvent": eventTypeReviewRequested,
	"MergedEvent":          eventTypeMergeThread,
	"ClosedEvent":          eventTypeCloseThread,
	"ReopenedEvent":        eventTypeReopenThread,
}

type bitbucketServerPullRequestTimelineData struct {
	Values []bitbucketServerEvent `json:"values"`
}

type bitbucketServerEvent struct {
	ID            int                  `json:"id"`
	Action        string               `json:"action"`                  // OPENED, COMMENTED
	CommentAction string               `json:"commentAction,omitempty"` // ADDED
	User          *bitbucketServerUser `json:"user,omitempty"`
	CreatedDate   int                  `json:"createdDate"`
}

func (c *bitbucketServerExternalServiceClient) getBitbucketServerPullRequestOrPullRequestTimelineItems(ctx context.Context, bitbucketServerPullRequestID graphql.ID) (pull *bitbucketServerPullRequestTimelineData, err error) {
	return &bitbucketServerPullRequestTimelineData{}, nil
}
