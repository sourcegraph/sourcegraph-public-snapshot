package campaigns

import (
	"context"
	"time"

	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/testing"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
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

	githubPR := buildGithubPR(clock(), "12345", "Remote title", "Remote body", "head-ref-on-github")

	type testCase struct {
		changeset    testChangesetOpts
		currentSpec  *testSpecOpts
		previousSpec *testSpecOpts

		sourcerMetadata interface{}
		// Whether or not the source responds to CreateChangeset with "already exists"
		alreadyExists bool

		wantCreateOnHostCode bool
		wantUpdateOnCodeHost bool
		wantGitserverCommit  bool

		wantChangeset changesetAssertions
	}

	tests := map[string]testCase{
		"published changeset without changesetSpec": {
			// Published changeset without a changesetSpec should be left
			// untouched.
			// But once we move syncing of changesets to the reconciler, we need to assert
			// that it's been synced.
			changeset: testChangesetOpts{
				publicationState: campaigns.ChangesetPublicationStatePublished,
				externalID:       "12345",
			},
			sourcerMetadata: githubPR,

			wantCreateOnHostCode: false,
			wantUpdateOnCodeHost: false,
			wantGitserverCommit:  false,

			wantChangeset: changesetAssertions{
				publicationState: campaigns.ChangesetPublicationStatePublished,
				externalID:       "12345",
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

			wantCreateOnHostCode: false,
			wantUpdateOnCodeHost: false,
			wantGitserverCommit:  false,

			wantChangeset: changesetAssertions{
				publicationState: campaigns.ChangesetPublicationStateUnpublished,
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
			wantUpdateOnCodeHost: false,
			wantGitserverCommit:  true,

			wantChangeset: changesetAssertions{
				publicationState: campaigns.ChangesetPublicationStatePublished,
				externalID:       "12345",
				externalBranch:   "head-ref-on-github",

				title: "Remote title",
				body:  "Remote body",
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
				externalID:       "12345",
				externalBranch:   "head-ref-on-github",

				title: "Remote title",
				body:  "Remote body",
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

			wantCreateOnHostCode: false,
			// We don't want a new commit, only an update on the code host.
			wantUpdateOnCodeHost: true,
			wantGitserverCommit:  false,

			wantChangeset: changesetAssertions{
				publicationState: campaigns.ChangesetPublicationStatePublished,
				externalID:       "12345",
				externalBranch:   "head-ref-on-github",

				// We update the title/body but want the title/body returned by the code host.
				title: "Remote title",
				body:  "Remote body",
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
				createdByCampaign: true,
			},
			sourcerMetadata: githubPR,

			wantCreateOnHostCode: false,
			// We don't want an update on the code host, only a new commit pushed.
			wantUpdateOnCodeHost: false,
			wantGitserverCommit:  true,

			wantChangeset: changesetAssertions{
				publicationState: campaigns.ChangesetPublicationStatePublished,
				externalID:       "12345",
				externalBranch:   "head-ref-on-github",
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
				externalID:        "12345",
				externalBranch:    "head-ref-on-github",
				createdByCampaign: true,
			},
			sourcerMetadata: githubPR,

			wantCreateOnHostCode: false,
			wantUpdateOnCodeHost: false,
			wantGitserverCommit:  false,

			wantChangeset: changesetAssertions{
				publicationState: campaigns.ChangesetPublicationStatePublished,
				externalID:       "12345",
				externalBranch:   "head-ref-on-github",
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

			// Setup the assertion: wantChangeset is what we want in the database
			// after the reconciler processed it.
			metadata := tc.sourcerMetadata

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
				FakeMetadata:    metadata,
			}
			if changesetSpec != nil {
				fakeSource.WantHeadRef = changesetSpec.Spec.HeadRef
				fakeSource.WantBaseRef = changesetSpec.Spec.BaseRef
			}

			sourcer := repos.NewFakeSourcer(nil, fakeSource)

			// Run the reconciler
			rec := reconciler{gitserverClient: gitClient, sourcer: sourcer, store: store}
			if err := rec.process(ctx, store, changeset); err != nil {
				t.Fatalf("reconciler process failed: %s", err)
			}

			// Assert that the changeset in the database looks like we want
			haveChangeset, err := store.GetChangeset(ctx, GetChangesetOpts{ID: changeset.ID})
			if err != nil {
				t.Fatal(err)
			}

			assertions := tc.wantChangeset
			assertions.repo = rs[0].ID
			if changesetSpec != nil {
				assertions.currentSpec = changesetSpec.ID
			}
			if previousSpec != nil {
				assertions.previousSpec = previousSpec.ID
			}
			assertChangeset(t, haveChangeset, assertions)

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
		})
	}
}

var (
	diffStatOne   int32 = 1
	diffStatTwo   int32 = 2
	diffStatThree int32 = 3
)

func buildGithubPR(now time.Time, externalID, title, body, headRef string) interface{} {
	return &github.PullRequest{
		ID:          externalID,
		Title:       title,
		Body:        body,
		HeadRefName: git.AbbreviateRef(headRef),
		Number:      12345,
		State:       "OPEN",
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
}
