package campaigns

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/testing"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater/protocol"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func TestReconcilerProcess_IntegrationTest(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)

	store := NewStore(dbconn.Global)

	admin := createTestUser(t, true)

	rs, extSvc := ct.CreateTestRepos(t, ctx, dbconn.Global, 1)

	state := ct.MockChangesetSyncState(&protocol.RepoInfo{
		Name: api.RepoName(rs[0].Name),
		VCS:  protocol.VCSInfo{URL: rs[0].URI},
	})
	defer state.Unmock()

	internalClient = &mockInternalClient{externalURL: "https://sourcegraph.test"}
	defer func() { internalClient = api.InternalClient }()

	githubPR := buildGithubPR(time.Now(), campaigns.ChangesetExternalStateOpen)
	githubHeadRef := git.EnsureRefPrefix(githubPR.HeadRefName)

	type testCase struct {
		changeset    testChangesetOpts
		currentSpec  *testSpecOpts
		previousSpec *testSpecOpts

		wantChangeset changesetAssertions
	}

	tests := map[string]testCase{
		"update a published changeset": {
			currentSpec: &testSpecOpts{
				headRef:   "refs/heads/head-ref-on-github",
				published: true,
			},

			previousSpec: &testSpecOpts{
				headRef:   "refs/heads/head-ref-on-github",
				published: true,

				title:         "old title",
				body:          "old body",
				commitDiff:    "old diff",
				commitMessage: "old message",
			},

			changeset: testChangesetOpts{
				publicationState: campaigns.ChangesetPublicationStatePublished,
				externalID:       "12345",
				externalBranch:   git.EnsureRefPrefix("head-ref-on-github"),
			},

			wantChangeset: changesetAssertions{
				publicationState: campaigns.ChangesetPublicationStatePublished,
				externalID:       githubPR.ID,
				externalBranch:   githubHeadRef,
				externalState:    campaigns.ChangesetExternalStateOpen,
				diffStat:         state.DiffStat,
				// We update the title/body but want the title/body returned by the code host.
				title: githubPR.Title,
				body:  githubPR.Body,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			// Create necessary associations.
			previousCampaignSpec := createCampaignSpec(t, ctx, store, "reconciler-test-campaign", admin.ID)
			campaignSpec := createCampaignSpec(t, ctx, store, "reconciler-test-campaign", admin.ID)
			campaign := createCampaign(t, ctx, store, "reconciler-test-campaign", admin.ID, campaignSpec.ID)

			// Create the specs.
			specOpts := *tc.currentSpec
			specOpts.user = admin.ID
			specOpts.repo = rs[0].ID
			specOpts.campaignSpec = campaignSpec.ID
			changesetSpec := createChangesetSpec(t, ctx, store, specOpts)

			previousSpecOpts := *tc.previousSpec
			previousSpecOpts.user = admin.ID
			previousSpecOpts.repo = rs[0].ID
			previousSpecOpts.campaignSpec = previousCampaignSpec.ID
			previousSpec := createChangesetSpec(t, ctx, store, previousSpecOpts)

			// Create the changeset with correct associations.
			changesetOpts := tc.changeset
			changesetOpts.repo = rs[0].ID
			changesetOpts.campaignIDs = []int64{campaign.ID}
			changesetOpts.ownedByCampaign = campaign.ID
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
			fakeSource := &ct.FakeChangesetSource{Svc: extSvc, FakeMetadata: githubPR}
			if changesetSpec != nil {
				fakeSource.WantHeadRef = changesetSpec.Spec.HeadRef
				fakeSource.WantBaseRef = changesetSpec.Spec.BaseRef
			}

			sourcer := repos.NewFakeSourcer(nil, fakeSource)

			// Run the reconciler
			rec := Reconciler{
				noSleepBeforeSync: true,
				GitserverClient:   gitClient,
				Sourcer:           sourcer,
				Store:             store,
			}
			err := rec.process(ctx, store, changeset)
			if err != nil {
				t.Fatalf("reconciler process failed: %s", err)
			}

			// Assert that the changeset in the database looks like we want
			assertions := tc.wantChangeset
			assertions.repo = rs[0].ID
			assertions.ownedByCampaign = changesetOpts.ownedByCampaign
			assertions.currentSpec = changesetSpec.ID
			assertions.previousSpec = previousSpec.ID
			reloadAndAssertChangeset(t, ctx, store, changeset, assertions)
		})

		// Clean up database.
		truncateTables(t, dbconn.Global, "changeset_events", "changesets", "campaigns", "campaign_specs", "changeset_specs")
	}
}

