package campaigns

import (
	"sort"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	cmpgn "github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
)

// ChangesetEvents is a collection of changeset events
type ChangesetEvents []*cmpgn.ChangesetEvent

func (ce ChangesetEvents) Len() int      { return len(ce) }
func (ce ChangesetEvents) Swap(i, j int) { ce[i], ce[j] = ce[j], ce[i] }

// Less sorts changeset events by their Timestamps
func (ce ChangesetEvents) Less(i, j int) bool {
	return ce[i].Timestamp().Before(ce[j].Timestamp())
}

// reviewState returns the overall review state of the review events in the
// slice.
// It should only be called by ComputeChangesetReviewState.
func (ce ChangesetEvents) reviewState() (cmpgn.ChangesetReviewState, error) {
	reviewsByAuthor := map[string]cmpgn.ChangesetReviewState{}

	for _, e := range ce {
		author, err := e.ReviewAuthor()
		if err != nil {
			return "", err
		}
		if author == "" {
			continue
		}
		s, err := e.ReviewState()
		if err != nil {
			return "", err
		}

		switch s {
		case cmpgn.ChangesetReviewStateApproved,
			cmpgn.ChangesetReviewStateChangesRequested:
			reviewsByAuthor[author] = s
		case cmpgn.ChangesetReviewStateDismissed:
			delete(reviewsByAuthor, author)
		}
	}

	return computeReviewState(reviewsByAuthor), nil
}

// State returns the  state of the changeset to which the events belong and assumes the events
// are sorted by ChangesetEvent.Timestamp().
func (ce ChangesetEvents) State() cmpgn.ChangesetState {
	state := cmpgn.ChangesetStateOpen
	for _, e := range ce {
		switch e.Kind {
		case cmpgn.ChangesetEventKindGitHubClosed, cmpgn.ChangesetEventKindBitbucketServerDeclined:
			state = cmpgn.ChangesetStateClosed
		case cmpgn.ChangesetEventKindGitHubMerged, cmpgn.ChangesetEventKindBitbucketServerMerged:
			// Merged is a final state. We can ignore everything after.
			return cmpgn.ChangesetStateMerged
		case cmpgn.ChangesetEventKindGitHubReopened, cmpgn.ChangesetEventKindBitbucketServerReopened:
			state = cmpgn.ChangesetStateOpen
		}
	}
	return state
}

type changesetStatesAtTime struct {
	t           time.Time
	state       cmpgn.ChangesetState
	reviewState cmpgn.ChangesetReviewState
}

func computeHistory(ch *cmpgn.Changeset, ce ChangesetEvents) ([]changesetStatesAtTime, error) {
	states := []changesetStatesAtTime{}

	currentState := cmpgn.ChangesetStateOpen
	currentReviewState := cmpgn.ChangesetReviewStatePending
	lastReviewByAuthor := map[string]campaigns.ChangesetReviewState{}

	pushStates := func(t time.Time) {
		states = append(states, changesetStatesAtTime{
			t:           t,
			state:       currentState,
			reviewState: currentReviewState,
		})
	}

	openedAt := ch.ExternalCreatedAt()
	if openedAt.IsZero() {
		return nil, errors.New("changeset ExternalCreatedAt has zero value")
	}
	pushStates(openedAt)

	for _, e := range ce {
		et := e.Timestamp()
		if et.IsZero() {
			continue
		}

		switch e.Kind {
		case cmpgn.ChangesetEventKindGitHubClosed, cmpgn.ChangesetEventKindBitbucketServerDeclined:
			// Merged is a final state. We can ignore everything after.
			if currentState != cmpgn.ChangesetStateMerged {
				currentState = cmpgn.ChangesetStateClosed
				pushStates(et)
			}

		case cmpgn.ChangesetEventKindGitHubMerged, cmpgn.ChangesetEventKindBitbucketServerMerged:
			currentState = cmpgn.ChangesetStateMerged
			pushStates(et)

		case cmpgn.ChangesetEventKindGitHubReopened, cmpgn.ChangesetEventKindBitbucketServerReopened:
			// Merged is a final state. We can ignore everything after.
			if currentState != cmpgn.ChangesetStateMerged {
				currentState = cmpgn.ChangesetStateOpen
				pushStates(et)
			}

		case campaigns.ChangesetEventKindGitHubReviewed,
			campaigns.ChangesetEventKindBitbucketServerApproved,
			campaigns.ChangesetEventKindBitbucketServerReviewed:

			s, err := e.ReviewState()
			if err != nil {
				return nil, err
			}

			// We only care about "Approved", "ChangesRequested" or "Dismissed" reviews
			if s != campaigns.ChangesetReviewStateApproved &&
				s != campaigns.ChangesetReviewStateChangesRequested &&
				s != campaigns.ChangesetReviewStateDismissed {
				continue
			}

			author, err := e.ReviewAuthor()
			if err != nil {
				return nil, err
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
				currentReviewState = newReviewState
				pushStates(et)
			}

		case campaigns.ChangesetEventKindBitbucketServerUnapproved:
			// We specifically ignore ChangesetEventKindGitHubReviewDismissed
			// events since GitHub updates the original
			// ChangesetEventKindGitHubReviewed event when a review has been
			// dismissed.

			author, err := e.ReviewAuthor()
			if err != nil {
				return nil, err
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
				currentReviewState = newReviewState
				pushStates(et)
			}
		}
	}

	return states, nil
}

// UpdateLabelsSince returns the set of current labels based the starting set of labels and looking at events
// that have occurred after "since".
func (ce *ChangesetEvents) UpdateLabelsSince(cs *cmpgn.Changeset) []cmpgn.ChangesetLabel {
	var current []cmpgn.ChangesetLabel
	var since time.Time
	if cs != nil {
		current = cs.Labels()
		since = cs.UpdatedAt
	}
	// Copy slice so that we don't mutate ce
	sorted := make(ChangesetEvents, len(*ce))
	copy(sorted, *ce)
	sort.Sort(sorted)

	// Iterate through all label events to get the current set
	set := make(map[string]cmpgn.ChangesetLabel)
	for _, l := range current {
		set[l.Name] = l
	}
	for _, event := range sorted {
		switch e := event.Metadata.(type) {
		case *github.LabelEvent:
			if e.CreatedAt.Before(since) {
				continue
			}
			if e.Removed {
				delete(set, e.Label.Name)
				continue
			}
			set[e.Label.Name] = cmpgn.ChangesetLabel{
				Name:        e.Label.Name,
				Color:       e.Label.Color,
				Description: e.Label.Description,
			}
		}
	}
	labels := make([]cmpgn.ChangesetLabel, 0, len(set))
	for _, label := range set {
		labels = append(labels, label)
	}
	return labels
}
