package campaigns

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/inconshreveable/log15"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func init() {
	dbtesting.DBNameSuffix = "campaignsenterpriserdb"
}

func TestMain(m *testing.M) {
	flag.Parse()
	if !testing.Verbose() {
		log15.Root().SetHandler(log15.DiscardHandler())
	}
	os.Exit(m.Run())
}

var createTestUser = func() func(*testing.T, bool) *types.User {
	count := 0

	// This function replicates the minium amount of work required by
	// db.Users.Create to create a new user, but it doesn't require passing in
	// a full db.NewUser every time.
	return func(t *testing.T, siteAdmin bool) *types.User {
		t.Helper()

		user := &types.User{
			Username:    fmt.Sprintf("testuser-%d", count),
			DisplayName: "testuser",
		}

		q := sqlf.Sprintf("INSERT INTO users (username, site_admin) VALUES (%s, %t) RETURNING id, site_admin", user.Username, siteAdmin)
		err := dbconn.Global.QueryRow(q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&user.ID, &user.SiteAdmin)
		if err != nil {
			t.Fatal(err)
		}

		if user.SiteAdmin != siteAdmin {
			t.Fatalf("user.SiteAdmin=%t, but expected is %t", user.SiteAdmin, siteAdmin)
		}

		_, err = dbconn.Global.Exec("INSERT INTO names(name, user_id) VALUES($1, $2)", user.Username, user.ID)
		if err != nil {
			t.Fatalf("failed to create name: %s", err)
		}

		count += 1

		return user
	}
}()

func truncateTables(t *testing.T, db *sql.DB, tables ...string) {
	t.Helper()

	_, err := db.Exec("TRUNCATE " + strings.Join(tables, ", ") + " RESTART IDENTITY")
	if err != nil {
		t.Fatal(err)
	}
}

func insertTestOrg(t *testing.T, db *sql.DB) (orgID int32) {
	t.Helper()

	err := db.QueryRow("INSERT INTO orgs (name) VALUES ('bbs-org') RETURNING id").Scan(&orgID)
	if err != nil {
		t.Fatal(err)
	}

	return orgID
}

type testChangesetOpts struct {
	repo         api.RepoID
	campaign     int64
	currentSpec  int64
	previousSpec int64
	campaignIDs  []int64

	externalServiceType string
	externalID          string
	externalBranch      string
	externalState       campaigns.ChangesetExternalState

	publicationState campaigns.ChangesetPublicationState

	reconcilerState campaigns.ReconcilerState
	failureMessage  string
	numFailures     int64

	ownedByCampaign int64

	unsynced bool
	closing  bool
}

func createChangeset(
	t *testing.T,
	ctx context.Context,
	store *Store,
	opts testChangesetOpts,
) *campaigns.Changeset {
	t.Helper()

	changeset := buildChangeset(opts)

	if err := store.CreateChangeset(ctx, changeset); err != nil {
		t.Fatalf("creating changeset failed: %s", err)
	}

	return changeset
}

func buildChangeset(opts testChangesetOpts) *campaigns.Changeset {
	if opts.externalServiceType == "" {
		opts.externalServiceType = extsvc.TypeGitHub
	}

	changeset := &campaigns.Changeset{
		RepoID:         opts.repo,
		CurrentSpecID:  opts.currentSpec,
		PreviousSpecID: opts.previousSpec,
		CampaignIDs:    opts.campaignIDs,

		ExternalServiceType: opts.externalServiceType,
		ExternalID:          opts.externalID,
		ExternalBranch:      opts.externalBranch,
		ExternalState:       opts.externalState,

		PublicationState: opts.publicationState,

		OwnedByCampaignID: opts.ownedByCampaign,

		Unsynced: opts.unsynced,
		Closing:  opts.closing,

		ReconcilerState: opts.reconcilerState,
		NumFailures:     opts.numFailures,
	}

	if opts.failureMessage != "" {
		changeset.FailureMessage = &opts.failureMessage
	}

	if opts.campaign != 0 {
		changeset.CampaignIDs = []int64{opts.campaign}
	}

	return changeset
}

type changesetAssertions struct {
	repo             api.RepoID
	currentSpec      int64
	previousSpec     int64
	ownedByCampaign  int64
	reconcilerState  campaigns.ReconcilerState
	publicationState campaigns.ChangesetPublicationState
	externalState    campaigns.ChangesetExternalState
	externalID       string
	externalBranch   string
	diffStat         *diff.Stat
	unsynced         bool
	closing          bool

	title string
	body  string

	failureMessage *string
	numFailures    int64
}

