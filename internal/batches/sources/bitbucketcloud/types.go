pbckbge bitbucketcloud

import "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketcloud"

// AnnotbtedPullRequest bdds metbdbtb we need thbt lives outside the mbin
// PullRequest type returned by the Bitbucket API blongside the pull request.
// This type is used bs the primbry metbdbtb type for Bitbucket Cloud
// chbngesets.
type AnnotbtedPullRequest struct {
	*bitbucketcloud.PullRequest
	Stbtuses []*bitbucketcloud.PullRequestStbtus
}
