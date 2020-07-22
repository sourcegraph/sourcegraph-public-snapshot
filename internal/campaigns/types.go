package campaigns

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
	"github.com/sourcegraph/sourcegraph/schema"
	"github.com/xeipuuv/gojsonschema"

	yamlv3 "gopkg.in/yaml.v3"
)

// SupportedExternalServices are the external service types currently supported
// by the campaigns feature. Repos that are associated with external services
// whose type is not in this list will simply be filtered out from the search
// results.
var SupportedExternalServices = map[string]struct{}{
	extsvc.TypeGitHub:          {},
	extsvc.TypeBitbucketServer: {},
	extsvc.TypeGitLab:          {},
}

// IsRepoSupported returns whether the given ExternalRepoSpec is supported by
// the campaigns feature, based on the external service type.
func IsRepoSupported(spec *api.ExternalRepoSpec) bool {
	_, ok := SupportedExternalServices[spec.ServiceType]
	return ok
}

// IsKindSupported returns whether the given extsvc Kind is supported by
// campaigns.
func IsKindSupported(extSvcKind string) bool {
	_, ok := SupportedExternalServices[extsvc.KindToType(extSvcKind)]
	return ok
}

// A Campaign of changesets over multiple Repos over time.
type Campaign struct {
	ID              int64
	Name            string
	Description     string
	Branch          string
	AuthorID        int32
	NamespaceUserID int32
	NamespaceOrgID  int32
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ChangesetIDs    []int64
	ClosedAt        time.Time
	CampaignSpecID  int64
}

// Clone returns a clone of a Campaign.
func (c *Campaign) Clone() *Campaign {
	cc := *c
	cc.ChangesetIDs = c.ChangesetIDs[:len(c.ChangesetIDs):len(c.ChangesetIDs)]
	return &cc
}

// RemoveChangesetID removes the given id from the Campaigns ChangesetIDs slice.
// If the id is not in ChangesetIDs calling this method doesn't have an effect.
func (c *Campaign) RemoveChangesetID(id int64) {
	for i := len(c.ChangesetIDs) - 1; i >= 0; i-- {
		if c.ChangesetIDs[i] == id {
			c.ChangesetIDs = append(c.ChangesetIDs[:i], c.ChangesetIDs[i+1:]...)
		}
	}
}

// GenChangesetBody creates the markdown to be used as the body of a changeset.
// It includes a URL back to the campaign on the Sourcegraph instance.
func (c *Campaign) GenChangesetBody(externalURL string) string {
	campaignID := MarshalCampaignID(c.ID)
	campaignURL := fmt.Sprintf("%s/campaigns/%s", externalURL, string(campaignID))
	description := fmt.Sprintf("%s\n\n---\n\nThis pull request was created by a Sourcegraph campaign. [Click here to see the campaign](%s).", c.Description, campaignURL)
	return description
}

// ChangesetState defines the possible states of a Changeset.
type ChangesetState string

// ChangesetState constants.
const (
	ChangesetStateUnpublished ChangesetState = "UNPUBLISHED"
	ChangesetStatePublishing  ChangesetState = "PUBLISHING"
	ChangesetStateErrored     ChangesetState = "ERRORED"
	ChangesetStateSynced      ChangesetState = "SYNCED"
)

// Valid returns true if the given ChangesetState is valid.
func (s ChangesetState) Valid() bool {
	switch s {
	case ChangesetStateUnpublished,
		ChangesetStatePublishing,
		ChangesetStateErrored,
		ChangesetStateSynced:
		return true
	default:
		return false
	}
}

// ChangesetExternalState defines the possible states of a Changeset on a code host.
type ChangesetExternalState string

// ChangesetExternalState constants.
const (
	ChangesetExternalStateOpen    ChangesetExternalState = "OPEN"
	ChangesetExternalStateClosed  ChangesetExternalState = "CLOSED"
	ChangesetExternalStateMerged  ChangesetExternalState = "MERGED"
	ChangesetExternalStateDeleted ChangesetExternalState = "DELETED"
)

// Valid returns true if the given ChangesetExternalState is valid.
func (s ChangesetExternalState) Valid() bool {
	switch s {
	case ChangesetExternalStateOpen,
		ChangesetExternalStateClosed,
		ChangesetExternalStateMerged,
		ChangesetExternalStateDeleted:
		return true
	default:
		return false
	}
}

// ChangesetLabel represents a label applied to a changeset
type ChangesetLabel struct {
	Name        string
	Color       string
	Description string
}

// CampaignState defines the possible states of a Campaign
type CampaignState string

const (
	CampaignStateAny    CampaignState = "ANY"
	CampaignStateOpen   CampaignState = "OPEN"
	CampaignStateClosed CampaignState = "CLOSED"
)

// ChangesetReviewState defines the possible states of a Changeset's review.
type ChangesetReviewState string

// ChangesetReviewState constants.
const (
	ChangesetReviewStateApproved         ChangesetReviewState = "APPROVED"
	ChangesetReviewStateChangesRequested ChangesetReviewState = "CHANGES_REQUESTED"
	ChangesetReviewStatePending          ChangesetReviewState = "PENDING"
	ChangesetReviewStateCommented        ChangesetReviewState = "COMMENTED"
	ChangesetReviewStateDismissed        ChangesetReviewState = "DISMISSED"
)

// Valid returns true if the given Changeset review state is valid.
func (s ChangesetReviewState) Valid() bool {
	switch s {
	case ChangesetReviewStateApproved,
		ChangesetReviewStateChangesRequested,
		ChangesetReviewStatePending,
		ChangesetReviewStateCommented,
		ChangesetReviewStateDismissed:
		return true
	default:
		return false
	}
}

// ChangesetCheckState constants.
type ChangesetCheckState string

const (
	ChangesetCheckStateUnknown ChangesetCheckState = "UNKNOWN"
	ChangesetCheckStatePending ChangesetCheckState = "PENDING"
	ChangesetCheckStatePassed  ChangesetCheckState = "PASSED"
	ChangesetCheckStateFailed  ChangesetCheckState = "FAILED"
)

// Valid returns true if the given Changeset check state is valid.
func (s ChangesetCheckState) Valid() bool {
	switch s {
	case ChangesetCheckStateUnknown,
		ChangesetCheckStatePending,
		ChangesetCheckStatePassed,
		ChangesetCheckStateFailed:
		return true
	default:
		return false
	}
}

// A Changeset is a changeset on a code host belonging to a Repository and many
// Campaigns.
type Changeset struct {
	ID                  int64
	RepoID              api.RepoID
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Metadata            interface{}
	CampaignIDs         []int64
	ExternalID          string
	ExternalServiceType string
	ExternalBranch      string
	ExternalDeletedAt   time.Time
	ExternalUpdatedAt   time.Time
	ExternalState       ChangesetExternalState
	ExternalReviewState ChangesetReviewState
	ExternalCheckState  ChangesetCheckState
	CreatedByCampaign   bool
	AddedToCampaign     bool
	DiffStatAdded       *int32
	DiffStatChanged     *int32
	DiffStatDeleted     *int32
	SyncState           ChangesetSyncState
}

