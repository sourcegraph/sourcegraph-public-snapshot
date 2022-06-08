package reconciler

import (
	"context"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/sources"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/lib/log/logtest"
)

func TestReconcilerProcess_IntegrationTest(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := actor.WithInternalActor(context.Background())
	db := database.NewDB(dbtest.NewDB(t))
	log := logtest.Scoped(t)

	store := store.New(db, &observation.TestContext, nil)

	admin := ct.CreateTestUser(t, db, true)

	repo, extSvc := ct.CreateTestRepo(t, ctx, db)
	ct.CreateTestSiteCredential(t, store, repo)

	state := ct.MockChangesetSyncState(&protocol.RepoInfo{
		Name: repo.Name,
		VCS:  protocol.VCSInfo{URL: repo.URI},
	})
	defer state.Unmock()

	internalClient = &mockInternalClient{externalURL: "https://sourcegraph.test"}
	defer func() { internalClient = internalapi.Client }()

	githubPR := buildGithubPR(time.Now(), btypes.ChangesetExternalStateOpen)
	githubHeadRef := gitdomain.EnsureRefPrefix(githubPR.HeadRefName)

	type testCase struct {
		changeset    ct.TestChangesetOpts
		currentSpec  *ct.TestSpecOpts
		previousSpec *ct.TestSpecOpts

		wantChangeset ct.ChangesetAssertions
	}

	tests := map[string]testCase{
		"update a published changeset": {
			currentSpec: &ct.TestSpecOpts{
				HeadRef:   "refs/heads/head-ref-on-github",
				Published: true,
			},

			previousSpec: &ct.TestSpecOpts{
				HeadRef:   "refs/heads/head-ref-on-github",
				Published: true,

				Title:         "old title",
				Body:          "old body",
				CommitDiff:    "old diff",
				CommitMessage: "old message",
			},

			changeset: ct.TestChangesetOpts{
				PublicationState: btypes.ChangesetPublicationStatePublished,
				ExternalID:       "12345",
				ExternalBranch:   "head-ref-on-github",
			},

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
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Create necessary associations.
			previousBatchSpec := ct.CreateBatchSpec(t, ctx, store, "reconciler-test-batch-change", admin.ID)
			batchSpec := ct.CreateBatchSpec(t, ctx, store, "reconciler-test-batch-change", admin.ID)
			batchChange := ct.CreateBatchChange(t, ctx, store, "reconciler-test-batch-change", admin.ID, batchSpec.ID)

			// Create the specs.
			specOpts := *tc.currentSpec
			specOpts.User = admin.ID
			specOpts.Repo = repo.ID
			specOpts.BatchSpec = batchSpec.ID
			changesetSpec := ct.CreateChangesetSpec(t, ctx, store, specOpts)

			previousSpecOpts := *tc.previousSpec
			previousSpecOpts.User = admin.ID
			previousSpecOpts.Repo = repo.ID
			previousSpecOpts.BatchSpec = previousBatchSpec.ID
			previousSpec := ct.CreateChangesetSpec(t, ctx, store, previousSpecOpts)

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
			changeset := ct.CreateChangeset(t, ctx, store, changesetOpts)

			// Setup gitserver dependency.
			gitClient := &ct.FakeGitserverClient{ResponseErr: nil}
			if changesetSpec != nil {
				gitClient.Response = changesetSpec.Spec.HeadRef
			}

			// Setup the sourcer that's used to create a Source with which
			// to create/update a changeset.
			fakeSource := &sources.FakeChangesetSource{Svc: extSvc, FakeMetadata: githubPR}
			if changesetSpec != nil {
				fakeSource.WantHeadRef = changesetSpec.Spec.HeadRef
				fakeSource.WantBaseRef = changesetSpec.Spec.BaseRef
			}

			sourcer := sources.NewFakeSourcer(nil, fakeSource)

			// Run the reconciler
			rec := Reconciler{
				noSleepBeforeSync: true,
				gitserverClient:   gitClient,
				sourcer:           sourcer,
				store:             store,
			}
			err := rec.process(ctx, log, store, changeset)
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
			ct.ReloadAndAssertChangeset(t, ctx, store, changeset, assertions)
		})

		// Clean up database.
		ct.TruncateTables(t, db, "changeset_events", "changesets", "batch_changes", "batch_specs", "changeset_specs")
	}
}
