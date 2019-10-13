package a8n

import (
	"fmt"
	"sort"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
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

// IncReviewStateCount increments the corresponding count for a given
// ChangesetReviewState
func (c *ChangesetCounts) IncReviewStateCount(s a8n.ChangesetReviewState) {
	switch s {
	case a8n.ChangesetReviewStatePending:
		c.OpenPending++
	case a8n.ChangesetReviewStateApproved:
		c.OpenApproved++
	case a8n.ChangesetReviewStateChangesRequested:
		c.OpenChangesRequested++
	}
}

// IncReviewStateCount decrements the corresponding count for a given
// ChangesetReviewState
func (c *ChangesetCounts) DecReviewStateCount(s a8n.ChangesetReviewState) {
	switch s {
	case a8n.ChangesetReviewStatePending:
		c.OpenPending--
	case a8n.ChangesetReviewStateApproved:
		c.OpenApproved--
	case a8n.ChangesetReviewStateChangesRequested:
		c.OpenChangesRequested--
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

// Event is a single event that happened in the lifetime of a single Changeset,
// for example a review or a merge.
type Event interface {
	Timestamp() time.Time
	Type() a8n.ChangesetEventKind
	Changeset() int64
}

// Events is a collection of Events that can be sorted by their Timestamps
type Events []Event

func (es Events) Len() int      { return len(es) }
func (es Events) Swap(i, j int) { es[i], es[j] = es[j], es[i] }

// Less sorts events by their timestamps
func (es Events) Less(i, j int) bool {
	return es[i].Timestamp().Before(es[j].Timestamp())
}

// CalcCounts calculates ChangesetCounts for the given Changesets and their
// Events in the timeframe specified by the start and end parameters.
// The number of ChangesetCounts returns is the number of 1 day intervals
// between start and end, with each ChangesetCounts representing a point in
// time.
func CalcCounts(start, end time.Time, cs []*a8n.Changeset, es ...Event) ([]*ChangesetCounts, error) {
	ts := generateTimestamps(start, end)
	counts := make([]*ChangesetCounts, len(ts))
	for i, t := range ts {
		counts[i] = &ChangesetCounts{Time: t}
	}

	// Sort all events once by their timestamps
	events := Events(es)
	sort.Sort(events)

	// Map sorted events to their changesets
	byChangeset := make(map[*a8n.Changeset]Events)
	for _, c := range cs {
		group := Events{}
		for _, e := range events {
			if e.Changeset() == c.ID {
				group = append(group, e)
			}
		}
		byChangeset[c] = group
	}

	for c, csEvents := range byChangeset {
		// We don't have an event for "open", so we check when it was
		// created on codehost
		openedAt := c.ExternalCreatedAt()
		if openedAt.IsZero() {
			continue
		}

		// For each changeset and its events, go through every point in time we
		// want to record and reconstruct the state of the changeset at that
		// point in time
		for _, c := range counts {
			if openedAt.After(c.Time) {
				// No need to look at events if changeset was not created yet
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

func computeCounts(c *ChangesetCounts, csEvents Events) error {
	var (
		// Since "Merged" and "Closed" are exclusive events and cancel each others
		// effects on ChangesetCounts out, we need to keep track of when a
		// changeset was closed, so we can undo the effect of the "Closed" event
		// when we come across a "Merge" (since, on GitHub, a PR can be closed AND
		// merged)
		closed = false

		lastReviewByAuthor = map[string]a8n.ChangesetReviewState{}
	)

	c.Total++
	c.Open++
	c.OpenPending++

	for _, e := range csEvents {
		// Event happened after point in time we're looking at, no need to look
		// at the events in future
		et := e.Timestamp()
		if et.IsZero() || et.After(c.Time) {
			return nil
		}

		// Compute current overall review state
		currentReviewState := computeReviewState(lastReviewByAuthor)

		switch e.Type() {
		case a8n.ChangesetEventKindGitHubClosed:
			c.Open--
			c.Closed++
			closed = true

			c.DecReviewStateCount(currentReviewState)

		case a8n.ChangesetEventKindGitHubReopened:
			c.Open++
			c.Closed--
			closed = false

			c.IncReviewStateCount(currentReviewState)

		case a8n.ChangesetEventKindGitHubMerged:
			// If it was closed, all "review counts" have been updated by the
			// closed events and we just need to reverse these two counts
			if closed {
				c.Closed--
				c.Merged++
				return nil
			}

			c.DecReviewStateCount(currentReviewState)

			c.Merged++
			c.Open--

			// Merged is a final state, we return here and don't need to look at
			// other events
			return nil

		case a8n.ChangesetEventKindGitHubReviewed:
			s, err := reviewState(e)
			if err != nil {
				return err
			}

			// We only care about "Approved" or "ChangesRequested" reviews
			if s != a8n.ChangesetReviewStateApproved && s != a8n.ChangesetReviewStateChangesRequested {
				continue
			}

			author, err := reviewAuthor(e)
			if err != nil {
				return err
			}

			// Save current review state, then insert new review and recompute
			// overall review state
			oldReviewState := currentReviewState
			lastReviewByAuthor[author] = s
			newReviewState := computeReviewState(lastReviewByAuthor)

			if newReviewState != oldReviewState {
				// Decrement the counts increased by old review state
				c.DecReviewStateCount(oldReviewState)

				// Increase the counts for new review state
				c.IncReviewStateCount(newReviewState)
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

func reviewState(e Event) (a8n.ChangesetReviewState, error) {
	var s a8n.ChangesetReviewState
	changesetEvent, ok := e.(*a8n.ChangesetEvent)
	if !ok {
		return s, errors.New("Reviewed event not ChangesetEvent")
	}

	review, ok := changesetEvent.Metadata.(*github.PullRequestReview)
	if !ok {
		return s, errors.New("ChangesetEvent metadata event not PullRequestReview")
	}

	s = a8n.ChangesetReviewState(review.State)
	if !s.Valid() {
		return s, fmt.Errorf("invalid review state: %s", review.State)
	}
	return s, nil
}

func reviewAuthor(e Event) (string, error) {
	changesetEvent, ok := e.(*a8n.ChangesetEvent)
	if !ok {
		return "", errors.New("Reviewed event not ChangesetEvent")
	}

	review, ok := changesetEvent.Metadata.(*github.PullRequestReview)
	if !ok {
		return "", errors.New("ChangesetEvent metadata event not PullRequestReview")
	}

	login := review.Author.Login
	if login == "" {
		return "", errors.New("review author is blank")
	}

	return login, nil
}

func computeReviewState(statesByAuthor map[string]a8n.ChangesetReviewState) a8n.ChangesetReviewState {
	states := make(map[a8n.ChangesetReviewState]bool)
	for _, s := range statesByAuthor {
		states[s] = true
	}
	return a8n.SelectReviewState(states)
}