func TestDetermineReconcilerPlan(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name           string
		previousSpec   testSpecOpts
		currentSpec    testSpecOpts
		changeset      testChangesetOpts
		wantOperations ReconcilerOperations
	}{
		{
			name:        "publish true",
			currentSpec: testSpecOpts{published: true},
			changeset: testChangesetOpts{
				publicationState: campaigns.ChangesetPublicationStateUnpublished,
			},
			wantOperations: ReconcilerOperations{
				campaigns.ReconcilerOperationPush,
				campaigns.ReconcilerOperationPublish,
			},
		},
		{
			name:        "publish as draft",
			currentSpec: testSpecOpts{published: "draft"},
			changeset: testChangesetOpts{
				publicationState: campaigns.ChangesetPublicationStateUnpublished,
			},
			wantOperations: ReconcilerOperations{campaigns.ReconcilerOperationPush, campaigns.ReconcilerOperationPublishDraft},
		},
		{
			name:        "publish false",
			currentSpec: testSpecOpts{published: false},
			changeset: testChangesetOpts{
				publicationState: campaigns.ChangesetPublicationStateUnpublished,
			},
			wantOperations: ReconcilerOperations{},
		},
		{
			name:        "draft but unsupported",
			currentSpec: testSpecOpts{published: "draft"},
			changeset: testChangesetOpts{
				externalServiceType: extsvc.TypeBitbucketServer,
				publicationState:    campaigns.ChangesetPublicationStateUnpublished,
			},
			// should be a noop
			wantOperations: ReconcilerOperations{},
		},
		{
			name:         "draft to publish true",
			previousSpec: testSpecOpts{published: "draft"},
			currentSpec:  testSpecOpts{published: true},
			changeset: testChangesetOpts{
				publicationState: campaigns.ChangesetPublicationStatePublished,
			},
			wantOperations: ReconcilerOperations{campaigns.ReconcilerOperationUndraft},
		},
		{
			name:         "draft to publish true on unpublished changeset",
			previousSpec: testSpecOpts{published: "draft"},
			currentSpec:  testSpecOpts{published: true},
			changeset: testChangesetOpts{
				publicationState: campaigns.ChangesetPublicationStateUnpublished,
			},
			wantOperations: ReconcilerOperations{campaigns.ReconcilerOperationPush, campaigns.ReconcilerOperationPublish},
		},
		{
			name:         "title changed on published changeset",
			previousSpec: testSpecOpts{published: true, title: "Before"},
			currentSpec:  testSpecOpts{published: true, title: "After"},
			changeset: testChangesetOpts{
				publicationState: campaigns.ChangesetPublicationStatePublished,
			},
			wantOperations: ReconcilerOperations{campaigns.ReconcilerOperationUpdate},
		},
		{
			name:         "commit diff changed on published changeset",
			previousSpec: testSpecOpts{published: true, commitDiff: "testDiff"},
			currentSpec:  testSpecOpts{published: true, commitDiff: "newTestDiff"},
			changeset: testChangesetOpts{
				publicationState: campaigns.ChangesetPublicationStatePublished,
			},
			wantOperations: ReconcilerOperations{
				campaigns.ReconcilerOperationPush,
				campaigns.ReconcilerOperationSleep,
				campaigns.ReconcilerOperationSync,
			},
		},
		{
			name:         "commit message changed on published changeset",
			previousSpec: testSpecOpts{published: true, commitMessage: "old message"},
			currentSpec:  testSpecOpts{published: true, commitMessage: "new message"},
			changeset: testChangesetOpts{
				publicationState: campaigns.ChangesetPublicationStatePublished,
			},
			wantOperations: ReconcilerOperations{
				campaigns.ReconcilerOperationPush,
				campaigns.ReconcilerOperationSleep,
				campaigns.ReconcilerOperationSync,
			},
		},
		{
			name:         "commit diff changed on merge changeset",
			previousSpec: testSpecOpts{published: true, commitDiff: "testDiff"},
			currentSpec:  testSpecOpts{published: true, commitDiff: "newTestDiff"},
			changeset: testChangesetOpts{
				publicationState: campaigns.ChangesetPublicationStatePublished,
				externalState:    campaigns.ChangesetExternalStateMerged,
			},
			// should be a noop
			wantOperations: ReconcilerOperations{},
		},
		{
			name:         "changeset closed-and-detached will reopen",
			previousSpec: testSpecOpts{published: true},
			currentSpec:  testSpecOpts{published: true},
			changeset: testChangesetOpts{
				publicationState: campaigns.ChangesetPublicationStatePublished,
				externalState:    campaigns.ChangesetExternalStateClosed,
				ownedByCampaign:  1234,
				campaignIDs:      []int64{1234},
			},
			wantOperations: ReconcilerOperations{
				campaigns.ReconcilerOperationReopen,
			},
		},
		{
			name:         "closing",
			previousSpec: testSpecOpts{published: true},
			currentSpec:  testSpecOpts{published: true},
			changeset: testChangesetOpts{
				publicationState: campaigns.ChangesetPublicationStatePublished,
				externalState:    campaigns.ChangesetExternalStateOpen,
				ownedByCampaign:  1234,
				campaignIDs:      []int64{1234},
				// Important bit:
				closing: true,
			},
			wantOperations: ReconcilerOperations{
				campaigns.ReconcilerOperationClose,
			},
		},
		{
			name:         "closing already-closed changeset",
			previousSpec: testSpecOpts{published: true},
			currentSpec:  testSpecOpts{published: true},
			changeset: testChangesetOpts{
				publicationState: campaigns.ChangesetPublicationStatePublished,
				externalState:    campaigns.ChangesetExternalStateClosed,
				ownedByCampaign:  1234,
				campaignIDs:      []int64{1234},
				// Important bit:
				closing: true,
			},
			wantOperations: ReconcilerOperations{
				// TODO: This should probably be a noop in the future
				campaigns.ReconcilerOperationClose,
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			var previousSpec *campaigns.ChangesetSpec
			if tc.previousSpec != (testSpecOpts{}) {
				previousSpec = buildChangesetSpec(t, tc.previousSpec)
			}

			currentSpec := buildChangesetSpec(t, tc.currentSpec)
			cs := buildChangeset(tc.changeset)

			plan, err := DetermineReconcilerPlan(previousSpec, currentSpec, cs)
			if err != nil {
				t.Fatal(err)
			}
			if have, want := plan.Ops, tc.wantOperations; !have.Equal(want) {
				t.Fatalf("incorrect plan determined, want=%v have=%v", want, have)
			}
		})
	}
}

