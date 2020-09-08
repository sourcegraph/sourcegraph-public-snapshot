package campaigns

import (
	"context"
	"strings"
	"time"

	"testing"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/testing"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func TestReconcilerProcess(t *testing.T) {
	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)

	now := time.Now().UTC().Truncate(time.Microsecond)
	clock := func() time.Time {
		return now.UTC().Truncate(time.Microsecond)
	}
	store := NewStoreWithClock(dbconn.Global, clock)

	admin := createTestUser(ctx, t)
	if !admin.SiteAdmin {
		t.Fatalf("admin is not site admin")
	}

	rs, extSvc := createTestRepos(t, ctx, dbconn.Global, 1)

	state := ct.MockChangesetSyncState(&protocol.RepoInfo{
		Name: api.RepoName(rs[0].Name),
		VCS:  protocol.VCSInfo{URL: rs[0].URI},
	})
	defer state.Unmock()

	githubPR := buildGithubPR(clock(), "OPEN")
	closedGitHubPR := buildGithubPR(clock(), "CLOSED")

	type testCase struct {
		changeset    testChangesetOpts
		currentSpec  *testSpecOpts
		previousSpec *testSpecOpts

		sourcerMetadata interface{}
		// Whether or not the source responds to CreateChangeset with "already exists"
		alreadyExists bool

		wantCreateOnHostCode bool
		wantUpdateOnCodeHost bool
		wantCloseOnCodeHost  bool
		wantLoadFromCodeHost bool
		wantGitserverCommit  bool

		wantChangeset changesetAssertions
	}

	tests := map[string]testCase{
		"published unsynced changeset without changesetSpec": {
			changeset: testChangesetOpts{
				publicationState: campaigns.ChangesetPublicationStatePublished,
				externalID:       githubPR.ID,
				unsynced:         true,
			},
			sourcerMetadata: githubPR,

			wantLoadFromCodeHost: true,

			wantChangeset: changesetAssertions{
				publicationState: campaigns.ChangesetPublicationStatePublished,
				externalID:       githubPR.ID,
				externalBranch:   githubPR.HeadRefName,
				externalState:    campaigns.ChangesetExternalStateOpen,
				unsynced:         false,
				title:            githubPR.Title,
				body:             githubPR.Body,
				diffStat:         state.DiffStat,
			},
		},
		"unpublished changeset stay unpublished": {
			currentSpec: &testSpecOpts{
				headRef:   "refs/heads/repo-1-branch-1",
				published: false,
			},
			changeset: testChangesetOpts{
				publicationState: campaigns.ChangesetPublicationStateUnpublished,
			},
			sourcerMetadata: githubPR,

			wantChangeset: changesetAssertions{
				publicationState: campaigns.ChangesetPublicationStateUnpublished,
				externalState:    "",
				externalID:       "",
				externalBranch:   "",
			},
		},
		"publish changeset": {
			currentSpec: &testSpecOpts{
				headRef:   "refs/heads/head-ref-on-github",
				published: true,
			},
			changeset: testChangesetOpts{
				publicationState: campaigns.ChangesetPublicationStateUnpublished,
			},
			sourcerMetadata: githubPR,

			wantCreateOnHostCode: true,
			wantGitserverCommit:  true,

			wantChangeset: changesetAssertions{
				publicationState: campaigns.ChangesetPublicationStatePublished,
				externalID:       githubPR.ID,
				externalBranch:   githubPR.HeadRefName,
				externalState:    campaigns.ChangesetExternalStateOpen,
				title:            githubPR.Title,
				body:             githubPR.Body,
				diffStat:         state.DiffStat,
			},
		},
		"retry publish changeset": {
			// This test case makes sure that everything works when the code host says
			// that the changeset already exists.
			alreadyExists: true,
			currentSpec: &testSpecOpts{
				headRef:   "refs/heads/head-ref-on-github",
				published: true,
			},
			changeset: testChangesetOpts{
				failureMessage:   "publication failed",
				publicationState: campaigns.ChangesetPublicationStateUnpublished,
			},
			sourcerMetadata: githubPR,

			// We first do a create and since that fails with "already exists"
			// we update.
			wantCreateOnHostCode: true,
			wantUpdateOnCodeHost: true,
			wantGitserverCommit:  true,

			wantChangeset: changesetAssertions{
				publicationState: campaigns.ChangesetPublicationStatePublished,
				externalID:       githubPR.ID,
				externalBranch:   githubPR.HeadRefName,
				externalState:    campaigns.ChangesetExternalStateOpen,
				title:            githubPR.Title,
				body:             githubPR.Body,
				diffStat:         state.DiffStat,
			},
		},
		"update published changeset metadata": {
			currentSpec: &testSpecOpts{
				headRef:   "refs/heads/head-ref-on-github",
				published: true,

				title: "new title",
				body:  "new body",
			},
			previousSpec: &testSpecOpts{
				headRef:   "refs/heads/head-ref-on-github",
				published: true,

				title: "old title",
				body:  "old body",
			},
			changeset: testChangesetOpts{
				publicationState:  campaigns.ChangesetPublicationStatePublished,
				externalID:        "12345",
				externalBranch:    "head-ref-on-github",
				createdByCampaign: true,
			},
			sourcerMetadata: githubPR,

			// We don't want a new commit, only an update on the code host.
			wantUpdateOnCodeHost: true,

			wantChangeset: changesetAssertions{
				publicationState: campaigns.ChangesetPublicationStatePublished,
				externalID:       githubPR.ID,
				externalBranch:   githubPR.HeadRefName,
				externalState:    campaigns.ChangesetExternalStateOpen,
				diffStat:         state.DiffStat,
				// We update the title/body but want the title/body returned by the code host.
				title: githubPR.Title,
				body:  githubPR.Body,
			},
		},
		"retry update published changeset metadata": {
			currentSpec: &testSpecOpts{
				headRef:   "refs/heads/head-ref-on-github",
				published: true,

				title: "new title",
				body:  "new body",
			},
			previousSpec: &testSpecOpts{
				headRef:   "refs/heads/head-ref-on-github",
				published: true,

				title: "old title",
				body:  "old body",
			},
			changeset: testChangesetOpts{
				publicationState:  campaigns.ChangesetPublicationStatePublished,
				externalID:        githubPR.ID,
				externalBranch:    githubPR.HeadRefName,
				externalState:     campaigns.ChangesetExternalStateOpen,
				createdByCampaign: true,
				// Previous update failed:
				failureMessage: "failed to update changeset metadata",
			},
			sourcerMetadata: githubPR,

			wantUpdateOnCodeHost: true,

			wantChangeset: changesetAssertions{
				publicationState: campaigns.ChangesetPublicationStatePublished,
				externalID:       githubPR.ID,
				externalBranch:   githubPR.HeadRefName,
				externalState:    campaigns.ChangesetExternalStateOpen,
				title:            githubPR.Title,
				body:             githubPR.Body,
				diffStat:         state.DiffStat,
				// failureMessage should be nil
			},
		},
		"update published changeset commit": {
			currentSpec: &testSpecOpts{
				headRef:   "refs/heads/head-ref-on-github",
				published: true,

				// Title and body the same, but commit changed
				commitDiff:    "new diff",
				commitMessage: "new message",
			},
			previousSpec: &testSpecOpts{
				headRef:   "refs/heads/head-ref-on-github",
				published: true,

				commitDiff:    "old diff",
				commitMessage: "old message",
			},
			changeset: testChangesetOpts{
				publicationState:  campaigns.ChangesetPublicationStatePublished,
				externalID:        "12345",
				externalBranch:    "head-ref-on-github",
				externalState:     campaigns.ChangesetExternalStateOpen,
				createdByCampaign: true,
			},
			sourcerMetadata: githubPR,

			// We don't want an update on the code host, only a new commit pushed.
			wantGitserverCommit: true,
			// And we want the changeset to be synced after pushing the commit.
			wantLoadFromCodeHost: true,

			wantChangeset: changesetAssertions{
				publicationState: campaigns.ChangesetPublicationStatePublished,
				externalState:    campaigns.ChangesetExternalStateOpen,
				externalID:       githubPR.ID,
				externalBranch:   githubPR.HeadRefName,
				diffStat:         state.DiffStat,
			},
		},
		"retry update published changeset commit": {
			currentSpec: &testSpecOpts{
				headRef:       "refs/heads/head-ref-on-github",
				published:     true,
				commitDiff:    "new diff",
				commitMessage: "new message",
			},
			previousSpec: &testSpecOpts{
				headRef:   "refs/heads/head-ref-on-github",
				published: true,

				commitDiff:    "old diff",
				commitMessage: "old message",
			},
			changeset: testChangesetOpts{
				publicationState:  campaigns.ChangesetPublicationStatePublished,
				externalID:        "12345",
				externalBranch:    "head-ref-on-github",
				externalState:     campaigns.ChangesetExternalStateOpen,
				createdByCampaign: true,

				// Previous update failed:
				failureMessage: "failed to update changeset commit",
			},
			sourcerMetadata: githubPR,

			wantGitserverCommit:  true,
			wantLoadFromCodeHost: true,

			wantChangeset: changesetAssertions{
				publicationState: campaigns.ChangesetPublicationStatePublished,
				externalState:    campaigns.ChangesetExternalStateOpen,
				externalID:       githubPR.ID,
				externalBranch:   githubPR.HeadRefName,
				diffStat:         state.DiffStat,
				// failureMessage should be nil
			},
		},
		"reprocess published changeset without changes": {
			// ChangesetSpec is already published and has no previous spec.
			// Simply a reprocessing of the same changeset.
			currentSpec: &testSpecOpts{
				headRef:   "refs/heads/head-ref-on-github",
				published: true,

				title: "title",
				body:  "body",
			},
			changeset: testChangesetOpts{
				publicationState:  campaigns.ChangesetPublicationStatePublished,
				externalID:        githubPR.ID,
				externalBranch:    githubPR.HeadRefName,
				externalState:     campaigns.ChangesetExternalStateOpen,
				createdByCampaign: true,
			},
			sourcerMetadata: githubPR,

			wantChangeset: changesetAssertions{
				publicationState: campaigns.ChangesetPublicationStatePublished,
				externalID:       githubPR.ID,
				externalBranch:   githubPR.HeadRefName,
				externalState:    campaigns.ChangesetExternalStateOpen,
			},
		},
		"closing published open changeset": {
			currentSpec: &testSpecOpts{
				headRef:   "refs/heads/head-ref-on-github",
				published: true,

				title: "title",
				body:  "body",
			},
			changeset: testChangesetOpts{
				publicationState: campaigns.ChangesetPublicationStatePublished,
				externalID:       githubPR.ID,
				externalBranch:   githubPR.HeadRefName,
				externalState:    campaigns.ChangesetExternalStateOpen,
				closing:          true,
			},
			// We return a closed GitHub PR here
			sourcerMetadata: closedGitHubPR,

			wantCloseOnCodeHost: true,
			// We want to also sync the changeset after closing it
			wantLoadFromCodeHost: true,

			wantChangeset: changesetAssertions{
				publicationState: campaigns.ChangesetPublicationStatePublished,
				closing:          false,

				externalID:     closedGitHubPR.ID,
				externalBranch: closedGitHubPR.HeadRefName,
				externalState:  campaigns.ChangesetExternalStateClosed,

				title:    closedGitHubPR.Title,
				body:     closedGitHubPR.Body,
				diffStat: state.DiffStat,
			},
		},
		"closing non-open changeset": {
			currentSpec: &testSpecOpts{
				headRef:   "refs/heads/head-ref-on-github",
				published: true,

				title: "title",
				body:  "body",
			},
			changeset: testChangesetOpts{
				publicationState: campaigns.ChangesetPublicationStatePublished,
				externalID:       githubPR.ID,
				externalBranch:   githubPR.HeadRefName,
				externalState:    campaigns.ChangesetExternalStateClosed,
				closing:          true,
			},
			// We return a closed GitHub PR here, but since it's a noop, we
			// don't sync and thus don't set its attributes on the changeset.
			sourcerMetadata: closedGitHubPR,

			// Should be a noop
			wantCloseOnCodeHost: false,

			wantChangeset: changesetAssertions{
				publicationState: campaigns.ChangesetPublicationStatePublished,
				closing:          false,

				externalID:     closedGitHubPR.ID,
				externalBranch: closedGitHubPR.HeadRefName,
				externalState:  campaigns.ChangesetExternalStateClosed,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Clean up database.
			truncateTables(t, dbconn.Global, "changeset_events", "changesets", "campaigns", "campaign_specs", "changeset_specs")

			// Create necessary associations.
			campaignSpec := createCampaignSpec(t, ctx, store, "reconciler-test-campaign", admin.ID)
			campaign := createCampaign(t, ctx, store, "reconciler-test-campaign", admin.ID, campaignSpec.ID)

			// Create the changesetSpec with associations wired up correctly.
			var changesetSpec *campaigns.ChangesetSpec
			if tc.currentSpec != nil {
				specOpts := *tc.currentSpec
				specOpts.user = admin.ID
				specOpts.repo = rs[0].ID
				specOpts.campaignSpec = campaignSpec.ID
				changesetSpec = createChangesetSpec(t, ctx, store, specOpts)
			}

			// If we need a previous spec, we need to set that up too.
			var previousSpec *campaigns.ChangesetSpec
			if tc.previousSpec != nil {
				previousCampaignSpec := createCampaignSpec(t, ctx, store, "previous-campaign-spec", admin.ID)
				specOpts := *tc.previousSpec
				specOpts.user = admin.ID
				specOpts.repo = rs[0].ID
				specOpts.campaignSpec = previousCampaignSpec.ID
				previousSpec = createChangesetSpec(t, ctx, store, specOpts)
			}

			// Create the changeset with correct associations.
			changesetOpts := tc.changeset
			changesetOpts.repo = rs[0].ID
			changesetOpts.campaign = campaign.ID
			if changesetSpec != nil {
				changesetOpts.currentSpec = changesetSpec.ID
			}
			if previousSpec != nil {
				changesetOpts.previousSpec = previousSpec.ID
			}
			changeset := createChangeset(t, ctx, store, changesetOpts)

			// Setup gitserver dependency.
			gitClient := &ct.FakeGitserverClient{ResponseErr: nil}
			if changesetSpec != nil {
				gitClient.Response = changesetSpec.Spec.HeadRef
			}

			// Setup the sourcer that's used to create a Source with which
			// to create/update a changeset.
			fakeSource := &ct.FakeChangesetSource{
				Svc:             extSvc,
				Err:             nil,
				ChangesetExists: tc.alreadyExists,
				FakeMetadata:    tc.sourcerMetadata,
			}
			if changesetSpec != nil {
				fakeSource.WantHeadRef = changesetSpec.Spec.HeadRef
				fakeSource.WantBaseRef = changesetSpec.Spec.BaseRef
			}

			sourcer := repos.NewFakeSourcer(nil, fakeSource)

			// Run the reconciler
			rec := reconciler{
				noSleepBeforeSync: true,
				gitserverClient:   gitClient,
				sourcer:           sourcer,
				store:             store,
			}
			if err := rec.process(ctx, store, changeset); err != nil {
				t.Fatalf("reconciler process failed: %s", err)
			}

			// Assert that all the calls happened
			if have, want := gitClient.CreateCommitFromPatchCalled, tc.wantGitserverCommit; have != want {
				t.Fatalf("wrong CreateCommitFromPatch call. wantCalled=%t, wasCalled=%t", want, have)
			}

			if have, want := fakeSource.CreateChangesetCalled, tc.wantCreateOnHostCode; have != want {
				t.Fatalf("wrong CreateChangeset call. wantCalled=%t, wasCalled=%t", want, have)
			}

			if have, want := fakeSource.UpdateChangesetCalled, tc.wantUpdateOnCodeHost; have != want {
				t.Fatalf("wrong UpdateChangeset call. wantCalled=%t, wasCalled=%t", want, have)
			}

			if have, want := fakeSource.LoadChangesetsCalled, tc.wantLoadFromCodeHost; have != want {
				t.Fatalf("wrong LoadChangesets call. wantCalled=%t, wasCalled=%t", want, have)
			}

			if have, want := fakeSource.CloseChangesetCalled, tc.wantCloseOnCodeHost; have != want {
				t.Fatalf("wrong CloseChangeset call. wantCalled=%t, wasCalled=%t", want, have)
			}

			// Assert that the changeset in the database looks like we want
			assertions := tc.wantChangeset
			assertions.repo = rs[0].ID
			if changesetSpec != nil {
				assertions.currentSpec = changesetSpec.ID
			}
			if previousSpec != nil {
				assertions.previousSpec = previousSpec.ID
			}
			reloadAndAssertChangeset(t, ctx, store, changeset, assertions)
		})
	}
}

