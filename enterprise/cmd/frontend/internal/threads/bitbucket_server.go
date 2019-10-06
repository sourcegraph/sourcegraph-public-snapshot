package threads

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/events"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments"
	commentobjectdb "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func newBitbucketServerExternalThread(result *bitbucketServerPullRequest, resultComments []bitbucketServerPullRequestComment, repoID api.RepoID, externalServiceID int64) externalThread {
	thread, threadComment := bitbucketServerPullRequestToThread(result)
	thread.RepositoryID = repoID
	thread.ExternalServiceID = externalServiceID

	replyComments := make([]comments.ExternalComment, len(resultComments))
	for i, c := range resultComments {
		replyComments[i] = bitbucketServerPullRequestCommentToExternalComment(c)
	}
	return externalThread{
		thread:        thread,
		threadComment: threadComment,
		comments:      replyComments,
	}
}

func fromBitbucketServerDate(date int64) time.Time {
	return time.Unix(0, date*int64(time.Millisecond))
}

func bitbucketServerPullRequestToThread(v *bitbucketServerPullRequest) (*DBThread, commentobjectdb.DBObjectCommentFields) {
	thread := &DBThread{
		Title:      v.Title,
		State:      v.State,
		CreatedAt:  fromBitbucketServerDate(v.CreatedDate),
		UpdatedAt:  fromBitbucketServerDate(v.UpdatedDate),
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
		CreatedAt: fromBitbucketServerDate(v.CreatedDate),
		UpdatedAt: fromBitbucketServerDate(v.UpdatedDate),
	}
	bitbucketServerUserSetDBObjectCommentFields(v.Author.User, &comment)
	return thread, comment
}

func bitbucketServerUserSetDBObjectCommentFields(user *bitbucketServerUser, f *commentobjectdb.DBObjectCommentFields) {
	// TODO!(sqs): map to sourcegraph user if possible
	f.Author.ExternalActorUsername = user.Name
	f.Author.ExternalActorURL = user.Links.Self[0].Href
}

func bitbucketServerPullRequestCommentToExternalComment(v bitbucketServerPullRequestComment) comments.ExternalComment {
	comment := comments.ExternalComment{}
	bitbucketServerUserSetDBObjectCommentFields(v.Author, &comment.DBObjectCommentFields)
	comment.CreatedAt = fromBitbucketServerDate(v.CreatedDate)
	comment.UpdatedAt = fromBitbucketServerDate(v.UpdatedDate)
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
	ID          int                  `json:"id"`
	Text        string               `json:"text"`
	Author      *bitbucketServerUser `json:"author"`
	CreatedDate int64                `json:"createdDate"`
	UpdatedDate int64                `json:"updatedDate"`
}

type bitbucketServerUser struct {
	ID           int                     `json:"id"`
	Name         string                  `json:"name"`
	DisplayName  string                  `json:"displayName"`
	EmailAddress string                  `json:"emailAddress"`
	Links        bitbucketServerSelfLink `json:"links"`
}

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
	pull, err := c.createBitbucketServerPullRequest(ctx, bitbucketServerRepository, data)
	if err != nil && strings.Contains(err.Error(), "Only one pull request may be open for a given source and target") {
		pull, err = c.getExistingBitbucketServerPullRequest(ctx, bitbucketServerRepository, data)
	}
	if err != nil {
		return 0, err
	}
	// TODO!(sqs): doesnt actually update title/body/etc.

	comments, err := c.getPullRequestComments(ctx, strconv.Itoa(pull.ID), repoID)
	if err != nil {
		return 0, err
	}

	return ensureExternalThreadIsPersisted(ctx, newBitbucketServerExternalThread(pull, comments, repoID, c.src.ExternalServices()[0].ID), data.ExistingThreadID)
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
	params.Set("direction", "OUTGOING")

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

