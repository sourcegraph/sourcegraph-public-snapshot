package campaigns

import (
	"sort"
	"time"

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

// UpdateLabelsSince returns the set of current labels based the starting set of labels and looking at events
// that have occurred after "since".
func (ce *ChangesetEvents) UpdateLabelsSince(cs *cmpgn.Changeset) []cmpgn.ChangesetLabel {
	var current []cmpgn.ChangesetLabel
	var since time.Time
	if cs != nil {
		current = cs.Labels()
		since = cs.UpdatedAt
	}
	// Copy slice so that we don't mutate ce
	sorted := make(ChangesetEvents, len(*ce))
	copy(sorted, *ce)
	sort.Sort(sorted)

	// Iterate through all label events to get the current set
	set := make(map[string]cmpgn.ChangesetLabel)
	for _, l := range current {
		set[l.Name] = l
	}
	for _, event := range sorted {
		switch e := event.Metadata.(type) {
		case *github.LabelEvent:
			if e.CreatedAt.Before(since) {
				continue
			}
			if e.Removed {
				delete(set, e.Label.Name)
				continue
			}
			set[e.Label.Name] = cmpgn.ChangesetLabel{
				Name:        e.Label.Name,
				Color:       e.Label.Color,
				Description: e.Label.Description,
			}
		}
	}
	labels := make([]cmpgn.ChangesetLabel, 0, len(set))
	for _, label := range set {
		labels = append(labels, label)
	}
	return labels
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
