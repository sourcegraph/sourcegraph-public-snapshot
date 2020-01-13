package a8n

import (
	"fmt"
	"sort"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
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

// AddReviewState adds n to the corresponding counter for a given
// ChangesetReviewState
func (c *ChangesetCounts) AddReviewState(s a8n.ChangesetReviewState, n int32) {
	switch s {
	case a8n.ChangesetReviewStatePending:
		c.OpenPending += n
	case a8n.ChangesetReviewStateApproved:
		c.OpenApproved += n
	case a8n.ChangesetReviewStateChangesRequested:
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
// Events in the timeframe specified by the start and end parameters. The
// number of ChangesetCounts returned is the number of 1 day intervals between
// start and end, with each ChangesetCounts representing a point in time at the
// boundary of each 24h interval.
func CalcCounts(start, end time.Time, cs []*a8n.Changeset, es ...Event) ([]*ChangesetCounts, error) {
	ts := generateTimestamps(start, end)
	counts := make([]*ChangesetCounts, len(ts))
	for i, t := range ts {
		counts[i] = &ChangesetCounts{Time: t}
	}

	// Sort all events once by their timestamps
	events := Events(es)
	sort.Sort(events)

	// Grouping Events by their Changeset ID
	byChangesetID := make(map[int64]Events)
	for _, e := range events {
		id := e.Changeset()
		byChangesetID[id] = append(byChangesetID[id], e)
	}

	// Map Events to their Changeset
	byChangeset := make(map[*a8n.Changeset]Events)
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
		case a8n.ChangesetEventKindGitHubClosed,
			a8n.ChangesetEventKindBitbucketServerDeclined:

			c.Open--
			c.Closed++
			closed = true

			c.AddReviewState(currentReviewState, -1)

		case a8n.ChangesetEventKindGitHubReopened,
			a8n.ChangesetEventKindBitbucketServerReopened:

			c.Open++
			c.Closed--
			closed = false

			c.AddReviewState(currentReviewState, 1)

		case a8n.ChangesetEventKindGitHubMerged,
			a8n.ChangesetEventKindBitbucketServerMerged:

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

		case a8n.ChangesetEventKindGitHubReviewed,
			a8n.ChangesetEventKindBitbucketServerApproved,
			a8n.ChangesetEventKindBitbucketServerReviewed:

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
				c.AddReviewState(oldReviewState, -1)

				// Increase the counts for new review state
				c.AddReviewState(newReviewState, 1)
			}

		case a8n.ChangesetEventKindBitbucketServerUnapproved:
			author, err := reviewAuthor(e)
			if err != nil {
				return err
			}

			// A BitbucketServer Unapproved can only follow a previous Approved by
			// the same author.
			lastReview, ok := lastReviewByAuthor[author]
			if !ok || lastReview != a8n.ChangesetReviewStateApproved {
				return errors.New("Bitbucket Server Unapproval not following an Approval")
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

func reviewState(e Event) (a8n.ChangesetReviewState, error) {
	var s a8n.ChangesetReviewState
	changesetEvent, ok := e.(*a8n.ChangesetEvent)
	if !ok {
		return s, errors.New("Reviewed event not ChangesetEvent")
	}

	switch changesetEvent.Kind {
	case a8n.ChangesetEventKindBitbucketServerApproved:
		return a8n.ChangesetReviewStateApproved, nil

	case a8n.ChangesetEventKindBitbucketServerReviewed:
		return a8n.ChangesetReviewStateChangesRequested, nil

	case a8n.ChangesetEventKindGitHubReviewed:
		review, ok := changesetEvent.Metadata.(*github.PullRequestReview)
		if !ok {
			return s, errors.New("ChangesetEvent metadata event not PullRequestReview")
		}

		s = a8n.ChangesetReviewState(review.State)
		if !s.Valid() {
			return s, fmt.Errorf("invalid review state: %s", review.State)
		}
		return s, nil

	default:
		return s, fmt.Errorf("unsupported changeset event kind: %s", changesetEvent.Kind)
	}
}

func reviewAuthor(e Event) (string, error) {
	changesetEvent, ok := e.(*a8n.ChangesetEvent)
	if !ok {
		return "", errors.New("Reviewed event not ChangesetEvent")
	}

	switch meta := changesetEvent.Metadata.(type) {
	case *github.PullRequestReview:
		login := meta.Author.Login
		if login == "" {
			return "", errors.New("review author is blank")
		}
		return login, nil

	case *bitbucketserver.Activity:
		username := meta.User.Name
		if username == "" {
			return "", errors.New("activity user is blank")
		}
		return username, nil
	default:
		return "", errors.New("ChangesetEvent metadata is of unsupported type")
	}
}

func computeReviewState(statesByAuthor map[string]a8n.ChangesetReviewState) a8n.ChangesetReviewState {
	states := make(map[a8n.ChangesetReviewState]bool)
	for _, s := range statesByAuthor {
		states[s] = true
	}
	return a8n.SelectReviewState(states)
}
