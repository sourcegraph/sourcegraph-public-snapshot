package threads

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/actor"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
)

func init() {
	graphqlbackend.ForceRefreshRepositoryThreads = ImportGitHubRepositoryThreads
}

func ImportGitHubRepositoryThreads(ctx context.Context, repoID api.RepoID, extRepo api.ExternalRepoSpec) error {
	client, externalServiceID, err := getClientForRepo(ctx, repoID)
	if err != nil {
		return err
	}

	results, err := listGitHubRepositoryIssuesAndPullRequests(ctx, client, graphql.ID(extRepo.ID))
	if err != nil {
		return err
	}

	toImport := make([]*externalThread, 0, len(results))
	for _, result := range results {
		// Skip cross-repository PRs because we don't handle those yet.
		if result.IsCrossRepository {
			continue
		}
		// HACK TODO!(sqs): omit renovate PRs
		if strings.HasPrefix(result.Title, "Update dependency ") {
			continue
		}
		toImport = append(toImport, newExternalThread(result, repoID, externalServiceID))
	}
	return ImportExternalThreads(ctx, repoID, externalServiceID, toImport)
}

func newExternalThread(result *githubIssueOrPullRequest, repoID api.RepoID, externalServiceID int64) *externalThread {
	thread, threadComment := githubIssueOrPullRequestToThread(result)
	thread.RepositoryID = repoID
	thread.ImportedFromExternalServiceID = externalServiceID

	replyComments := make([]*comments.ExternalComment, len(result.Comments.Nodes))
	for i, c := range result.Comments.Nodes {
		replyComments[i] = githubIssueCommentToExternalComment(c)
	}
	return &externalThread{
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
		ExternalID: string(v.ID),
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

func githubIssueCommentToExternalComment(v *githubIssueComment) *comments.ExternalComment {
	comment := comments.ExternalComment{}
	githubActorSetDBObjectCommentFields(v.Author, &comment.DBObjectCommentFields)
	comment.CreatedAt = v.CreatedAt
	comment.UpdatedAt = v.UpdatedAt
	comment.Body = v.Body
	return &comment
}

type githubIssueOrPullRequest struct {
	Typename          string       `json:"__typename"`
	ID                graphql.ID   `json:"id"`
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

func listGitHubRepositoryIssuesAndPullRequests(ctx context.Context, client *github.Client, githubRepositoryID graphql.ID) (results []*githubIssueOrPullRequest, err error) {
	var data struct {
		Node *struct {
			PullRequests, Issues struct {
				Nodes []*githubIssueOrPullRequest
			}
		}
	}
	if err := client.RequestGraphQL(ctx, "", `
query ImportGitHubRepositoryIssuesAndPullRequests($repository: ID!) {
	node(id: $repository) {
		... on Repository {
			issues(first: 100, orderBy: { field: UPDATED_AT, direction: DESC }) {
				nodes {
`+githubIssueOrPullRequestCommonQuery+`
				}
			}
			pullRequests(first: 100, orderBy: { field: UPDATED_AT, direction: DESC }) {
				nodes {
`+githubIssueOrPullRequestCommonQuery+`
`+githubPullRequestQuery+`
				}
			}
		}
	}
}
`+githubActorFieldsFragment, map[string]interface{}{
		"repository": githubRepositoryID,
	}, &data); err != nil {
		return nil, err
	}
	if data.Node == nil {
		return nil, fmt.Errorf("github repository with ID %q not found", githubRepositoryID)
	}
	return append(data.Node.Issues.Nodes, data.Node.PullRequests.Nodes...), nil
}
