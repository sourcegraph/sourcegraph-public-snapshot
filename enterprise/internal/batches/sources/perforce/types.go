package perforce

import "github.com/sourcegraph/sourcegraph/internal/extsvc/perforce"

// from @varsanojidan, in a comment on the PR:
// You only really need this if the Changelist model doesn't contain information about changelist builds or reviews
// (if perforce has such a concept), if you can get all the info from Changelist directly,
// then you can just use the Changelist type instead.
//
// at some point we will probably want to add info about CLs that we would get from Swarm or elsewhere
// so I'll leave this in here
type AnnotatedChangelist struct {
	*perforce.Changelist
}