// Clone returns a clone of a Changeset.
func (c *Changeset) Clone() *Changeset {
	tt := *c
	tt.CampaignIDs = c.CampaignIDs[:len(c.CampaignIDs):len(c.CampaignIDs)]
	return &tt
}

// DiffStat returns a *diff.Stat if DiffStatAdded, DiffStatChanged, and
// DiffStatDeleted are set, or nil if one or more is not.
func (c *Changeset) DiffStat() *diff.Stat {
	if c.DiffStatAdded == nil || c.DiffStatChanged == nil || c.DiffStatDeleted == nil {
		return nil
	}

	return &diff.Stat{
		Added:   *c.DiffStatAdded,
		Changed: *c.DiffStatChanged,
		Deleted: *c.DiffStatDeleted,
	}
}

func (c *Changeset) SetDiffStat(stat *diff.Stat) {
	if stat == nil {
		c.DiffStatAdded = nil
		c.DiffStatChanged = nil
		c.DiffStatDeleted = nil
	} else {
		added := stat.Added
		c.DiffStatAdded = &added

		changed := stat.Changed
		c.DiffStatChanged = &changed

		deleted := stat.Deleted
		c.DiffStatDeleted = &deleted
	}
}

func (c *Changeset) SetMetadata(meta interface{}) error {
	switch pr := meta.(type) {
	case *github.PullRequest:
		c.Metadata = pr
		c.ExternalID = strconv.FormatInt(pr.Number, 10)
		c.ExternalServiceType = extsvc.TypeGitHub
		c.ExternalBranch = pr.HeadRefName
		c.ExternalUpdatedAt = pr.UpdatedAt
	case *bitbucketserver.PullRequest:
		c.Metadata = pr
		c.ExternalID = strconv.FormatInt(int64(pr.ID), 10)
		c.ExternalServiceType = extsvc.TypeBitbucketServer
		c.ExternalBranch = git.AbbreviateRef(pr.FromRef.ID)
		c.ExternalUpdatedAt = unixMilliToTime(int64(pr.UpdatedDate))
	case *gitlab.MergeRequest:
		c.Metadata = pr
		c.ExternalID = strconv.FormatInt(int64(pr.IID), 10)
		c.ExternalServiceType = extsvc.TypeGitLab
		c.ExternalBranch = pr.SourceBranch
		c.ExternalUpdatedAt = pr.UpdatedAt
	default:
		return errors.New("unknown changeset type")
	}
	return nil
}

// RemoveCampaignID removes the given id from the Changesets CampaignIDs slice.
// If the id is not in CampaignIDs calling this method doesn't have an effect.
func (c *Changeset) RemoveCampaignID(id int64) {
	for i := len(c.CampaignIDs) - 1; i >= 0; i-- {
		if c.CampaignIDs[i] == id {
			c.CampaignIDs = append(c.CampaignIDs[:i], c.CampaignIDs[i+1:]...)
		}
	}
}

// Title of the Changeset.
func (c *Changeset) Title() (string, error) {
	switch m := c.Metadata.(type) {
	case *github.PullRequest:
		return m.Title, nil
	case *bitbucketserver.PullRequest:
		return m.Title, nil
	case *gitlab.MergeRequest:
		return m.Title, nil
	default:
		return "", errors.New("unknown changeset type")
	}
}

// ExternalCreatedAt is when the Changeset was created on the codehost. When it
// cannot be determined when the changeset was created, a zero-value timestamp
// is returned.
func (c *Changeset) ExternalCreatedAt() time.Time {
	switch m := c.Metadata.(type) {
	case *github.PullRequest:
		return m.CreatedAt
	case *bitbucketserver.PullRequest:
		return unixMilliToTime(int64(m.CreatedDate))
	case *gitlab.MergeRequest:
		return m.CreatedAt
	default:
		return time.Time{}
	}
}

// Body of the Changeset.
func (c *Changeset) Body() (string, error) {
	switch m := c.Metadata.(type) {
	case *github.PullRequest:
		return m.Body, nil
	case *bitbucketserver.PullRequest:
		return m.Description, nil
	case *gitlab.MergeRequest:
		return m.Description, nil
	default:
		return "", errors.New("unknown changeset type")
	}
}

// SetDeleted sets the internal state of a Changeset so that its State is
// ChangesetStateDeleted.
func (c *Changeset) SetDeleted() {
	c.ExternalDeletedAt = time.Now().UTC().Truncate(time.Microsecond)
}

// IsDeleted returns true when the Changeset's ExternalDeletedAt is a non-zero
// timestamp.
func (c *Changeset) IsDeleted() bool {
	return !c.ExternalDeletedAt.IsZero()
}

// externalState of a Changeset based on the metadata.
// It does NOT reflect the final calculated externalState, use `ExternalState` instead.
func (c *Changeset) externalState() (s ChangesetExternalState, err error) {
	if !c.ExternalDeletedAt.IsZero() {
		return ChangesetExternalStateDeleted, nil
	}

	switch m := c.Metadata.(type) {
	case *github.PullRequest:
		s = ChangesetExternalState(m.State)
	case *bitbucketserver.PullRequest:
		if m.State == "DECLINED" {
			s = ChangesetExternalStateClosed
		} else {
			s = ChangesetExternalState(m.State)
		}
	case *gitlab.MergeRequest:
		switch m.State {
		case gitlab.MergeRequestStateOpened:
			s = ChangesetExternalStateOpen
		case gitlab.MergeRequestStateClosed, gitlab.MergeRequestStateLocked:
			s = ChangesetExternalStateClosed
		case gitlab.MergeRequestStateMerged:
			s = ChangesetExternalStateMerged
		default:
			return "", errors.Errorf("unknown merge request state: %s", m.State)
		}
	default:
		return "", errors.New("unknown changeset type")
	}

	if !s.Valid() {
		return "", errors.Errorf("changeset state %q invalid", s)
	}

	return s, nil
}

// URL of a Changeset.
func (c *Changeset) URL() (s string, err error) {
	switch m := c.Metadata.(type) {
	case *github.PullRequest:
		return m.URL, nil
	case *bitbucketserver.PullRequest:
		if len(m.Links.Self) < 1 {
			return "", errors.New("bitbucketserver pull request has no self links")
		}
		selfLink := m.Links.Self[0]
		return selfLink.Href, nil
	case *gitlab.MergeRequest:
		return m.WebURL, nil
	default:
		return "", errors.New("unknown changeset type")
	}
}

// Changesets is a slice of *Changesets.
type Changesets []*Changeset

// IDs returns the IDs of all changesets in the slice.
func (cs Changesets) IDs() []int64 {
	ids := make([]int64, len(cs))
	for i, c := range cs {
		ids[i] = c.ID
	}
	return ids
}

// IDs returns the RepoIDs of all changesets in the slice.
func (cs Changesets) RepoIDs() []api.RepoID {
	repoIDs := make([]api.RepoID, len(cs))
	for i, c := range cs {
		repoIDs[i] = c.RepoID
	}
	return repoIDs
}

