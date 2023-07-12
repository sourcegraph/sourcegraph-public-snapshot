package types

import (
	"sort"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// RewirerMapping maps a connection between ChangesetSpec and Changeset.
// If the ChangesetSpec doesn't match a Changeset (ie. it describes a to-be-created Changeset), ChangesetID is 0.
// If the ChangesetSpec is 0, the Changeset will be non-zero and means "to be closed".
// If both are non-zero values, the changeset should be updated with the changeset spec in the mapping.
type RewirerMapping struct {
	ChangesetSpecID int64
	ChangesetSpec   *ChangesetSpec
	ChangesetID     int64
	Changeset       *Changeset
	RepoID          api.RepoID
	Repo            *types.Repo
}

type RewirerMappings []*RewirerMapping

// ChangesetIDs returns a list of unique changeset IDs in the slice of mappings.
func (rm RewirerMappings) ChangesetIDs() []int64 {
	changesetIDMap := make(map[int64]struct{})
	for _, m := range rm {
		if m.ChangesetID != 0 {
			changesetIDMap[m.ChangesetID] = struct{}{}
		}
	}
	changesetIDs := make([]int64, 0, len(changesetIDMap))
	for id := range changesetIDMap {
		changesetIDs = append(changesetIDs, id)
	}
	sort.Slice(changesetIDs, func(i, j int) bool { return changesetIDs[i] < changesetIDs[j] })
	return changesetIDs
}

// ChangesetSpecIDs returns a list of unique changeset spec IDs in the slice of mappings.
func (rm RewirerMappings) ChangesetSpecIDs() []int64 {
	changesetSpecIDMap := make(map[int64]struct{})
	for _, m := range rm {
		if m.ChangesetSpecID != 0 {
			changesetSpecIDMap[m.ChangesetSpecID] = struct{}{}
		}
	}
	changesetSpecIDs := make([]int64, 0, len(changesetSpecIDMap))
	for id := range changesetSpecIDMap {
		changesetSpecIDs = append(changesetSpecIDs, id)
	}
	sort.Slice(changesetSpecIDs, func(i, j int) bool { return changesetSpecIDs[i] < changesetSpecIDs[j] })
	return changesetSpecIDs
}

// RepoIDs returns a list of unique repo IDs in the slice of mappings.
func (rm RewirerMappings) RepoIDs() []api.RepoID {
	repoIDMap := make(map[api.RepoID]struct{})
	for _, m := range rm {
		repoIDMap[m.RepoID] = struct{}{}
	}
	repoIDs := make([]api.RepoID, 0, len(repoIDMap))
	for id := range repoIDMap {
		repoIDs = append(repoIDs, id)
	}
	sort.Slice(repoIDs, func(i, j int) bool { return repoIDs[i] < repoIDs[j] })
	return repoIDs
}
