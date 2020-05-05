package campaigns

import (
	"fmt"
	"sort"
	"time"

	"github.com/inconshreveable/log15"
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

			err := computeCounts(c, csEvents)
			if err != nil {
				return counts, err
			}
		}
	}

	return counts, nil
}

func computeCounts(c *ChangesetCounts, csEvents ChangesetEvents) error {
	var (
		// Since "Merged" and "Closed" are exclusive events and cancel each others
		// effects on ChangesetCounts out, we need to keep track of when a
		// changeset was closed, so we can undo the effect of the "Closed" event
		// when we come across a "Merge" (since, on GitHub, a PR can be closed AND
		// merged)
		closed = false

		lastReviewByAuthor = map[string]campaigns.ChangesetReviewState{}
	)

	c.Total++
	c.Open++
	c.OpenPending++

	for _, e := range csEvents {
		et := e.Timestamp()
		if et.IsZero() {
			continue
		}
		// Event happened after point in time we're looking at, no need to look
		// at the events in future
		if et.After(c.Time) {
			return nil
		}

		// Compute current overall review state
		currentReviewState := computeReviewState(lastReviewByAuthor)

		switch e.Type() {
		case campaigns.ChangesetEventKindGitHubClosed,
			campaigns.ChangesetEventKindBitbucketServerDeclined:

			c.Open--
			c.Closed++
			closed = true

			c.AddReviewState(currentReviewState, -1)

		case campaigns.ChangesetEventKindGitHubReopened,
			campaigns.ChangesetEventKindBitbucketServerReopened:

			c.Open++
			c.Closed--
			closed = false

			c.AddReviewState(currentReviewState, 1)

		case campaigns.ChangesetEventKindGitHubMerged,
			campaigns.ChangesetEventKindBitbucketServerMerged:

			// If it was closed, all "review counts" have been updated by the
			// closed events and we just need to reverse these two counts
			if closed {
				c.Closed--
				c.Merged++
				return nil
			}

			c.AddReviewState(currentReviewState, -1)

			c.Merged++
			c.Open--

			// Merged is a final state, we return here and don't need to look at
			// other events
			return nil

		case campaigns.ChangesetEventKindGitHubReviewed,
			campaigns.ChangesetEventKindBitbucketServerApproved,
			campaigns.ChangesetEventKindBitbucketServerReviewed:

			s, err := e.ReviewState()
			if err != nil {
				return err
			}

			// We only care about "Approved", "ChangesRequested" or "Dismissed" reviews
			if s != campaigns.ChangesetReviewStateApproved &&
				s != campaigns.ChangesetReviewStateChangesRequested &&
				s != campaigns.ChangesetReviewStateDismissed {
				continue
			}

			author, err := e.ReviewAuthor()
			if err != nil {
				return err
			}
			if author == "" {
				continue
			}

			// Save current review state, then insert new review or delete
			// dismissed review, then recompute overall review state
			oldReviewState := currentReviewState

			if s == campaigns.ChangesetReviewStateDismissed {
				// In case of a dismissed review we dismiss _all_ of the
				// previous reviews by the author, since that is what GitHub
				// does in its UI.
				delete(lastReviewByAuthor, author)
			} else {
				lastReviewByAuthor[author] = s
			}

			newReviewState := computeReviewState(lastReviewByAuthor)

			if newReviewState != oldReviewState {
				// Decrement the counts increased by old review state
				c.AddReviewState(oldReviewState, -1)

				// Increase the counts for new review state
				c.AddReviewState(newReviewState, 1)
			}

		case campaigns.ChangesetEventKindBitbucketServerUnapproved:
			// We specifically ignore ChangesetEventKindGitHubReviewDismissed
			// events since GitHub updates the original
			// ChangesetEventKindGitHubReviewed event when a review has been
			// dismissed.

			author, err := e.ReviewAuthor()
			if err != nil {
				return err
			}
			if author == "" {
				continue
			}

			if e.Type() == campaigns.ChangesetEventKindBitbucketServerUnapproved {
				// A BitbucketServer Unapproved can only follow a previous Approved by
				// the same author.
				lastReview, ok := lastReviewByAuthor[author]
				if !ok || lastReview != campaigns.ChangesetReviewStateApproved {
					log15.Warn("Bitbucket Server Unapproval not following an Approval", "event", e)
					continue
				}
			}

			if e.Type() == campaigns.ChangesetEventKindGitHubReviewDismissed {
				// A GitHub Review Dismissed can only follow a previous review by
				// the author of the review included in the event.
				_, ok := lastReviewByAuthor[author]
				if !ok {
					log15.Warn("GitHub review dismissal not following a review", "event", e)
					continue
				}
			}

			// Save current review state, then remove last approval and
			// recompute overall review state
			oldReviewState := currentReviewState
			delete(lastReviewByAuthor, author)
			newReviewState := computeReviewState(lastReviewByAuthor)
			if newReviewState != oldReviewState {
				// Decrement the counts increased by old review state
				c.AddReviewState(oldReviewState, -1)

				// Increase the counts for new review state
				c.AddReviewState(newReviewState, 1)
			}
		}
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