// Filter returns a new Changesets slice in which changesets have been filtered
// out for which the predicate didn't return true.
func (cs Changesets) Filter(predicate func(*Changeset) bool) (filtered Changesets) {
	for _, c := range cs {
		if predicate(c) {
			filtered = append(filtered, c)
		}
	}

	return filtered
}

// Keyer represents items that return a unique key
type Keyer interface {
	Key() string
}

// Events returns the list of ChangesetEvents from the Changeset's metadata.
func (c *Changeset) Events() (events []*ChangesetEvent) {
	switch m := c.Metadata.(type) {
	case *github.PullRequest:
		events = make([]*ChangesetEvent, 0, len(m.TimelineItems))
		for _, ti := range m.TimelineItems {
			ev := ChangesetEvent{ChangesetID: c.ID}

			switch e := ti.Item.(type) {
			case *github.PullRequestReviewThread:
				for _, c := range e.Comments {
					ev := ev
					ev.Key = c.Key()
					ev.Kind = ChangesetEventKindFor(c)
					ev.Metadata = c
					events = append(events, &ev)
				}

			case *github.ReviewRequestedEvent:
				// If the reviewer of a ReviewRequestedEvent has been deleted,
				// the fields are blank and we cannot match the event to an
				// entry in the database and/or reliably use it, so we drop it.
				if e.ReviewerDeleted() {
					continue
				}
				ev.Key = e.Key()
				ev.Kind = ChangesetEventKindFor(e)
				ev.Metadata = e
				events = append(events, &ev)

			default:
				ev.Key = ti.Item.(Keyer).Key()
				ev.Kind = ChangesetEventKindFor(ti.Item)
				ev.Metadata = ti.Item
				events = append(events, &ev)
			}
		}

	case *bitbucketserver.PullRequest:
		events = make([]*ChangesetEvent, 0, len(m.Activities)+len(m.CommitStatus))
		addEvent := func(e Keyer) {
			events = append(events, &ChangesetEvent{
				ChangesetID: c.ID,
				Key:         e.Key(),
				Kind:        ChangesetEventKindFor(e),
				Metadata:    e,
			})
		}
		for _, a := range m.Activities {
			addEvent(a)
		}
		for _, s := range m.CommitStatus {
			addEvent(s)
		}

	case *gitlab.MergeRequest:
		events = make([]*ChangesetEvent, 0, len(m.Notes)+len(m.Pipelines))

		for _, note := range m.Notes {
			if review := note.ToReview(); review != nil {
				events = append(events, &ChangesetEvent{
					ChangesetID: c.ID,
					Key:         review.(Keyer).Key(),
					Kind:        ChangesetEventKindFor(review),
					Metadata:    review,
				})
			}
		}

		for _, pipeline := range m.Pipelines {
			events = append(events, &ChangesetEvent{
				ChangesetID: c.ID,
				Key:         pipeline.Key(),
				Kind:        ChangesetEventKindFor(pipeline),
				Metadata:    pipeline,
			})
		}
	}
	return events
}

// HeadRefOid returns the git ObjectID of the HEAD reference associated with
// Changeset on the codehost. If the codehost doesn't include the ObjectID, an
// empty string is returned.
func (c *Changeset) HeadRefOid() (string, error) {
	switch m := c.Metadata.(type) {
	case *github.PullRequest:
		return m.HeadRefOid, nil
	case *bitbucketserver.PullRequest:
		return "", nil
	case *gitlab.MergeRequest:
		return m.DiffRefs.HeadSHA, nil
	default:
		return "", errors.New("unknown changeset type")
	}
}

// HeadRef returns the full ref (e.g. `refs/heads/my-branch`) of the
// HEAD reference associated with the Changeset on the codehost.
func (c *Changeset) HeadRef() (string, error) {
	switch m := c.Metadata.(type) {
	case *github.PullRequest:
		return "refs/heads/" + m.HeadRefName, nil
	case *bitbucketserver.PullRequest:
		return m.FromRef.ID, nil
	case *gitlab.MergeRequest:
		return "refs/heads/" + m.SourceBranch, nil
	default:
		return "", errors.New("unknown changeset type")
	}
}

// BaseRefOid returns the git ObjectID of the base reference associated with the
// Changeset on the codehost. If the codehost doesn't include the ObjectID, an
// empty string is returned.
func (c *Changeset) BaseRefOid() (string, error) {
	switch m := c.Metadata.(type) {
	case *github.PullRequest:
		return m.BaseRefOid, nil
	case *bitbucketserver.PullRequest:
		return "", nil
	case *gitlab.MergeRequest:
		return m.DiffRefs.BaseSHA, nil
	default:
		return "", errors.New("unknown changeset type")
	}
}

// BaseRef returns the full ref (e.g. `refs/heads/my-branch`) of the base ref
// associated with the Changeset on the codehost.
func (c *Changeset) BaseRef() (string, error) {
	switch m := c.Metadata.(type) {
	case *github.PullRequest:
		return "refs/heads/" + m.BaseRefName, nil
	case *bitbucketserver.PullRequest:
		return m.ToRef.ID, nil
	case *gitlab.MergeRequest:
		return "refs/heads/" + m.TargetBranch, nil
	default:
		return "", errors.New("unknown changeset type")
	}
}

// SupportsLabels returns whether the code host on which the changeset is
// hosted supports labels and whether it's safe to call the
// (*Changeset).Labels() method.
func (c *Changeset) SupportsLabels() bool {
	switch c.Metadata.(type) {
	case *github.PullRequest, *gitlab.MergeRequest:
		return true
	default:
		return false
	}
}

func (c *Changeset) Labels() []ChangesetLabel {
	switch m := c.Metadata.(type) {
	case *github.PullRequest:
		labels := make([]ChangesetLabel, len(m.Labels.Nodes))
		for i, l := range m.Labels.Nodes {
			labels[i] = ChangesetLabel{
				Name:        l.Name,
				Color:       l.Color,
				Description: l.Description,
			}
		}
		return labels
	case *gitlab.MergeRequest:
		// Similarly to GitHub above, GitLab labels can have colors (foreground
		// _and_ background, in fact) and descriptions. Unfortunately, the REST
		// API only returns this level of detail on the list endpoint (with an
		// option added in GitLab 12.7), and not when retrieving individual MRs.
		//
		// When our minimum GitLab version is 12.0, we should be able to switch
		// to retrieving MRs via GraphQL, and then we can start retrieving
		// richer label data.
		labels := make([]ChangesetLabel, len(m.Labels))
		for i, l := range m.Labels {
			labels[i] = ChangesetLabel{Name: l, Color: "000000"}
		}
		return labels
	default:
		return []ChangesetLabel{}
	}
}

type ChangesetSyncState struct {
	BaseRefOid string
	HeadRefOid string

	// This is essentially the result of c.ExternalState != CampaignStateOpen
	// the last time a sync occured. We use this to short circuit computing the
	// sync state if the changeset remains closed.
	IsComplete bool
}

