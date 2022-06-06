package reconciler

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	mockassert "github.com/derision-test/go-mockgen/testutil/assert"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/sources"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	et "github.com/sourcegraph/sourcegraph/internal/encryption/testing"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	gitprotocol "github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestExecutor_ExecutePlan(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	db := database.NewDB(dbtest.NewDB(t))

	now := timeutil.Now()
	clock := func() time.Time { return now }
	cstore := store.NewWithClock(db, &observation.TestContext, et.TestKey{}, clock)

	admin := ct.CreateTestUser(t, db, true)

	repo, extSvc := ct.CreateTestRepo(t, ctx, db)
	ct.CreateTestSiteCredential(t, cstore, repo)

	state := ct.MockChangesetSyncState(&protocol.RepoInfo{
		Name: repo.Name,
		VCS:  protocol.VCSInfo{URL: repo.URI},
	})
	defer state.Unmock()

	internalClient = &mockInternalClient{externalURL: "https://sourcegraph.test"}
	defer func() { internalClient = internalapi.Client }()

	githubPR := buildGithubPR(clock(), btypes.ChangesetExternalStateOpen)
	githubHeadRef := gitdomain.EnsureRefPrefix(githubPR.HeadRefName)
	draftGithubPR := buildGithubPR(clock(), btypes.ChangesetExternalStateDraft)
	closedGitHubPR := buildGithubPR(clock(), btypes.ChangesetExternalStateClosed)

	notFoundErr := sources.ChangesetNotFoundError{
		Changeset: &sources.Changeset{
			Changeset: &btypes.Changeset{ExternalID: "100000"},
		},
	}

	type testCase struct {
		changeset      ct.TestChangesetOpts
		hasCurrentSpec bool
		plan           *Plan

		sourcerMetadata any
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
			plan:           &Plan{Ops: Operations{}},

			wantChangeset: ct.ChangesetAssertions{},
		},
		"import": {
			changeset: ct.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStateUnpublished,
				ExternalID:       githubPR.ID,
			},
			plan: &Plan{
				Ops: Operations{btypes.ReconcilerOperationImport},
			},

			wantLoadFromCodeHost: true,

			wantChangeset: ct.ChangesetAssertions{
				PublicationState: btypes.ChangesetPublicationStatePublished,
				ExternalID:       githubPR.ID,
				ExternalBranch:   githubHeadRef,
				ExternalState:    btypes.ChangesetExternalStateOpen,
				Title:            githubPR.Title,
				Body:             githubPR.Body,
				DiffStat:         state.DiffStat,
			},
		},
		"import and not-found error": {
			changeset: ct.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStateUnpublished,
				ExternalID:       githubPR.ID,
			},
			plan: &Plan{
				Ops: Operations{btypes.ReconcilerOperationImport},
			},
			sourcerErr: notFoundErr,

			wantLoadFromCodeHost: true,

			wantNonRetryableErr: true,
		},
		"push and publish": {
			hasCurrentSpec: true,
			changeset: ct.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStateUnpublished,
			},
			plan: &Plan{
				Ops: Operations{
					btypes.ReconcilerOperationPush,
					btypes.ReconcilerOperationPublish,
				},
			},

			wantCreateOnCodeHost: true,
			wantGitserverCommit:  true,

			wantChangeset: ct.ChangesetAssertions{
				PublicationState: btypes.ChangesetPublicationStatePublished,
				ExternalID:       githubPR.ID,
				ExternalBranch:   githubHeadRef,
				ExternalState:    btypes.ChangesetExternalStateOpen,
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
				PublicationState: btypes.ChangesetPublicationStateUnpublished,
			},
			plan: &Plan{
				Ops: Operations{
					btypes.ReconcilerOperationPush,
					btypes.ReconcilerOperationPublish,
				},
			},

			// We first do a create and since that fails with "already exists"
			// we update.
			wantCreateOnCodeHost: true,
			wantUpdateOnCodeHost: true,
			wantGitserverCommit:  true,

			wantChangeset: ct.ChangesetAssertions{
				PublicationState: btypes.ChangesetPublicationStatePublished,
				ExternalID:       githubPR.ID,
				ExternalBranch:   githubHeadRef,
				ExternalState:    btypes.ChangesetExternalStateOpen,
				Title:            githubPR.Title,
				Body:             githubPR.Body,
				DiffStat:         state.DiffStat,
			},
		},
		"update": {
			hasCurrentSpec: true,
			changeset: ct.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStatePublished,
				ExternalID:       "12345",
				ExternalBranch:   "head-ref-on-github",
			},

			plan: &Plan{
				Ops: Operations{
					btypes.ReconcilerOperationUpdate,
				},
			},

			// We don't want a new commit, only an update on the code host.
			wantUpdateOnCodeHost: true,

			wantChangeset: ct.ChangesetAssertions{
				PublicationState: btypes.ChangesetPublicationStatePublished,
				ExternalID:       githubPR.ID,
				ExternalBranch:   githubHeadRef,
				ExternalState:    btypes.ChangesetExternalStateOpen,
				DiffStat:         state.DiffStat,
				// We update the title/body but want the title/body returned by the code host.
				Title: githubPR.Title,
				Body:  githubPR.Body,
			},
		},
		"push sleep sync": {
			hasCurrentSpec: true,
			changeset: ct.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStatePublished,
				ExternalID:       "12345",
				ExternalBranch:   gitdomain.EnsureRefPrefix("head-ref-on-github"),
				ExternalState:    btypes.ChangesetExternalStateOpen,
			},

			plan: &Plan{
				Ops: Operations{
					btypes.ReconcilerOperationPush,
					btypes.ReconcilerOperationSleep,
					btypes.ReconcilerOperationSync,
				},
			},

			wantGitserverCommit:  true,
			wantLoadFromCodeHost: true,

			wantChangeset: ct.ChangesetAssertions{
				PublicationState: btypes.ChangesetPublicationStatePublished,
				ExternalState:    btypes.ChangesetExternalStateOpen,
				ExternalID:       githubPR.ID,
				ExternalBranch:   githubHeadRef,
				DiffStat:         state.DiffStat,
			},
		},
		"close open changeset": {
			hasCurrentSpec: true,
			changeset: ct.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStatePublished,
				ExternalID:       githubPR.ID,
				ExternalBranch:   githubHeadRef,
				ExternalState:    btypes.ChangesetExternalStateOpen,
				Closing:          true,
			},
			plan: &Plan{
				Ops: Operations{
					btypes.ReconcilerOperationClose,
				},
			},
			// We return a closed GitHub PR here
			sourcerMetadata: closedGitHubPR,

			wantCloseOnCodeHost: true,

			wantChangeset: ct.ChangesetAssertions{
				PublicationState: btypes.ChangesetPublicationStatePublished,
				Closing:          false,

				ExternalID:     closedGitHubPR.ID,
				ExternalBranch: gitdomain.EnsureRefPrefix(closedGitHubPR.HeadRefName),
				ExternalState:  btypes.ChangesetExternalStateClosed,

				Title:    closedGitHubPR.Title,
				Body:     closedGitHubPR.Body,
				DiffStat: state.DiffStat,
			},
		},
		"close closed changeset": {
			hasCurrentSpec: true,
			changeset: ct.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStatePublished,
				ExternalID:       githubPR.ID,
				ExternalBranch:   githubHeadRef,
				ExternalState:    btypes.ChangesetExternalStateClosed,
				Closing:          true,
			},
			plan: &Plan{
				Ops: Operations{
					btypes.ReconcilerOperationClose,
				},
			},

			// We return a closed GitHub PR here, but since it's a noop, we
			// don't sync and thus don't set its attributes on the changeset.
			sourcerMetadata: closedGitHubPR,

			// Should be a noop
			wantCloseOnCodeHost: false,

			wantChangeset: ct.ChangesetAssertions{
				PublicationState: btypes.ChangesetPublicationStatePublished,
				Closing:          false,

				ExternalID:     closedGitHubPR.ID,
				ExternalBranch: gitdomain.EnsureRefPrefix(closedGitHubPR.HeadRefName),
				ExternalState:  btypes.ChangesetExternalStateClosed,
			},
		},
		"reopening closed changeset without updates": {
			hasCurrentSpec: true,
			changeset: ct.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStatePublished,
				ExternalID:       githubPR.ID,
				ExternalBranch:   githubHeadRef,
				ExternalState:    btypes.ChangesetExternalStateClosed,
			},
			plan: &Plan{
				Ops: Operations{
					btypes.ReconcilerOperationReopen,
				},
			},

			wantReopenOnCodeHost: true,

			wantChangeset: ct.ChangesetAssertions{
				PublicationState: btypes.ChangesetPublicationStatePublished,

				ExternalID:     githubPR.ID,
				ExternalBranch: githubHeadRef,
				ExternalState:  btypes.ChangesetExternalStateOpen,

				Title:    githubPR.Title,
				Body:     githubPR.Body,
				DiffStat: state.DiffStat,
			},
		},
		"push and publishdraft": {
			hasCurrentSpec: true,
			changeset: ct.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStateUnpublished,
			},

			plan: &Plan{
				Ops: Operations{
					btypes.ReconcilerOperationPush,
					btypes.ReconcilerOperationPublishDraft,
				},
			},

			sourcerMetadata: draftGithubPR,

			wantCreateDraftOnCodeHost: true,
			wantGitserverCommit:       true,

			wantChangeset: ct.ChangesetAssertions{
				PublicationState: btypes.ChangesetPublicationStatePublished,

				ExternalID:     draftGithubPR.ID,
				ExternalBranch: gitdomain.EnsureRefPrefix(draftGithubPR.HeadRefName),
				ExternalState:  btypes.ChangesetExternalStateDraft,

				Title:    draftGithubPR.Title,
				Body:     draftGithubPR.Body,
				DiffStat: state.DiffStat,
			},
		},
		"undraft": {
			hasCurrentSpec: true,
			changeset: ct.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStatePublished,
				ExternalState:    btypes.ChangesetExternalStateDraft,
			},

			plan: &Plan{
				Ops: Operations{
					btypes.ReconcilerOperationUndraft,
				},
			},

			wantUndraftOnCodeHost: true,

			wantChangeset: ct.ChangesetAssertions{
				PublicationState: btypes.ChangesetPublicationStatePublished,

				ExternalID:     githubPR.ID,
				ExternalBranch: githubHeadRef,
				ExternalState:  btypes.ChangesetExternalStateOpen,

				Title:    githubPR.Title,
				Body:     githubPR.Body,
				DiffStat: state.DiffStat,
			},
		},
		"archive open changeset": {
			hasCurrentSpec: false,
			changeset: ct.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStatePublished,
				ExternalID:       githubPR.ID,
				ExternalBranch:   githubHeadRef,
				ExternalState:    btypes.ChangesetExternalStateOpen,
				Closing:          true,
				BatchChanges: []btypes.BatchChangeAssoc{{
					BatchChangeID: 1234, Archive: true, IsArchived: false,
				}},
			},
			plan: &Plan{
				Ops: Operations{
					btypes.ReconcilerOperationClose,
					btypes.ReconcilerOperationArchive,
				},
			},
			// We return a closed GitHub PR here
			sourcerMetadata: closedGitHubPR,

			wantCloseOnCodeHost: true,

			wantChangeset: ct.ChangesetAssertions{
				PublicationState: btypes.ChangesetPublicationStatePublished,
				Closing:          false,

				ExternalID:     closedGitHubPR.ID,
				ExternalBranch: gitdomain.EnsureRefPrefix(closedGitHubPR.HeadRefName),
				ExternalState:  btypes.ChangesetExternalStateClosed,

				Title:    closedGitHubPR.Title,
				Body:     closedGitHubPR.Body,
				DiffStat: state.DiffStat,

				ArchivedInOwnerBatchChange: true,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Create necessary associations.
			batchSpec := ct.CreateBatchSpec(t, ctx, cstore, "executor-test-batch-change", admin.ID)
			batchChange := ct.CreateBatchChange(t, ctx, cstore, "executor-test-batch-change", admin.ID, batchSpec.ID)

			// Create the changesetSpec with associations wired up correctly.
			var changesetSpec *btypes.ChangesetSpec
			if tc.hasCurrentSpec {
				// The attributes of the spec don't really matter, but the
				// associations do.
				specOpts := ct.TestSpecOpts{}
				specOpts.User = admin.ID
				specOpts.Repo = repo.ID
				specOpts.BatchSpec = batchSpec.ID
				changesetSpec = ct.CreateChangesetSpec(t, ctx, cstore, specOpts)
			}

			// Create the changeset with correct associations.
			changesetOpts := tc.changeset
			changesetOpts.Repo = repo.ID
			if len(changesetOpts.BatchChanges) != 0 {
				for i := range changesetOpts.BatchChanges {
					changesetOpts.BatchChanges[i].BatchChangeID = batchChange.ID
				}
			} else {
				changesetOpts.BatchChanges = []btypes.BatchChangeAssoc{{BatchChangeID: batchChange.ID}}
			}
			changesetOpts.OwnedByBatchChange = batchChange.ID
			if changesetSpec != nil {
				changesetOpts.CurrentSpec = changesetSpec.ID
			}
			changeset := ct.CreateChangeset(t, ctx, cstore, changesetOpts)

			// Setup gitserver dependency.
			gitClient := &ct.FakeGitserverClient{ResponseErr: nil}
			if changesetSpec != nil {
				gitClient.Response = changesetSpec.Spec.HeadRef
			}

			// Setup the sourcer that's used to create a Source with which
			// to create/update a changeset.
			fakeSource := &sources.FakeChangesetSource{
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

			sourcer := sources.NewFakeSourcer(nil, fakeSource)

			tc.plan.Changeset = changeset
			tc.plan.ChangesetSpec = changesetSpec

			// Execute the plan
			err := executePlan(
				ctx,
				gitClient,
				sourcer,
				// Don't actually sleep for the sake of testing.
				true,
				cstore,
				tc.plan,
			)
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
			assertions.Repo = repo.ID
			assertions.OwnedByBatchChange = changesetOpts.OwnedByBatchChange
			assertions.AttachedTo = []int64{batchChange.ID}
			if changesetSpec != nil {
				assertions.CurrentSpec = changesetSpec.ID
			}
			ct.ReloadAndAssertChangeset(t, ctx, cstore, changeset, assertions)

			// Assert that the body included a backlink if needed. We'll do
			// more detailed unit tests of decorateChangesetBody elsewhere;
			// we're just looking for a basic marker here that _something_
			// happened.
			var rcs *sources.Changeset
			if tc.wantCreateOnCodeHost && fakeSource.CreateChangesetCalled {
				rcs = fakeSource.CreatedChangesets[0]
			} else if tc.wantUpdateOnCodeHost && fakeSource.UpdateChangesetCalled {
				rcs = fakeSource.UpdatedChangesets[0]
			}

			if rcs != nil {
				if !strings.Contains(rcs.Body, "Created by Sourcegraph batch change") {
					t.Errorf("did not find backlink in body: %q", rcs.Body)
				}
			}
		})

		// After each test: clean up database.
		ct.TruncateTables(t, db, "changeset_events", "changesets", "batch_changes", "batch_specs", "changeset_specs")
	}
}

