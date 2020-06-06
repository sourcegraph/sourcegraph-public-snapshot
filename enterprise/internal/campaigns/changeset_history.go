package campaigns

import (
	"sort"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	cmpgn "github.com/sourcegraph/sourcegraph/internal/campaigns"
)

// changesetHistory is a collection of a changesets states (open/closed/merged
// state and review state) over time.
type changesetHistory []changesetStatesAtTime

// StatesAtTime returns the changeset's states valid at the given time. If the
// changeset didn't exist yet, the second parameter is false.
func (h changesetHistory) StatesAtTime(t time.Time) (changesetStatesAtTime, bool) {
	if len(h) == 0 {
		return changesetStatesAtTime{}, false
	}

	var (
		states changesetStatesAtTime
		found  bool
	)

	for _, s := range h {
		if s.t.After(t) {
			break
		}
		states = s
		found = true
	}

	return states, found
}

type changesetStatesAtTime struct {
	t           time.Time
	state       cmpgn.ChangesetState
	reviewState cmpgn.ChangesetReviewState
}

// computeHistory calculates the changesetHistory for the given Changeset and
// its ChangesetEvents.
// The ChangesetEvents MUST be sorted by their Timestamp.
func computeHistory(ch *cmpgn.Changeset, ce ChangesetEvents) (changesetHistory, error) {
	if !sort.IsSorted(ce) {
		return nil, errors.New("changeset events not sorted")
	}

	var (
		states = []changesetStatesAtTime{}

		currentState       = cmpgn.ChangesetStateOpen
		currentReviewState = cmpgn.ChangesetReviewStatePending

		lastReviewByAuthor = map[string]campaigns.ChangesetReviewState{}
	)

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

		case campaigns.ChangesetEventKindGitHubReviewDismissed:
			// We specifically ignore ChangesetEventKindGitHubReviewDismissed
			// events since GitHub updates the original
			// ChangesetEventKindGitHubReviewed event when a review has been
			// dismissed.
			// See: https://github.com/sourcegraph/sourcegraph/pull/9461
			continue

		case campaigns.ChangesetEventKindBitbucketServerUnapproved,
			campaigns.ChangesetEventKindBitbucketServerDismissed:
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

			if e.Type() == campaigns.ChangesetEventKindBitbucketServerDismissed {
				// A BitbucketServer Dismissed event can only follow a previous "Changes Requested" review by
				// the same author.
				lastReview, ok := lastReviewByAuthor[author]
				if !ok || lastReview != campaigns.ChangesetReviewStateChangesRequested {
					log15.Warn("Bitbucket Server Dismissal not following a Review", "event", e)
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

	// We don't have an event for the deletion of a Changeset, but we set
	// ExternalDeletedAt manually in the Syncer.
	deletedAt := ch.ExternalDeletedAt
	if !deletedAt.IsZero() {
		currentState = cmpgn.ChangesetStateClosed
		pushStates(deletedAt)
	}

	return states, nil
}