func (state *ChangesetSyncState) Equals(old *ChangesetSyncState) bool {
	return state.BaseRefOid == old.BaseRefOid && state.HeadRefOid == old.HeadRefOid && state.IsComplete == old.IsComplete
}

// A ChangesetEvent is an event that happened in the lifetime
// and context of a Changeset.
type ChangesetEvent struct {
	ID          int64
	ChangesetID int64
	Kind        ChangesetEventKind
	Key         string // Deduplication key
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Metadata    interface{}
}

// Clone returns a clone of a ChangesetEvent.
func (e *ChangesetEvent) Clone() *ChangesetEvent {
	ee := *e
	return &ee
}

// Actor returns the actor of the ChangesetEvent.
func (e *ChangesetEvent) Actor() string {
	var a string

	switch e := e.Metadata.(type) {
	case *github.AssignedEvent:
		a = e.Actor.Login
	case *github.ClosedEvent:
		a = e.Actor.Login
	case *github.IssueComment:
		a = e.Author.Login
	case *github.RenamedTitleEvent:
		a = e.Actor.Login
	case *github.MergedEvent:
		a = e.Actor.Login
	case *github.PullRequestReview:
		a = e.Author.Login
	case *github.PullRequestReviewComment:
		a = e.Author.Login
	case *github.ReopenedEvent:
		a = e.Actor.Login
	case *github.ReviewDismissedEvent:
		a = e.Actor.Login
	case *github.ReviewRequestRemovedEvent:
		a = e.Actor.Login
	case *github.ReviewRequestedEvent:
		a = e.Actor.Login
	case *github.UnassignedEvent:
		a = e.Actor.Login
	case *github.LabelEvent:
		a = e.Actor.Login
	}

	return a
}

// ReviewAuthor returns the author of the review if the ChangesetEvent is related to a review.
func (e *ChangesetEvent) ReviewAuthor() (string, error) {
	switch meta := e.Metadata.(type) {
	case *github.PullRequestReview:
		login := meta.Author.Login
		if login == "" {
			return "", errors.New("review author is blank")
		}
		return login, nil

	case *github.ReviewDismissedEvent:
		login := meta.Review.Author.Login
		if login == "" {
			return "", errors.New("review author in dismissed event is blank")
		}
		return login, nil

	case *bitbucketserver.Activity:
		username := meta.User.Name
		if username == "" {
			return "", errors.New("activity user is blank")
		}
		return username, nil

	case *bitbucketserver.ParticipantStatusEvent:
		username := meta.User.Name
		if username == "" {
			return "", errors.New("activity user is blank")
		}
		return username, nil

	case *gitlab.ReviewApproved:
		username := meta.Author.Username
		if username == "" {
			return "", errors.New("review user is blank")
		}
		return username, nil

	case *gitlab.ReviewUnapproved:
		username := meta.Author.Username
		if username == "" {
			return "", errors.New("review user is blank")
		}
		return username, nil

	default:
		return "", nil
	}
}

// ReviewState returns the review state of the ChangesetEvent if it is a review event.
func (e *ChangesetEvent) ReviewState() (ChangesetReviewState, error) {
	switch e.Kind {
	case ChangesetEventKindBitbucketServerApproved,
		ChangesetEventKindGitLabApproved:
		return ChangesetReviewStateApproved, nil

	// BitbucketServer's "REVIEWED" activity is created when someone clicks
	// the "Needs work" button in the UI, which is why we map it to "Changes Requested"
	case ChangesetEventKindBitbucketServerReviewed:
		return ChangesetReviewStateChangesRequested, nil

	case ChangesetEventKindGitHubReviewed:
		review, ok := e.Metadata.(*github.PullRequestReview)
		if !ok {
			return "", errors.New("ChangesetEvent metadata event not PullRequestReview")
		}

		s := ChangesetReviewState(strings.ToUpper(review.State))
		if !s.Valid() {
			// Ignore invalid states
			log15.Warn("invalid review state", "state", review.State)
			return ChangesetReviewStatePending, nil
		}
		return s, nil

	case ChangesetEventKindGitHubReviewDismissed,
		ChangesetEventKindBitbucketServerUnapproved,
		ChangesetEventKindBitbucketServerDismissed,
		ChangesetEventKindGitLabUnapproved:
		return ChangesetReviewStateDismissed, nil

	default:
		return ChangesetReviewStatePending, nil
	}
}

// Type returns the ChangesetEventKind of the ChangesetEvent.
func (e *ChangesetEvent) Type() ChangesetEventKind {
	return e.Kind
}

// Changeset returns the changeset ID of the ChangesetEvent.
func (e *ChangesetEvent) Changeset() int64 {
	return e.ChangesetID
}

// Timestamp returns the time when the ChangesetEvent happened (or was updated)
// on the codehost, not when it was created in Sourcegraph's database.
func (e *ChangesetEvent) Timestamp() time.Time {
	var t time.Time

	switch e := e.Metadata.(type) {
	case *github.AssignedEvent:
		t = e.CreatedAt
	case *github.ClosedEvent:
		t = e.CreatedAt
	case *github.IssueComment:
		t = e.UpdatedAt
	case *github.RenamedTitleEvent:
		t = e.CreatedAt
	case *github.MergedEvent:
		t = e.CreatedAt
	case *github.PullRequestReview:
		t = e.UpdatedAt
	case *github.PullRequestReviewComment:
		t = e.UpdatedAt
	case *github.ReopenedEvent:
		t = e.CreatedAt
	case *github.ReviewDismissedEvent:
		t = e.CreatedAt
	case *github.ReviewRequestRemovedEvent:
		t = e.CreatedAt
	case *github.ReviewRequestedEvent:
		t = e.CreatedAt
	case *github.UnassignedEvent:
		t = e.CreatedAt
	case *github.LabelEvent:
		t = e.CreatedAt
	case *github.CommitStatus:
		t = e.ReceivedAt
	case *github.CheckSuite:
		return e.ReceivedAt
	case *github.CheckRun:
		return e.ReceivedAt
	case *bitbucketserver.Activity:
		t = unixMilliToTime(int64(e.CreatedDate))
	case *bitbucketserver.ParticipantStatusEvent:
		t = unixMilliToTime(int64(e.CreatedDate))
	case *bitbucketserver.CommitStatus:
		t = unixMilliToTime(int64(e.Status.DateAdded))
	case *gitlab.ReviewApproved:
		return e.CreatedAt
	case *gitlab.ReviewUnapproved:
		return e.CreatedAt
	}

	return t
}