func TestReconcilerProcess_PublishedChangesetDuplicateBranch(t *testing.T) {
	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)

	store := NewStore(dbconn.Global)

	admin := createTestUser(ctx, t)
	if !admin.SiteAdmin {
		t.Fatalf("admin is not site admin")
	}

	rs, _ := createTestRepos(t, ctx, dbconn.Global, 1)

	state := ct.MockChangesetSyncState(&protocol.RepoInfo{
		Name: api.RepoName(rs[0].Name),
		VCS:  protocol.VCSInfo{URL: rs[0].URI},
	})
	defer state.Unmock()

	commonHeadRef := "refs/heads/collision"

	// Create a published changeset.
	campaignSpec := createCampaignSpec(t, ctx, store, "reconciler-test-campaign", admin.ID)
	campaign := createCampaign(t, ctx, store, "reconciler-test-campaign", admin.ID, campaignSpec.ID)
	changesetSpec := createChangesetSpec(t, ctx, store, testSpecOpts{
		user:         admin.ID,
		repo:         rs[0].ID,
		campaignSpec: campaignSpec.ID,
		headRef:      commonHeadRef,
	})
	createChangeset(t, ctx, store, testChangesetOpts{
		repo:             rs[0].ID,
		publicationState: campaigns.ChangesetPublicationStatePublished,
		campaign:         campaign.ID,
		ownedByCampaign:  campaign.ID,
		currentSpec:      changesetSpec.ID,
		externalBranch:   git.AbbreviateRef(commonHeadRef),
		externalID:       "123",
	})

	// Try to publish a changeset on the same HeadRef/ExternalBranch.
	otherCampaignSpec := createCampaignSpec(t, ctx, store, "other-test-campaign", admin.ID)
	otherCampaign := createCampaign(t, ctx, store, "other-test-campaign", admin.ID, otherCampaignSpec.ID)
	otherChangesetSpec := createChangesetSpec(t, ctx, store, testSpecOpts{
		user:         admin.ID,
		repo:         rs[0].ID,
		campaignSpec: otherCampaignSpec.ID,
		headRef:      commonHeadRef,
		published:    true,
	})
	otherChangeset := createChangeset(t, ctx, store, testChangesetOpts{
		repo:             rs[0].ID,
		publicationState: campaigns.ChangesetPublicationStateUnpublished,
		campaign:         otherCampaign.ID,
		ownedByCampaign:  otherCampaign.ID,
		currentSpec:      otherChangesetSpec.ID,
	})

	// Run the reconciler
	rec := reconciler{store: store}
	haveErr := rec.process(ctx, store, otherChangeset)
	if !errors.Is(haveErr, ErrPublishSameBranch) {
		t.Fatalf("reconciler process failed with wrong error: %s", haveErr)
	}
}

