package testing

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/go-diff/diff"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type TestChangesetOpts struct {
	Repo         api.RepoID
	Campaign     int64
	CurrentSpec  int64
	PreviousSpec int64
	CampaignIDs  []int64

	ExternalServiceType string
	ExternalID          string
	ExternalBranch      string
	ExternalState       campaigns.ChangesetExternalState
	ExternalReviewState campaigns.ChangesetReviewState
	ExternalCheckState  campaigns.ChangesetCheckState

	PublicationState campaigns.ChangesetPublicationState

	ReconcilerState campaigns.ReconcilerState
	FailureMessage  string
	NumFailures     int64

	OwnedByCampaign int64

	Closing bool

	Metadata interface{}
}

type CreateChangeseter interface {
	CreateChangeset(ctx context.Context, changeset *campaigns.Changeset) error
}

func CreateChangeset(
	t *testing.T,
	ctx context.Context,
	store CreateChangeseter,
	opts TestChangesetOpts,
) *campaigns.Changeset {
	t.Helper()

	changeset := BuildChangeset(opts)

	if err := store.CreateChangeset(ctx, changeset); err != nil {
		t.Fatalf("creating changeset failed: %s", err)
	}

	return changeset
}

func BuildChangeset(opts TestChangesetOpts) *campaigns.Changeset {
	if opts.ExternalServiceType == "" {
		opts.ExternalServiceType = extsvc.TypeGitHub
	}

	changeset := &campaigns.Changeset{
		RepoID:         opts.Repo,
		CurrentSpecID:  opts.CurrentSpec,
		PreviousSpecID: opts.PreviousSpec,
		CampaignIDs:    opts.CampaignIDs,

		ExternalServiceType: opts.ExternalServiceType,
		ExternalID:          opts.ExternalID,
		ExternalState:       opts.ExternalState,
		ExternalReviewState: opts.ExternalReviewState,
		ExternalCheckState:  opts.ExternalCheckState,

		PublicationState: opts.PublicationState,

		OwnedByCampaignID: opts.OwnedByCampaign,

		Closing: opts.Closing,

		ReconcilerState: opts.ReconcilerState,
		NumFailures:     opts.NumFailures,

		Metadata: opts.Metadata,
	}

	if opts.ExternalBranch != "" {
		changeset.ExternalBranch = git.EnsureRefPrefix(opts.ExternalBranch)
	}

	if opts.FailureMessage != "" {
		changeset.FailureMessage = &opts.FailureMessage
	}

	if opts.Campaign != 0 {
		changeset.CampaignIDs = []int64{opts.Campaign}
	}

	return changeset
}

type ChangesetAssertions struct {
	Repo             api.RepoID
	CurrentSpec      int64
	PreviousSpec     int64
	OwnedByCampaign  int64
	ReconcilerState  campaigns.ReconcilerState
	PublicationState campaigns.ChangesetPublicationState
	ExternalState    campaigns.ChangesetExternalState
	ExternalID       string
	ExternalBranch   string
	DiffStat         *diff.Stat
	Closing          bool

	Title string
	Body  string

	FailureMessage *string
	NumFailures    int64
}

