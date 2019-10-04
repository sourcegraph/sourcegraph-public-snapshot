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

// Events returns the list of ChangesetEvents from the Changeset's metadata.
func (t *Changeset) Events() (events []*ChangesetEvent) {
	switch m := t.Metadata.(type) {
	case *github.PullRequest:
		events = make([]*ChangesetEvent, 0, len(m.TimelineItems))
		for _, ti := range m.TimelineItems {
			events = append(events, &ChangesetEvent{
				ChangesetID: t.ID,
				Source:      ChangesetEventSourceGitHubAPI,
				Kind:        ChangesetEventKind(ti.Item),
				Key:         ti.Item.(interface{ Key() string }).Key(),
				Metadata:    ti.Item,
			})
		}
	}
	return events
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

// ChangesetEventKind returns the changesetEventKind for the given
// specific code host event.
func ChangesetEventKind(e interface{}) changesetEventKind {
	switch e := e.(type) {
	case *github.AssignedEvent:
		return ChangesetEventKindGitHubAssigned
	case *github.ClosedEvent:
		return ChangesetEventKindGitHubClosed
	case *github.IssueComment:
		return ChangesetEventKindGitHubCommented
	case *github.RenamedTitleEvent:
		return ChangesetEventKindGitHubRenamedTitle
	case *github.MergedEvent:
		return ChangesetEventKindGitHubMerged
	case *github.PullRequestReview:
		return ChangesetEventKindGitHubReviewed
	case *github.ReopenedEvent:
		return ChangesetEventKindGitHubReopened
	case *github.ReviewDismissedEvent:
		return ChangesetEventKindGitHubReviewDismissed
	case *github.ReviewRequestRemovedEvent:
		return ChangesetEventKindGitHubReviewRequestRemoved
	case *github.ReviewRequestedEvent:
		return ChangesetEventKindGitHubReviewRequested
	case *github.UnassignedEvent:
		return ChangesetEventKindGitHubUnassigned
	default:
		panic(errors.Errorf("unknown changeset event kind for %T", e))
	}
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
	ChangesetEventKindGitHubAssigned             changesetEventKind = "github:assigned"
	ChangesetEventKindGitHubClosed               changesetEventKind = "github:closed"
	ChangesetEventKindGitHubCommented            changesetEventKind = "github:commented"
	ChangesetEventKindGitHubRenamedTitle         changesetEventKind = "github:renamed"
	ChangesetEventKindGitHubMerged               changesetEventKind = "github:merged"
	ChangesetEventKindGitHubReviewed             changesetEventKind = "github:reviewed"
	ChangesetEventKindGitHubReopened             changesetEventKind = "github:reopened"
	ChangesetEventKindGitHubReviewDismissed      changesetEventKind = "github:review_dismissed"
	ChangesetEventKindGitHubReviewRequestRemoved changesetEventKind = "github:review_request_removed"
	ChangesetEventKindGitHubReviewRequested      changesetEventKind = "github:review_requested"
	ChangesetEventKindGitHubUnassigned           changesetEventKind = "github:unassigned"

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
