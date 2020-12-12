package campaigns

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/store"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/testing"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	gitprotocol "github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func TestExecutor_ExecutePlan(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	dbtesting.SetupGlobalTestDB(t)

	now := timeutil.Now()
	clock := func() time.Time { return now }
	store := store.NewWithClock(dbconn.Global, clock)

	admin := ct.CreateTestUser(t, true)

	rs, extSvc := ct.CreateTestRepos(t, ctx, dbconn.Global, 1)

	state := ct.MockChangesetSyncState(&protocol.RepoInfo{
		Name: api.RepoName(rs[0].Name),
		VCS:  protocol.VCSInfo{URL: rs[0].URI},
	})
	defer state.Unmock()

	internalClient = &mockInternalClient{externalURL: "https://sourcegraph.test"}
	defer func() { internalClient = api.InternalClient }()

	githubPR := buildGithubPR(clock(), campaigns.ChangesetExternalStateOpen)
	githubHeadRef := git.EnsureRefPrefix(githubPR.HeadRefName)
	draftGithubPR := buildGithubPR(clock(), campaigns.ChangesetExternalStateDraft)
	closedGitHubPR := buildGithubPR(clock(), campaigns.ChangesetExternalStateClosed)

	notFoundErr := repos.ChangesetNotFoundError{
		Changeset: &repos.Changeset{
			Changeset: &campaigns.Changeset{ExternalID: "100000"},
		},
	}

	type testCase struct {
		changeset      ct.TestChangesetOpts
		hasCurrentSpec bool
		plan           *ReconcilerPlan

		sourcerMetadata interface{}
		sourcerErr      error
		// Whether or not the source responds to CreateChangeset with "already exists"
		alreadyExists bool

		wantCreateOnCodeHost      bool
		wantCreateDraftOnCodeHost bool
		wantUndraftOnCodeHost     bool
		wantUpdateOnCodeHost      bool
		wantCloseOnCodeHost       bool
		wantLoadFromCodeHost      bool
		wantReopenOnCodeHost      bool

		wantGitserverCommit bool

		wantChangeset       ct.ChangesetAssertions
		wantNonRetryableErr bool
	}

	tests := map[string]testCase{
		"noop": {
			hasCurrentSpec: true,
			changeset:      ct.TestChangesetOpts{},
			plan:           &ReconcilerPlan{Ops: ReconcilerOperations{}},

			wantChangeset: ct.ChangesetAssertions{},
		},
		"import": {
			changeset: ct.TestChangesetOpts{
				PublicationState: campaigns.ChangesetPublicationStatePublished,
				ExternalID:       githubPR.ID,
				Unsynced:         true,
			},
			plan: &ReconcilerPlan{
				Ops: ReconcilerOperations{campaigns.ReconcilerOperationImport},
			},

			wantLoadFromCodeHost: true,

			wantChangeset: ct.ChangesetAssertions{
				PublicationState: campaigns.ChangesetPublicationStatePublished,
				ExternalID:       githubPR.ID,
				ExternalBranch:   githubHeadRef,
				ExternalState:    campaigns.ChangesetExternalStateOpen,
				Unsynced:         false,
				Title:            githubPR.Title,
				Body:             githubPR.Body,
				DiffStat:         state.DiffStat,
			},
		},
		"import and not-found error": {
			changeset: ct.TestChangesetOpts{
				PublicationState: campaigns.ChangesetPublicationStatePublished,
				ExternalID:       githubPR.ID,
				Unsynced:         true,
			},
			plan: &ReconcilerPlan{
				Ops: ReconcilerOperations{campaigns.ReconcilerOperationImport},
			},
			sourcerErr: notFoundErr,

			wantLoadFromCodeHost: true,

			wantNonRetryableErr: true,
		},
		"push and publish": {
			hasCurrentSpec: true,
			changeset: ct.TestChangesetOpts{
				PublicationState: campaigns.ChangesetPublicationStateUnpublished,
			},
			plan: &ReconcilerPlan{
				Ops: ReconcilerOperations{
					campaigns.ReconcilerOperationPush,
					campaigns.ReconcilerOperationPublish,
				},
			},

			wantCreateOnCodeHost: true,
			wantGitserverCommit:  true,

			wantChangeset: ct.ChangesetAssertions{
				PublicationState: campaigns.ChangesetPublicationStatePublished,
				ExternalID:       githubPR.ID,
				ExternalBranch:   githubHeadRef,
				ExternalState:    campaigns.ChangesetExternalStateOpen,
				Title:            githubPR.Title,
				Body:             githubPR.Body,
				DiffStat:         state.DiffStat,
			},
		},
		"retry push and publish": {
			// This test case makes sure that everything works when the code host says
			// that the changeset already exists.
			alreadyExists:  true,
			hasCurrentSpec: true,
			changeset: ct.TestChangesetOpts{
				// The reconciler resets the failure message before passing the
				// changeset to the executor.
				// We simulate that here by not setting FailureMessage.
				PublicationState: campaigns.ChangesetPublicationStateUnpublished,
			},
			plan: &ReconcilerPlan{
				Ops: ReconcilerOperations{
					campaigns.ReconcilerOperationPush,
					campaigns.ReconcilerOperationPublish,
				},
			},

			// We first do a create and since that fails with "already exists"
			// we update.
			wantCreateOnCodeHost: true,
			wantUpdateOnCodeHost: true,
			wantGitserverCommit:  true,

			wantChangeset: ct.ChangesetAssertions{
				PublicationState: campaigns.ChangesetPublicationStatePublished,
				ExternalID:       githubPR.ID,
				ExternalBranch:   githubHeadRef,
				ExternalState:    campaigns.ChangesetExternalStateOpen,
				Title:            githubPR.Title,
				Body:             githubPR.Body,
				DiffStat:         state.DiffStat,
			},
		},
		"update": {
			hasCurrentSpec: true,
			changeset: ct.TestChangesetOpts{
				PublicationState: campaigns.ChangesetPublicationStatePublished,
				ExternalID:       "12345",
				ExternalBranch:   "head-ref-on-github",
			},

			plan: &ReconcilerPlan{
				Ops: ReconcilerOperations{
					campaigns.ReconcilerOperationUpdate,
				},
			},

			// We don't want a new commit, only an update on the code host.
			wantUpdateOnCodeHost: true,

			wantChangeset: ct.ChangesetAssertions{
				PublicationState: campaigns.ChangesetPublicationStatePublished,
				ExternalID:       githubPR.ID,
				ExternalBranch:   githubHeadRef,
				ExternalState:    campaigns.ChangesetExternalStateOpen,
				DiffStat:         state.DiffStat,
				// We update the title/body but want the title/body returned by the code host.
				Title: githubPR.Title,
				Body:  githubPR.Body,
			},
		},
		"push sleep sync": {
			hasCurrentSpec: true,
			changeset: ct.TestChangesetOpts{
				PublicationState: campaigns.ChangesetPublicationStatePublished,
				ExternalID:       "12345",
				ExternalBranch:   git.EnsureRefPrefix("head-ref-on-github"),
				ExternalState:    campaigns.ChangesetExternalStateOpen,
			},

			plan: &ReconcilerPlan{
				Ops: ReconcilerOperations{
					campaigns.ReconcilerOperationPush,
					campaigns.ReconcilerOperationSleep,
					campaigns.ReconcilerOperationSync,
				},
			},

			wantGitserverCommit:  true,
			wantLoadFromCodeHost: true,

			wantChangeset: ct.ChangesetAssertions{
				PublicationState: campaigns.ChangesetPublicationStatePublished,
				ExternalState:    campaigns.ChangesetExternalStateOpen,
				ExternalID:       githubPR.ID,
				ExternalBranch:   githubHeadRef,
				DiffStat:         state.DiffStat,
			},
		},
		"close open changeset": {
			hasCurrentSpec: true,
			changeset: ct.TestChangesetOpts{
				PublicationState: campaigns.ChangesetPublicationStatePublished,
				ExternalID:       githubPR.ID,
				ExternalBranch:   githubHeadRef,
				ExternalState:    campaigns.ChangesetExternalStateOpen,
				Closing:          true,
			},
			plan: &ReconcilerPlan{
				Ops: ReconcilerOperations{
					campaigns.ReconcilerOperationClose,
				},
			},
			// We return a closed GitHub PR here
			sourcerMetadata: closedGitHubPR,

			wantCloseOnCodeHost: true,

			wantChangeset: ct.ChangesetAssertions{
				PublicationState: campaigns.ChangesetPublicationStatePublished,
				Closing:          false,

				ExternalID:     closedGitHubPR.ID,
				ExternalBranch: git.EnsureRefPrefix(closedGitHubPR.HeadRefName),
				ExternalState:  campaigns.ChangesetExternalStateClosed,

				Title:    closedGitHubPR.Title,
				Body:     closedGitHubPR.Body,
				DiffStat: state.DiffStat,
			},
		},
		"close closed changeset": {
			hasCurrentSpec: true,
			changeset: ct.TestChangesetOpts{
				PublicationState: campaigns.ChangesetPublicationStatePublished,
				ExternalID:       githubPR.ID,
				ExternalBranch:   githubHeadRef,
				ExternalState:    campaigns.ChangesetExternalStateClosed,
				Closing:          true,
			},
			plan: &ReconcilerPlan{
				Ops: ReconcilerOperations{
					campaigns.ReconcilerOperationClose,
				},
			},

			// We return a closed GitHub PR here, but since it's a noop, we
			// don't sync and thus don't set its attributes on the changeset.
			sourcerMetadata: closedGitHubPR,

			// Should be a noop
			wantCloseOnCodeHost: false,

			wantChangeset: ct.ChangesetAssertions{
				PublicationState: campaigns.ChangesetPublicationStatePublished,
				Closing:          false,

				ExternalID:     closedGitHubPR.ID,
				ExternalBranch: git.EnsureRefPrefix(closedGitHubPR.HeadRefName),
				ExternalState:  campaigns.ChangesetExternalStateClosed,
			},
		},
		"reopening closed changeset without updates": {
			hasCurrentSpec: true,
			changeset: ct.TestChangesetOpts{
				PublicationState: campaigns.ChangesetPublicationStatePublished,
				ExternalID:       githubPR.ID,
				ExternalBranch:   githubHeadRef,
				ExternalState:    campaigns.ChangesetExternalStateClosed,
			},
			plan: &ReconcilerPlan{
				Ops: ReconcilerOperations{
					campaigns.ReconcilerOperationReopen,
				},
			},

			wantReopenOnCodeHost: true,

			wantChangeset: ct.ChangesetAssertions{
				PublicationState: campaigns.ChangesetPublicationStatePublished,

				ExternalID:     githubPR.ID,
				ExternalBranch: githubHeadRef,
				ExternalState:  campaigns.ChangesetExternalStateOpen,

				Title:    githubPR.Title,
				Body:     githubPR.Body,
				DiffStat: state.DiffStat,
			},
		},
		"push and publishdraft": {
			hasCurrentSpec: true,
			changeset: ct.TestChangesetOpts{
				PublicationState: campaigns.ChangesetPublicationStateUnpublished,
			},

			plan: &ReconcilerPlan{
				Ops: ReconcilerOperations{
					campaigns.ReconcilerOperationPush,
					campaigns.ReconcilerOperationPublishDraft,
				},
			},

			sourcerMetadata: draftGithubPR,

			wantCreateDraftOnCodeHost: true,
			wantGitserverCommit:       true,

			wantChangeset: ct.ChangesetAssertions{
				PublicationState: campaigns.ChangesetPublicationStatePublished,

				ExternalID:     draftGithubPR.ID,
				ExternalBranch: git.EnsureRefPrefix(draftGithubPR.HeadRefName),
				ExternalState:  campaigns.ChangesetExternalStateDraft,

				Title:    draftGithubPR.Title,
				Body:     draftGithubPR.Body,
				DiffStat: state.DiffStat,
			},
		},
		"undraft": {
			hasCurrentSpec: true,
			changeset: ct.TestChangesetOpts{
				PublicationState: campaigns.ChangesetPublicationStatePublished,
				ExternalState:    campaigns.ChangesetExternalStateDraft,
			},

			plan: &ReconcilerPlan{
				Ops: ReconcilerOperations{
					campaigns.ReconcilerOperationUndraft,
				},
			},

			wantUndraftOnCodeHost: true,

			wantChangeset: ct.ChangesetAssertions{
				PublicationState: campaigns.ChangesetPublicationStatePublished,

				ExternalID:     githubPR.ID,
				ExternalBranch: githubHeadRef,
				ExternalState:  campaigns.ChangesetExternalStateOpen,

				Title:    githubPR.Title,
				Body:     githubPR.Body,
				DiffStat: state.DiffStat,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Create necessary associations.
			campaignSpec := createCampaignSpec(t, ctx, store, "executor-test-campaign", admin.ID)
			campaign := createCampaign(t, ctx, store, "executor-test-campaign", admin.ID, campaignSpec.ID)

			// Create the changesetSpec with associations wired up correctly.
			var changesetSpec *campaigns.ChangesetSpec
			if tc.hasCurrentSpec {
				// The attributes of the spec don't really matter, but the
				// associations do.
				specOpts := ct.TestSpecOpts{}
				specOpts.User = admin.ID
				specOpts.Repo = rs[0].ID
				specOpts.CampaignSpec = campaignSpec.ID
				changesetSpec = ct.CreateChangesetSpec(t, ctx, store, specOpts)
			}

			// Create the changeset with correct associations.
			changesetOpts := tc.changeset
			changesetOpts.Repo = rs[0].ID
			changesetOpts.CampaignIDs = []int64{campaign.ID}
			changesetOpts.OwnedByCampaign = campaign.ID
			if changesetSpec != nil {
				changesetOpts.CurrentSpec = changesetSpec.ID
			}
			changeset := ct.CreateChangeset(t, ctx, store, changesetOpts)

			// Setup gitserver dependency.
			gitClient := &ct.FakeGitserverClient{ResponseErr: nil}
			if changesetSpec != nil {
				gitClient.Response = changesetSpec.Spec.HeadRef
			}

			// Setup the sourcer that's used to create a Source with which
			// to create/update a changeset.
			fakeSource := &ct.FakeChangesetSource{
				Svc:             extSvc,
				Err:             tc.sourcerErr,
				ChangesetExists: tc.alreadyExists,
			}

			if tc.sourcerMetadata != nil {
				fakeSource.FakeMetadata = tc.sourcerMetadata
			} else {
				fakeSource.FakeMetadata = githubPR
			}
			if changesetSpec != nil {
				fakeSource.WantHeadRef = changesetSpec.Spec.HeadRef
				fakeSource.WantBaseRef = changesetSpec.Spec.BaseRef
			}

			sourcer := repos.NewFakeSourcer(nil, fakeSource)

			// Run the reconciler
			executor := &executor{
				tx: store,

				// Don't actually sleep for the sake of testing.
				noSleepBeforeSync: true,

				gitserverClient: gitClient,
				sourcer:         sourcer,

				ch:   changeset,
				spec: changesetSpec,
			}

			// Execute the plan
			err := executor.ExecutePlan(ctx, tc.plan)
			if err != nil {
				if tc.wantNonRetryableErr && errcode.IsNonRetryable(err) {
					// all good
				} else {
					t.Fatalf("ExecutePlan failed: %s", err)
				}
			}

			// Assert that all the calls happened
			if have, want := gitClient.CreateCommitFromPatchCalled, tc.wantGitserverCommit; have != want {
				t.Fatalf("wrong CreateCommitFromPatch call. wantCalled=%t, wasCalled=%t", want, have)
			}

			if have, want := fakeSource.CreateDraftChangesetCalled, tc.wantCreateDraftOnCodeHost; have != want {
				t.Fatalf("wrong CreateDraftChangeset call. wantCalled=%t, wasCalled=%t", want, have)
			}

			if have, want := fakeSource.UndraftedChangesetsCalled, tc.wantUndraftOnCodeHost; have != want {
				t.Fatalf("wrong UndraftChangeset call. wantCalled=%t, wasCalled=%t", want, have)
			}

			if have, want := fakeSource.CreateChangesetCalled, tc.wantCreateOnCodeHost; have != want {
				t.Fatalf("wrong CreateChangeset call. wantCalled=%t, wasCalled=%t", want, have)
			}

			if have, want := fakeSource.UpdateChangesetCalled, tc.wantUpdateOnCodeHost; have != want {
				t.Fatalf("wrong UpdateChangeset call. wantCalled=%t, wasCalled=%t", want, have)
			}

			if have, want := fakeSource.ReopenChangesetCalled, tc.wantReopenOnCodeHost; have != want {
				t.Fatalf("wrong ReopenChangeset call. wantCalled=%t, wasCalled=%t", want, have)
			}

			if have, want := fakeSource.LoadChangesetCalled, tc.wantLoadFromCodeHost; have != want {
				t.Fatalf("wrong LoadChangeset call. wantCalled=%t, wasCalled=%t", want, have)
			}

			if have, want := fakeSource.CloseChangesetCalled, tc.wantCloseOnCodeHost; have != want {
				t.Fatalf("wrong CloseChangeset call. wantCalled=%t, wasCalled=%t", want, have)
			}

			if tc.wantNonRetryableErr {
				return
			}

			// Assert that the changeset in the database looks like we want
			assertions := tc.wantChangeset
			assertions.Repo = rs[0].ID
			assertions.OwnedByCampaign = changesetOpts.OwnedByCampaign
			if changesetSpec != nil {
				assertions.CurrentSpec = changesetSpec.ID
			}
			ct.ReloadAndAssertChangeset(t, ctx, store, changeset, assertions)

			// Assert that the body included a backlink if needed. We'll do
			// more detailed unit tests of decorateChangesetBody elsewhere;
			// we're just looking for a basic marker here that _something_
			// happened.
			var rcs *repos.Changeset
			if tc.wantCreateOnCodeHost && fakeSource.CreateChangesetCalled {
				rcs = fakeSource.CreatedChangesets[0]
			} else if tc.wantUpdateOnCodeHost && fakeSource.UpdateChangesetCalled {
				rcs = fakeSource.UpdatedChangesets[0]
			}

			if rcs != nil {
				if !strings.Contains(rcs.Body, "Created by Sourcegraph campaign") {
					t.Errorf("did not find backlink in body: %q", rcs.Body)
				}
			}
		})

		// After each test: clean up database.
		truncateTables(t, dbconn.Global, "changeset_events", "changesets", "campaigns", "campaign_specs", "changeset_specs")
	}
}

