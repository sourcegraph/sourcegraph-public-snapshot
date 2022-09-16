package reconciler

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/log/logtest"

	stesting "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/sources/testing"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	bt "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
)

func TestReconcilerProcess_IntegrationTest(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := actor.WithInternalActor(context.Background())
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))

	store := store.New(db, &observation.TestContext, nil)

	admin := bt.CreateTestUser(t, db, true)

	repo, extSvc := bt.CreateTestRepo(t, ctx, db)
	bt.CreateTestSiteCredential(t, store, repo)

	state := bt.MockChangesetSyncState(&protocol.RepoInfo{
		Name: repo.Name,
		VCS:  protocol.VCSInfo{URL: repo.URI},
	})
	defer state.Unmock()

	internalClient = &mockInternalClient{externalURL: "https://sourcegraph.test"}
	defer func() { internalClient = internalapi.Client }()

	githubPR := buildGithubPR(time.Now(), btypes.ChangesetExternalStateOpen)
	githubHeadRef := gitdomain.EnsureRefPrefix(githubPR.HeadRefName)

	type testCase struct {
		changeset    bt.TestChangesetOpts
		currentSpec  *bt.TestSpecOpts
		previousSpec *bt.TestSpecOpts

		wantChangeset bt.ChangesetAssertions
	}

	tests := map[string]testCase{
		"update a published changeset": {
			currentSpec: &bt.TestSpecOpts{
				HeadRef:   "refs/heads/head-ref-on-github",
				Typ:       btypes.ChangesetSpecTypeBranch,
				Published: true,
			},

			previousSpec: &bt.TestSpecOpts{
				HeadRef:   "refs/heads/head-ref-on-github",
				Typ:       btypes.ChangesetSpecTypeBranch,
				Published: true,

				Title:         "old title",
				Body:          "old body",
				CommitDiff:    "old diff",
				CommitMessage: "old message",
			},

			changeset: bt.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStatePublished,
				ExternalID:       "12345",
				ExternalBranch:   "head-ref-on-github",
			},

			wantChangeset: bt.ChangesetAssertions{
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
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Create necessary associations.
			previousBatchSpec := bt.CreateBatchSpec(t, ctx, store, "reconciler-test-batch-change", admin.ID, 0)
			batchSpec := bt.CreateBatchSpec(t, ctx, store, "reconciler-test-batch-change", admin.ID, 0)
			batchChange := bt.CreateBatchChange(t, ctx, store, "reconciler-test-batch-change", admin.ID, batchSpec.ID)

			// Create the specs.
			specOpts := *tc.currentSpec
			specOpts.User = admin.ID
			specOpts.Repo = repo.ID
			specOpts.BatchSpec = batchSpec.ID
			changesetSpec := bt.CreateChangesetSpec(t, ctx, store, specOpts)

			previousSpecOpts := *tc.previousSpec
			previousSpecOpts.User = admin.ID
			previousSpecOpts.Repo = repo.ID
			previousSpecOpts.BatchSpec = previousBatchSpec.ID
			previousSpec := bt.CreateChangesetSpec(t, ctx, store, previousSpecOpts)

			// Create the changeset with correct associations.
			changesetOpts := tc.changeset
			changesetOpts.Repo = repo.ID
			changesetOpts.BatchChanges = []btypes.BatchChangeAssoc{{BatchChangeID: batchChange.ID}}
			changesetOpts.OwnedByBatchChange = batchChange.ID
			if changesetSpec != nil {
				changesetOpts.CurrentSpec = changesetSpec.ID
			}
			if previousSpec != nil {
				changesetOpts.PreviousSpec = previousSpec.ID
			}
			changeset := bt.CreateChangeset(t, ctx, store, changesetOpts)

			// Setup gitserver dependency.
			gitClient := &bt.FakeGitserverClient{ResponseErr: nil}
			if changesetSpec != nil {
				gitClient.Response = changesetSpec.HeadRef
			}

			// Setup the sourcer that's used to create a Source with which
			// to create/update a changeset.
			fakeSource := &stesting.FakeChangesetSource{
				Svc:                  extSvc,
				FakeMetadata:         githubPR,
				CurrentAuthenticator: &auth.OAuthBearerTokenWithSSH{},
			}
			if changesetSpec != nil {
				fakeSource.WantHeadRef = changesetSpec.HeadRef
				fakeSource.WantBaseRef = changesetSpec.BaseRef
			}

			sourcer := stesting.NewFakeSourcer(nil, fakeSource)

			// Run the reconciler
			rec := Reconciler{
				noSleepBeforeSync: true,
				gitserverClient:   gitClient,
				sourcer:           sourcer,
				store:             store,
			}
			err := rec.process(ctx, logger, store, changeset)
			if err != nil {
				t.Fatalf("reconciler process failed: %s", err)
			}

			// Assert that the changeset in the database looks like we want
			assertions := tc.wantChangeset
			assertions.Repo = repo.ID
			assertions.OwnedByBatchChange = changesetOpts.OwnedByBatchChange
			assertions.AttachedTo = []int64{batchChange.ID}
			assertions.CurrentSpec = changesetSpec.ID
			assertions.PreviousSpec = previousSpec.ID
			bt.ReloadAndAssertChangeset(t, ctx, store, changeset, assertions)
		})

		// Clean up database.
		bt.TruncateTables(t, db, "changeset_events", "changesets", "batch_changes", "batch_specs", "changeset_specs")
	}
}
