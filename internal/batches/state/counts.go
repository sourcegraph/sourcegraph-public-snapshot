package state

import (
	"fmt"
	"time"

	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
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
func CalcCounts(start, end time.Time, cs []*btypes.Changeset, es ...*btypes.ChangesetEvent) ([]*ChangesetCounts, error) {
	ts := GenerateTimestamps(start, end)
	counts := make([]*ChangesetCounts, len(ts))
	for i, t := range ts {
		counts[i] = &ChangesetCounts{Time: t}
	}

	// Grouping Events by their Changeset ID
	byChangesetID := make(map[int64]ChangesetEvents)
	for _, e := range es {
		id := e.Changeset()
		byChangesetID[id] = append(byChangesetID[id], e)
	}

	// Map Events to their Changeset
	byChangeset := make(map[*btypes.Changeset]ChangesetEvents)
	for _, c := range cs {
		byChangeset[c] = byChangesetID[c.ID]
	}

	for changeset, csEvents := range byChangeset {
		// Compute history of changeset
		history, err := computeHistory(changeset, csEvents)
		if err != nil {
			return counts, err
		}

		// Go through every point in time we want to record and check the
		// states of the changeset at that point in time
		for _, c := range counts {
			states, ok := history.StatesAtTime(c.Time)
			if !ok {
				// Changeset didn't exist yet
				continue
			}

			c.Total++
			switch states.externalState {
			case btypes.ChangesetExternalStateDraft:
				c.Draft++
			case btypes.ChangesetExternalStateOpen:
				c.Open++
				switch states.reviewState {
				case btypes.ChangesetReviewStatePending:
					c.OpenPending++
				case btypes.ChangesetReviewStateApproved:
					c.OpenApproved++
				case btypes.ChangesetReviewStateChangesRequested:
					c.OpenChangesRequested++
				}

			case btypes.ChangesetExternalStateMerged:
				c.Merged++
			case btypes.ChangesetExternalStateClosed,
				btypes.ChangesetExternalStateReadOnly:
				// We'll lump read-only into closed, rather than trying to add another
				// state.
				c.Closed++
			}
		}
	}

	return counts, nil
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