func TestExecutor_ExecutePlan_PublishedChangesetDuplicateBranch(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	db := database.NewDB(dbtest.NewDB(t))

	cstore := store.New(db, &observation.TestContext, et.TestKey{})

	repo, _ := ct.CreateTestRepo(t, ctx, db)

	commonHeadRef := "refs/heads/collision"

	// Create a published changeset.
	ct.CreateChangeset(t, ctx, cstore, ct.TestChangesetOpts{
		Repo:             repo.ID,
		PublicationState: btypes.ChangesetPublicationStatePublished,
		ExternalBranch:   commonHeadRef,
		ExternalID:       "123",
	})

	// Plan only needs a push operation, since that's where we check
	plan := &Plan{}
	plan.AddOp(btypes.ReconcilerOperationPush)

	// Build a changeset that would be pushed on the same HeadRef/ExternalBranch.
	plan.ChangesetSpec = ct.BuildChangesetSpec(t, ct.TestSpecOpts{
		Repo:      repo.ID,
		HeadRef:   commonHeadRef,
		Published: true,
	})
	plan.Changeset = ct.BuildChangeset(ct.TestChangesetOpts{Repo: repo.ID})

	err := executePlan(ctx, nil, sources.NewFakeSourcer(nil, &sources.FakeChangesetSource{}), true, cstore, plan)
	if err == nil {
		t.Fatal("reconciler did not return error")
	}

	// We expect a non-retryable error to be returned.
	if !errcode.IsNonRetryable(err) {
		t.Fatalf("error is not non-retryabe. have=%s", err)
	}
}

