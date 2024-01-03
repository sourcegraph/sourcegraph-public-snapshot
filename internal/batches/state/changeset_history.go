package state

import (
	"sort"
	"time"

	"github.com/inconshreveable/log15" //nolint:logging // TODO move all logging to sourcegraph/log

	adobatches "github.com/sourcegraph/sourcegraph/internal/batches/sources/azuredevops"
	gerritbatches "github.com/sourcegraph/sourcegraph/internal/batches/sources/gerrit"

	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// changesetHistory is a collection of external changeset states
// (open/closed/merged state and review state) over time.
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

// RequiredEventTypesForHistory keeps track of all event kinds required for calculating the history of a changeset.
//
// We specifically ignore ChangesetEventKindGitHubReviewDismissed
// events since GitHub updates the original
// ChangesetEventKindGitHubReviewed event when a review has been
// dismissed.
// See: https://github.com/sourcegraph/sourcegraph/pull/9461
var RequiredEventTypesForHistory = []btypes.ChangesetEventKind{
	// Undraft.
	btypes.ChangesetEventKindGitHubReadyForReview,
	btypes.ChangesetEventKindGitLabUnmarkWorkInProgress,

	// Draft.
	btypes.ChangesetEventKindGitHubConvertToDraft,
	btypes.ChangesetEventKindGitLabMarkWorkInProgress,

	// Closed, unmerged.
	btypes.ChangesetEventKindBitbucketCloudPullRequestRejected,
	btypes.ChangesetEventKindBitbucketServerDeclined,
	btypes.ChangesetEventKindGitHubClosed,
	btypes.ChangesetEventKindGitLabClosed,

	// Closed, merged.
	btypes.ChangesetEventKindBitbucketCloudPullRequestFulfilled,
	btypes.ChangesetEventKindBitbucketServerMerged,
	btypes.ChangesetEventKindGitHubMerged,
	btypes.ChangesetEventKindGitLabMerged,
	btypes.ChangesetEventKindAzureDevOpsPullRequestMerged,

	// Reopened
	btypes.ChangesetEventKindBitbucketServerReopened,
	btypes.ChangesetEventKindGitHubReopened,
	btypes.ChangesetEventKindGitLabReopened,

	// Reviewed, indeterminate status.
	btypes.ChangesetEventKindGitHubReviewed,

	// Reviewed, approved.
	btypes.ChangesetEventKindBitbucketCloudApproved,
	btypes.ChangesetEventKindBitbucketCloudPullRequestApproved,
	btypes.ChangesetEventKindBitbucketServerApproved,
	btypes.ChangesetEventKindBitbucketServerReviewed,
	btypes.ChangesetEventKindGitLabApproved,
	btypes.ChangesetEventKindAzureDevOpsPullRequestApproved,
	btypes.ChangesetEventKindAzureDevOpsPullRequestApprovedWithSuggestions,

	// Reviewed, not approved.
	btypes.ChangesetEventKindBitbucketCloudPullRequestChangesRequestRemoved,
	btypes.ChangesetEventKindBitbucketCloudPullRequestUnapproved,
	btypes.ChangesetEventKindBitbucketServerUnapproved,
	btypes.ChangesetEventKindBitbucketServerDismissed,
	btypes.ChangesetEventKindGitLabUnapproved,
	btypes.ChangesetEventKindAzureDevOpsPullRequestWaitingForAuthor,
	btypes.ChangesetEventKindAzureDevOpsPullRequestRejected,
}

type changesetStatesAtTime struct {
	t             time.Time
	externalState btypes.ChangesetExternalState
	reviewState   btypes.ChangesetReviewState
}

// computeHistory calculates the changesetHistory for the given Changeset and
// its ChangesetEvents.
// The ChangesetEvents MUST be sorted by their Timestamp.
func computeHistory(ch *btypes.Changeset, ce ChangesetEvents) (changesetHistory, error) {
	if !sort.IsSorted(ce) {
		return nil, errors.New("changeset events not sorted")
	}

	var (
		states = []changesetStatesAtTime{}

		currentExtState    = initialExternalState(ch, ce)
		currentReviewState = btypes.ChangesetReviewStatePending

		lastReviewByAuthor = map[string]btypes.ChangesetReviewState{}
		// The draft state is tracked alongside the "external state" on GitHub and GitLab,
		// that means we need to take changes to this state into account separately. On reopen,
		// we cannot simply say it's open, because it could be it was converted to a draft while
		// it was closed. Hence, we need to track the state using this variable.
		isDraft = currentExtState == btypes.ChangesetExternalStateDraft
	)

	pushStates := func(t time.Time) {
		states = append(states, changesetStatesAtTime{
			t:             t,
			externalState: currentExtState,
			reviewState:   currentReviewState,
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

		// NOTE: If you add any kinds here, make sure they also appear in `RequiredEventTypesForHistory`.
		switch e.Kind {
		case btypes.ChangesetEventKindGitHubClosed,
			btypes.ChangesetEventKindBitbucketServerDeclined,
			btypes.ChangesetEventKindGitLabClosed,
			btypes.ChangesetEventKindBitbucketCloudPullRequestRejected:
			// Merged and ReadOnly are final states. We can ignore everything after.
			if currentExtState != btypes.ChangesetExternalStateMerged &&
				currentExtState != btypes.ChangesetExternalStateReadOnly {
				currentExtState = btypes.ChangesetExternalStateClosed
				pushStates(et)
			}

		case btypes.ChangesetEventKindGitHubMerged,
			btypes.ChangesetEventKindBitbucketServerMerged,
			btypes.ChangesetEventKindGitLabMerged,
			btypes.ChangesetEventKindBitbucketCloudPullRequestFulfilled,
			btypes.ChangesetEventKindAzureDevOpsPullRequestMerged:
			currentExtState = btypes.ChangesetExternalStateMerged
			pushStates(et)

		case btypes.ChangesetEventKindGitLabMarkWorkInProgress:
			isDraft = true
			// This event only matters when the changeset is open, otherwise a change in the title won't change the overall external state.
			if currentExtState == btypes.ChangesetExternalStateOpen {
				currentExtState = btypes.ChangesetExternalStateDraft
				pushStates(et)
			}

		case btypes.ChangesetEventKindGitHubConvertToDraft:
			isDraft = true
			// Merged and ReadOnly are final states. We can ignore everything after.
			if currentExtState != btypes.ChangesetExternalStateMerged &&
				currentExtState != btypes.ChangesetExternalStateReadOnly {
				currentExtState = btypes.ChangesetExternalStateDraft
				pushStates(et)
			}

		case btypes.ChangesetEventKindGitLabUnmarkWorkInProgress,
			btypes.ChangesetEventKindGitHubReadyForReview:
			isDraft = false
			// This event only matters when the changeset is open, otherwise a change in the title won't change the overall external state.
			if currentExtState == btypes.ChangesetExternalStateDraft {
				currentExtState = btypes.ChangesetExternalStateOpen
				pushStates(et)
			}

		case btypes.ChangesetEventKindGitHubReopened,
			btypes.ChangesetEventKindBitbucketServerReopened,
			btypes.ChangesetEventKindGitLabReopened:
			// Merged and ReadOnly are final states. We can ignore everything after.
			if currentExtState != btypes.ChangesetExternalStateMerged &&
				currentExtState != btypes.ChangesetExternalStateReadOnly {
				if isDraft {
					currentExtState = btypes.ChangesetExternalStateDraft
				} else {
					currentExtState = btypes.ChangesetExternalStateOpen
				}
				pushStates(et)
			}

		case btypes.ChangesetEventKindGitHubReviewed,
			btypes.ChangesetEventKindBitbucketServerApproved,
			btypes.ChangesetEventKindBitbucketServerReviewed,
			btypes.ChangesetEventKindGitLabApproved,
			btypes.ChangesetEventKindBitbucketCloudApproved,
			btypes.ChangesetEventKindBitbucketCloudPullRequestApproved,
			btypes.ChangesetEventKindAzureDevOpsPullRequestApproved:
			s, err := e.ReviewState()
			if err != nil {
				return nil, err
			}

			// We only care about "Approved", "ChangesRequested" or "Dismissed" reviews
			if s != btypes.ChangesetReviewStateApproved &&
				s != btypes.ChangesetReviewStateChangesRequested &&
				s != btypes.ChangesetReviewStateDismissed {
				continue
			}

			author := e.ReviewAuthor()
			// If the user has been deleted, skip their reviews, as they don't count towards the final state anymore.
			if author == "" {
				continue
			}

			// Save current review state, then insert new review or delete
			// dismissed review, then recompute overall review state
			oldReviewState := currentReviewState

			if s == btypes.ChangesetReviewStateDismissed {
				// In case of a dismissed review we dismiss _all_ of the
				// previous reviews by the author, since that is what GitHub
				// does in its UI.
				delete(lastReviewByAuthor, author)
			} else {
				lastReviewByAuthor[author] = s
			}

			newReviewState := reduceReviewStates(lastReviewByAuthor)

			if newReviewState != oldReviewState {
				currentReviewState = newReviewState
				pushStates(et)
			}

		case btypes.ChangesetEventKindBitbucketServerUnapproved,
			btypes.ChangesetEventKindBitbucketServerDismissed,
			btypes.ChangesetEventKindGitLabUnapproved,
			btypes.ChangesetEventKindBitbucketCloudPullRequestChangesRequestRemoved,
			btypes.ChangesetEventKindBitbucketCloudPullRequestUnapproved:

			author := e.ReviewAuthor()
			// If the user has been deleted, skip their reviews, as they don't count towards the final state anymore.
			if author == "" {
				continue
			}

			if e.Type() == btypes.ChangesetEventKindBitbucketServerUnapproved {
				// A BitbucketServer Unapproved can only follow a previous Approved by
				// the same author.
				lastReview, ok := lastReviewByAuthor[author]
				if !ok || lastReview != btypes.ChangesetReviewStateApproved {
					log15.Warn("Bitbucket Server Unapproval not following an Approval", "event", e)
					continue
				}
			}

			if e.Type() == btypes.ChangesetEventKindBitbucketServerDismissed {
				// A BitbucketServer Dismissed event can only follow a previous "Changes Requested" review by
				// the same author.
				lastReview, ok := lastReviewByAuthor[author]
				if !ok || lastReview != btypes.ChangesetReviewStateChangesRequested {
					log15.Warn("Bitbucket Server Dismissal not following a Review", "event", e)
					continue
				}
			}

			// Save current review state, then remove last approval and
			// recompute overall review state
			oldReviewState := currentReviewState
			delete(lastReviewByAuthor, author)
			newReviewState := reduceReviewStates(lastReviewByAuthor)

			if newReviewState != oldReviewState {
				currentReviewState = newReviewState
				pushStates(et)
			}
		case btypes.ChangesetEventKindAzureDevOpsPullRequestRejected,
			btypes.ChangesetEventKindAzureDevOpsPullRequestApprovedWithSuggestions,
			btypes.ChangesetEventKindAzureDevOpsPullRequestWaitingForAuthor:
			currentReviewState = btypes.ChangesetReviewStateChangesRequested
			author := e.ReviewAuthor()
			lastReviewByAuthor[author] = currentReviewState
			pushStates(et)
		}
	}

	// We don't have an event for the deletion of a Changeset, but we set
	// ExternalDeletedAt manually in the Syncer.
	deletedAt := ch.ExternalDeletedAt
	if !deletedAt.IsZero() {
		currentExtState = btypes.ChangesetExternalStateClosed
		pushStates(deletedAt)
	}

	return states, nil
}

// reduceReviewStates reduces the given a map of review per author down to a
// single overall ChangesetReviewState.
func reduceReviewStates(statesByAuthor map[string]btypes.ChangesetReviewState) btypes.ChangesetReviewState {
	states := make(map[btypes.ChangesetReviewState]bool)
	for _, s := range statesByAuthor {
		states[s] = true
	}
	return selectReviewState(states)
}

// initialExternalState infers from the changeset state and the list of events in which
// ChangesetExternalState the changeset must have been when it has been created.
func initialExternalState(ch *btypes.Changeset, ce ChangesetEvents) btypes.ChangesetExternalState {
	open := true
	switch m := ch.Metadata.(type) {
	case *github.PullRequest:
		if m.IsDraft {
			open = false
		}

	case *gitlab.MergeRequest:
		if m.WorkInProgress {
			open = false
		}
	case *adobatches.AnnotatedPullRequest:
		if m.IsDraft {
			open = false
		}
	case *gerritbatches.AnnotatedChange:
		if m.Change.WorkInProgress {
			open = false
		}
	default:
		return btypes.ChangesetExternalStateOpen
	}
	// Walk the events backwards, since we need to look from the current time to the past.
	for i := len(ce) - 1; i >= 0; i-- {
		e := ce[i]
		switch e.Metadata.(type) {
		case *gitlab.UnmarkWorkInProgressEvent, *github.ReadyForReviewEvent:
			open = false
		case *gitlab.MarkWorkInProgressEvent, *github.ConvertToDraftEvent:
			open = true
		}
	}
	if open {
		return btypes.ChangesetExternalStateOpen
	}
	return btypes.ChangesetExternalStateDraft
}
