package campaigns

import (
	cmpgn "github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

// ChangesetEvents is a collection of changeset events
type ChangesetEvents []*cmpgn.ChangesetEvent

func (ce ChangesetEvents) Len() int      { return len(ce) }
func (ce ChangesetEvents) Swap(i, j int) { ce[i], ce[j] = ce[j], ce[i] }

// Less sorts changeset events by their Timestamps
func (ce ChangesetEvents) Less(i, j int) bool {
	return ce[i].Timestamp().Before(ce[j].Timestamp())
}

// FindMergeCommit will return the merge commit from the given set of events, stopping
// on the first it finds.
// It returns an empty string if none are found.
func (ce ChangesetEvents) FindMergeCommitID() string {
	for _, event := range ce {
		switch m := event.Metadata.(type) {
		case *bitbucketserver.Activity:
			if event.Kind != cmpgn.ChangesetEventKindBitbucketServerMerged {
				continue
			}
			if m.Commit == nil {
				continue
			}
			return m.Commit.ID

		case *github.MergedEvent:
			return m.Commit.OID
		}
	}
	return ""
}