// Update updates the metadata of e with new metadata in o.
func (e *ChangesetEvent) Update(o *ChangesetEvent) {
	if e.ChangesetID != o.ChangesetID || e.Kind != o.Kind || e.Key != o.Key {
		return
	}

	switch e := e.Metadata.(type) {
	case *github.LabelEvent:
		o := o.Metadata.(*github.LabelEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}

		if e.Label == (github.Label{}) {
			e.Label = o.Label
		}

	case *github.AssignedEvent:
		o := o.Metadata.(*github.AssignedEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if e.Assignee == (github.Actor{}) {
			e.Assignee = o.Assignee
		}

		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}

	case *github.ClosedEvent:
		o := o.Metadata.(*github.ClosedEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if o.URL != "" && e.URL != o.URL {
			e.URL = o.URL
		}

		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}

	case *github.IssueComment:
		o := o.Metadata.(*github.IssueComment)

		if e.DatabaseID == 0 {
			e.DatabaseID = o.DatabaseID
		}

		if e.Author == (github.Actor{}) {
			e.Author = o.Author
		}

		if e.Editor == nil {
			e.Editor = o.Editor
		}

		if o.AuthorAssociation != "" && e.AuthorAssociation != o.AuthorAssociation {
			e.AuthorAssociation = o.AuthorAssociation
		}

		if o.Body != "" && e.Body != o.Body {
			e.Body = o.Body
		}

		if o.URL != "" && e.URL != o.URL {
			e.URL = o.URL
		}

		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}

		if e.UpdatedAt.Before(o.UpdatedAt) {
			e.UpdatedAt = o.UpdatedAt
		}

		if o.IncludesCreatedEdit {
			e.IncludesCreatedEdit = true
		}

	case *github.RenamedTitleEvent:
		o := o.Metadata.(*github.RenamedTitleEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if o.PreviousTitle != "" && e.PreviousTitle != o.PreviousTitle {
			e.PreviousTitle = o.PreviousTitle
		}

		if o.CurrentTitle != "" && e.CurrentTitle != o.CurrentTitle {
			e.CurrentTitle = o.CurrentTitle
		}

		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}

	case *github.MergedEvent:
		o := o.Metadata.(*github.MergedEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if o.MergeRefName != "" && e.MergeRefName != o.MergeRefName {
			e.MergeRefName = o.MergeRefName
		}

		if o.URL != "" && e.URL != o.URL {
			e.URL = o.URL
		}

		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}

		updateGitHubCommit(&e.Commit, &o.Commit)

	case *github.PullRequestReview:
		o := o.Metadata.(*github.PullRequestReview)

		updateGitHubPullRequestReview(e, o)

	case *github.PullRequestReviewComment:
		o := o.Metadata.(*github.PullRequestReviewComment)

		if e.DatabaseID == 0 {
			e.DatabaseID = o.DatabaseID
		}

		if e.Author == (github.Actor{}) {
			e.Author = o.Author
		}

		if o.AuthorAssociation != "" && e.AuthorAssociation != o.AuthorAssociation {
			e.AuthorAssociation = o.AuthorAssociation
		}

		if e.Editor == (github.Actor{}) {
			e.Editor = o.Editor
		}

		if o.Body != "" && e.Body != o.Body {
			e.Body = o.Body
		}

		if o.State != "" && e.State != o.State {
			e.State = o.State
		}

		if o.URL != "" && e.URL != o.URL {
			e.URL = o.URL
		}

		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}

		if e.UpdatedAt.Before(o.UpdatedAt) {
			e.UpdatedAt = o.UpdatedAt
		}

		if e, o := e.Commit, o.Commit; !reflect.DeepEqual(e, o) {
			updateGitHubCommit(&e, &o)
		}

		if o.IncludesCreatedEdit {
			e.IncludesCreatedEdit = true
		}

	case *github.ReopenedEvent:
		o := o.Metadata.(*github.ReopenedEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}
	case *github.ReviewDismissedEvent:
		o := o.Metadata.(*github.ReviewDismissedEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if o.DismissalMessage != "" && e.DismissalMessage != o.DismissalMessage {
			e.DismissalMessage = o.DismissalMessage
		}

		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}

		updateGitHubPullRequestReview(&e.Review, &o.Review)

	case *github.ReviewRequestRemovedEvent:
		o := o.Metadata.(*github.ReviewRequestRemovedEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if e.RequestedReviewer == (github.Actor{}) {
			e.RequestedReviewer = o.RequestedReviewer
		}

		if e.RequestedTeam == (github.Team{}) {
			e.RequestedTeam = o.RequestedTeam
		}

		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}

	case *github.ReviewRequestedEvent:
		o := o.Metadata.(*github.ReviewRequestedEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if e.RequestedReviewer == (github.Actor{}) {
			e.RequestedReviewer = o.RequestedReviewer
		}

		if e.RequestedTeam == (github.Team{}) {
			e.RequestedTeam = o.RequestedTeam
		}

		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}

	case *github.UnassignedEvent:
		o := o.Metadata.(*github.UnassignedEvent)

		if e.Actor == (github.Actor{}) {
			e.Actor = o.Actor
		}

		if e.Assignee == (github.Actor{}) {
			e.Assignee = o.Assignee
		}

		if e.CreatedAt.IsZero() {
			e.CreatedAt = o.CreatedAt
		}
	case *bitbucketserver.Activity:
		o := o.Metadata.(*bitbucketserver.Activity)

		if e.CreatedDate == 0 {
			e.CreatedDate = o.CreatedDate
		}

		if e.User == (bitbucketserver.User{}) {
			e.User = o.User
		}

		if e.Action == "" {
			e.Action = o.Action
		}

		if e.CommentAction == "" {
			e.CommentAction = o.CommentAction
		}

		if e.Comment == nil && o.Comment != nil {
			e.Comment = o.Comment
		}

		if len(e.AddedReviewers) == 0 {
			e.AddedReviewers = o.AddedReviewers
		}

		if len(e.RemovedReviewers) == 0 {
			e.RemovedReviewers = o.RemovedReviewers
		}

		if e.Commit == nil && o.Commit != nil {
			e.Commit = o.Commit
		}

	case *bitbucketserver.ParticipantStatusEvent:
		o := o.Metadata.(*bitbucketserver.ParticipantStatusEvent)

		if e.CreatedDate == 0 {
			e.CreatedDate = o.CreatedDate
		}

		if e.Action == "" {
			e.Action = o.Action
		}

		if e.User == (bitbucketserver.User{}) {
			e.User = o.User
		}

	case *bitbucketserver.CommitStatus:
		o := o.Metadata.(*bitbucketserver.CommitStatus)
		// We always get the full event, so safe to replace it
		*e = *o

	case *github.CheckRun:
		o := o.Metadata.(*github.CheckRun)
		updateGithubCheckRun(e, o)

	case *github.CheckSuite:
		o := o.Metadata.(*github.CheckSuite)
		if e.Status == "" {
			e.Status = o.Status
		}
		if e.Conclusion == "" {
			e.Conclusion = o.Conclusion
		}
		e.CheckRuns = o.CheckRuns

	case *gitlab.ReviewApproved:
		o := o.Metadata.(*gitlab.ReviewApproved)
		// We always get the full event, so safe to replace it
		*e = *o

	case *gitlab.ReviewUnapproved:
		o := o.Metadata.(*gitlab.ReviewUnapproved)
		// We always get the full event, so safe to replace it
		*e = *o

	default:
		panic(errors.Errorf("unknown changeset event metadata %T", e))
	}
}

func updateGithubCheckRun(e, o *github.CheckRun) {
	if e.Status == "" {
		e.Status = o.Status
	}
	if e.Conclusion == "" {
		e.Conclusion = o.Conclusion
	}
}

