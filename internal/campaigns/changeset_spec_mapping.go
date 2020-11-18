package campaigns

import "github.com/sourcegraph/sourcegraph/internal/api"

// ChangesetSpecRewire holds a mapping from a changeset spec to a changeset.
// If the changeset spec doesn't target any changeset (ie. it's new), ChangesetID is 0.
type ChangesetSpecRewire struct {
	ChangesetSpecID int64
	ChangesetID     int64
	RepoID          api.RepoID
}
