package extsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/internal"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
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

	pulls, err := listGitHubPullRequestsForRepository(ctx, client, graphql.ID(extRepo.ID))
	if err != nil {
		return err
	}

	toImport := make(map[*internal.DBThread]commentobjectdb.DBObjectCommentFields, len(pulls))
	for _, ghPull := range pulls {
		toImport[&internal.DBThread{
			Type:                          internal.DBThreadTypeChangeset,
			RepositoryID:                  repoID,
			Title:                         ghPull.Title,
			State:                         ghPull.State,
			IsPreview:                     false,
			CreatedAt:                     ghPull.CreatedAt,
			BaseRef:                       ghPull.BaseRefName,
			HeadRef:                       ghPull.HeadRefName,
			ImportedFromExternalServiceID: externalServiceID,
			ExternalID:                    string(ghPull.ID),
		}] = commentobjectdb.DBObjectCommentFields{
			AuthorUserID: actor.FromContext(ctx).UID, // TODO!(sqs): map to github user, and dont always just use current user
			Body:         ghPull.Body,
		}
	}
	return internal.ImportExternalThreads(ctx, repoID, externalServiceID, toImport)
}

type githubPullRequest struct {
	ID                graphql.ID `json:"id"`
	Number            int        `json:"number"`
	Title             string     `json:"title"`
	Body              string     `json:"body"`
	CreatedAt         time.Time  `json:"createdAt"`
	BaseRefName       string     `json:"baseRefName"`
	HeadRefName       string     `json:"headRefName"`
	IsCrossRepository bool       `json:"isCrossRepository"`
	Permalink         string     `json:"permalink"`
	State             string     `json:"state"`
}

func listGitHubPullRequestsForRepository(ctx context.Context, client *github.Client, githubRepositoryID graphql.ID) (pulls []*githubPullRequest, err error) {
	var data struct {
		Node *struct {
			PullRequests []*githubPullRequest
		}
	}
	if err := client.RequestGraphQL(ctx, "", `
query ImportGitHubThreads($repository: ID!) {
	node(id: $repository) {
		... on Repository {
			pullRequests(first: 10) {
				id
				number
				title
				body
				createdAt
				baseRefName
				headRefName
				isCrossRepository
				permalink
				state
			}
		}
	}
}
`, map[string]interface{}{
		"repository": githubRepositoryID,
	}, &data); err != nil {
		return nil, err
	}
	if data.Node == nil {
		return nil, fmt.Errorf("github repository with ID %q not found", githubRepositoryID)
	}
	return data.Node.PullRequests, nil
}