func AssertChangeset(t *testing.T, c *campaigns.Changeset, a ChangesetAssertions) {
	t.Helper()

	if c == nil {
		t.Fatalf("changeset is nil")
	}

	if have, want := c.RepoID, a.Repo; have != want {
		t.Fatalf("changeset RepoID wrong. want=%d, have=%d", want, have)
	}

	if have, want := c.CurrentSpecID, a.CurrentSpec; have != want {
		t.Fatalf("changeset CurrentSpecID wrong. want=%d, have=%d", want, have)
	}

	if have, want := c.PreviousSpecID, a.PreviousSpec; have != want {
		t.Fatalf("changeset PreviousSpecID wrong. want=%d, have=%d", want, have)
	}

	if have, want := c.OwnedByCampaignID, a.OwnedByCampaign; have != want {
		t.Fatalf("changeset OwnedByCampaignID wrong. want=%d, have=%d", want, have)
	}

	if have, want := c.ReconcilerState, a.ReconcilerState; have != want {
		t.Fatalf("changeset ReconcilerState wrong. want=%s, have=%s", want, have)
	}

	if have, want := c.PublicationState, a.PublicationState; have != want {
		t.Fatalf("changeset PublicationState wrong. want=%s, have=%s", want, have)
	}

	if have, want := c.ExternalState, a.ExternalState; have != want {
		t.Fatalf("changeset ExternalState wrong. want=%s, have=%s", want, have)
	}

	if have, want := c.ExternalID, a.ExternalID; have != want {
		t.Fatalf("changeset ExternalID wrong. want=%s, have=%s", want, have)
	}

	if have, want := c.ExternalBranch, a.ExternalBranch; have != want {
		t.Fatalf("changeset ExternalBranch wrong. want=%s, have=%s", want, have)
	}

	if want, have := a.FailureMessage, c.FailureMessage; want == nil && have != nil {
		t.Fatalf("expected no failure message, but have=%q", *have)
	}

	if diff := cmp.Diff(a.DiffStat, c.DiffStat()); diff != "" {
		t.Fatalf("changeset DiffStat wrong. (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff(a.Closing, c.Closing); diff != "" {
		t.Fatalf("changeset Closing wrong. (-want +got):\n%s", diff)
	}

	if want := c.FailureMessage; want != nil {
		if c.FailureMessage == nil {
			t.Fatalf("expected failure message %q but have none", *want)
		}
		if want, have := *a.FailureMessage, *c.FailureMessage; have != want {
			t.Fatalf("wrong failure message. want=%q, have=%q", want, have)
		}
	}

	if have, want := c.NumFailures, a.NumFailures; have != want {
		t.Fatalf("changeset NumFailures wrong. want=%d, have=%d", want, have)
	}

	if have, want := c.ExternalBranch, a.ExternalBranch; have != want {
		t.Fatalf("changeset ExternalBranch wrong. want=%s, have=%s", want, have)
	}

	if want := a.Title; want != "" {
		have, err := c.Title()
		if err != nil {
			t.Fatalf("changeset.Title failed: %s", err)
		}

		if have != want {
			t.Fatalf("changeset Title wrong. want=%s, have=%s", want, have)
		}
	}

	if want := a.Body; want != "" {
		have, err := c.Body()
		if err != nil {
			t.Fatalf("changeset.Body failed: %s", err)
		}

		if have != want {
			t.Fatalf("changeset Body wrong. want=%s, have=%s", want, have)
		}
	}
}

type GetChangesetByIDer interface {
	GetChangesetByID(ctx context.Context, id int64) (*campaigns.Changeset, error)
}

func ReloadAndAssertChangeset(t *testing.T, ctx context.Context, s GetChangesetByIDer, c *campaigns.Changeset, a ChangesetAssertions) (reloaded *campaigns.Changeset) {
	t.Helper()

	reloaded, err := s.GetChangesetByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("reloading changeset %d failed: %s", c.ID, err)
	}

	AssertChangeset(t, reloaded, a)

	return reloaded
}

type UpdateChangeseter interface {
	UpdateChangeset(ctx context.Context, changeset *campaigns.Changeset) error
}

func SetChangesetPublished(t *testing.T, ctx context.Context, s UpdateChangeseter, c *campaigns.Changeset, externalID, externalBranch string) {
	t.Helper()

	c.ExternalBranch = externalBranch
	c.ExternalID = externalID
	c.PublicationState = campaigns.ChangesetPublicationStatePublished
	c.ReconcilerState = campaigns.ReconcilerStateCompleted
	c.ExternalState = campaigns.ChangesetExternalStateOpen

	if err := s.UpdateChangeset(ctx, c); err != nil {
		t.Fatalf("failed to update changeset: %s", err)
	}
}

var FailedChangesetFailureMessage = "Failed test"

func SetChangesetFailed(t *testing.T, ctx context.Context, s UpdateChangeseter, c *campaigns.Changeset) {
	t.Helper()

	c.ReconcilerState = campaigns.ReconcilerStateErrored
	c.FailureMessage = &FailedChangesetFailureMessage
	c.NumFailures = 5

	if err := s.UpdateChangeset(ctx, c); err != nil {
		t.Fatalf("failed to update changeset: %s", err)
	}
}

func SetChangesetClosed(t *testing.T, ctx context.Context, s UpdateChangeseter, c *campaigns.Changeset) {
	t.Helper()

	c.PublicationState = campaigns.ChangesetPublicationStatePublished
	c.ReconcilerState = campaigns.ReconcilerStateCompleted
	c.Closing = false
	c.ExternalState = campaigns.ChangesetExternalStateClosed

	if err := s.UpdateChangeset(ctx, c); err != nil {
		t.Fatalf("failed to update changeset: %s", err)
	}
}