func assertChangeset(t *testing.T, c *campaigns.Changeset, a changesetAssertions) {
	t.Helper()

	if c == nil {
		t.Fatalf("changeset is nil")
	}

	if have, want := c.RepoID, a.repo; have != want {
		t.Fatalf("changeset RepoID wrong. want=%d, have=%d", want, have)
	}

	if have, want := c.CurrentSpecID, a.currentSpec; have != want {
		t.Fatalf("changeset CurrentSpecID wrong. want=%d, have=%d", want, have)
	}

	if have, want := c.PreviousSpecID, a.previousSpec; have != want {
		t.Fatalf("changeset PreviousSpecID wrong. want=%d, have=%d", want, have)
	}

	if have, want := c.OwnedByCampaignID, a.ownedByCampaign; have != want {
		t.Fatalf("changeset OwnedByCampaignID wrong. want=%d, have=%d", want, have)
	}

	if have, want := c.ReconcilerState, a.reconcilerState; have != want {
		t.Fatalf("changeset ReconcilerState wrong. want=%s, have=%s", want, have)
	}

	if have, want := c.PublicationState, a.publicationState; have != want {
		t.Fatalf("changeset PublicationState wrong. want=%s, have=%s", want, have)
	}

	if have, want := c.ExternalState, a.externalState; have != want {
		t.Fatalf("changeset ExternalState wrong. want=%s, have=%s", want, have)
	}

	if have, want := c.ExternalID, a.externalID; have != want {
		t.Fatalf("changeset ExternalID wrong. want=%s, have=%s", want, have)
	}

	if have, want := c.ExternalBranch, a.externalBranch; have != want {
		t.Fatalf("changeset ExternalBranch wrong. want=%s, have=%s", want, have)
	}

	if want, have := a.failureMessage, c.FailureMessage; want == nil && have != nil {
		t.Fatalf("expected no failure message, but have=%q", *have)
	}

	if diff := cmp.Diff(a.diffStat, c.DiffStat()); diff != "" {
		t.Fatalf("changeset DiffStat wrong. (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff(a.unsynced, c.Unsynced); diff != "" {
		t.Fatalf("changeset Unsynced wrong. (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff(a.closing, c.Closing); diff != "" {
		t.Fatalf("changeset Closing wrong. (-want +got):\n%s", diff)
	}

	if want := c.FailureMessage; want != nil {
		if c.FailureMessage == nil {
			t.Fatalf("expected failure message %q but have none", *want)
		}
		if want, have := *a.failureMessage, *c.FailureMessage; have != want {
			t.Fatalf("wrong failure message. want=%q, have=%q", want, have)
		}
	}

	if have, want := c.NumFailures, a.numFailures; have != want {
		t.Fatalf("changeset NumFailures wrong. want=%d, have=%d", want, have)
	}

	if have, want := c.ExternalBranch, a.externalBranch; have != want {
		t.Fatalf("changeset ExternalBranch wrong. want=%s, have=%s", want, have)
	}

	if want := a.title; want != "" {
		have, err := c.Title()
		if err != nil {
			t.Fatalf("changeset.Title failed: %s", err)
		}

		if have != want {
			t.Fatalf("changeset Title wrong. want=%s, have=%s", want, have)
		}
	}

	if want := a.body; want != "" {
		have, err := c.Body()
		if err != nil {
			t.Fatalf("changeset.Body failed: %s", err)
		}

		if have != want {
			t.Fatalf("changeset Body wrong. want=%s, have=%s", want, have)
		}
	}
}

func reloadAndAssertChangeset(t *testing.T, ctx context.Context, s *Store, c *campaigns.Changeset, a changesetAssertions) (reloaded *campaigns.Changeset) {
	t.Helper()

	reloaded, err := s.GetChangeset(ctx, GetChangesetOpts{ID: c.ID})
	if err != nil {
		t.Fatalf("reloading changeset %d failed: %s", c.ID, err)
	}

	assertChangeset(t, reloaded, a)

	return reloaded
}

func applyAndListChangesets(ctx context.Context, t *testing.T, svc *Service, campaignSpecRandID string, wantChangesets int) (*campaigns.Campaign, campaigns.Changesets) {
	t.Helper()

	campaign, err := svc.ApplyCampaign(ctx, ApplyCampaignOpts{
		CampaignSpecRandID: campaignSpecRandID,
	})
	if err != nil {
		t.Fatalf("failed to apply campaign: %s", err)
	}

	if campaign.ID == 0 {
		t.Fatalf("campaign ID is zero")
	}

	changesets, _, err := svc.store.ListChangesets(ctx, ListChangesetsOpts{CampaignID: campaign.ID})
	if err != nil {
		t.Fatal(err)
	}

	if have, want := len(changesets), wantChangesets; have != want {
		t.Fatalf("wrong number of changesets. want=%d, have=%d", want, have)
	}

	return campaign, changesets
}

func setChangesetPublished(t *testing.T, ctx context.Context, s *Store, c *campaigns.Changeset, externalID, externalBranch string) {
	t.Helper()

	c.ExternalBranch = externalBranch
	c.ExternalID = externalID
	c.PublicationState = campaigns.ChangesetPublicationStatePublished
	c.ReconcilerState = campaigns.ReconcilerStateCompleted
	c.ExternalState = campaigns.ChangesetExternalStateOpen
	c.Unsynced = false

	if err := s.UpdateChangeset(ctx, c); err != nil {
		t.Fatalf("failed to update changeset: %s", err)
	}
}

func setChangesetFailed(t *testing.T, ctx context.Context, s *Store, c *campaigns.Changeset) {
	t.Helper()

	c.ReconcilerState = campaigns.ReconcilerStateErrored
	c.FailureMessage = &canceledChangesetFailureMessage
	c.NumFailures = 5

	if err := s.UpdateChangeset(ctx, c); err != nil {
		t.Fatalf("failed to update changeset: %s", err)
	}
}

func setChangesetClosed(t *testing.T, ctx context.Context, s *Store, c *campaigns.Changeset) {
	t.Helper()

	c.PublicationState = campaigns.ChangesetPublicationStatePublished
	c.ReconcilerState = campaigns.ReconcilerStateCompleted
	c.Closing = false
	c.ExternalState = campaigns.ChangesetExternalStateClosed

	if err := s.UpdateChangeset(ctx, c); err != nil {
		t.Fatalf("failed to update changeset: %s", err)
	}
}

type testSpecOpts struct {
	user         int32
	repo         api.RepoID
	campaignSpec int64

	// If this is non-blank, the changesetSpec will be an import/track spec for
	// the changeset with the given externalID in the given repo.
	externalID string

	// If this is set, the changesetSpec will be a "create commit on this
	// branch" changeset spec.
	headRef string

	// If this is set along with headRef, the changesetSpec will have published
	// set.
	published interface{}

	title             string
	body              string
	commitMessage     string
	commitDiff        string
	commitAuthorEmail string
	commitAuthorName  string
}

var testChangsetSpecDiffStat = &diff.Stat{Added: 10, Changed: 5, Deleted: 2}

func buildChangesetSpec(t *testing.T, opts testSpecOpts) *campaigns.ChangesetSpec {
	t.Helper()

	published := campaigns.PublishedValue{Val: opts.published}
	if opts.published == nil {
		// Set false as the default.
		published.Val = false
	}
	if !published.Valid() {
		t.Fatalf("invalid value for published passed, got %v (%T)", opts.published, opts.published)
	}

	spec := &campaigns.ChangesetSpec{
		UserID:         opts.user,
		RepoID:         opts.repo,
		CampaignSpecID: opts.campaignSpec,
		Spec: &campaigns.ChangesetSpecDescription{
			BaseRepository: graphqlbackend.MarshalRepositoryID(opts.repo),

			ExternalID: opts.externalID,
			HeadRef:    opts.headRef,
			Published:  published,

			Title: opts.title,
			Body:  opts.body,

			Commits: []campaigns.GitCommitDescription{
				{
					Message:     opts.commitMessage,
					Diff:        opts.commitDiff,
					AuthorEmail: opts.commitAuthorEmail,
					AuthorName:  opts.commitAuthorName,
				},
			},
		},
		DiffStatAdded:   testChangsetSpecDiffStat.Added,
		DiffStatChanged: testChangsetSpecDiffStat.Changed,
		DiffStatDeleted: testChangsetSpecDiffStat.Deleted,
	}

	return spec
}

func createChangesetSpec(
	t *testing.T,
	ctx context.Context,
	store *Store,
	opts testSpecOpts,
) *campaigns.ChangesetSpec {
	t.Helper()

	spec := buildChangesetSpec(t, opts)

	if err := store.CreateChangesetSpec(ctx, spec); err != nil {
		t.Fatal(err)
	}

	return spec
}

func createCampaignSpec(t *testing.T, ctx context.Context, store *Store, name string, userID int32) *campaigns.CampaignSpec {
	t.Helper()

	s := &campaigns.CampaignSpec{
		UserID:          userID,
		NamespaceUserID: userID,
		Spec: campaigns.CampaignSpecFields{
			Name:        name,
			Description: "the description",
			ChangesetTemplate: campaigns.ChangesetTemplate{
				Branch: "branch-name",
			},
		},
	}

	if err := store.CreateCampaignSpec(ctx, s); err != nil {
		t.Fatal(err)
	}

	return s
}

func createCampaign(t *testing.T, ctx context.Context, store *Store, name string, userID int32, spec int64) *campaigns.Campaign {
	t.Helper()

	c := &campaigns.Campaign{
		InitialApplierID: userID,
		LastApplierID:    userID,
		LastAppliedAt:    store.Clock()(),
		NamespaceUserID:  userID,
		CampaignSpecID:   spec,
		Name:             name,
		Description:      "campaign description",
	}

	if err := store.CreateCampaign(ctx, c); err != nil {
		t.Fatal(err)
	}

	return c
}