func TestExecutor_ExecutePlan_AvoidLoadingChangesetSource(t *testing.T) {
	ctx := context.Background()
	db := database.NewDB(dbtest.NewDB(t))
	cstore := store.New(db, &observation.TestContext, et.TestKey{})
	repo, _ := ct.CreateTestRepo(t, ctx, db)

	changesetSpec := ct.BuildChangesetSpec(t, ct.TestSpecOpts{
		Repo:      repo.ID,
		HeadRef:   "refs/heads/my-pr",
		Published: true,
	})
	changeset := ct.BuildChangeset(ct.TestChangesetOpts{ExternalState: "OPEN", Repo: repo.ID})

	ourError := errors.New("this should not be returned")
	sourcer := sources.NewFakeSourcer(ourError, &sources.FakeChangesetSource{})

	t.Run("plan requires changeset source", func(t *testing.T) {
		plan := &Plan{}
		plan.ChangesetSpec = changesetSpec
		plan.Changeset = changeset

		plan.AddOp(btypes.ReconcilerOperationClose)

		err := executePlan(ctx, nil, sourcer, true, cstore, plan)
		if err != ourError {
			t.Fatalf("executePlan did not return expected error: %s", err)
		}
	})

	t.Run("plan does not require changeset source", func(t *testing.T) {
		plan := &Plan{}
		plan.ChangesetSpec = changesetSpec
		plan.Changeset = changeset

		plan.AddOp(btypes.ReconcilerOperationDetach)

		err := executePlan(ctx, nil, sourcer, true, cstore, plan)
		if err != nil {
			t.Fatalf("executePlan returned unexpected error: %s", err)
		}
	})
}