func buildGithubPR(now time.Time, state string) *github.PullRequest {
	pr := &github.PullRequest{
		ID:          "12345",
		Number:      12345,
		Title:       state + " GitHub PR",
		Body:        state + " GitHub PR",
		HeadRefName: git.AbbreviateRef("head-ref-on-github"),
		State:       state,
		TimelineItems: []github.TimelineItem{
			{Type: "PullRequestCommit", Item: &github.PullRequestCommit{
				Commit: github.Commit{
					OID:           "new-f00bar",
					PushedDate:    now,
					CommittedDate: now,
				},
			}},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	if strings.ToLower(state) == "closed" {
		pr.State = "CLOSED"
		// We add a "ClosedEvent" so that the SyncChangesets call that happens after closing
		// the PR has the "correct" state to set the ExternalState
		pr.TimelineItems = append(pr.TimelineItems, github.TimelineItem{
			Type: "ClosedEvent",
			Item: &github.ClosedEvent{CreatedAt: now.Add(1 * time.Hour)},
		})
		pr.UpdatedAt = now.Add(1 * time.Hour)
	}

	return pr
}

type testChangesetOpts struct {
	repo         api.RepoID
	campaign     int64
	currentSpec  int64
	previousSpec int64

	externalServiceType string
	externalID          string
	externalBranch      string
	externalState       campaigns.ChangesetExternalState

	publicationState campaigns.ChangesetPublicationState

	reconcilerState campaigns.ReconcilerState
	failureMessage  string
	numFailures     int64

	createdByCampaign bool
	ownedByCampaign   int64

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

	if opts.externalServiceType == "" {
		opts.externalServiceType = extsvc.TypeGitHub
	}

	changeset := &campaigns.Changeset{
		RepoID:         opts.repo,
		CurrentSpecID:  opts.currentSpec,
		PreviousSpecID: opts.previousSpec,

		ExternalServiceType: opts.externalServiceType,
		ExternalID:          opts.externalID,
		ExternalBranch:      opts.externalBranch,
		ExternalState:       opts.externalState,

		PublicationState: opts.publicationState,

		CreatedByCampaign: opts.createdByCampaign,
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

	if err := store.CreateChangeset(ctx, changeset); err != nil {
		t.Fatalf("creating changeset failed: %s", err)
	}

	return changeset
}