func TestExecutor_ExecutePlan_PublishedChangesetDuplicateBranch(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	dbtesting.SetupGlobalTestDB(t)

	store := store.New(dbconn.Global)

	rs, _ := ct.CreateTestRepos(t, ctx, dbconn.Global, 1)

	commonHeadRef := "refs/heads/collision"

	// Create a published changeset.
	ct.CreateChangeset(t, ctx, store, ct.TestChangesetOpts{
		Repo:             rs[0].ID,
		PublicationState: campaigns.ChangesetPublicationStatePublished,
		ExternalBranch:   commonHeadRef,
		ExternalID:       "123",
	})

	// Build a changeset that would be pushed on the same HeadRef/ExternalBranch.
	spec := ct.BuildChangesetSpec(t, ct.TestSpecOpts{
		Repo:      rs[0].ID,
		HeadRef:   commonHeadRef,
		Published: true,
	})
	changeset := ct.BuildChangeset(ct.TestChangesetOpts{Repo: rs[0].ID})

	executor := &executor{
		tx:      store,
		sourcer: repos.NewFakeSourcer(nil, &ct.FakeChangesetSource{}),

		ch:   changeset,
		spec: spec,
	}

	// Plan only needs a push operation, since that's where we check
	plan := &ReconcilerPlan{}
	plan.AddOp(campaigns.ReconcilerOperationPush)

	err := executor.ExecutePlan(ctx, plan)
	if err == nil {
		t.Fatal("reconciler did not return error")
	}

	// We expect a non-retryable error to be returned.
	if !errcode.IsNonRetryable(err) {
		t.Fatalf("error is not non-retryabe. have=%s", err)
	}
}