func updateGitHubPullRequestReview(e, o *github.PullRequestReview) {
	if e.DatabaseID == 0 {
		e.DatabaseID = o.DatabaseID
	}

	if e.Author == (github.Actor{}) {
		e.Author = o.Author
	}

	if o.AuthorAssociation != "" && e.AuthorAssociation != o.AuthorAssociation {
		e.AuthorAssociation = o.AuthorAssociation
	}

	if o.Body != "" && e.Body != o.Body {
		e.Body = o.Body
	}

	if o.State != "" && e.State != o.State {
		e.State = o.State
	}

	if o.URL != "" && e.URL != o.URL {
		e.URL = o.URL
	}

	if e.CreatedAt.IsZero() {
		e.CreatedAt = o.CreatedAt
	}

	if e.UpdatedAt.Before(o.UpdatedAt) {
		e.UpdatedAt = o.UpdatedAt
	}

	if e, o := e.Commit, o.Commit; !reflect.DeepEqual(e, o) {
		updateGitHubCommit(&e, &o)
	}

	if o.IncludesCreatedEdit {
		e.IncludesCreatedEdit = true
	}
}

func updateGitHubCommit(e, o *github.Commit) {
	if o.OID != "" && e.OID != o.OID {
		e.OID = o.OID
	}

	if o.Message != "" && e.Message != o.Message {
		e.Message = o.Message
	}

	if o.MessageHeadline != "" && e.MessageHeadline != o.MessageHeadline {
		e.MessageHeadline = o.MessageHeadline
	}

	if o.URL != "" && e.URL != o.URL {
		e.URL = o.URL
	}

	if e.Committer != (github.GitActor{}) && e.Committer != o.Committer {
		e.Committer = o.Committer
	}

	if e.CommittedDate.IsZero() {
		e.CommittedDate = o.CommittedDate
	}

	if e.PushedDate.IsZero() {
		e.PushedDate = o.PushedDate
	}
}

// ChangesetEventKindFor returns the ChangesetEventKind for the given
// specific code host event.
func ChangesetEventKindFor(e interface{}) ChangesetEventKind {
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
	case *github.PullRequestReviewComment:
		return ChangesetEventKindGitHubReviewCommented
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
	case *github.PullRequestCommit:
		return ChangesetEventKindGitHubCommit
	case *github.LabelEvent:
		if e.Removed {
			return ChangesetEventKindGitHubUnlabeled
		}
		return ChangesetEventKindGitHubLabeled
	case *github.CommitStatus:
		return ChangesetEventKindCommitStatus
	case *github.CheckSuite:
		return ChangesetEventKindCheckSuite
	case *github.CheckRun:
		return ChangesetEventKindCheckRun
	case *bitbucketserver.Activity:
		return ChangesetEventKind("bitbucketserver:" + strings.ToLower(string(e.Action)))
	case *bitbucketserver.ParticipantStatusEvent:
		return ChangesetEventKind("bitbucketserver:participant_status:" + strings.ToLower(string(e.Action)))
	case *bitbucketserver.CommitStatus:
		return ChangesetEventKindBitbucketServerCommitStatus
	case *gitlab.Pipeline:
		return ChangesetEventKindGitLabPipeline
	case *gitlab.ReviewApproved:
		return ChangesetEventKindGitLabApproved
	case *gitlab.ReviewUnapproved:
		return ChangesetEventKindGitLabUnapproved
	default:
		panic(errors.Errorf("unknown changeset event kind for %T", e))
	}
}

// NewChangesetEventMetadata returns a new metadata object for the given
// ChangesetEventKind.
func NewChangesetEventMetadata(k ChangesetEventKind) (interface{}, error) {
	switch {
	case strings.HasPrefix(string(k), "bitbucketserver"):
		switch k {
		case ChangesetEventKindBitbucketServerCommitStatus:
			return new(bitbucketserver.CommitStatus), nil
		case ChangesetEventKindBitbucketServerDismissed:
			return new(bitbucketserver.ParticipantStatusEvent), nil
		default:
			return new(bitbucketserver.Activity), nil
		}
	case strings.HasPrefix(string(k), "github"):
		switch k {
		case ChangesetEventKindGitHubAssigned:
			return new(github.AssignedEvent), nil
		case ChangesetEventKindGitHubClosed:
			return new(github.ClosedEvent), nil
		case ChangesetEventKindGitHubCommented:
			return new(github.IssueComment), nil
		case ChangesetEventKindGitHubRenamedTitle:
			return new(github.RenamedTitleEvent), nil
		case ChangesetEventKindGitHubMerged:
			return new(github.MergedEvent), nil
		case ChangesetEventKindGitHubReviewed:
			return new(github.PullRequestReview), nil
		case ChangesetEventKindGitHubReviewCommented:
			return new(github.PullRequestReviewComment), nil
		case ChangesetEventKindGitHubReopened:
			return new(github.ReopenedEvent), nil
		case ChangesetEventKindGitHubReviewDismissed:
			return new(github.ReviewDismissedEvent), nil
		case ChangesetEventKindGitHubReviewRequestRemoved:
			return new(github.ReviewRequestRemovedEvent), nil
		case ChangesetEventKindGitHubReviewRequested:
			return new(github.ReviewRequestedEvent), nil
		case ChangesetEventKindGitHubUnassigned:
			return new(github.UnassignedEvent), nil
		case ChangesetEventKindGitHubCommit:
			return new(github.PullRequestCommit), nil
		case ChangesetEventKindGitHubLabeled:
			return new(github.LabelEvent), nil
		case ChangesetEventKindGitHubUnlabeled:
			return &github.LabelEvent{Removed: true}, nil
		case ChangesetEventKindCommitStatus:
			return new(github.CommitStatus), nil
		case ChangesetEventKindCheckSuite:
			return new(github.CheckSuite), nil
		case ChangesetEventKindCheckRun:
			return new(github.CheckRun), nil
		}
	case strings.HasPrefix(string(k), "gitlab"):
		switch k {
		case ChangesetEventKindGitLabApproved:
			return new(gitlab.ReviewApproved), nil
		case ChangesetEventKindGitLabPipeline:
			return new(gitlab.Pipeline), nil
		case ChangesetEventKindGitLabUnapproved:
			return new(gitlab.ReviewUnapproved), nil
		}
	}
	return nil, errors.Errorf("unknown changeset event kind %q", k)
}

// ChangesetEventKind defines the kind of a ChangesetEvent. This type is unexported
// so that users of ChangesetEvent can't instantiate it with a Kind being an arbitrary
// string.
type ChangesetEventKind string

