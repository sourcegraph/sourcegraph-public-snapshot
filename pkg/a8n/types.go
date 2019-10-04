package a8n

import (
	"time"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
)

// A Campaign of changesets over multiple Repos over time.
type Campaign struct {
	ID              int64
	Name            string
	Description     string
	AuthorID        int32
	NamespaceUserID int32
	NamespaceOrgID  int32
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ChangesetIDs    []int64
}

// Clone returns a clone of a Campaign.
func (c *Campaign) Clone() *Campaign {
	cc := *c
	cc.ChangesetIDs = c.ChangesetIDs[:len(c.ChangesetIDs):len(c.ChangesetIDs)]
	return &cc
}

// ChangesetState defines the possible states of a Changeset.
type ChangesetState string

// ChangesetState constants.
const (
	ChangesetStateOpen   ChangesetState = "OPEN"
	ChangesetStateClosed ChangesetState = "CLOSED"
	ChangesetStateMerged ChangesetState = "MERGED"
)

// Valid returns true if the given Changeset is valid.
func (s ChangesetState) Valid() bool {
	switch s {
	case ChangesetStateOpen,
		ChangesetStateClosed,
		ChangesetStateMerged:
		return true
	default:
		return false
	}
}

// ChangesetReviewState defines the possible states of a Changeset's review.
type ChangesetReviewState string

// ChangesetReviewState constants.
const (
	ChangesetReviewStateApproved         ChangesetReviewState = "APPROVED"
	ChangesetReviewStateChangesRequested ChangesetReviewState = "CHANGES_REQUESTED"
	ChangesetReviewStatePending          ChangesetReviewState = "PENDING"
)

// Valid returns true if the given Changeset is valid.
func (s ChangesetReviewState) Valid() bool {
	switch s {
	case ChangesetReviewStateApproved,
		ChangesetReviewStateChangesRequested,
		ChangesetReviewStatePending:
		return true
	default:
		return false
	}
}

// A Changeset is a changeset on a code host belonging to a Repository and many
// Campaigns.
type Changeset struct {
	ID                  int64
	RepoID              int32
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Metadata            interface{}
	CampaignIDs         []int64
	ExternalID          string
	ExternalServiceType string
	Events              []*ChangesetEvent
}

// Clone returns a clone of a Changeset.
func (t *Changeset) Clone() *Changeset {
	tt := *t
	tt.CampaignIDs = t.CampaignIDs[:len(t.CampaignIDs):len(t.CampaignIDs)]
	return &tt
}

// Title of the Changeset.
func (t *Changeset) Title() (string, error) {
	switch m := t.Metadata.(type) {
	case *github.PullRequest:
		return m.Title, nil
	case *bitbucketserver.PullRequest:
		return m.Title, nil
	default:
		return "", errors.New("unknown changeset type")
	}
}

// Body of the Changeset.
func (t *Changeset) Body() (string, error) {
	switch m := t.Metadata.(type) {
	case *github.PullRequest:
		return m.Body, nil
	case *bitbucketserver.PullRequest:
		return m.Description, nil
	default:
		return "", errors.New("unknown changeset type")
	}
}

// State of a Changeset.
func (t *Changeset) State() (s ChangesetState, err error) {
	switch m := t.Metadata.(type) {
	case *github.PullRequest:
		s = ChangesetState(m.State)
	case *bitbucketserver.PullRequest:
		s = ChangesetState(m.State)
	default:
		return "", errors.New("unknown changeset type")
	}

	if !s.Valid() {
		return "", errors.Errorf("changeset state %q invalid", s)
	}

	return s, nil
}

// URL of a Changeset.
func (t *Changeset) URL() (s string, err error) {
	switch m := t.Metadata.(type) {
	case *github.PullRequest:
		return m.URL, nil
	case *bitbucketserver.PullRequest:
		if len(m.Links.Self) < 1 {
			return "", errors.New("bitbucketserver pull request has no self links")
		}
		selfLink := m.Links.Self[0]
		return selfLink.Href, nil
	default:
		return "", errors.New("unknown changeset type")
	}
}

// ReviewState of a Changeset.
func (t *Changeset) ReviewState() (s ChangesetReviewState, err error) {
	states := map[ChangesetReviewState]bool{}

	switch m := t.Metadata.(type) {
	case *github.PullRequest:
		for _, ti := range m.TimelineItems {
			if r, ok := ti.Item.(*github.PullRequestReview); ok {
				states[ChangesetReviewState(r.State)] = true
			}
		}
	case *bitbucketserver.PullRequest:
		for _, r := range m.Reviewers {
			switch r.Status {
			case "UNAPPROVED":
				states[ChangesetReviewStatePending] = true
			case "NEEDS_WORK":
				states[ChangesetReviewStateChangesRequested] = true
			case "APPROVED":
				states[ChangesetReviewStateApproved] = true
			}
		}
	default:
		return "", errors.New("unknown changeset type")
	}

	return selectReviewState(states), nil
}