func TestExecutor_LoadAuthenticator(t *testing.T) {
	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)

	store := store.New(dbconn.Global)

	admin := ct.CreateTestUser(t, true)
	user := ct.CreateTestUser(t, false)

	rs, _ := ct.CreateTestRepos(t, ctx, dbconn.Global, 1)
	repo := rs[0]

	campaignSpec := createCampaignSpec(t, ctx, store, "reconciler-test-campaign", admin.ID)
	adminCampaign := createCampaign(t, ctx, store, "reconciler-test-campaign", admin.ID, campaignSpec.ID)
	userCampaign := createCampaign(t, ctx, store, "reconciler-test-campaign", user.ID, campaignSpec.ID)

	t.Run("imported changeset uses global token", func(t *testing.T) {
		a, err := (&executor{
			ch: &campaigns.Changeset{
				OwnedByCampaignID: 0,
			},
		}).loadAuthenticator(ctx)
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}
		if a != nil {
			t.Errorf("unexpected non-nil authenticator: %v", a)
		}
	})

	t.Run("owned by missing campaign", func(t *testing.T) {
		_, err := (&executor{
			ch: &campaigns.Changeset{
				OwnedByCampaignID: 1234,
			},
			tx: store,
		}).loadAuthenticator(ctx)
		if err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("owned by admin user without credential", func(t *testing.T) {
		a, err := (&executor{
			ch: &campaigns.Changeset{
				OwnedByCampaignID: adminCampaign.ID,
			},
			repo: repo,
			tx:   store,
		}).loadAuthenticator(ctx)
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}
		if a != nil {
			t.Errorf("unexpected non-nil authenticator: %v", a)
		}
	})

	t.Run("owned by normal user without credential", func(t *testing.T) {
		_, err := (&executor{
			ch: &campaigns.Changeset{
				OwnedByCampaignID: userCampaign.ID,
			},
			repo: repo,
			tx:   store,
		}).loadAuthenticator(ctx)
		if err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("owned by admin user with credential", func(t *testing.T) {
		token := &auth.OAuthBearerToken{Token: "abcdef"}
		if _, err := db.UserCredentials.Create(ctx, db.UserCredentialScope{
			Domain:              db.UserCredentialDomainCampaigns,
			UserID:              admin.ID,
			ExternalServiceType: repo.ExternalRepo.ServiceType,
			ExternalServiceID:   repo.ExternalRepo.ServiceID,
		}, token); err != nil {
			t.Fatal(err)
		}

		a, err := (&executor{
			ch: &campaigns.Changeset{
				OwnedByCampaignID: adminCampaign.ID,
			},
			repo: repo,
			tx:   store,
		}).loadAuthenticator(ctx)
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}
		if diff := cmp.Diff(token, a); diff != "" {
			t.Errorf("unexpected authenticator:\n%s", diff)
		}
	})

	t.Run("owned by normal user with credential", func(t *testing.T) {
		token := &auth.OAuthBearerToken{Token: "abcdef"}
		if _, err := db.UserCredentials.Create(ctx, db.UserCredentialScope{
			Domain:              db.UserCredentialDomainCampaigns,
			UserID:              user.ID,
			ExternalServiceType: repo.ExternalRepo.ServiceType,
			ExternalServiceID:   repo.ExternalRepo.ServiceID,
		}, token); err != nil {
			t.Fatal(err)
		}

		a, err := (&executor{
			ch: &campaigns.Changeset{
				OwnedByCampaignID: userCampaign.ID,
			},
			repo: repo,
			tx:   store,
		}).loadAuthenticator(ctx)
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}
		if diff := cmp.Diff(token, a); diff != "" {
			t.Errorf("unexpected authenticator:\n%s", diff)
		}
	})
}

