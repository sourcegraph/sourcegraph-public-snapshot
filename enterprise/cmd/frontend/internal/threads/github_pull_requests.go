package threads

import (
	"context"
	"fmt"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
)

type CreateChangesetData struct {
	BaseRefName, HeadRefName string
	Title                    string
	Body                     string
}

func CreateOrGetExistingGitHubPullRequest(ctx context.Context, repoID api.RepoID, extRepo api.ExternalRepoSpec, data CreateChangesetData) (threadID int64, err error) {
	{
		// TODO!(sqs): demo only, make sure we dont accidentally make PRs to other repos
		repo, err := backend.Repos.Get(ctx, repoID)
		if err != nil {
			return 0, err
		}
		if !strings.Contains(string(repo.Name), "/sd9/") && !strings.Contains(string(repo.Name), "/sd9org/") {
			return 0, fmt.Errorf("refusing to create issue/PR for demo on non-demo repo %q", repo.Name)
		}
	}

	client, externalServiceID, err := getClientForRepo(ctx, repoID)
	if err != nil {
		return 0, err
	}

	pull, err := createOrGetExistingGitHubPullRequest(ctx, client, graphql.ID(extRepo.ID), data)
	if err != nil {
		return 0, err
	}
	externalThread := newExternalThread(pull, repoID, externalServiceID)

	thread, err := dbThreads{}.GetByExternal(ctx, externalServiceID, externalThread.thread.ExternalID)
	if err == nil {
		threadID = thread.ID
	} else if err == errThreadNotFound {
		threadID, err = dbCreateExternalThread(ctx, nil, externalThread)
	}
	return threadID, err
}

func createOrGetExistingGitHubPullRequest(ctx context.Context, client *github.Client, githubRepositoryID graphql.ID, data CreateChangesetData) (*githubIssueOrPullRequest, error) {
	pull, err := createGitHubPullRequest(ctx, client, githubRepositoryID, data)
	if err != nil && strings.Contains(err.Error(), "A pull request already exists") {
		return getExistingGitHubPullRequest(ctx, client, githubRepositoryID, data)
	}
	return pull, err
}

func createGitHubPullRequest(ctx context.Context, client *github.Client, githubRepositoryID graphql.ID, data CreateChangesetData) (*githubIssueOrPullRequest, error) {
	var resp struct {
		CreatePullRequest struct {
			PullRequest githubIssueOrPullRequest
		}
	}
	if err := client.RequestGraphQL(ctx, "", `
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

func getExistingGitHubPullRequest(ctx context.Context, client *github.Client, githubRepositoryID graphql.ID, data CreateChangesetData) (*githubIssueOrPullRequest, error) {
	var resp struct {
		Node *struct {
			PullRequests struct {
				Nodes []*githubIssueOrPullRequest
			}
		}
	}
	if err := client.RequestGraphQL(ctx, "", `
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
