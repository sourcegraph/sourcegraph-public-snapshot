package azuredevops

import "github.com/sourcegraph/sourcegraph/internal/extsvc/azuredevops"

// AzureDevOpsAnnotatedPullRequest adds metadata we need that lives outside the main
// PullRequest type returned by the Azure DevOps API alongside the pull request.
// This type is used as the primary metadata type for Azure DevOps
// changesets.
type AzureDevOpsAnnotatedPullRequest struct {
	*azuredevops.PullRequest
	Statuses []*azuredevops.PullRequestBuildStatus
}