// Valid ChangesetEvent kinds
const (
	ChangesetEventKindGitHubAssigned             ChangesetEventKind = "github:assigned"
	ChangesetEventKindGitHubClosed               ChangesetEventKind = "github:closed"
	ChangesetEventKindGitHubCommented            ChangesetEventKind = "github:commented"
	ChangesetEventKindGitHubRenamedTitle         ChangesetEventKind = "github:renamed"
	ChangesetEventKindGitHubMerged               ChangesetEventKind = "github:merged"
	ChangesetEventKindGitHubReviewed             ChangesetEventKind = "github:reviewed"
	ChangesetEventKindGitHubReopened             ChangesetEventKind = "github:reopened"
	ChangesetEventKindGitHubReviewDismissed      ChangesetEventKind = "github:review_dismissed"
	ChangesetEventKindGitHubReviewRequestRemoved ChangesetEventKind = "github:review_request_removed"
	ChangesetEventKindGitHubReviewRequested      ChangesetEventKind = "github:review_requested"
	ChangesetEventKindGitHubReviewCommented      ChangesetEventKind = "github:review_commented"
	ChangesetEventKindGitHubUnassigned           ChangesetEventKind = "github:unassigned"
	ChangesetEventKindGitHubCommit               ChangesetEventKind = "github:commit"
	ChangesetEventKindGitHubLabeled              ChangesetEventKind = "github:labeled"
	ChangesetEventKindGitHubUnlabeled            ChangesetEventKind = "github:unlabeled"
	ChangesetEventKindCommitStatus               ChangesetEventKind = "github:commit_status"
	ChangesetEventKindCheckSuite                 ChangesetEventKind = "github:check_suite"
	ChangesetEventKindCheckRun                   ChangesetEventKind = "github:check_run"

	ChangesetEventKindBitbucketServerApproved     ChangesetEventKind = "bitbucketserver:approved"
	ChangesetEventKindBitbucketServerUnapproved   ChangesetEventKind = "bitbucketserver:unapproved"
	ChangesetEventKindBitbucketServerDeclined     ChangesetEventKind = "bitbucketserver:declined"
	ChangesetEventKindBitbucketServerReviewed     ChangesetEventKind = "bitbucketserver:reviewed"
	ChangesetEventKindBitbucketServerOpened       ChangesetEventKind = "bitbucketserver:opened"
	ChangesetEventKindBitbucketServerReopened     ChangesetEventKind = "bitbucketserver:reopened"
	ChangesetEventKindBitbucketServerRescoped     ChangesetEventKind = "bitbucketserver:rescoped"
	ChangesetEventKindBitbucketServerUpdated      ChangesetEventKind = "bitbucketserver:updated"
	ChangesetEventKindBitbucketServerCommented    ChangesetEventKind = "bitbucketserver:commented"
	ChangesetEventKindBitbucketServerMerged       ChangesetEventKind = "bitbucketserver:merged"
	ChangesetEventKindBitbucketServerCommitStatus ChangesetEventKind = "bitbucketserver:commit_status"

	// BitbucketServer calls this an Unapprove event but we've called it Dismissed to more
	// clearly convey that it only occurs when a request for changes has been dismissed.
	ChangesetEventKindBitbucketServerDismissed ChangesetEventKind = "bitbucketserver:participant_status:unapproved"

	ChangesetEventKindGitLabApproved   ChangesetEventKind = "gitlab:approved"
	ChangesetEventKindGitLabPipeline   ChangesetEventKind = "gitlab:pipeline"
	ChangesetEventKindGitLabUnapproved ChangesetEventKind = "gitlab:unapproved"
)

// ChangesetSyncData represents data about the sync status of a changeset
type ChangesetSyncData struct {
	ChangesetID int64
	// UpdatedAt is the time we last updated / synced the changeset in our DB
	UpdatedAt time.Time
	// LatestEvent is the time we received the most recent changeset event
	LatestEvent time.Time
	// ExternalUpdatedAt is the time the external changeset last changed
	ExternalUpdatedAt time.Time
	// RepoExternalServiceID is the external_service_id in the repo table, usually
	// represented by the code host URL
	RepoExternalServiceID string
}

func MarshalCampaignID(id int64) graphql.ID {
	return relay.MarshalID("Campaign", id)
}

func UnmarshalCampaignID(id graphql.ID) (campaignID int64, err error) {
	err = relay.UnmarshalSpec(id, &campaignID)
	return
}

func unixMilliToTime(ms int64) time.Time {
	return time.Unix(0, ms*int64(time.Millisecond))
}

// ****************************
// TODO: NEW CAMPAIGNS WORKFLOW BELOW
// ****************************

func NewCampaignSpecFromRaw(rawSpec string) (*CampaignSpec, error) {
	c := &CampaignSpec{RawSpec: rawSpec}

	return c, c.UnmarshalValidate()
}