func TestLoadChangesetSource(t *testing.T) {
	ctx := actor.WithInternalActor(context.Background())
	db := database.NewDB(dbtest.NewDB(t))
	token := &auth.OAuthBearerToken{Token: "abcdef"}

	cstore := store.New(db, &observation.TestContext, et.TestKey{})

	admin := ct.CreateTestUser(t, db, true)
	user := ct.CreateTestUser(t, db, false)

	repo, _ := ct.CreateTestRepo(t, ctx, db)

	batchSpec := ct.CreateBatchSpec(t, ctx, cstore, "reconciler-test-batch-change", admin.ID)
	adminBatchChange := ct.CreateBatchChange(t, ctx, cstore, "reconciler-test-batch-change", admin.ID, batchSpec.ID)
	userBatchChange := ct.CreateBatchChange(t, ctx, cstore, "reconciler-test-batch-change", user.ID, batchSpec.ID)

	t.Run("imported changeset uses global token when no site-credential exists", func(t *testing.T) {
		fakeSource := &sources.FakeChangesetSource{}
		sourcer := sources.NewFakeSourcer(nil, fakeSource)
		_, err := loadChangesetSource(ctx, cstore, sourcer, &btypes.Changeset{
			OwnedByBatchChangeID: 0,
		}, repo)
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}
		if fakeSource.CurrentAuthenticator != nil {
			t.Errorf("unexpected non-nil authenticator: %v", fakeSource.CurrentAuthenticator)
		}
	})

	t.Run("imported changeset uses site-credential when exists", func(t *testing.T) {
		if err := cstore.CreateSiteCredential(ctx, &btypes.SiteCredential{
			ExternalServiceType: repo.ExternalRepo.ServiceType,
			ExternalServiceID:   repo.ExternalRepo.ServiceID,
		}, token); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			ct.TruncateTables(t, db, "batch_changes_site_credentials")
		})
		fakeSource := &sources.FakeChangesetSource{}
		sourcer := sources.NewFakeSourcer(nil, fakeSource)
		_, err := loadChangesetSource(ctx, cstore, sourcer, &btypes.Changeset{
			OwnedByBatchChangeID: 0,
		}, repo)
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}
		if diff := cmp.Diff(token, fakeSource.CurrentAuthenticator); diff != "" {
			t.Errorf("unexpected authenticator:\n%s", diff)
		}
	})

	t.Run("owned by missing batch change", func(t *testing.T) {
		fakeSource := &sources.FakeChangesetSource{}
		sourcer := sources.NewFakeSourcer(nil, fakeSource)
		_, err := loadChangesetSource(ctx, cstore, sourcer, &btypes.Changeset{
			OwnedByBatchChangeID: 1234,
		}, repo)
		if err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("owned by admin user without credential", func(t *testing.T) {
		fakeSource := &sources.FakeChangesetSource{}
		sourcer := sources.NewFakeSourcer(nil, fakeSource)
		_, err := loadChangesetSource(ctx, cstore, sourcer, &btypes.Changeset{
			OwnedByBatchChangeID: adminBatchChange.ID,
		}, repo)
		if !errors.Is(err, errMissingCredentials{repo: string(repo.Name)}) {
			t.Fatalf("unexpected error %v", err)
		}
	})

	t.Run("owned by normal user without credential", func(t *testing.T) {
		fakeSource := &sources.FakeChangesetSource{}
		sourcer := sources.NewFakeSourcer(nil, fakeSource)
		_, err := loadChangesetSource(ctx, cstore, sourcer, &btypes.Changeset{
			OwnedByBatchChangeID: userBatchChange.ID,
		}, repo)
		if err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("owned by admin user with credential", func(t *testing.T) {
		if _, err := cstore.UserCredentials().Create(ctx, database.UserCredentialScope{
			Domain:              database.UserCredentialDomainBatches,
			UserID:              admin.ID,
			ExternalServiceType: repo.ExternalRepo.ServiceType,
			ExternalServiceID:   repo.ExternalRepo.ServiceID,
		}, token); err != nil {
			t.Fatal(err)
		}

		fakeSource := &sources.FakeChangesetSource{}
		sourcer := sources.NewFakeSourcer(nil, fakeSource)
		_, err := loadChangesetSource(ctx, cstore, sourcer, &btypes.Changeset{
			OwnedByBatchChangeID: adminBatchChange.ID,
		}, repo)
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}
		if diff := cmp.Diff(token, fakeSource.CurrentAuthenticator); diff != "" {
			t.Errorf("unexpected authenticator:\n%s", diff)
		}
	})

	t.Run("owned by normal user with credential", func(t *testing.T) {
		if _, err := cstore.UserCredentials().Create(ctx, database.UserCredentialScope{
			Domain:              database.UserCredentialDomainBatches,
			UserID:              user.ID,
			ExternalServiceType: repo.ExternalRepo.ServiceType,
			ExternalServiceID:   repo.ExternalRepo.ServiceID,
		}, token); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			ct.TruncateTables(t, db, "user_credentials")
		})

		fakeSource := &sources.FakeChangesetSource{}
		sourcer := sources.NewFakeSourcer(nil, fakeSource)
		_, err := loadChangesetSource(ctx, cstore, sourcer, &btypes.Changeset{
			OwnedByBatchChangeID: userBatchChange.ID,
		}, repo)
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}
		if diff := cmp.Diff(token, fakeSource.CurrentAuthenticator); diff != "" {
			t.Errorf("unexpected authenticator:\n%s", diff)
		}
	})

	t.Run("owned by user without credential falls back to site-credential", func(t *testing.T) {
		if err := cstore.CreateSiteCredential(ctx, &btypes.SiteCredential{
			ExternalServiceType: repo.ExternalRepo.ServiceType,
			ExternalServiceID:   repo.ExternalRepo.ServiceID,
		}, token); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			ct.TruncateTables(t, db, "batch_changes_site_credentials")
		})

		fakeSource := &sources.FakeChangesetSource{}
		sourcer := sources.NewFakeSourcer(nil, fakeSource)
		_, err := loadChangesetSource(ctx, cstore, sourcer, &btypes.Changeset{
			OwnedByBatchChangeID: userBatchChange.ID,
		}, repo)
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}
		if diff := cmp.Diff(token, fakeSource.CurrentAuthenticator); diff != "" {
			t.Errorf("unexpected authenticator:\n%s", diff)
		}
	})
}

