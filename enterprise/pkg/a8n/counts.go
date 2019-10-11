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
	Timestamp() (time.Time, error)
	Type() a8n.ChangesetEventKind
	Changeset() int64
}

type Events []Event

func (es Events) Len() int      { return len(es) }
func (es Events) Swap(i, j int) { es[i], es[j] = es[j], es[i] }

// Less sorts events by their timestamps
func (es Events) Less(i, j int) bool {
	t1, _ := es[i].Timestamp()
	t2, _ := es[j].Timestamp()
	return t1.Before(t2)
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
		for _, count := range counts {
			t := count.Time

			if openedAt.Before(t) || openedAt.Equal(t) {
				count.Total++
				count.Open++
				count.OpenPending++
			} else {
				// No need to look at events if changeset was not created yet
				continue
			}

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

			for _, e := range csEvents {
				et, err := e.Timestamp()
				if err != nil {
					return nil, err
				}
				// Event happened after point in time we're looking at, ignore
				if et.After(t) {
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
					count.Open--
					count.Closed++
					if !reviewed {
						count.OpenPending--
					}
					closed = true

				case a8n.ChangesetEventKindGitHubReopened:
					count.Open++
					count.Closed--
					if !reviewed {
						count.OpenPending++
					}
					closed = false

				case a8n.ChangesetEventKindGitHubMerged:
					// Reverse effects of closed for counting purposes
					if closed {
						count.Closed--
						count.Open++
						if !reviewed {
							count.OpenPending++
						}
					}
					if approved {
						count.OpenApproved--
					}
					if changesRequested {
						count.OpenChangesRequested--
					}
					if !reviewed {
						count.OpenPending--
					}
					count.Merged++
					count.Open--
					merged = true

				case a8n.ChangesetEventKindGitHubReviewed:
					s, err := reviewState(e)
					if err != nil {
						return nil, err
					}

					switch s {
					case a8n.ChangesetReviewStateApproved:
						if !approved {
							approved = true
							count.OpenApproved++
							reviewed = true
							count.OpenPending--
						}
					case a8n.ChangesetReviewStateChangesRequested:
						if !changesRequested {
							changesRequested = true
							count.OpenChangesRequested++
							reviewed = true
							count.OpenPending--
						}
					case a8n.ChangesetReviewStatePending:
					case a8n.ChangesetReviewStateCommented:
						// Ignore
					}
				}
			}
		}
	}

	return counts, nil
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