func (c *bitbucketServerExternalServiceClient) pullRequestURL(ctx context.Context, threadExternalID string, repoID api.RepoID) (string, error) {
	repoObj, err := backend.Repos.Get(ctx, repoID)
	if err != nil {
		return "", err
	}
	repo := getBitbucketServerRepositoryInput(repoObj.Name)

	pullRequestID, err := strconv.Atoi(threadExternalID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("rest/api/1.0/projects/%s/repos/%s/pull-requests/%d", repo.Project.Key, repo.Slug, pullRequestID), nil
}

func (c *bitbucketServerExternalServiceClient) RefreshThreadMetadata(ctx context.Context, threadID, threadExternalServiceID int64, externalID string, repoID api.RepoID) error {
	pull, comments, err := c.getPullRequestAndComments(ctx, externalID, repoID)
	if err != nil {
		return err
	}
	externalThread := newBitbucketServerExternalThread(pull, comments, repoID, threadExternalServiceID)
	return dbUpdateExternalThread(ctx, threadID, externalThread)
}

func (c *bitbucketServerExternalServiceClient) getPullRequestAndComments(ctx context.Context, threadExternalID string, repoID api.RepoID) (*bitbucketServerPullRequest, []bitbucketServerPullRequestComment, error) {
	pullRequestURL, err := c.pullRequestURL(ctx, threadExternalID, repoID)
	if err != nil {
		return nil, nil, err
	}
	var pull bitbucketServerPullRequest
	if err := c.src.Client().Send(ctx, "GET", pullRequestURL, nil, nil, &pull); err != nil {
		return nil, nil, err
	}

	comments, err := c.getPullRequestComments(ctx, threadExternalID, repoID)
	if err != nil {
		return nil, nil, err
	}

	return &pull, comments, nil
}

func (c *bitbucketServerExternalServiceClient) getPullRequestComments(ctx context.Context, threadExternalID string, repoID api.RepoID) ([]bitbucketServerPullRequestComment, error) {
	activities, err := c.getBitbucketServerPullRequestActivities(ctx, threadExternalID, repoID)
	if err != nil {
		return nil, err
	}
	var cs []bitbucketServerPullRequestComment
	for _, a := range activities {
		if bitbucketServerActionToEventType[a.Action] == comments.EventTypeComment && a.Comment != nil {
			cs = append(cs, *a.Comment)
		}
	}
	return cs, nil
}

func (c *bitbucketServerExternalServiceClient) GetThreadTimelineItems(ctx context.Context, threadExternalID string, repoID api.RepoID) ([]events.CreationData, error) {
	activities, err := c.getBitbucketServerPullRequestActivities(ctx, threadExternalID, repoID)
	if err != nil {
		return nil, err
	}

	// Bitbucket Server timeline events.
	items := make([]events.CreationData, 0, len(activities))
	for _, e := range activities {
		if eventType, ok := bitbucketServerActionToEventType[e.Action]; ok {
			if eventType == comments.EventTypeComment {
				// Skip because these will be added to the events table AND the comments table at
				// the same time.
				continue
			}

			data := events.CreationData{
				Type:      eventType,
				Data:      e,
				CreatedAt: fromBitbucketServerDate(e.CreatedDate),
			}

			var actor *bitbucketServerUser
			if e.User != nil {
				actor = e.User
			}
			if actor != nil {
				data.ExternalActorUsername = actor.Name
				data.ExternalActorURL = actor.Links.Self[0].Href
			}

			items = append(items, data)
		}
	}
	return items, nil
}

var bitbucketServerActionToEventType = map[string]events.Type{
	"OPENED":    eventTypeCreateThread,
	"COMMENTED": comments.EventTypeComment,
	"MERGED":    eventTypeMergeThread,
	"CLOSED":    eventTypeCloseThread,
	// TODO!(sqs): add review events
}

type bitbucketServerActivity struct {
	ID            int                                `json:"id"`
	Action        string                             `json:"action"`                  // OPENED, COMMENTED
	CommentAction string                             `json:"commentAction,omitempty"` // ADDED
	User          *bitbucketServerUser               `json:"user,omitempty"`
	CreatedDate   int64                              `json:"createdDate"`
	Comment       *bitbucketServerPullRequestComment `json:"comment,omitempty"`
}

func (c *bitbucketServerExternalServiceClient) getBitbucketServerPullRequestActivities(ctx context.Context, threadExternalID string, repoID api.RepoID) (activities []bitbucketServerActivity, err error) {
	pullRequestURL, err := c.pullRequestURL(ctx, threadExternalID, repoID)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("limit", "100")
	// TODO!(sqs): get all pages
	if _, err := c.src.Client().Page(ctx, fmt.Sprintf("%s/activities", pullRequestURL), params, nil, &activities); err != nil {
		return nil, err
	}
	return activities, nil
}
