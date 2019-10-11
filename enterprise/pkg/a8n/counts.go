package a8n

import (
	"fmt"
	"sort"
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/a8n"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

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

type Event interface {
	Timestamp() time.Time
	Type() a8n.ChangesetEventKind
	Changeset() int64
}

type Events []Event

func (es Events) Len() int      { return len(es) }
func (es Events) Swap(i, j int) { es[i], es[j] = es[j], es[i] }

// Less sorts events by their timestamps
func (es Events) Less(i, j int) bool {
	return es[i].Timestamp().Before(es[j].Timestamp())
}

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
		openedAt, err := c.ExternalCreatedAt()
		if err != nil {
			return nil, err
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
	// Since some events cancel out another events effects we need to keep track of the
	// changesets state up until an event so we know what to revert
	// i.e. "merge" decrements OpenApproved counts, but only if
	// changeset was previously approved
	var (
		merged           = false
		closed           = false
		reviewed         = false
		approved         = false
		changesRequested = false
	)

	c.Total++
	c.Open++
	c.OpenPending++

	lastReviewByAuthor := map[string]a8n.ChangesetReviewState{}

	for _, e := range csEvents {
		// Event happened after point in time we're looking at, ignore
		et := e.Timestamp()
		if et.IsZero() || et.After(c.Time) {
			continue
		}

		switch e.Type() {
		case a8n.ChangesetEventKindGitHubClosed:
			// GitHub emits Closed/Merged events at the same time when a PR is
			// merged. We want to count that as a single Merged, not Closed
			// See: https://github.com/sourcegraph/sourcegraph/pull/5847#discussion_r332477806
			if merged {
				continue
			}
			c.Open--
			c.Closed++
			if !reviewed {
				c.OpenPending--
			}
			closed = true

		case a8n.ChangesetEventKindGitHubReopened:
			c.Open++
			c.Closed--
			if !reviewed {
				c.OpenPending++
			}
			closed = false

		case a8n.ChangesetEventKindGitHubMerged:
			// Reverse effects of closed for counting purposes
			if closed {
				c.Closed--
				c.Open++
				if !reviewed {
					c.OpenPending++
				}
			}
			if approved {
				c.OpenApproved--
			}
			if changesRequested {
				c.OpenChangesRequested--
			}
			if !reviewed {
				c.OpenPending--
			}
			c.Merged++
			c.Open--
			merged = true

		case a8n.ChangesetEventKindGitHubReviewed:
			s, err := reviewState(e)
			if err != nil {
				return err
			}

			author, err := reviewAuthor(e)
			if err != nil {
				return err
			}

			// Compute previous overall review state
			previousOverallState := computeReviewState(lastReviewByAuthor)

			// Insert new review, potentially replacing old review, but only if
			// it's not "PENDING" or "COMMENTED"
			if s == a8n.ChangesetReviewStateApproved || s == a8n.ChangesetReviewStateChangesRequested {
				lastReviewByAuthor[author] = s
			}

			// Compute new overall review state
			newOverallState := computeReviewState(lastReviewByAuthor)

			switch newOverallState {
			case a8n.ChangesetReviewStateApproved:
				switch previousOverallState {
				case a8n.ChangesetReviewStatePending:
					approved = true
					c.OpenApproved++
					reviewed = true
					c.OpenPending--
				case a8n.ChangesetReviewStateChangesRequested:
					changesRequested = false
					approved = true
					c.OpenChangesRequested--
					c.OpenApproved++
				}

			case a8n.ChangesetReviewStateChangesRequested:
				switch previousOverallState {
				case a8n.ChangesetReviewStatePending:
					changesRequested = true
					c.OpenChangesRequested++
					reviewed = true
					c.OpenPending--
				case a8n.ChangesetReviewStateApproved:
					approved = false
					changesRequested = true
					c.OpenChangesRequested++
					c.OpenApproved--
				}
			case a8n.ChangesetReviewStatePending:
			case a8n.ChangesetReviewStateCommented:
				// Ignore
			}
		}
	}

	return nil
}

func generateTimestamps(start, end time.Time) []time.Time {
	// Walk backwards from `end` to >= `start` in 1 day intervals
	// Backwards so we always end exactly on `end`
	ts := []time.Time{}
	for t := end; t.After(start) || t.Equal(start); t = t.Add(-24 * time.Hour) {
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
