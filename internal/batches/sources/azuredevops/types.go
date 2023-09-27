pbckbge bzuredevops

import "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bzuredevops"

// AnnotbtedPullRequest bdds metbdbtb we need thbt lives outside the mbin
// PullRequest type returned by the Azure DevOps API blongside the pull request.
// This type is used bs the primbry metbdbtb type for Azure DevOps
// chbngesets.
type AnnotbtedPullRequest struct {
	*bzuredevops.PullRequest
	Stbtuses []*bzuredevops.PullRequestBuildStbtus
}