type CampaignSpec struct {
	ID     int64
	RandID string

	RawSpec string
	Spec    CampaignSpecFields

	NamespaceUserID int32
	NamespaceOrgID  int32

	UserID int32

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Clone returns a clone of a CampaignSpec.
func (cs *CampaignSpec) Clone() *CampaignSpec {
	cc := *cs
	return &cc
}

// UnmarshalValidate unmarshals the RawSpec into Spec and validates it against
// the CampaignSpec schema and does additional semantic validation.
func (cs *CampaignSpec) UnmarshalValidate() error {
	return unmarshalValidate(schema.CampaignSpecSchemaJSON, []byte(cs.RawSpec), &cs.Spec)
}

// CampaignSpecTTL specifies the TTL of CampaignSpecs that haven't been applied
// yet.
const CampaignSpecTTL = 7 * 24 * time.Hour

// ExpiresAt returns the time when the CampaignSpec will be deleted if not
// applied.
func (cs *CampaignSpec) ExpiresAt() time.Time {
	return cs.CreatedAt.Add(CampaignSpecTTL)
}

type CampaignSpecFields struct {
	Name              string             `json:"name"`
	Description       string             `json:"description"`
	On                []CampaignSpecOn   `json:"on"`
	Steps             []CampaignSpecStep `json:"steps"`
	ChangesetTemplate ChangesetTemplate  `json:"changesetTemplate"`
}

type CampaignSpecOn struct {
	RepositoriesMatchingQuery string `json:"repositoriesMatchingQuery,omitempty"`
	Repository                string `json:"repository,omitempty"`
}

type CampaignSpecStep struct {
	Run       string            `json:"run"`
	Container string            `json:"container"`
	Env       map[string]string `json:"env"`
}

type ChangesetTemplate struct {
	Title     string         `json:"title"`
	Body      string         `json:"body"`
	Branch    string         `json:"branch"`
	Commit    CommitTemplate `json:"commit"`
	Published bool           `json:"published"`
}

type CommitTemplate struct {
	Message string `json:"message"`
}

func NewChangesetSpecFromRaw(rawSpec string) (*ChangesetSpec, error) {
	c := &ChangesetSpec{RawSpec: rawSpec}
	err := c.UnmarshalValidate()
	if err != nil {
		return nil, err
	}

	c.computeDiffStat()

	return c, nil
}

type ChangesetSpec struct {
	ID     int64
	RandID string

	RawSpec string
	// TODO(mrnugget): should we rename the "spec" column to "description"?
	Spec ChangesetSpecDescription

	DiffStatAdded   int32
	DiffStatChanged int32
	DiffStatDeleted int32

	CampaignSpecID int64
	RepoID         api.RepoID
	UserID         int32

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Clone returns a clone of a ChangesetSpec.
func (cs *ChangesetSpec) Clone() *ChangesetSpec {
	cc := *cs
	return &cc
}

// computeDiffStat parses the Diff of the ChangesetSpecDescription and sets the
// diff stat fields that can be retrieved with DiffStat().
// If the Diff is invalid or parsing failed, an error is returned.
func (cs *ChangesetSpec) computeDiffStat() error {
	if cs.Spec.IsExisting() {
		return nil
	}

	d, err := cs.Spec.Diff()
	if err != nil {
		return err
	}

	stats := diff.Stat{}
	reader := diff.NewMultiFileDiffReader(strings.NewReader(d))
	for {
		fileDiff, err := reader.ReadFile()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		stat := fileDiff.Stat()
		stats.Added += stat.Added
		stats.Deleted += stat.Deleted
		stats.Changed += stat.Changed
	}

	cs.DiffStatAdded = stats.Added
	cs.DiffStatDeleted = stats.Deleted
	cs.DiffStatChanged = stats.Changed

	return nil
}

// DiffStat returns a *diff.Stat.
func (cs *ChangesetSpec) DiffStat() diff.Stat {
	return diff.Stat{
		Added:   cs.DiffStatAdded,
		Deleted: cs.DiffStatDeleted,
		Changed: cs.DiffStatChanged,
	}
}

// UnmarshalValidate unmarshals the RawSpec into Spec and validates it against
// the ChangesetSpec schema and does additional semantic validation.
func (cs *ChangesetSpec) UnmarshalValidate() error {
	err := unmarshalValidate(schema.ChangesetSpecSchemaJSON, []byte(cs.RawSpec), &cs.Spec)
	if err != nil {
		return err
	}

	headRepo := cs.Spec.HeadRepository
	baseRepo := cs.Spec.BaseRepository
	if headRepo != "" && baseRepo != "" && headRepo != baseRepo {
		return ErrHeadBaseMismatch
	}

	return nil
}

// ChangesetSpecTTL specifies the TTL of ChangesetSpecs that haven't been
// attached to a CampaignSpec.
// It's lower than CampaignSpecTTL because ChangesetSpecs should be attached to
// a CampaignSpec immediately after having been created, whereas a CampaignSpec
// might take a while to be complete and might also go through a lengthy review
// phase.
const ChangesetSpecTTL = 2 * 24 * time.Hour

// ExpiresAt returns the time when the ChangesetSpec will be deleted if not
// attached to a CampaignSpec.
func (cs *ChangesetSpec) ExpiresAt() time.Time {
	return cs.CreatedAt.Add(ChangesetSpecTTL)
}

// ErrHeadBaseMismatch is returned by (*ChangesetSpec).UnmarshalValidate() if
// the head and base repositories do not match (a case which we do not support
// yet).
var ErrHeadBaseMismatch = errors.New("headRepository does not match baseRepository")

type ChangesetSpecDescription struct {
	BaseRepository graphql.ID `json:"baseRepository,omitempty"`

	// If this is not empty, the description is a reference to an existing
	// changeset and the rest of these fields are empty.
	// TODO(mrnugget): Id or ID, that is the question?
	ExternalID string `json:"externalId,omitempty"`

	BaseRev string `json:"baseRev,omitempty"`
	BaseRef string `json:"baseRef,omitempty"`

	HeadRepository graphql.ID `json:"headRepository,omitempty"`
	HeadRef        string     `json:"headRef,omitempty"`

	Title string `json:"title,omitempty"`
	Body  string `json:"body,omitempty"`

	Commits []GitCommitDescription `json:"commits,omitempty"`

	Published bool `json:"published,omitempty"`
}

// Type returns the ChangesetSpecDescriptionType of the ChangesetSpecDescription.
func (d *ChangesetSpecDescription) Type() ChangesetSpecDescriptionType {
	if d.ExternalID != "" {
		return ChangesetSpecDescriptionTypeExisting
	}
	return ChangesetSpecDescriptionTypeBranch
}

// IsExisting returns whether the description is of type
// ChangesetSpecDescriptionTypeExisting.
func (d *ChangesetSpecDescription) IsExisting() bool {
	return d.Type() == ChangesetSpecDescriptionTypeExisting
}

// IsBranch returns whether the description is of type
// ChangesetSpecDescriptionTypeBranch.
func (d *ChangesetSpecDescription) IsBranch() bool {
	return d.Type() == ChangesetSpecDescriptionTypeBranch
}

// ChangesetSpecDescriptionType tells the consumer what the type of a
// ChangesetSpecDescription is without having to look into the description.
// Useful in the GraphQL when a HiddenChangesetSpec is returned.
type ChangesetSpecDescriptionType string

// Valid ChangesetSpecDescriptionTypes kinds
const (
	ChangesetSpecDescriptionTypeExisting ChangesetSpecDescriptionType = "EXISTING"
	ChangesetSpecDescriptionTypeBranch   ChangesetSpecDescriptionType = "BRANCH"
)

// ErrNoCommits is returned by (*ChangesetSpecDescription).Diff if the
// description doesn't have any commits descriptions.
var ErrNoCommits = errors.New("changeset description doesn't contain commit descriptions")

// Diff returns the Diff of the first GitCommitDescription in Commits. If the
// ChangesetSpecDescription doesn't have Commits it returns ErrNoCommits.
//
// We currently only support a single commit in Commits. Once we support more,
// this method will need to be revisited.
func (d *ChangesetSpecDescription) Diff() (string, error) {
	if len(d.Commits) == 0 {
		return "", ErrNoCommits
	}
	return d.Commits[0].Diff, nil
}

type GitCommitDescription struct {
	Message string `json:"message,omitempty"`
	Diff    string `json:"diff,omitempty"`
}

// unmarshalValidate validates the input, which can be YAML or JSON, against
// the provided JSON schema. If the validation is successful is unmarshals the
// validated input into the target.
func unmarshalValidate(schema string, input []byte, target interface{}) error {
	sl := gojsonschema.NewSchemaLoader()
	sc, err := sl.Compile(gojsonschema.NewStringLoader(schema))
	if err != nil {
		return errors.Wrap(err, "failed to compile JSON schema")
	}

	normalized, err := yaml.YAMLToJSONCustom(input, yamlv3.Unmarshal)
	if err != nil {
		return errors.Wrapf(err, "failed to normalize JSON")
	}

	res, err := sc.Validate(gojsonschema.NewBytesLoader(normalized))
	if err != nil {
		return errors.Wrap(err, "failed to validate input against schema")
	}

	var errs *multierror.Error
	for _, err := range res.Errors() {
		e := err.String()
		// Remove `(root): ` from error formatting since these errors are
		// presented to users.
		e = strings.TrimPrefix(e, "(root): ")
		errs = multierror.Append(errs, errors.New(e))
	}

	if err := json.Unmarshal(normalized, target); err != nil {
		errs = multierror.Append(errs, err)
	}

	return errs.ErrorOrNil()
}