func TestExecutor_UserCredentialsForGitserver(t *testing.T) {
	ctx := actor.WithInternalActor(context.Background())
	db := database.NewDB(dbtest.NewDB(t))

	cstore := store.New(db, &observation.TestContext, et.TestKey{})

	admin := ct.CreateTestUser(t, db, true)
	user := ct.CreateTestUser(t, db, false)

	gitHubRepo, gitHubExtSvc := ct.CreateTestRepo(t, ctx, db)

	gitLabRepos, gitLabExtSvc := ct.CreateGitlabTestRepos(t, ctx, db, 1)
	gitLabRepo := gitLabRepos[0]

	bbsRepos, bbsExtSvc := ct.CreateBbsTestRepos(t, ctx, db, 1)
	bbsRepo := bbsRepos[0]

	bbsSSHRepos, bbsSSHExtsvc := ct.CreateBbsSSHTestRepos(t, ctx, db, 1)
	bbsSSHRepo := bbsSSHRepos[0]

	plan := &Plan{}
	plan.AddOp(btypes.ReconcilerOperationPush)

	tests := []struct {
		name           string
		user           *types.User
		extSvc         *types.ExternalService
		repo           *types.Repo
		credentials    auth.Authenticator
		wantErr        bool
		wantPushConfig *gitprotocol.PushConfig
	}{
		{
			name:        "github OAuthBearerToken",
			user:        user,
			extSvc:      gitHubExtSvc,
			repo:        gitHubRepo,
			credentials: &auth.OAuthBearerToken{Token: "my-secret-github-token"},
			wantPushConfig: &gitprotocol.PushConfig{
				RemoteURL: "https://my-secret-github-token@github.com/sourcegraph/" + string(gitHubRepo.Name),
			},
		},
		{
			name:    "github no credentials",
			user:    user,
			extSvc:  gitHubExtSvc,
			repo:    gitHubRepo,
			wantErr: true,
		},
		{
			name:    "github site-admin and no credentials",
			extSvc:  gitHubExtSvc,
			repo:    gitHubRepo,
			user:    admin,
			wantErr: true,
		},
		{
			name:        "gitlab OAuthBearerToken",
			user:        user,
			extSvc:      gitLabExtSvc,
			repo:        gitLabRepo,
			credentials: &auth.OAuthBearerToken{Token: "my-secret-gitlab-token"},
			wantPushConfig: &gitprotocol.PushConfig{
				RemoteURL: "https://git:my-secret-gitlab-token@gitlab.com/sourcegraph/" + string(gitLabRepo.Name),
			},
		},
		{
			name:    "gitlab no credentials",
			user:    user,
			extSvc:  gitLabExtSvc,
			repo:    gitLabRepo,
			wantErr: true,
		},
		{
			name:    "gitlab site-admin and no credentials",
			user:    admin,
			extSvc:  gitLabExtSvc,
			repo:    gitLabRepo,
			wantErr: true,
		},
		{
			name:        "bitbucketServer BasicAuth",
			user:        user,
			extSvc:      bbsExtSvc,
			repo:        bbsRepo,
			credentials: &auth.BasicAuth{Username: "fredwoard johnssen", Password: "my-secret-bbs-token"},
			wantPushConfig: &gitprotocol.PushConfig{
				RemoteURL: "https://fredwoard%20johnssen:my-secret-bbs-token@bitbucket.sourcegraph.com/scm/" + string(bbsRepo.Name),
			},
		},
		{
			name:    "bitbucketServer no credentials",
			user:    user,
			extSvc:  bbsExtSvc,
			repo:    bbsRepo,
			wantErr: true,
		},
		{
			name:    "bitbucketServer site-admin and no credentials",
			user:    admin,
			extSvc:  bbsExtSvc,
			repo:    bbsRepo,
			wantErr: true,
		},
		{
			name:    "ssh clone URL no credentials",
			user:    user,
			extSvc:  bbsSSHExtsvc,
			repo:    bbsSSHRepo,
			wantErr: true,
		},
		{
			name:    "ssh clone URL no credentials admin",
			user:    admin,
			extSvc:  bbsSSHExtsvc,
			repo:    bbsSSHRepo,
			wantErr: true,
		},
		{
			name:   "ssh clone URL SSH credential",
			user:   admin,
			extSvc: bbsSSHExtsvc,
			repo:   bbsSSHRepo,
			credentials: &auth.OAuthBearerTokenWithSSH{
				OAuthBearerToken: auth.OAuthBearerToken{Token: "test"},
				PrivateKey:       "private key",
				PublicKey:        "public key",
				Passphrase:       "passphrase",
			},
			wantPushConfig: &gitprotocol.PushConfig{
				RemoteURL:  "ssh://git@bitbucket.sgdev.org:7999/" + string(bbsSSHRepo.Name),
				PrivateKey: "private key",
				Passphrase: "passphrase",
			},
		},
		{
			name:        "ssh clone URL non-SSH credential",
			user:        admin,
			extSvc:      bbsSSHExtsvc,
			repo:        bbsSSHRepo,
			credentials: &auth.OAuthBearerToken{Token: "test"},
			wantErr:     true,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.credentials != nil {
				cred, err := cstore.UserCredentials().Create(ctx, database.UserCredentialScope{
					Domain:              database.UserCredentialDomainBatches,
					UserID:              tt.user.ID,
					ExternalServiceType: tt.repo.ExternalRepo.ServiceType,
					ExternalServiceID:   tt.repo.ExternalRepo.ServiceID,
				}, tt.credentials)
				if err != nil {
					t.Fatal(err)
				}
				defer func() { cstore.UserCredentials().Delete(ctx, cred.ID) }()
			}

			batchSpec := ct.CreateBatchSpec(t, ctx, cstore, fmt.Sprintf("reconciler-credentials-%d", i), tt.user.ID)
			batchChange := ct.CreateBatchChange(t, ctx, cstore, fmt.Sprintf("reconciler-credentials-%d", i), tt.user.ID, batchSpec.ID)

			plan.Changeset = &btypes.Changeset{
				OwnedByBatchChangeID: batchChange.ID,
				RepoID:               tt.repo.ID,
			}
			plan.ChangesetSpec = ct.BuildChangesetSpec(t, ct.TestSpecOpts{
				HeadRef:    "refs/heads/my-branch",
				Published:  true,
				CommitDiff: "testdiff",
			})

			gitClient := &ct.FakeGitserverClient{ResponseErr: nil}
			fakeSource := &sources.FakeChangesetSource{Svc: tt.extSvc}
			sourcer := sources.NewFakeSourcer(nil, fakeSource)

			err := executePlan(
				context.Background(),
				gitClient,
				sourcer,
				true,
				cstore,
				plan,
			)

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

func TestLoadRemoteRepo(t *testing.T) {
	ctx := context.Background()
	targetRepo := &types.Repo{}

	t.Run("forks disabled", func(t *testing.T) {
		t.Run("unforked changeset", func(t *testing.T) {
			// Set up a changeset source that will panic if any methods are invoked.
			css := NewStrictMockChangesetSource()

			// This should succeed, since loadRemoteRepo() should early return with
			// forks disabled.
			remoteRepo, err := loadRemoteRepo(ctx, css, targetRepo, &btypes.Changeset{}, &btypes.ChangesetSpec{
				ForkNamespace: nil,
			})
			assert.Nil(t, err)
			assert.Same(t, targetRepo, remoteRepo)
		})

		t.Run("forked changeset", func(t *testing.T) {
			forkNamespace := "fork"
			want := &types.Repo{}
			css := NewMockForkableChangesetSource()
			css.GetNamespaceForkFunc.SetDefaultReturn(want, nil)

			// This should succeed, since loadRemoteRepo() should early return with
			// forks disabled.
			remoteRepo, err := loadRemoteRepo(ctx, css, targetRepo, &btypes.Changeset{}, &btypes.ChangesetSpec{
				ForkNamespace: &forkNamespace,
			})
			assert.Nil(t, err)
			assert.Same(t, want, remoteRepo)
			mockassert.CalledOnce(t, css.GetNamespaceForkFunc)
		})
	})

	t.Run("forks enabled", func(t *testing.T) {
		forkNamespace := "<user>"

		t.Run("unforkable changeset source", func(t *testing.T) {
			css := NewMockChangesetSource()

			repo, err := loadRemoteRepo(ctx, css, targetRepo, &btypes.Changeset{}, &btypes.ChangesetSpec{
				ForkNamespace: &forkNamespace,
			})
			assert.Nil(t, repo)
			assert.ErrorIs(t, err, errChangesetSourceCannotFork)
		})

		t.Run("forkable changeset source", func(t *testing.T) {
			t.Run("success", func(t *testing.T) {
				want := &types.Repo{}
				css := NewMockForkableChangesetSource()
				css.GetUserForkFunc.SetDefaultReturn(want, nil)

				have, err := loadRemoteRepo(ctx, css, targetRepo, &btypes.Changeset{}, &btypes.ChangesetSpec{
					ForkNamespace: &forkNamespace,
				})
				assert.Nil(t, err)
				assert.Same(t, want, have)
				mockassert.CalledOnce(t, css.GetUserForkFunc)
			})

			t.Run("error from the source", func(t *testing.T) {
				want := errors.New("source error")
				css := NewMockForkableChangesetSource()
				css.GetUserForkFunc.SetDefaultReturn(nil, want)

				repo, err := loadRemoteRepo(ctx, css, targetRepo, &btypes.Changeset{}, &btypes.ChangesetSpec{
					ForkNamespace: &forkNamespace,
				})
				assert.Nil(t, repo)
				assert.Same(t, want, err)
				mockassert.CalledOnce(t, css.GetUserForkFunc)
			})
		})
	})
}

func TestDecorateChangesetBody(t *testing.T) {
	ns := database.NewMockNamespaceStore()
	ns.GetByIDFunc.SetDefaultHook(func(_ context.Context, _ int32, user int32) (*database.Namespace, error) {
		return &database.Namespace{Name: "my-user", User: user}, nil
	})

	internalClient = &mockInternalClient{externalURL: "https://sourcegraph.test"}
	defer func() { internalClient = internalapi.Client }()

	fs := &FakeStore{
		GetBatchChangeMock: func(ctx context.Context, opts store.GetBatchChangeOpts) (*btypes.BatchChange, error) {
			return &btypes.BatchChange{ID: 1234, Name: "reconciler-test-batch-change"}, nil
		},
	}

	cs := ct.BuildChangeset(ct.TestChangesetOpts{OwnedByBatchChange: 1234})

	body := "body"
	rcs := &sources.Changeset{Body: body, Changeset: cs}
	if err := decorateChangesetBody(context.Background(), fs, ns, rcs); err != nil {
		t.Errorf("unexpected non-nil error: %v", err)
	}
	if want := body + "\n\n[_Created by Sourcegraph batch change `my-user/reconciler-test-batch-change`._](https://sourcegraph.test/users/my-user/batch-changes/reconciler-test-batch-change)"; rcs.Body != want {
		t.Errorf("repos.Changeset body unexpectedly changed:\nhave=%q\nwant=%q", rcs.Body, want)
	}
}

func TestBatchChangeURL(t *testing.T) {
	ctx := context.Background()

	t.Run("errors", func(t *testing.T) {
		for name, tc := range map[string]*mockInternalClient{
			"ExternalURL error": {err: errors.New("foo")},
			"invalid URL":       {externalURL: "foo://:bar"},
		} {
			t.Run(name, func(t *testing.T) {
				internalClient = tc
				defer func() { internalClient = internalapi.Client }()

				if _, err := batchChangeURL(ctx, nil, nil); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		internalClient = &mockInternalClient{externalURL: "https://sourcegraph.test"}
		defer func() { internalClient = internalapi.Client }()

		url, err := batchChangeURL(
			ctx,
			&database.Namespace{Name: "foo", Organization: 123},
			&btypes.BatchChange{Name: "bar"},
		)
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}
		if want := "https://sourcegraph.test/organizations/foo/batch-changes/bar"; url != want {
			t.Errorf("unexpected URL: have=%q want=%q", url, want)
		}
	})
}

func TestNamespaceURL(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		ns   *database.Namespace
		want string
	}{
		"user": {
			ns:   &database.Namespace{User: 123, Name: "user"},
			want: "/users/user",
		},
		"org": {
			ns:   &database.Namespace{Organization: 123, Name: "org"},
			want: "/organizations/org",
		},
		"neither": {
			ns:   &database.Namespace{Name: "user"},
			want: "/users/user",
		},
	} {
		t.Run(name, func(t *testing.T) {
			if have := namespaceURL(tc.ns); have != tc.want {
				t.Errorf("unexpected URL: have=%q want=%q", have, tc.want)
			}
		})
	}
}

type mockInternalClient struct {
	externalURL string
	err         error
}

func (c *mockInternalClient) ExternalURL(ctx context.Context) (string, error) {
	return c.externalURL, c.err
}
