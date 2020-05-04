package campaigns

import (
	"fmt"
	"sort"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/campaigns"
)

// ChangesetCounts represents the states in which a given set of Changesets was
// at a given point in time
type ChangesetCounts struct {
	Time                 time.Time
	Total                int32
	Merged               int32
	Closed               int32
	Open                 int32
	OpenApproved         int32
	OpenChangesRequested int32
	OpenPending          int32
}

// AddReviewState adds n to the corresponding counter for a given
// ChangesetReviewState
func (c *ChangesetCounts) AddReviewState(s campaigns.ChangesetReviewState, n int32) {
	switch s {
	case campaigns.ChangesetReviewStatePending:
		c.OpenPending += n
	case campaigns.ChangesetReviewStateApproved:
		c.OpenApproved += n
	case campaigns.ChangesetReviewStateChangesRequested:
		c.OpenChangesRequested += n
	}
}

func (cc *ChangesetCounts) String() string {
	return fmt.Sprintf("%s (Total: %d, Merged: %d, Closed: %d, Open: %d, OpenApproved: %d, OpenChangesRequested: %d, OpenPending: %d)",
		cc.Time.String(),
		cc.Total,
		cc.Merged,
		cc.Closed,
		cc.Open,
		cc.OpenApproved,
		cc.OpenChangesRequested,
		cc.OpenPending,
	)
}

// CalcCounts calculates ChangesetCounts for the given Changesets and their
// ChangesetEvents in the timeframe specified by the start and end parameters.
// The number of ChangesetCounts returned is the number of 1 day intervals
// between start and end, with each ChangesetCounts representing a point in
// time at the boundary of each 24h interval.
func CalcCounts(start, end time.Time, cs []*campaigns.Changeset, es ...*campaigns.ChangesetEvent) ([]*ChangesetCounts, error) {
	ts := generateTimestamps(start, end)
	counts := make([]*ChangesetCounts, len(ts))
	for i, t := range ts {
		counts[i] = &ChangesetCounts{Time: t}
	}

	// Sort all events once by their timestamps
	events := ChangesetEvents(es)
	sort.Sort(events)

	// Grouping Events by their Changeset ID
	byChangesetID := make(map[int64]ChangesetEvents)
	for _, e := range events {
		id := e.Changeset()
		byChangesetID[id] = append(byChangesetID[id], e)
	}

	// Map Events to their Changeset
	byChangeset := make(map[*campaigns.Changeset]ChangesetEvents)
	for _, c := range cs {
		byChangeset[c] = byChangesetID[c.ID]
	}

	for changeset, csEvents := range byChangeset {
		// We don't have an event for "open", so we check when it was
		// created on codehost
		openedAt := changeset.ExternalCreatedAt()
		if openedAt.IsZero() {
			continue
		}

		// We don't have an event for the deletion of a Changeset, but we set
		// ExternalDeletedAt manually in the Syncer.
		deletedAt := changeset.ExternalDeletedAt

		// For each changeset and its events, go through every point in time we
		// want to record and reconstruct the state of the changeset at that
		// point in time
		for _, c := range counts {
			if openedAt.After(c.Time) {
				// No need to look at events if changeset was not created yet
				continue
			}

			if !deletedAt.IsZero() && (deletedAt.Before(c.Time) || deletedAt.Equal(c.Time)) {
				c.Total++
				c.Closed++
				continue
			}

			err := computeCounts(c, changeset, csEvents)
			if err != nil {
				return counts, err
			}
		}
	}

	return counts, nil
}

func computeCounts(c *ChangesetCounts, ch *campaigns.Changeset, csEvents ChangesetEvents) error {
	history, err := computeHistory(ch, csEvents)
	if err != nil {
		return err
	}

	var states changesetStatesAtTime

	for _, s := range history {
		if s.t.After(c.Time) {
			break
		}
		states = s
	}

	if states.t.After(c.Time) {
		return nil
	}

	c.Total += 1
	switch states.state {
	case campaigns.ChangesetStateOpen:
		c.Open += 1
		c.AddReviewState(states.reviewState, 1)
	case campaigns.ChangesetStateMerged:
		c.Merged += 1
	case campaigns.ChangesetStateClosed:
		c.Closed += 1
	}

	return nil
}

func generateTimestamps(start, end time.Time) []time.Time {
	// Walk backwards from `end` to >= `start` in 1 day intervals
	// Backwards so we always end exactly on `end`
	ts := []time.Time{}
	for t := end; !t.Before(start); t = t.AddDate(0, 0, -1) {
		ts = append(ts, t)
	}

	// Now reverse so we go from oldest to newest in slice
	for i := len(ts)/2 - 1; i >= 0; i-- {
		opp := len(ts) - 1 - i
		ts[i], ts[opp] = ts[opp], ts[i]
	}

	return ts
}
