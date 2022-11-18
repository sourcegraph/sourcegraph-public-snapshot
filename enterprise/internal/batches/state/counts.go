package state

import (
	"fmt"
	"time"

	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
)

// timestampCount defines how many timestamps we will return for a given dateframe.
const timestampCount = 150

// ChangesetCounts represents the states in which a given set of Changesets was
// at a given point in time
type ChangesetCounts struct {
	Time                 time.Time
	Total                int32
	Merged               int32
	Closed               int32
	Draft                int32
	Open                 int32
	OpenApproved         int32
	OpenChangesRequested int32
	OpenPending          int32
}

func (cc *ChangesetCounts) String() string {
	return fmt.Sprintf("%s (Total: %d, Merged: %d, Closed: %d, Draft: %d, Open: %d, OpenApproved: %d, OpenChangesRequested: %d, OpenPending: %d)",
		cc.Time.String(),
		cc.Total,
		cc.Merged,
		cc.Closed,
		cc.Draft,
		cc.Open,
		cc.OpenApproved,
		cc.OpenChangesRequested,
		cc.OpenPending,
	)
}

// CalcCounts calculates ChangesetCounts for the given Changesets and their
// ChangesetEvents in the timeframe specified by the start and end parameters.
// The number of ChangesetCounts returned is always `timestampCount`. Between
// start and end, it generates `timestampCount` datapoints with each ChangesetCounts
// representing a point in time. `es` are expected to be pre-sorted.
func CalcCounts(start, end time.Time, cs []*btypes.Changeset) ([]*ChangesetCounts, error) {
	// todo: reimplement

	return []*ChangesetCounts{}, nil
}

func GenerateTimestamps(start, end time.Time) []time.Time {
	timeStep := end.Sub(start) / timestampCount
	// Walk backwards from `end` to >= `start` in equal intervals.
	// Backwards so we always end exactly on `end`.
	ts := []time.Time{}
	for t := end; !t.Before(start); t = t.Add(-timeStep) {
		ts = append(ts, t)
	}

	// Now reverse so we go from oldest to newest in slice
	for i := len(ts)/2 - 1; i >= 0; i-- {
		opp := len(ts) - 1 - i
		ts[i], ts[opp] = ts[opp], ts[i]
	}

	return ts
}