func selectReviewState(states map[ChangesetReviewState]bool) ChangesetReviewState {
	// If any review requested changes, that state takes precedence over all
	// other review states, followed by explicit approval. Everything else is
	// considered pending.
	for _, state := range [...]ChangesetReviewState{
		ChangesetReviewStateChangesRequested,
		ChangesetReviewStateApproved,
	} {
		if states[state] {
			return state
		}
	}

	return ChangesetReviewStatePending
}

// A ChangesetEvent is an event that happened in the lifetime
// and context of a Changeset.
type ChangesetEvent struct {
	ID          int64
	ChangesetID int64
	Kind        changesetEventKind
	Source      changesetEventSource
	Key         string // Deduplication key
	CreatedAt   time.Time
	Metadata    interface{}
}

// Clone returns a clone of a ChangesetEvent.
func (e *ChangesetEvent) Clone() *ChangesetEvent {
	ee := *e
	return &ee
}

// changesetEventSource defines the source of a ChangesetEvent. This type is unexported
// so that users of ChangesetEvent can't instantiate it with a Source being an arbitrary
// string.
type changesetEventSource string

// Valid ChangesetEvent sources
const (
	ChangesetEventSourceGitHubAPI              changesetEventSource = "github:api"
	ChangesetEventSourceGitHubWebhook          changesetEventSource = "github:webhook"
	ChangesetEventSourceBitbucketServerAPI     changesetEventSource = "bitbucketserver:api"
	ChangesetEventSourceBitbucketServerWebhook changesetEventSource = "bitbucketserver:webhook"
)

// changesetEventKind defines the kind of a ChangesetEvent. This type is unexported
// so that users of ChangesetEvent can't instantiate it with a Kind being an arbitrary
// string.
type changesetEventKind string

// Valid ChangesetEvent kinds
const (
	// ChangesetEventKindAssigned is AssignedEvent on GitHub and doesnt exist on BBS
	ChangesetEventKindAssigned changesetEventKind = "assigned"

	// ChangesetEventKindClosed is ClosedEvent on GitHub and DECLINED on BBS
	ChangesetEventKindClosed changesetEventKind = "closed"

	// ChangesetEventKindCommented is IssueComment on GitHub and COMMENTED on BBS
	ChangesetEventKindCommented changesetEventKind = "commented"

	// ChangesetEventKindRenamedTitle is RenamedTitleEvent on GitHub and doesn't exist on BBS
	ChangesetEventKindRenamedTitle changesetEventKind = "renamed"

	// ChangesetEventKindMerged is MergedEvent on GitHub and MERGED on BBS
	ChangesetEventKindMerged changesetEventKind = "merged"

	// ChangesetEventKindApproved is PullRequestReviewEvent(APPROVED) on GitHub and APPROVED on BBS
	ChangesetEventKindApproved changesetEventKind = "approved"

	// ChangesetEventKindReopened is ReopenedEvent on GitHub and doesn't exist on BBS
	ChangesetEventKindReopened changesetEventKind = "reopened"

	// ChangesetEventKindReviewDismissed is ReviewDismissedEvent on GitHub and doesn't exist on BBS
	ChangesetEventKindReviewDismissed changesetEventKind = "review_dismissed"

	// ChangesetEventKindReviewRequestRemoved is ReviewRequestRemovedEvent on GitHub and doesn't exist on BBS
	ChangesetEventKindReviewRequestRemoved changesetEventKind = "review_request_removed"

	// ChangesetEventKindReviewRequested is ReviewRequestedEvent on GitHub and doesn't exist on BBS
	ChangesetEventKindReviewRequested changesetEventKind = "review_requested"

	// ChangesetEventKindUnassigned is UnassignedEvent on GitHub and doesnt exist on BBS
	ChangesetEventKindUnassigned changesetEventKind = "unassigned"

	// TODO: Full set of Bitbucket Server pull request actions:
	//   - APPROVED
	//   - COMMENTED
	//   - DECLINED
	//   - MERGED
	//   - OPENED
	//   - REOPENED
	//   - RESCOPED
	//   - UNAPPROVED
	//   - UPDATED
)
