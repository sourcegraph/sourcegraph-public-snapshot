package bitbucketcloud

import "github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketcloud"

// AnnotatedPullRequest adds metadata we need that lives outside the main
// PullRequest type returned by the Bitbucket API alongside the pull request.
// This type is used as the primary metadata type for Bitbucket Cloud
// changesets.
type AnnotatedPullRequest struct {
	*bitbucketcloud.PullRequest
	Statuses []*bitbucketcloud.PullRequestStatus
}