func buildGithubPR(now time.Time, externalState campaigns.ChangesetExternalState) *github.PullRequest {
	state := string(externalState)

	pr := &github.PullRequest{
		ID:          "12345",
		Number:      12345,
		Title:       state + " GitHub PR",
		Body:        state + " GitHub PR",
		State:       state,
		HeadRefName: git.AbbreviateRef("head-ref-on-github"),
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

	if externalState == campaigns.ChangesetExternalStateDraft {
		pr.State = "OPEN"
		pr.IsDraft = true
	}

	if externalState == campaigns.ChangesetExternalStateClosed {
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

func TestDecorateChangesetBody(t *testing.T) {
	db.Mocks.Namespaces.GetByID = func(ctx context.Context, org, user int32) (*db.Namespace, error) {
		return &db.Namespace{Name: "my-user", User: user}, nil
	}
	defer func() { db.Mocks.Namespaces.GetByID = nil }()

	internalClient = &mockInternalClient{externalURL: "https://sourcegraph.test"}
	defer func() { internalClient = api.InternalClient }()

	fs := &FakeStore{
		GetCampaignMock: func(ctx context.Context, opts GetCampaignOpts) (*campaigns.Campaign, error) {
			return &campaigns.Campaign{ID: 1234, Name: "reconciler-test-campaign"}, nil
		},
	}

	cs := buildChangeset(testChangesetOpts{ownedByCampaign: 1234})

	body := "body"
	rcs := &repos.Changeset{Body: body, Changeset: cs}
	if err := decorateChangesetBody(context.Background(), fs, rcs); err != nil {
		t.Errorf("unexpected non-nil error: %v", err)
	}
	if want := body + "\n\n[_Created by Sourcegraph campaign `my-user/reconciler-test-campaign`._](https://sourcegraph.test/users/my-user/campaigns/reconciler-test-campaign)"; rcs.Body != want {
		t.Errorf("repos.Changeset body unexpectedly changed:\nhave=%q\nwant=%q", rcs.Body, want)
	}
}

func TestCampaignURL(t *testing.T) {
	ctx := context.Background()

	t.Run("errors", func(t *testing.T) {
		for name, tc := range map[string]*mockInternalClient{
			"ExternalURL error": {err: errors.New("foo")},
			"invalid URL":       {externalURL: "foo://:bar"},
		} {
			t.Run(name, func(t *testing.T) {
				internalClient = tc
				defer func() { internalClient = api.InternalClient }()

				if _, err := campaignURL(ctx, nil, nil); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		internalClient = &mockInternalClient{externalURL: "https://sourcegraph.test"}
		defer func() { internalClient = api.InternalClient }()

		url, err := campaignURL(
			ctx,
			&db.Namespace{Name: "foo", Organization: 123},
			&campaigns.Campaign{Name: "bar"},
		)
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}
		if want := "https://sourcegraph.test/organizations/foo/campaigns/bar"; url != want {
			t.Errorf("unexpected URL: have=%q want=%q", url, want)
		}
	})
}

func TestNamespaceURL(t *testing.T) {
	t.Parallel()

	for name, tc := range map[string]struct {
		ns   *db.Namespace
		want string
	}{
		"user": {
			ns:   &db.Namespace{User: 123, Name: "user"},
			want: "/users/user",
		},
		"org": {
			ns:   &db.Namespace{Organization: 123, Name: "org"},
			want: "/organizations/org",
		},
		"neither": {
			ns:   &db.Namespace{Name: "user"},
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
