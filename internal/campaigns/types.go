package campaigns

import (
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/vcs/git"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// SupportedExternalServices are the external service types currently supported
// by the campaigns feature. Repos that are associated with external services
// whose type is not in this list will simply be filtered out from the search
// results.
var SupportedExternalServices = map[string]struct{}{
	github.ServiceType:          {},
	bitbucketserver.ServiceType: {},
}

// IsRepoSupported returns whether the given ExternalRepoSpec is supported by
// the campaigns feature, based on the external service type.
func IsRepoSupported(spec *api.ExternalRepoSpec) bool {
	_, ok := SupportedExternalServices[spec.ServiceType]
	return ok
}

// CampaignPlanPatch is a patch applied to a repository (to create a new branch).
type CampaignPlanPatch struct {
	Repo api.RepoID
	// The commit SHA this patch is based on (e.g.: "4095572721c6234cd72013fd49dff4fb48f0f8a4").
	BaseRevision api.CommitID
	// The ref name that pointed to the BaseRevision at the time of patch creation (e.g.: "refs/heads/master").
	BaseRef string
	Patch   string
}

// A CampaignPlan is a collection of multiple CampaignJobs.
type CampaignPlan struct {
	ID int64

	UserID int32

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Clone returns a clone of a CampaignPlan.
func (c *CampaignPlan) Clone() *CampaignPlan {
	cc := *c
	return &cc
}

// A CampaignJob is the application of a CampaignType over CampaignPlan arguments in
// a specific repository at a specific revision.
type CampaignJob struct {
	ID             int64
	CampaignPlanID int64

	RepoID  api.RepoID
	Rev     api.CommitID
	BaseRef string

	Diff string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Clone returns a clone of a CampaignJob.
func (c *CampaignJob) Clone() *CampaignJob {
	cc := *c
	return &cc
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
	CampaignPlanID  int64
	ClosedAt        time.Time
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

// ChangesetState defines the possible states of a Changeset.
type ChangesetState string

// ChangesetState constants.
const (
	ChangesetStateOpen    ChangesetState = "OPEN"
	ChangesetStateClosed  ChangesetState = "CLOSED"
	ChangesetStateMerged  ChangesetState = "MERGED"
	ChangesetStateDeleted ChangesetState = "DELETED"
)

// Valid returns true if the given Changeset is valid.
func (s ChangesetState) Valid() bool {
	switch s {
	case ChangesetStateOpen,
		ChangesetStateClosed,
		ChangesetStateMerged,
		ChangesetStateDeleted:
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

// BackgroundProcessStatus defines the status of a background process.
type BackgroundProcessStatus struct {
	Canceled      bool
	Total         int32
	Completed     int32
	Pending       int32
	ProcessState  BackgroundProcessState
	ProcessErrors []string
}

func (b BackgroundProcessStatus) CompletedCount() int32         { return b.Completed }
func (b BackgroundProcessStatus) PendingCount() int32           { return b.Pending }
func (b BackgroundProcessStatus) State() BackgroundProcessState { return b.ProcessState }
func (b BackgroundProcessStatus) Errors() []string              { return b.ProcessErrors }
func (b BackgroundProcessStatus) Finished() bool {
	return b.ProcessState != BackgroundProcessStateProcessing
}
func (b BackgroundProcessStatus) Processing() bool {
	return b.ProcessState == BackgroundProcessStateProcessing
}

// BackgroundProcessState defines the possible states of a background process.
type BackgroundProcessState string

// BackgroundProcessState constants
const (
	BackgroundProcessStateProcessing BackgroundProcessState = "PROCESSING"
	BackgroundProcessStateErrored    BackgroundProcessState = "ERRORED"
	BackgroundProcessStateCompleted  BackgroundProcessState = "COMPLETED"
	BackgroundProcessStateCanceled   BackgroundProcessState = "CANCELED"

	// Remember to update Finished() above if a new state is added
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

// A ChangesetJob is the creation of a Changeset on an external host from a
// local CampaignJob for a given Campaign.
type ChangesetJob struct {
	ID            int64
	CampaignID    int64
	CampaignJobID int64

	// Only set once the ChangesetJob has successfully finished.
	ChangesetID int64

	Branch string

	Error string

	StartedAt  time.Time
	FinishedAt time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

// Clone returns a clone of a ChangesetJob.
func (c *ChangesetJob) Clone() *ChangesetJob {
	cc := *c
	return &cc
}

// SuccessfullyCompleted returns true for jobs that have already successfully run
func (c *ChangesetJob) SuccessfullyCompleted() bool {
	return c.Error == "" && !c.FinishedAt.IsZero() && c.ChangesetID != 0
}

// Reset sets the Error, StartedAt and FinishedAt fields to their respective
// zero values, so that the ChangesetJob can be executed again.
func (c *ChangesetJob) Reset() {
	c.Error = ""
	c.StartedAt = time.Time{}
	c.FinishedAt = time.Time{}
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
	ExternalState       ChangesetState
	ExternalReviewState ChangesetReviewState
	ExternalCheckState  ChangesetCheckState
}

// Clone returns a clone of a Changeset.
func (c *Changeset) Clone() *Changeset {
	tt := *c
	tt.CampaignIDs = c.CampaignIDs[:len(c.CampaignIDs):len(c.CampaignIDs)]
	return &tt
}

func (c *Changeset) SetMetadata(meta interface{}) error {
	switch pr := meta.(type) {
	case *github.PullRequest:
		c.Metadata = pr
		c.ExternalID = strconv.FormatInt(pr.Number, 10)
		c.ExternalServiceType = github.ServiceType
		c.ExternalBranch = pr.HeadRefName
		c.ExternalUpdatedAt = pr.UpdatedAt
	case *bitbucketserver.PullRequest:
		c.Metadata = pr
		c.ExternalID = strconv.FormatInt(int64(pr.ID), 10)
		c.ExternalServiceType = bitbucketserver.ServiceType
		c.ExternalBranch = git.AbbreviateRef(pr.FromRef.ID)
		c.ExternalUpdatedAt = unixMilliToTime(int64(pr.UpdatedDate))
	default:
		return errors.New("unknown changeset type")
	}
	return nil
}

// SetDerivedState will update the external state fields on c based on the current
// state of the changeset and associated events.
func (c *Changeset) SetDerivedState(es []*ChangesetEvent) {
	// Copy so that we can sort without mutating the argument
	events := make(ChangesetEvents, len(es))
	copy(events, es)
	sort.Sort(events)

	if state, err := ComputeChangesetState(c, events); err != nil {
		log15.Warn("Computing changeset state", "err", err)
	} else {
		c.ExternalState = state
	}
	if state, err := ComputeReviewState(c, events); err != nil {
		log15.Warn("Computing changeset review state", "err", err)
	} else {
		c.ExternalReviewState = state
	}
	c.ExternalCheckState = ComputeCheckState(c, events)
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

// state of a Changeset based on the metadata.
// It does NOT reflect the final calculated state, use `ExternalState` instead.
func (c *Changeset) state() (s ChangesetState, err error) {
	if !c.ExternalDeletedAt.IsZero() {
		return ChangesetStateDeleted, nil
	}

	switch m := c.Metadata.(type) {
	case *github.PullRequest:
		s = ChangesetState(m.State)
	case *bitbucketserver.PullRequest:
		if m.State == "DECLINED" {
			s = ChangesetStateClosed
		} else {
			s = ChangesetState(m.State)
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
	default:
		return "", errors.New("unknown changeset type")
	}
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
			default:
				ev.Key = ti.Item.(interface{ Key() string }).Key()
				ev.Kind = ChangesetEventKindFor(ti.Item)
				ev.Metadata = ti.Item
				events = append(events, &ev)
			}
		}

	case *bitbucketserver.PullRequest:
		events = make([]*ChangesetEvent, 0, len(m.Activities))
		for _, a := range m.Activities {
			events = append(events, &ChangesetEvent{
				ChangesetID: c.ID,
				Key:         a.Key(),
				Kind:        ChangesetEventKindFor(a),
				Metadata:    a,
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
	default:
		return "", errors.New("unknown changeset type")
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
	default:
		return []ChangesetLabel{}
	}
}

// reviewState of a Changeset. GitHub doesn't keep the review state on a
// changeset, so a GitHub Changeset will always return
// ChangesetReviewStatePending.
// This method should not be called directly. Use ComputeReviewState instead.
func (c *Changeset) reviewState() (s ChangesetReviewState, err error) {
	states := map[ChangesetReviewState]bool{}

	switch m := c.Metadata.(type) {
	case *github.PullRequest:
		// For GitHub we need to use `ChangesetEvents.ReviewState`
		log15.Warn("Changeset.ReviewState() called, but GitHub review state is calculated through ChangesetEvents.ReviewState", "changeset", c)
		return ChangesetReviewStatePending, nil

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

	return SelectReviewState(states), nil
}

// SelectReviewState computes the single review state for a given set of
// ChangesetReviewStates. Since a pull request, for example, can have multiple
// reviews with different states, we need a function to determine what the
// state for the pull request is.
func SelectReviewState(states map[ChangesetReviewState]bool) ChangesetReviewState {
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

// ChangesetEvents is a collection of changeset events
type ChangesetEvents []*ChangesetEvent

func (ce ChangesetEvents) Len() int      { return len(ce) }
func (ce ChangesetEvents) Swap(i, j int) { ce[i], ce[j] = ce[j], ce[i] }

// Less sorts changeset events by their Timestamps
func (ce ChangesetEvents) Less(i, j int) bool {
	return ce[i].Timestamp().Before(ce[j].Timestamp())
}

// reviewState returns the overall review state of the review events in the
// slice.
// It should only be called by ComputeChangesetReviewState.
func (ce ChangesetEvents) reviewState() (ChangesetReviewState, error) {
	reviewsByAuthor := map[string]ChangesetReviewState{}

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
		case ChangesetReviewStateApproved,
			ChangesetReviewStateChangesRequested:
			reviewsByAuthor[author] = s
		case ChangesetReviewStateDismissed:
			delete(reviewsByAuthor, author)
		}
	}

	states := make(map[ChangesetReviewState]bool)
	for _, s := range reviewsByAuthor {
		states[s] = true
	}
	return SelectReviewState(states), nil
}

// State returns the  state of the changeset to which the events belong and assumes the events
// are sorted by ChangesetEvent.Timestamp().
func (ce ChangesetEvents) State() ChangesetState {
	state := ChangesetStateOpen
	for _, e := range ce {
		switch e.Kind {
		case ChangesetEventKindGitHubClosed, ChangesetEventKindBitbucketServerDeclined:
			state = ChangesetStateClosed
		case ChangesetEventKindGitHubMerged, ChangesetEventKindBitbucketServerMerged:
			state = ChangesetStateMerged
		case ChangesetEventKindGitHubReopened, ChangesetEventKindBitbucketServerReopened:
			state = ChangesetStateOpen
		}
	}
	return state
}

// ComputeCheckState computes the overall check state based on the current synced check state
// and any webhook events that have arrived after the most recent sync
func ComputeCheckState(c *Changeset, events ChangesetEvents) ChangesetCheckState {
	switch m := c.Metadata.(type) {
	case *github.PullRequest:
		return computeGitHubCheckState(c.UpdatedAt, m, events)

	case *bitbucketserver.PullRequest:
		return computeBitbucketBuildStatus(m)
	}

	return ChangesetCheckStateUnknown
}

// ComputeChangesetState computes the overall state for the changeset and its
// associated events. The events should be presorted.
func ComputeChangesetState(c *Changeset, events ChangesetEvents) (ChangesetState, error) {
	if len(events) == 0 {
		return c.state()
	}
	newestEvent := events[len(events)-1]
	if c.UpdatedAt.After(newestEvent.Timestamp()) {
		return c.state()
	}
	return events.State(), nil
}

// ComputeReviewState computes the review state for the changeset and its
// associated events. The events should be presorted.
func ComputeReviewState(c *Changeset, events ChangesetEvents) (ChangesetReviewState, error) {
	if len(events) == 0 {
		return c.reviewState()
	}

	// GitHub only stores the ReviewState in events, we can't look at the
	// Changeset.
	if c.ExternalServiceType == github.ServiceType {
		return events.reviewState()
	}

	// For other codehosts we check whether the Changeset is newer or the
	// events and use the newest entity to get the reviewstate.
	newestEvent := events[len(events)-1]
	if c.UpdatedAt.After(newestEvent.Timestamp()) {
		return c.reviewState()
	}
	return events.reviewState()
}

func computeBitbucketBuildStatus(pr *bitbucketserver.PullRequest) ChangesetCheckState {
	var states []ChangesetCheckState
	for _, status := range pr.BuildStatuses {
		states = append(states, parseBitbucketBuildState(status.State))
	}
	return combineCheckStates(states)
}

func parseBitbucketBuildState(s string) ChangesetCheckState {
	switch s {
	case "FAILED":
		return ChangesetCheckStateFailed
	case "INPROGRESS":
		return ChangesetCheckStatePending
	case "SUCCESSFUL":
		return ChangesetCheckStatePassed
	default:
		return ChangesetCheckStateUnknown
	}
}

func computeGitHubCheckState(lastSynced time.Time, pr *github.PullRequest, events []*ChangesetEvent) ChangesetCheckState {
	// We should only consider the latest commit. This could be from a sync or a webhook that
	// has occurred later
	var latestCommitTime time.Time
	var latestOID string
	statusPerContext := make(map[string]ChangesetCheckState)
	statusPerCheckSuite := make(map[string]ChangesetCheckState)
	statusPerCheckRun := make(map[string]ChangesetCheckState)

	if len(pr.Commits.Nodes) > 0 {
		// We only request the most recent commit
		commit := pr.Commits.Nodes[0]
		latestCommitTime = commit.Commit.CommittedDate
		latestOID = commit.Commit.OID
		// Calc status per context for the most recent synced commit
		for _, c := range commit.Commit.Status.Contexts {
			statusPerContext[c.Context] = parseGithubCheckState(c.State)
		}
		for _, c := range commit.Commit.CheckSuites.Nodes {
			if c.Status == "QUEUED" && len(c.CheckRuns.Nodes) == 0 {
				// Ignore queued suites with no runs.
				// It is common for suites to be created and then stay in the QUEUED state
				// forever with zero runs.
				continue
			}
			statusPerCheckSuite[c.ID] = parseGithubCheckSuiteState(c.Status, c.Conclusion)
			for _, r := range c.CheckRuns.Nodes {
				statusPerCheckRun[r.ID] = parseGithubCheckSuiteState(r.Status, r.Conclusion)
			}
		}
	}

	var statuses []*github.CommitStatus
	// Get all status updates that have happened since our last sync
	for _, e := range events {
		switch m := e.Metadata.(type) {
		case *github.CommitStatus:
			if m.ReceivedAt.After(lastSynced) {
				statuses = append(statuses, m)
			}
		case *github.PullRequestCommit:
			if m.Commit.CommittedDate.After(latestCommitTime) {
				latestCommitTime = m.Commit.CommittedDate
				latestOID = m.Commit.OID
				// statusPerContext is now out of date, reset it
				for k := range statusPerContext {
					delete(statusPerContext, k)
				}
			}
		case *github.CheckSuite:
			if m.Status == "QUEUED" && len(m.CheckRuns.Nodes) == 0 {
				// Ignore suites with no runs.
				// See previous comment.
				continue
			}
			if m.ReceivedAt.After(lastSynced) {
				statusPerCheckSuite[m.ID] = parseGithubCheckSuiteState(m.Status, m.Conclusion)
			}
		case *github.CheckRun:
			if m.ReceivedAt.After(lastSynced) {
				statusPerCheckRun[m.ID] = parseGithubCheckSuiteState(m.Status, m.Conclusion)
			}
		}
	}

	if len(statuses) > 0 {
		// Update the statuses using any new webhook events for the latest commit
		sort.Slice(statuses, func(i, j int) bool {
			return statuses[i].ReceivedAt.Before(statuses[j].ReceivedAt)
		})
		for _, s := range statuses {
			if s.SHA != latestOID {
				continue
			}
			statusPerContext[s.Context] = parseGithubCheckState(s.State)
		}
	}
	finalStates := make([]ChangesetCheckState, 0, len(statusPerContext))
	for k := range statusPerContext {
		finalStates = append(finalStates, statusPerContext[k])
	}
	for k := range statusPerCheckSuite {
		finalStates = append(finalStates, statusPerCheckSuite[k])
	}
	for k := range statusPerCheckRun {
		finalStates = append(finalStates, statusPerCheckRun[k])
	}
	return combineCheckStates(finalStates)
}

// combineCheckStates combines multiple check states into an overall state
// pending takes highest priority
// followed by error
// success return only if all successful
func combineCheckStates(states []ChangesetCheckState) ChangesetCheckState {
	if len(states) == 0 {
		return ChangesetCheckStateUnknown
	}
	stateMap := make(map[ChangesetCheckState]bool)
	for _, s := range states {
		stateMap[s] = true
	}

	switch {
	case stateMap[ChangesetCheckStateUnknown]:
		// If are pending, overall is Pending
		return ChangesetCheckStateUnknown
	case stateMap[ChangesetCheckStatePending]:
		// If are pending, overall is Pending
		return ChangesetCheckStatePending
	case stateMap[ChangesetCheckStateFailed]:
		// If no pending, but have errors then overall is Failed
		return ChangesetCheckStateFailed
	case stateMap[ChangesetCheckStatePassed]:
		// No pending or errors then overall is Passed
		return ChangesetCheckStatePassed
	}

	return ChangesetCheckStateUnknown
}

func parseGithubCheckState(s string) ChangesetCheckState {
	s = strings.ToUpper(s)
	switch s {
	case "ERROR", "FAILURE":
		return ChangesetCheckStateFailed
	case "EXPECTED", "PENDING":
		return ChangesetCheckStatePending
	case "SUCCESS":
		return ChangesetCheckStatePassed
	default:
		return ChangesetCheckStateUnknown
	}
}

func parseGithubCheckSuiteState(status, conclusion string) ChangesetCheckState {
	status = strings.ToUpper(status)
	conclusion = strings.ToUpper(conclusion)
	switch status {
	case "IN_PROGRESS", "QUEUED", "REQUESTED":
		return ChangesetCheckStatePending
	}
	if status != "COMPLETED" {
		return ChangesetCheckStateUnknown
	}
	switch conclusion {
	case "SUCCESS", "NEUTRAL":
		return ChangesetCheckStatePassed
	case "ACTION_REQUIRED":
		return ChangesetCheckStatePending
	case "CANCELLED", "FAILURE", "TIMED_OUT":
		return ChangesetCheckStateFailed
	}
	return ChangesetCheckStateUnknown
}

// UpdateLabelsSince returns the set of current labels based the starting set of labels and looking at events
// that have occurred after "since".
func (ce *ChangesetEvents) UpdateLabelsSince(cs *Changeset) []ChangesetLabel {
	var current []ChangesetLabel
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
	set := make(map[string]ChangesetLabel)
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
			set[e.Label.Name] = ChangesetLabel{
				Name:        e.Label.Name,
				Color:       e.Label.Color,
				Description: e.Label.Description,
			}
		}
	}
	labels := make([]ChangesetLabel, 0, len(set))
	for _, label := range set {
		labels = append(labels, label)
	}
	return labels
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
	default:
		return "", nil
	}
}

// ReviewState returns the review state of the ChangesetEvent if it is a review event.
func (e *ChangesetEvent) ReviewState() (ChangesetReviewState, error) {
	switch e.Kind {
	case ChangesetEventKindBitbucketServerApproved:
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
		ChangesetEventKindBitbucketServerUnapproved:
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
	default:
		panic(errors.Errorf("unknown changeset event kind for %T", e))
	}
}

// NewChangesetEventMetadata returns a new metadata object for the given
// ChangesetEventKind.
func NewChangesetEventMetadata(k ChangesetEventKind) (interface{}, error) {
	switch {
	case strings.HasPrefix(string(k), "bitbucketserver"):
		return new(bitbucketserver.Activity), nil
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

	ChangesetEventKindBitbucketServerApproved   ChangesetEventKind = "bitbucketserver:approved"
	ChangesetEventKindBitbucketServerUnapproved ChangesetEventKind = "bitbucketserver:unapproved"
	ChangesetEventKindBitbucketServerDeclined   ChangesetEventKind = "bitbucketserver:declined"
	ChangesetEventKindBitbucketServerReviewed   ChangesetEventKind = "bitbucketserver:reviewed"
	ChangesetEventKindBitbucketServerOpened     ChangesetEventKind = "bitbucketserver:opened"
	ChangesetEventKindBitbucketServerReopened   ChangesetEventKind = "bitbucketserver:reopened"
	ChangesetEventKindBitbucketServerRescoped   ChangesetEventKind = "bitbucketserver:rescoped"
	ChangesetEventKindBitbucketServerUpdated    ChangesetEventKind = "bitbucketserver:updated"
	ChangesetEventKindBitbucketServerCommented  ChangesetEventKind = "bitbucketserver:commented"
	ChangesetEventKindBitbucketServerMerged     ChangesetEventKind = "bitbucketserver:merged"
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
}

func unixMilliToTime(ms int64) time.Time {
	return time.Unix(0, ms*int64(time.Millisecond))
}
