package threads

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func UpdateGitHubThreadMetadata(ctx context.Context, threadID, threadExternalServiceID int64, externalID string, repoID api.RepoID) error {
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

	var data struct {
		Node *githubIssueOrPullRequest
	}
	if err := client.RequestGraphQL(ctx, "", `
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

	externalThread := newExternalThread(data.Node, repoID, externalServiceID)
	return dbUpdateExternalThread(ctx, threadID, externalThread)
}