func TestExecutor_UserCredentialsForGitserver(t *testing.T) {
	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)

	store := store.New(dbconn.Global)

	admin := ct.CreateTestUser(t, true)
	user := ct.CreateTestUser(t, false)

	rs, extSvc := ct.CreateTestRepos(t, ctx, dbconn.Global, 1)
	gitHubRepo := rs[0]
	gitHubRepoCloneURL := gitHubRepo.Sources[extSvc.URN()].CloneURL

	gitLabRepos, gitLabExtSvc := ct.CreateGitlabTestRepos(t, ctx, dbconn.Global, 1)
	gitLabRepo := gitLabRepos[0]
	gitLabRepoCloneURL := gitLabRepo.Sources[gitLabExtSvc.URN()].CloneURL

	bbsRepos, bbsExtSvc := ct.CreateBbsTestRepos(t, ctx, dbconn.Global, 1)
	bbsRepo := bbsRepos[0]
	bbsRepoCloneURL := bbsRepo.Sources[bbsExtSvc.URN()].CloneURL

	gitClient := &ct.FakeGitserverClient{ResponseErr: nil}
	fakeSource := &ct.FakeChangesetSource{Svc: extSvc}
	sourcer := repos.NewFakeSourcer(nil, fakeSource)

	plan := &ReconcilerPlan{}
	plan.AddOp(campaigns.ReconcilerOperationPush)

	tests := []struct {
		name           string
		user           *types.User
		repo           *types.Repo
		credentials    auth.Authenticator
		wantErr        bool
		wantPushConfig *gitprotocol.PushConfig
	}{
		{
			name:        "github OAuthBearerToken",
			user:        user,
			repo:        gitHubRepo,
			credentials: &auth.OAuthBearerToken{Token: "my-secret-github-token"},
			wantPushConfig: &gitprotocol.PushConfig{
				RemoteURL: "https://my-secret-github-token@github.com/" + string(gitHubRepo.Name),
			},
		},
		{
			name:    "github no credentials",
			user:    user,
			repo:    gitHubRepo,
			wantErr: true,
		},
		{
			name: "github site-admin and no credentials",
			repo: gitHubRepo,
			user: admin,
			wantPushConfig: &gitprotocol.PushConfig{
				RemoteURL: gitHubRepoCloneURL,
			},
		},
		{
			name:        "gitlab OAuthBearerToken",
			user:        user,
			repo:        gitLabRepo,
			credentials: &auth.OAuthBearerToken{Token: "my-secret-gitlab-token"},
			wantPushConfig: &gitprotocol.PushConfig{
				RemoteURL: "https://git:my-secret-gitlab-token@gitlab.com/" + string(gitLabRepo.Name),
			},
		},
		{
			name:    "gitlab no credentials",
			user:    user,
			repo:    gitLabRepo,
			wantErr: true,
		},
		{
			name: "gitlab site-admin and no credentials",
			user: admin,
			repo: gitLabRepo,
			wantPushConfig: &gitprotocol.PushConfig{
				RemoteURL: gitLabRepoCloneURL,
			},
		},
		{
			name:        "bitbucketServer BasicAuth",
			user:        user,
			repo:        bbsRepo,
			credentials: &auth.BasicAuth{Username: "fredwoard johnssen", Password: "my-secret-bbs-token"},
			wantPushConfig: &gitprotocol.PushConfig{
				RemoteURL: "https://fredwoard%20johnssen:my-secret-bbs-token@bitbucket.sourcegraph.com/scm/" + string(bbsRepo.Name),
			},
		},
		{
			name:    "bitbucketServer no credentials",
			user:    user,
			repo:    bbsRepo,
			wantErr: true,
		},
		{
			name: "bitbucketServer site-admin and no credentials",
			user: admin,
			repo: bbsRepo,
			wantPushConfig: &gitprotocol.PushConfig{
				RemoteURL: bbsRepoCloneURL,
			},
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.credentials != nil {
				cred, err := db.UserCredentials.Create(ctx, db.UserCredentialScope{
					Domain:              db.UserCredentialDomainCampaigns,
					UserID:              tt.user.ID,
					ExternalServiceType: tt.repo.ExternalRepo.ServiceType,
					ExternalServiceID:   tt.repo.ExternalRepo.ServiceID,
				}, tt.credentials)
				if err != nil {
					t.Fatal(err)
				}
				defer func() { db.UserCredentials.Delete(ctx, cred.ID) }()
			}

			campaignSpec := createCampaignSpec(t, ctx, store, fmt.Sprintf("reconciler-credentials-%d", i), tt.user.ID)
			campaign := createCampaign(t, ctx, store, fmt.Sprintf("reconciler-credentials-%d", i), tt.user.ID, campaignSpec.ID)

			ex := &executor{
				ch: &campaigns.Changeset{
					OwnedByCampaignID: campaign.ID,
					RepoID:            tt.repo.ID,
				},
				spec: ct.BuildChangesetSpec(t, ct.TestSpecOpts{
					HeadRef:    "refs/heads/my-branch",
					Published:  true,
					CommitDiff: "testdiff",
				}),
				sourcer:         sourcer,
				gitserverClient: gitClient,
				tx:              store,
			}

			err := ex.ExecutePlan(context.Background(), plan)
			if !tt.wantErr && err != nil {
				t.Fatalf("executing plan failed: %s", err)
			}
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error but got none")
				} else {
					return
				}
			}

			if diff := cmp.Diff(tt.wantPushConfig, gitClient.CreateCommitFromPatchReq.Push); diff != "" {
				t.Errorf("unexpected push options:\n%s", diff)
			}
		})
	}
}
