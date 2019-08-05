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
		// Skip cross-repository PRs because we don't handle those yet.
		if ghPull.IsCrossRepository {
			continue
		}

		thread, comment := githubPullToThread(ghPull)
		thread.RepositoryID = repoID
		thread.ImportedFromExternalServiceID = externalServiceID
		comment.AuthorUserID = actor.FromContext(ctx).UID // TODO!(sqs): map to github user, and dont always just use current user
		toImport[thread] = comment
	}
	return internal.ImportExternalThreads(ctx, repoID, externalServiceID, toImport)
}

func githubPullToThread(ghPull *githubPullRequest) (*internal.DBThread, commentobjectdb.DBObjectCommentFields) {
	getRefName := func(ref *githubRef, oid string) string {
		// If a base/head is deleted, point to its OID directly.
		if ref == nil || ghPull.State == "MERGED" {
			return oid
		}
		return ref.Prefix + ref.Name
	}

	thread := &internal.DBThread{
		Type:       internal.DBThreadTypeChangeset,
		Title:      ghPull.Title,
		State:      ghPull.State,
		IsPreview:  false,
		CreatedAt:  ghPull.CreatedAt,
		BaseRef:    getRefName(ghPull.BaseRef, ghPull.BaseRefOid),
		HeadRef:    getRefName(ghPull.HeadRef, ghPull.HeadRefOid),
		ExternalID: string(ghPull.ID),
	}

	comment := commentobjectdb.DBObjectCommentFields{
		Body: ghPull.Body,
	}
	return thread, comment
}

type githubPullRequest struct {
	ID                graphql.ID `json:"id"`
	Number            int        `json:"number"`
	Title             string     `json:"title"`
	Body              string     `json:"body"`
	CreatedAt         time.Time  `json:"createdAt"`
	BaseRef           *githubRef `json:"baseRef"`
	BaseRefOid        string     `json:"baseRefOid"`
	HeadRef           *githubRef `json:"headRef"`
	HeadRefOid        string     `json:"headRefOid"`
	IsCrossRepository bool       `json:"isCrossRepository"`
	Permalink         string     `json:"permalink"`
	State             string     `json:"state"`
}

type githubRef struct {
	Name   string `json:"name"`
	Prefix string `json:"prefix"`
}

func listGitHubPullRequestsForRepository(ctx context.Context, client *github.Client, githubRepositoryID graphql.ID) (pulls []*githubPullRequest, err error) {
	var data struct {
		Node *struct {
			PullRequests struct {
				Nodes []*githubPullRequest
			}
		}
	}
	if err := client.RequestGraphQL(ctx, "", `
query ImportGitHubThreads($repository: ID!) {
	node(id: $repository) {
		... on Repository {
			pullRequests(first: 50, orderBy: { field: UPDATED_AT, direction: DESC }) {
				nodes {
					id
					number
					title
					body
					createdAt
					baseRef { name prefix }
					baseRefOid
					headRef { name prefix }
					headRefOid
					isCrossRepository
					permalink
					state
				}
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
	return data.Node.PullRequests.Nodes, nil
}
