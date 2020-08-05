package campaigns

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/testing"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/schema"
)

func init() {
	dbtesting.DBNameSuffix = "campaignsenterpriserdb"
}

func TestServicePermissionLevels(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)

	store := NewStore(dbconn.Global)
	svc := NewService(store, nil)

	admin := createTestUser(ctx, t)
	if !admin.SiteAdmin {
		t.Fatalf("admin is not site admin")
	}

	user := createTestUser(ctx, t)
	if user.SiteAdmin {
		t.Fatalf("user cannot be site admin")
	}

	otherUser := createTestUser(ctx, t)
	if otherUser.SiteAdmin {
		t.Fatalf("user cannot be site admin")
	}

	rs, _ := createTestRepos(t, ctx, dbconn.Global, 1)

	createTestData := func(t *testing.T, s *Store, svc *Service, author int32) (*campaigns.Campaign, *campaigns.Changeset, *campaigns.CampaignSpec) {
		campaign := testCampaign(author)
		if err := s.CreateCampaign(ctx, campaign); err != nil {
			t.Fatal(err)
		}

		changeset := testChangeset(rs[0].ID, campaign.ID, campaigns.ChangesetExternalStateOpen)
		if err := s.CreateChangeset(ctx, changeset); err != nil {
			t.Fatal(err)
		}

		campaign.ChangesetIDs = append(campaign.ChangesetIDs, changeset.ID)
		if err := s.UpdateCampaign(ctx, campaign); err != nil {
			t.Fatal(err)
		}

		cs := &campaigns.CampaignSpec{UserID: author, NamespaceUserID: author}
		if err := s.CreateCampaignSpec(ctx, cs); err != nil {
			t.Fatal(err)
		}

		return campaign, changeset, cs
	}

	assertAuthError := func(t *testing.T, err error) {
		t.Helper()

		if err == nil {
			t.Fatalf("expected error. got none")
		}
		if err != nil {
			if _, ok := err.(*backend.InsufficientAuthorizationError); !ok {
				t.Fatalf("wrong error: %s (%T)", err, err)
			}
		}
	}

	assertNoAuthError := func(t *testing.T, err error) {
		t.Helper()

		// Ignore other errors, we only want to check whether it's an auth error
		if _, ok := err.(*backend.InsufficientAuthorizationError); ok {
			t.Fatalf("got auth error")
		}
	}

	tests := []struct {
		name           string
		campaignAuthor int32
		currentUser    int32
		assertFunc     func(t *testing.T, err error)
	}{
		{
			name:           "unauthorized user",
			campaignAuthor: user.ID,
			currentUser:    otherUser.ID,
			assertFunc:     assertAuthError,
		},
		{
			name:           "campaign author",
			campaignAuthor: user.ID,
			currentUser:    user.ID,
			assertFunc:     assertNoAuthError,
		},

		{
			name:           "site-admin",
			campaignAuthor: user.ID,
			currentUser:    admin.ID,
			assertFunc:     assertNoAuthError,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			campaign, changeset, campaignSpec := createTestData(t, store, svc, tc.campaignAuthor)
			// Fresh context.Background() because the previous one is wrapped in AuthzBypas
			currentUserCtx := actor.WithActor(context.Background(), actor.FromUser(tc.currentUser))

			t.Run("EnqueueChangesetSync", func(t *testing.T) {
				err := svc.EnqueueChangesetSync(currentUserCtx, changeset.ID)
				tc.assertFunc(t, err)
			})

			t.Run("CloseCampaign", func(t *testing.T) {
				_, err := svc.CloseCampaign(currentUserCtx, campaign.ID, false)
				tc.assertFunc(t, err)
			})

			t.Run("DeleteCampaign", func(t *testing.T) {
				err := svc.DeleteCampaign(currentUserCtx, campaign.ID)
				tc.assertFunc(t, err)
			})

			t.Run("MoveCampaign", func(t *testing.T) {
				_, err := svc.MoveCampaign(currentUserCtx, MoveCampaignOpts{
					CampaignID: campaign.ID,
					NewName:    "foobar2",
				})
				tc.assertFunc(t, err)
			})

			t.Run("ApplyCampaign", func(t *testing.T) {
				_, err := svc.ApplyCampaign(currentUserCtx, ApplyCampaignOpts{
					CampaignSpecRandID: campaignSpec.RandID,
				})
				tc.assertFunc(t, err)
			})
		})
	}
}

func TestService(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)

	now := time.Now().UTC().Truncate(time.Microsecond)
	clock := func() time.Time {
		return now.UTC().Truncate(time.Microsecond)
	}
	cf := httpcli.NewExternalHTTPClientFactory()

	admin := createTestUser(ctx, t)
	if !admin.SiteAdmin {
		t.Fatal("admin is not a site-admin")
	}

	user := createTestUser(ctx, t)
	if user.SiteAdmin {
		t.Fatal("user is admin, want non-admin")
	}

	store := NewStoreWithClock(dbconn.Global, clock)

	reposStore := repos.NewDBStore(dbconn.Global, sql.TxOptions{})

	ext := &repos.ExternalService{
		Kind:        extsvc.TypeGitHub,
		DisplayName: "GitHub",
		Config: marshalJSON(t, &schema.GitHubConnection{
			Url:   "https://github.com",
			Token: "SECRETTOKEN",
		}),
	}
	if err := reposStore.UpsertExternalServices(ctx, ext); err != nil {
		t.Fatal(err)
	}

	var rs []*repos.Repo
	for i := 0; i < 4; i++ {
		r := testRepo(i, extsvc.TypeGitHub)
		r.Sources = map[string]*repos.SourceInfo{ext.URN(): {ID: ext.URN()}}

		rs = append(rs, r)
	}

	awsCodeCommitRepoID := 4
	{
		r := testRepo(awsCodeCommitRepoID, extsvc.TypeAWSCodeCommit)
		r.Sources = map[string]*repos.SourceInfo{ext.URN(): {ID: ext.URN()}}
		rs = append(rs, r)
	}

	err := reposStore.UpsertRepos(ctx, rs...)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("CreateCampaign", func(t *testing.T) {
		campaign := testCampaign(admin.ID)
		svc := NewServiceWithClock(store, cf, clock)

		err = svc.CreateCampaign(ctx, campaign)
		if err != nil {
			t.Fatal(err)
		}

		_, err = store.GetCampaign(ctx, GetCampaignOpts{ID: campaign.ID})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("DeleteCampaign", func(t *testing.T) {
		campaign := testCampaign(admin.ID)

		svc := NewServiceWithClock(store, cf, clock)

		if err = svc.CreateCampaign(ctx, campaign); err != nil {
			t.Fatal(err)
		}
		if err := svc.DeleteCampaign(ctx, campaign.ID); err != nil {
			t.Fatalf("campaign not deleted: %s", err)
		}
	})

	t.Run("CloseCampaign", func(t *testing.T) {
		campaign := testCampaign(admin.ID)

		svc := NewServiceWithClock(store, cf, clock)

		if err = svc.CreateCampaign(ctx, campaign); err != nil {
			t.Fatal(err)
		}

		campaign, err = svc.CloseCampaign(ctx, campaign.ID, true)
		if err != nil {
			t.Fatalf("campaign not closed: %s", err)
		}
		if campaign.ClosedAt.IsZero() {
			t.Fatalf("campaign ClosedAt is zero")
		}
	})

	t.Run("EnqueueChangesetSync", func(t *testing.T) {
		svc := NewServiceWithClock(store, cf, clock)

		campaign := testCampaign(admin.ID)
		if err = store.CreateCampaign(ctx, campaign); err != nil {
			t.Fatal(err)
		}

		changeset := testChangeset(rs[0].ID, campaign.ID, campaigns.ChangesetExternalStateOpen)
		if err = store.CreateChangeset(ctx, changeset); err != nil {
			t.Fatal(err)
		}

		campaign.ChangesetIDs = []int64{changeset.ID}
		if err = store.UpdateCampaign(ctx, campaign); err != nil {
			t.Fatal(err)
		}

		called := false
		repoupdater.MockEnqueueChangesetSync = func(ctx context.Context, ids []int64) error {
			if len(ids) != 1 && ids[0] != changeset.ID {
				t.Fatalf("MockEnqueueChangesetSync received wrong ids: %+v", ids)
			}
			called = true
			return nil
		}
		t.Cleanup(func() { repoupdater.MockEnqueueChangesetSync = nil })

		if err := svc.EnqueueChangesetSync(ctx, changeset.ID); err != nil {
			t.Fatal(err)
		}

		if !called {
			t.Fatal("MockEnqueueChangesetSync not called")
		}

		// Repo filtered out by authzFilter
		ct.AuthzFilterRepos(t, rs[0].ID)

		// should result in a not found error
		if err := svc.EnqueueChangesetSync(ctx, changeset.ID); !errcode.IsNotFound(err) {
			t.Fatalf("expected not-found error but got %s", err)
		}
	})

	t.Run("CloseOpenChangesets", func(t *testing.T) {
		changeset1 := testChangeset(rs[0].ID, 0, campaigns.ChangesetExternalStateOpen)
		if err = store.CreateChangeset(ctx, changeset1); err != nil {
			t.Fatal(err)
		}
		changeset2 := testChangeset(rs[1].ID, 0, campaigns.ChangesetExternalStateOpen)
		if err = store.CreateChangeset(ctx, changeset2); err != nil {
			t.Fatal(err)
		}

		// Repo of changeset2 filtered out by authzFilter
		ct.AuthzFilterRepos(t, changeset2.RepoID)

		fakeSource := &ct.FakeChangesetSource{Err: nil}
		sourcer := repos.NewFakeSourcer(nil, fakeSource)

		svc := NewServiceWithClock(store, cf, clock)
		svc.sourcer = sourcer

		// Try to close open changesets
		err := svc.CloseOpenChangesets(ctx, []*campaigns.Changeset{changeset1, changeset2})
		if err != nil {
			t.Fatal(err)
		}

		// Only changeset1 should be closed
		if have, want := len(fakeSource.ClosedChangesets), 1; have != want {
			t.Fatalf("ClosedChangesets has wrong length. want=%d, have=%d", want, have)
		}

		if have, want := fakeSource.ClosedChangesets[0].RepoID, changeset1.RepoID; have != want {
			t.Fatalf("wrong changesets closed. want=%d, have=%d", want, have)
		}
	})

	t.Run("CreateCampaignSpec", func(t *testing.T) {
		svc := NewServiceWithClock(store, cf, clock)

		changesetSpecs := make([]*campaigns.ChangesetSpec, 0, len(rs))
		changesetSpecRandIDs := make([]string, 0, len(rs))
		for _, r := range rs {
			cs := &campaigns.ChangesetSpec{RepoID: r.ID, UserID: admin.ID}
			if err := store.CreateChangesetSpec(ctx, cs); err != nil {
				t.Fatal(err)
			}
			changesetSpecs = append(changesetSpecs, cs)
			changesetSpecRandIDs = append(changesetSpecRandIDs, cs.RandID)
		}

		adminCtx := actor.WithActor(context.Background(), actor.FromUser(admin.ID))

		t.Run("success", func(t *testing.T) {
			opts := CreateCampaignSpecOpts{
				NamespaceUserID:      admin.ID,
				RawSpec:              ct.TestRawCampaignSpec,
				ChangesetSpecRandIDs: changesetSpecRandIDs,
			}

			spec, err := svc.CreateCampaignSpec(adminCtx, opts)
			if err != nil {
				t.Fatal(err)
			}

			if spec.ID == 0 {
				t.Fatalf("CampaignSpec ID is 0")
			}

			if have, want := spec.UserID, admin.ID; have != want {
				t.Fatalf("UserID is %d, want %d", have, want)
			}

			var wantFields campaigns.CampaignSpecFields
			if err := json.Unmarshal([]byte(spec.RawSpec), &wantFields); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(wantFields, spec.Spec); diff != "" {
				t.Fatalf("wrong spec fields (-want +got):\n%s", diff)
			}

			for _, cs := range changesetSpecs {
				cs2, err := store.GetChangesetSpec(ctx, GetChangesetSpecOpts{ID: cs.ID})
				if err != nil {
					t.Fatal(err)
				}

				if have, want := cs2.CampaignSpecID, spec.ID; have != want {
					t.Fatalf("changesetSpec has wrong CampaignSpecID. want=%d, have=%d", want, have)
				}
			}
		})

		t.Run("success with YAML raw spec", func(t *testing.T) {
			opts := CreateCampaignSpecOpts{
				NamespaceUserID: admin.ID,
				RawSpec:         ct.TestRawCampaignSpecYAML,
			}

			spec, err := svc.CreateCampaignSpec(adminCtx, opts)
			if err != nil {
				t.Fatal(err)
			}

			if spec.ID == 0 {
				t.Fatalf("CampaignSpec ID is 0")
			}

			var wantFields campaigns.CampaignSpecFields
			if err := json.Unmarshal([]byte(ct.TestRawCampaignSpec), &wantFields); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(wantFields, spec.Spec); diff != "" {
				t.Fatalf("wrong spec fields (-want +got):\n%s", diff)
			}
		})

		t.Run("missing repository permissions", func(t *testing.T) {
			// Single repository filtered out by authzFilter
			ct.AuthzFilterRepos(t, changesetSpecs[0].RepoID)

			opts := CreateCampaignSpecOpts{
				NamespaceUserID:      admin.ID,
				RawSpec:              ct.TestRawCampaignSpec,
				ChangesetSpecRandIDs: changesetSpecRandIDs,
			}

			if _, err := svc.CreateCampaignSpec(adminCtx, opts); !errcode.IsNotFound(err) {
				t.Fatalf("expected not-found error but got %s", err)
			}
		})

		t.Run("invalid changesetspec id", func(t *testing.T) {
			containsInvalidID := []string{changesetSpecRandIDs[0], "foobar"}
			opts := CreateCampaignSpecOpts{
				NamespaceUserID:      admin.ID,
				RawSpec:              ct.TestRawCampaignSpec,
				ChangesetSpecRandIDs: containsInvalidID,
			}

			if _, err := svc.CreateCampaignSpec(adminCtx, opts); !errcode.IsNotFound(err) {
				t.Fatalf("expected not-found error but got %s", err)
			}
		})

		t.Run("namespace user is not admin and not creator", func(t *testing.T) {
			userCtx := actor.WithActor(context.Background(), actor.FromUser(user.ID))

			opts := CreateCampaignSpecOpts{
				NamespaceUserID: admin.ID,
				RawSpec:         ct.TestRawCampaignSpecYAML,
			}

			_, err := svc.CreateCampaignSpec(userCtx, opts)
			if !errcode.IsUnauthorized(err) {
				t.Fatalf("expected unauthorized error but got %s", err)
			}

			// Try again as admin
			adminCtx := actor.WithActor(context.Background(), actor.FromUser(admin.ID))

			opts.NamespaceUserID = user.ID

			_, err = svc.CreateCampaignSpec(adminCtx, opts)
			if err != nil {
				t.Fatalf("expected no error but got %s", err)
			}
		})

		t.Run("missing access to namespace org", func(t *testing.T) {
			org, err := db.Orgs.Create(ctx, "test-org", nil)
			if err != nil {
				t.Fatal(err)
			}

			opts := CreateCampaignSpecOpts{
				NamespaceOrgID:       org.ID,
				RawSpec:              ct.TestRawCampaignSpec,
				ChangesetSpecRandIDs: changesetSpecRandIDs,
			}

			userCtx := actor.WithActor(context.Background(), actor.FromUser(user.ID))

			_, err = svc.CreateCampaignSpec(userCtx, opts)
			if have, want := err, backend.ErrNotAnOrgMember; have != want {
				t.Fatalf("expected %s error but got %s", want, have)
			}

			// Create org membership and try again
			if _, err := db.OrgMembers.Create(ctx, org.ID, user.ID); err != nil {
				t.Fatal(err)
			}

			_, err = svc.CreateCampaignSpec(userCtx, opts)
			if err != nil {
				t.Fatalf("expected no error but got %s", err)
			}
		})
	})

	t.Run("CreateChangesetSpec", func(t *testing.T) {
		svc := NewServiceWithClock(store, cf, clock)

		repo := rs[0]
		rawSpec := ct.NewRawChangesetSpecGitBranch(graphqlbackend.MarshalRepositoryID(repo.ID), "d34db33f")

		t.Run("success", func(t *testing.T) {
			spec, err := svc.CreateChangesetSpec(ctx, rawSpec, admin.ID)
			if err != nil {
				t.Fatal(err)
			}

			if spec.ID == 0 {
				t.Fatalf("ChangesetSpec ID is 0")
			}

			wantFields := &campaigns.ChangesetSpecDescription{}
			if err := json.Unmarshal([]byte(spec.RawSpec), wantFields); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(wantFields, spec.Spec); diff != "" {
				t.Fatalf("wrong spec fields (-want +got):\n%s", diff)
			}

			wantDiffStat := diff.Stat{
				Added:   1,
				Changed: 2,
				Deleted: 1,
			}
			if diff := cmp.Diff(wantDiffStat, spec.DiffStat()); diff != "" {
				t.Fatalf("wrong diff stat (-want +got):\n%s", diff)
			}
		})

		t.Run("invalid raw spec", func(t *testing.T) {
			invalidRaw := `{"externalComputer": "beepboop"}`
			_, err := svc.CreateChangesetSpec(ctx, invalidRaw, admin.ID)
			if err == nil {
				t.Fatal("expected error but got nil")
			}

			haveErr := fmt.Sprintf("%v", err)
			wantErr := "4 errors occurred:\n\t* Must validate one and only one schema (oneOf)\n\t* baseRepository is required\n\t* externalID is required\n\t* Additional property externalComputer is not allowed\n\n"
			if diff := cmp.Diff(wantErr, haveErr); diff != "" {
				t.Fatalf("unexpected error (-want +got):\n%s", diff)
			}
		})

		t.Run("missing repository permissions", func(t *testing.T) {
			// Single repository filtered out by authzFilter
			ct.AuthzFilterRepos(t, repo.ID)

			_, err := svc.CreateChangesetSpec(ctx, rawSpec, admin.ID)
			if !errcode.IsNotFound(err) {
				t.Fatalf("expected not-found error but got %s", err)
			}
		})
	})

	t.Run("ApplyCampaign", func(t *testing.T) {
		// See TestServiceApplyCampaign
	})

	t.Run("MoveCampaign", func(t *testing.T) {
		svc := NewServiceWithClock(store, cf, clock)

		createCampaign := func(t *testing.T, name string, authorID, userID, orgID int32) *campaigns.Campaign {
			t.Helper()

			c := &campaigns.Campaign{
				AuthorID:        authorID,
				NamespaceUserID: userID,
				NamespaceOrgID:  orgID,
				Name:            name,
			}

			if err := store.CreateCampaign(ctx, c); err != nil {
				t.Fatal(err)
			}

			return c
		}

		t.Run("new name", func(t *testing.T) {
			campaign := createCampaign(t, "old-name", admin.ID, admin.ID, 0)

			opts := MoveCampaignOpts{CampaignID: campaign.ID, NewName: "new-name"}
			moved, err := svc.MoveCampaign(ctx, opts)
			if err != nil {
				t.Fatal(err)
			}

			if have, want := moved.Name, opts.NewName; have != want {
				t.Fatalf("wrong name. want=%q, have=%q", want, have)
			}
		})

		t.Run("new user namespace", func(t *testing.T) {
			campaign := createCampaign(t, "old-name", admin.ID, admin.ID, 0)

			user2 := createTestUser(ctx, t)

			opts := MoveCampaignOpts{CampaignID: campaign.ID, NewNamespaceUserID: user2.ID}
			moved, err := svc.MoveCampaign(ctx, opts)
			if err != nil {
				t.Fatal(err)
			}

			if have, want := moved.NamespaceUserID, opts.NewNamespaceUserID; have != want {
				t.Fatalf("wrong NamespaceUserID. want=%d, have=%d", want, have)
			}

			if have, want := moved.NamespaceOrgID, opts.NewNamespaceOrgID; have != want {
				t.Fatalf("wrong NamespaceOrgID. want=%d, have=%d", want, have)
			}
		})

		t.Run("new user namespace but current user is not admin", func(t *testing.T) {
			campaign := createCampaign(t, "old-name", user.ID, user.ID, 0)

			user2 := createTestUser(ctx, t)

			opts := MoveCampaignOpts{CampaignID: campaign.ID, NewNamespaceUserID: user2.ID}

			userCtx := actor.WithActor(context.Background(), actor.FromUser(user.ID))
			_, err := svc.MoveCampaign(userCtx, opts)
			if !errcode.IsUnauthorized(err) {
				t.Fatalf("expected unauthorized error but got %s", err)
			}
		})

		t.Run("new org namespace", func(t *testing.T) {
			campaign := createCampaign(t, "old-name", admin.ID, admin.ID, 0)

			org, err := db.Orgs.Create(ctx, "org", nil)
			if err != nil {
				t.Fatal(err)
			}

			opts := MoveCampaignOpts{CampaignID: campaign.ID, NewNamespaceOrgID: org.ID}
			moved, err := svc.MoveCampaign(ctx, opts)
			if err != nil {
				t.Fatal(err)
			}

			if have, want := moved.NamespaceUserID, opts.NewNamespaceUserID; have != want {
				t.Fatalf("wrong NamespaceUserID. want=%d, have=%d", want, have)
			}

			if have, want := moved.NamespaceOrgID, opts.NewNamespaceOrgID; have != want {
				t.Fatalf("wrong NamespaceOrgID. want=%d, have=%d", want, have)
			}
		})

		t.Run("new org namespace but current user is missing access", func(t *testing.T) {
			campaign := createCampaign(t, "old-name", user.ID, user.ID, 0)

			org, err := db.Orgs.Create(ctx, "org-no-access", nil)
			if err != nil {
				t.Fatal(err)
			}

			opts := MoveCampaignOpts{CampaignID: campaign.ID, NewNamespaceOrgID: org.ID}

			userCtx := actor.WithActor(context.Background(), actor.FromUser(user.ID))
			_, err = svc.MoveCampaign(userCtx, opts)
			if have, want := err, backend.ErrNotAnOrgMember; have != want {
				t.Fatalf("expected %s error but got %s", want, have)
			}
		})
	})
}

func TestServiceApplyCampaign(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := backend.WithAuthzBypass(context.Background())
	dbtesting.SetupGlobalTestDB(t)

	admin := createTestUser(ctx, t)
	if !admin.SiteAdmin {
		t.Fatal("admin is not a site-admin")
	}

	user := createTestUser(ctx, t)
	if user.SiteAdmin {
		t.Fatal("user is admin, want non-admin")
	}

	repos, _ := createTestRepos(t, ctx, dbconn.Global, 4)

	store := NewStore(dbconn.Global)
	svc := NewService(store, httpcli.NewExternalHTTPClientFactory())

	t.Run("campaignSpec without changesetSpecs", func(t *testing.T) {
		t.Run("new campaign", func(t *testing.T) {
			campaignSpec := createCampaignSpec(t, ctx, store, "campaign1", admin.ID)
			campaign, err := svc.ApplyCampaign(ctx, ApplyCampaignOpts{
				CampaignSpecRandID: campaignSpec.RandID,
			})
			if err != nil {
				t.Fatal(err)
			}

			if campaign.ID == 0 {
				t.Fatalf("campaign ID is 0")
			}

			want := &campaigns.Campaign{
				Name:            campaignSpec.Spec.Name,
				Description:     campaignSpec.Spec.Description,
				Branch:          campaignSpec.Spec.ChangesetTemplate.Branch,
				AuthorID:        campaignSpec.UserID,
				ChangesetIDs:    []int64{},
				NamespaceUserID: campaignSpec.NamespaceUserID,
				CampaignSpecID:  campaignSpec.ID,

				// Ignore these fields
				ID:        campaign.ID,
				UpdatedAt: campaign.UpdatedAt,
				CreatedAt: campaign.CreatedAt,
			}

			if diff := cmp.Diff(want, campaign); diff != "" {
				t.Fatalf("wrong spec fields (-want +got):\n%s", diff)
			}
		})

		t.Run("existing campaign", func(t *testing.T) {
			campaignSpec := createCampaignSpec(t, ctx, store, "campaign2", admin.ID)
			campaign := createCampaign(t, ctx, store, "campaign2", admin.ID, campaignSpec.ID)

			t.Run("apply same campaignSpec", func(t *testing.T) {
				campaign2, err := svc.ApplyCampaign(ctx, ApplyCampaignOpts{
					CampaignSpecRandID: campaignSpec.RandID,
				})
				if err != nil {
					t.Fatal(err)
				}

				if have, want := campaign2.ID, campaign.ID; have != want {
					t.Fatalf("campaign ID is wrong. want=%d, have=%d", want, have)
				}
			})

			t.Run("apply campaign spec with same name", func(t *testing.T) {
				campaignSpec2 := createCampaignSpec(t, ctx, store, "campaign2", admin.ID)
				campaign2, err := svc.ApplyCampaign(ctx, ApplyCampaignOpts{
					CampaignSpecRandID: campaignSpec2.RandID,
				})
				if err != nil {
					t.Fatal(err)
				}

				if have, want := campaign2.ID, campaign.ID; have != want {
					t.Fatalf("campaign ID is wrong. want=%d, have=%d", want, have)
				}
			})

			t.Run("apply campaign spec with same name but different namespace", func(t *testing.T) {
				user2 := createTestUser(ctx, t)
				campaignSpec2 := createCampaignSpec(t, ctx, store, "campaign2", user2.ID)

				campaign2, err := svc.ApplyCampaign(ctx, ApplyCampaignOpts{
					CampaignSpecRandID: campaignSpec2.RandID,
				})
				if err != nil {
					t.Fatal(err)
				}

				if campaign2.ID == 0 {
					t.Fatalf("campaign2 ID is 0")
				}

				if campaign2.ID == campaign.ID {
					t.Fatalf("campaign IDs are the same, but want different")
				}
			})

			t.Run("campaign spec with same name and same ensureCampaignID", func(t *testing.T) {
				campaignSpec2 := createCampaignSpec(t, ctx, store, "campaign2", admin.ID)

				campaign2, err := svc.ApplyCampaign(ctx, ApplyCampaignOpts{
					CampaignSpecRandID: campaignSpec2.RandID,
					EnsureCampaignID:   campaign.ID,
				})
				if err != nil {
					t.Fatal(err)
				}
				if have, want := campaign2.ID, campaign.ID; have != want {
					t.Fatalf("campaign has wrong ID. want=%d, have=%d", want, have)
				}
			})

			t.Run("campaign spec with same name but different ensureCampaignID", func(t *testing.T) {
				campaignSpec2 := createCampaignSpec(t, ctx, store, "campaign2", admin.ID)

				_, err := svc.ApplyCampaign(ctx, ApplyCampaignOpts{
					CampaignSpecRandID: campaignSpec2.RandID,
					EnsureCampaignID:   campaign.ID + 999,
				})
				if err != ErrEnsureCampaignFailed {
					t.Fatalf("wrong error: %s", err)
				}
			})
		})
	})

	// These tests focus on changesetSpecs and wiring them up with changesets.
	// The applying/re-applying of a campaignSpec to an existing campaign is
	// covered in the tests above.
	t.Run("campaignSpec with changesetSpecs", func(t *testing.T) {
		// We need to mock SyncChangesets because ApplyCampaign syncs
		// changesets. Once that moves to the background, we can remove this
		// mock.
		syncedBranchName := "refs/heads/synced-branch-name"
		MockSyncChangesets = func(_ context.Context, _ RepoStore, tx SyncStore, _ *httpcli.Factory, cs ...*campaigns.Changeset) error {
			for _, c := range cs {
				c.ExternalBranch = syncedBranchName
				if err := tx.UpdateChangeset(ctx, c); err != nil {
					return err
				}
			}
			return nil
		}
		t.Cleanup(func() { MockSyncChangesets = nil })

		t.Run("new campaign", func(t *testing.T) {
			campaignSpec := createCampaignSpec(t, ctx, store, "campaign3", admin.ID)

			spec1 := createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[0].ID,
				campaignSpec: campaignSpec.ID,
				externalID:   "1234",
			})

			spec2 := createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[1].ID,
				campaignSpec: campaignSpec.ID,
				headRef:      "refs/heads/my-branch",
			})

			campaign, cs := applyAndListChangesets(ctx, t, svc, campaignSpec.RandID, 2)

			if have, want := campaign.Name, "campaign3"; have != want {
				t.Fatalf("wrong campaign name. want=%s, have=%s", want, have)
			}

			c1 := cs.Find(campaigns.WithExternalID(spec1.Spec.ExternalID))
			assertChangeset(t, c1, changesetAssertions{
				repo:             spec1.RepoID,
				externalBranch:   syncedBranchName,
				externalID:       "1234",
				reconcilerState:  campaigns.ReconcilerStateCompleted,
				publicationState: campaigns.ChangesetPublicationStatePublished,
			})

			c2 := cs.Find(campaigns.WithCurrentSpecID(spec2.ID))
			assertChangeset(t, c2, changesetAssertions{
				repo:             spec2.RepoID,
				currentSpec:      spec2.ID,
				ownedByCampaign:  campaign.ID,
				reconcilerState:  campaigns.ReconcilerStateQueued,
				publicationState: campaigns.ChangesetPublicationStateUnpublished,
			})
		})

		t.Run("campaign with changesets", func(t *testing.T) {
			// First we create a campaignSpec and apply it, so that we have
			// changesets and changesetSpecs in the database, wired up
			// correctly.
			campaignSpec1 := createCampaignSpec(t, ctx, store, "campaign4", admin.ID)

			createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[0].ID,
				campaignSpec: campaignSpec1.ID,
				externalID:   "1234",
			})

			createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[0].ID,
				campaignSpec: campaignSpec1.ID,
				externalID:   "5678",
			})

			oldSpec3 := createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[1].ID,
				campaignSpec: campaignSpec1.ID,
				headRef:      "refs/heads/repo-1-branch-1",
			})

			oldSpec4 := createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[2].ID,
				campaignSpec: campaignSpec1.ID,
				headRef:      "refs/heads/repo-2-branch-1",
			})

			// Apply and expect 4 changesets
			_, oldChangesets := applyAndListChangesets(ctx, t, svc, campaignSpec1.RandID, 4)

			// Now we create another campaign spec with the same campaign name
			// and namespace.
			campaignSpec2 := createCampaignSpec(t, ctx, store, "campaign4", admin.ID)

			// Same
			spec1 := createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[0].ID,
				campaignSpec: campaignSpec2.ID,
				externalID:   "1234",
			})

			// DIFFERENT: Track #9999 in repo[0]
			spec2 := createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[0].ID,
				campaignSpec: campaignSpec2.ID,
				externalID:   "5678",
			})

			// Same
			spec3 := createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[1].ID,
				campaignSpec: campaignSpec2.ID,
				headRef:      "refs/heads/repo-1-branch-1",
			})

			// DIFFERENT: branch changed in repo[2]
			spec4 := createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[2].ID,
				campaignSpec: campaignSpec2.ID,
				headRef:      "refs/heads/repo-2-branch-2",
			})

			// NEW: repo[3]
			spec5 := createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[3].ID,
				campaignSpec: campaignSpec2.ID,
				headRef:      "refs/heads/repo-3-branch-1",
			})

			// Before we apply the new campaign spec, we set up the assertion
			// for changesets to be closed.
			wantClosed := oldChangesets.Find(campaigns.WithCurrentSpecID(oldSpec4.ID))
			// We need to make it look "published", otherwise it won't be closed.
			setChangesetPublished(t, ctx, store, wantClosed, oldSpec4.Spec.HeadRef, "98765")

			verifyClosed := assertChangesetsClose(t, wantClosed)

			// Apply and expect 5 changesets
			campaign, cs := applyAndListChangesets(ctx, t, svc, campaignSpec2.RandID, 5)

			verifyClosed()

			c1 := cs.Find(campaigns.WithExternalID(spec1.Spec.ExternalID))
			assertChangeset(t, c1, changesetAssertions{
				repo:             repos[0].ID,
				currentSpec:      0,
				previousSpec:     0,
				externalBranch:   syncedBranchName,
				externalID:       "1234",
				reconcilerState:  campaigns.ReconcilerStateCompleted,
				publicationState: campaigns.ChangesetPublicationStatePublished,
			})

			c2 := cs.Find(campaigns.WithExternalID(spec2.Spec.ExternalID))
			assertChangeset(t, c2, changesetAssertions{
				repo:             repos[0].ID,
				currentSpec:      0,
				previousSpec:     0,
				externalBranch:   syncedBranchName,
				externalID:       "5678",
				reconcilerState:  campaigns.ReconcilerStateCompleted,
				publicationState: campaigns.ChangesetPublicationStatePublished,
			})

			c3 := cs.Find(campaigns.WithCurrentSpecID(spec3.ID))
			assertChangeset(t, c3, changesetAssertions{
				repo:             repos[1].ID,
				currentSpec:      spec3.ID,
				previousSpec:     oldSpec3.ID,
				ownedByCampaign:  campaign.ID,
				reconcilerState:  campaigns.ReconcilerStateQueued,
				publicationState: campaigns.ChangesetPublicationStateUnpublished,
			})

			c4 := cs.Find(campaigns.WithCurrentSpecID(spec4.ID))
			assertChangeset(t, c4, changesetAssertions{
				repo:             repos[2].ID,
				currentSpec:      spec4.ID,
				ownedByCampaign:  campaign.ID,
				reconcilerState:  campaigns.ReconcilerStateQueued,
				publicationState: campaigns.ChangesetPublicationStateUnpublished,
			})

			c5 := cs.Find(campaigns.WithCurrentSpecID(spec5.ID))
			assertChangeset(t, c5, changesetAssertions{
				repo:             repos[3].ID,
				currentSpec:      spec5.ID,
				ownedByCampaign:  campaign.ID,
				reconcilerState:  campaigns.ReconcilerStateQueued,
				publicationState: campaigns.ChangesetPublicationStateUnpublished,
			})
		})

		t.Run("campaign tracking changesets owned by another campaign", func(t *testing.T) {
			campaignSpec1 := createCampaignSpec(t, ctx, store, "owner-campaign", admin.ID)

			oldSpec1 := createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[0].ID,
				campaignSpec: campaignSpec1.ID,
				headRef:      "refs/heads/repo-0-branch-0",
			})

			ownerCampaign, ownerChangesets := applyAndListChangesets(ctx, t, svc, campaignSpec1.RandID, 1)

			// Now we update the changeset so it looks like it's been published
			// on the code host.
			c := ownerChangesets[0]
			setChangesetPublished(t, ctx, store, c, "refs/heads/repo-0-branch-0", "88888")

			// This other campaign tracks the changeset created by the first one
			campaignSpec2 := createCampaignSpec(t, ctx, store, "tracking-campaign", admin.ID)
			createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         c.RepoID,
				campaignSpec: campaignSpec2.ID,
				externalID:   c.ExternalID,
			})

			_, trackedChangesets := applyAndListChangesets(ctx, t, svc, campaignSpec2.RandID, 1)
			// This should still point to the owner campaign
			c2 := trackedChangesets[0]
			assertChangeset(t, c2, changesetAssertions{
				repo:             c.RepoID,
				currentSpec:      oldSpec1.ID,
				ownedByCampaign:  ownerCampaign.ID,
				externalBranch:   c.ExternalBranch,
				externalID:       c.ExternalID,
				reconcilerState:  campaigns.ReconcilerStateCompleted,
				publicationState: campaigns.ChangesetPublicationStatePublished,
			})

			// Now we stop tracking it in the second campaign
			campaignSpec3 := createCampaignSpec(t, ctx, store, "tracking-campaign", admin.ID)

			// Campaign should have 0 changesets after applying, but the
			// tracked changeset should not be closed, since the campaign is
			// not the owner.
			//
			verifyClosed := assertChangesetsClose(t)
			applyAndListChangesets(ctx, t, svc, campaignSpec3.RandID, 0)
			verifyClosed()
		})

		t.Run("campaign with changeset that is unpublished", func(t *testing.T) {
			campaignSpec1 := createCampaignSpec(t, ctx, store, "unpublished-changesets", admin.ID)

			createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[3].ID,
				campaignSpec: campaignSpec1.ID,
				headRef:      "refs/heads/never-published",
			})

			// We apply the spec and expect 1 changeset
			_, changesets := applyAndListChangesets(ctx, t, svc, campaignSpec1.RandID, 1)

			// But the changeset was not published yet.
			// And now we apply a new spec without any changesets.
			campaignSpec2 := createCampaignSpec(t, ctx, store, "unpublished-changesets", admin.ID)

			// That should close no changesets, but leave the campaign with 0 changesets
			verifyClosed := assertChangesetsClose(t)
			applyAndListChangesets(ctx, t, svc, campaignSpec2.RandID, 0)
			verifyClosed()

			// And the unpublished changesets should be deleted
			toBeDeleted := changesets[0]
			_, err := store.GetChangeset(ctx, GetChangesetOpts{ID: toBeDeleted.ID})
			if err != ErrNoResults {
				t.Fatalf("expected changeset to be deleted but was not")
			}
		})

		t.Run("missing repository permissions", func(t *testing.T) {
			// Single repository filtered out by authzFilter
			ct.AuthzFilterRepos(t, repos[1].ID)

			campaignSpec := createCampaignSpec(t, ctx, store, "missing-permissions", admin.ID)

			createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[0].ID,
				campaignSpec: campaignSpec.ID,
				externalID:   "1234",
			})

			createChangesetSpec(t, ctx, store, testSpecOpts{
				user:         admin.ID,
				repo:         repos[1].ID, // Filtered out by authzFilter
				campaignSpec: campaignSpec.ID,
				headRef:      "refs/heads/my-branch",
			})

			_, err := svc.ApplyCampaign(ctx, ApplyCampaignOpts{
				CampaignSpecRandID: campaignSpec.RandID,
			})
			if err == nil {
				t.Fatal("expected error, but got none")
			}
			notFoundErr, ok := err.(*db.RepoNotFoundErr)
			if !ok {
				t.Fatalf("expected RepoNotFoundErr but got: %s", err)
			}
			if notFoundErr.ID != repos[1].ID {
				t.Fatalf("wrong repository ID in RepoNotFoundErr: %d", notFoundErr.ID)
			}
		})
	})
}

var testUser = db.NewUser{
	Email:                 "thorsten@sourcegraph.com",
	Username:              "thorsten",
	DisplayName:           "thorsten",
	Password:              "1234",
	EmailVerificationCode: "foobar",
}

var createTestUser = func() func(context.Context, *testing.T) *types.User {
	count := 0

	return func(ctx context.Context, t *testing.T) *types.User {
		t.Helper()

		u := testUser
		u.Username = fmt.Sprintf("%s-%d", u.Username, count)

		user, err := db.Users.Create(ctx, u)
		if err != nil {
			t.Fatal(err)
		}

		count += 1

		return user
	}
}()

func testCampaign(user int32) *campaigns.Campaign {
	c := &campaigns.Campaign{
		Name:            "test-campaign",
		AuthorID:        user,
		NamespaceUserID: user,
	}

	return c
}

func testChangeset(repoID api.RepoID, campaign int64, extState campaigns.ChangesetExternalState) *campaigns.Changeset {
	changeset := &campaigns.Changeset{
		RepoID:              repoID,
		ExternalServiceType: extsvc.TypeGitHub,
		ExternalID:          fmt.Sprintf("ext-id-%d", campaign),
		Metadata:            &github.PullRequest{State: string(extState)},
		ExternalState:       extState,
	}

	if campaign != 0 {
		changeset.CampaignIDs = []int64{campaign}
	}

	return changeset
}

func createCampaign(t *testing.T, ctx context.Context, store *Store, name string, userID int32, spec int64) *campaigns.Campaign {
	t.Helper()

	c := &campaigns.Campaign{
		AuthorID:        userID,
		NamespaceUserID: userID,
		CampaignSpecID:  spec,
		Name:            name,
		Description:     "campaign description",
	}

	if err := store.CreateCampaign(ctx, c); err != nil {
		t.Fatal(err)
	}

	return c
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
	published bool

	title         string
	body          string
	commitMessage string
	commitDiff    string
}

func createChangesetSpec(
	t *testing.T,
	ctx context.Context,
	store *Store,
	opts testSpecOpts,
) *campaigns.ChangesetSpec {
	t.Helper()

	spec := &campaigns.ChangesetSpec{
		UserID:         opts.user,
		RepoID:         opts.repo,
		CampaignSpecID: opts.campaignSpec,
		Spec: &campaigns.ChangesetSpecDescription{
			BaseRepository: graphqlbackend.MarshalRepositoryID(opts.repo),

			ExternalID: opts.externalID,
			HeadRef:    opts.headRef,
			Published:  opts.published,

			Title: opts.title,
			Body:  opts.body,

			Commits: []campaigns.GitCommitDescription{
				{
					Message: opts.commitMessage,
					Diff:    opts.commitDiff,
				},
			},
		},
	}

	if err := store.CreateChangesetSpec(ctx, spec); err != nil {
		t.Fatal(err)
	}

	return spec
}

func createTestRepos(t *testing.T, ctx context.Context, db *sql.DB, count int) ([]*repos.Repo, *repos.ExternalService) {
	t.Helper()

	rstore := repos.NewDBStore(db, sql.TxOptions{})

	ext := &repos.ExternalService{
		Kind:        extsvc.TypeGitHub,
		DisplayName: "GitHub",
		Config: marshalJSON(t, &schema.GitHubConnection{
			Url:   "https://github.com",
			Token: "SECRETTOKEN",
		}),
	}
	if err := rstore.UpsertExternalServices(ctx, ext); err != nil {
		t.Fatal(err)
	}

	var rs []*repos.Repo
	for i := 0; i < count; i++ {
		r := testRepo(i, extsvc.TypeGitHub)
		r.Sources = map[string]*repos.SourceInfo{ext.URN(): {ID: ext.URN()}}

		rs = append(rs, r)
	}

	err := rstore.UpsertRepos(ctx, rs...)
	if err != nil {
		t.Fatal(err)
	}

	return rs, ext
}

type testChangesetOpts struct {
	repo         api.RepoID
	campaign     int64
	currentSpec  int64
	previousSpec int64

	externalServiceType string
	externalID          string
	externalBranch      string

	publicationState campaigns.ChangesetPublicationState
	failureMessage   string

	createdByCampaign bool
	ownedByCampaign   int64
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

		PublicationState: opts.publicationState,

		CreatedByCampaign: opts.createdByCampaign,
		OwnedByCampaignID: opts.ownedByCampaign,
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

type changesetAssertions struct {
	repo             api.RepoID
	currentSpec      int64
	previousSpec     int64
	ownedByCampaign  int64
	reconcilerState  campaigns.ReconcilerState
	publicationState campaigns.ChangesetPublicationState
	externalID       string
	externalBranch   string

	title string
	body  string

	failureMessage *string
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

	if have, want := c.ExternalID, a.externalID; have != want {
		t.Fatalf("changeset ExternalID wrong. want=%s, have=%s", want, have)
	}

	if have, want := c.ExternalBranch, a.externalBranch; have != want {
		t.Fatalf("changeset ExternalBranch wrong. want=%s, have=%s", want, have)
	}

	if want, have := a.failureMessage, c.FailureMessage; want == nil && have != nil {
		t.Fatalf("expected no failure message, but have=%q", *have)
	}

	if want := c.FailureMessage; want != nil {
		if c.FailureMessage == nil {
			t.Fatalf("expected failure message %q but have none", *want)
		}
		if want, have := *a.failureMessage, *c.FailureMessage; have != want {
			t.Fatalf("wrong failure message. want=%q, have=%q", want, have)
		}
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

func applyAndListChangesets(ctx context.Context, t *testing.T, svc *Service, campaignSpecRandID string, wantChangesets int) (*campaigns.Campaign, campaigns.Changesets) {
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

func assertChangesetsClose(t *testing.T, want ...*campaigns.Changeset) (verify func()) {
	t.Helper()

	closedCalled := false

	mockApplyCampaignCloseChangesets = func(toClose campaigns.Changesets) {
		closedCalled = true
		if have, want := len(toClose), len(want); have != want {
			t.Fatalf("closing wrong number of changesets. want=%d, have=%d", want, have)
		}
		closedByID := map[int64]bool{}
		for _, c := range toClose {
			closedByID[c.ID] = true
		}
		for _, c := range want {
			if _, ok := closedByID[c.ID]; !ok {
				t.Fatalf("expected changeset %d to be closed but was not", c.ID)
			}
		}
	}

	t.Cleanup(func() { mockApplyCampaignCloseChangesets = nil })

	verify = func() {
		if !closedCalled {
			t.Fatalf("expected CloseOpenChangesets to be called but was not")
		}
	}

	return verify
}

func setChangesetPublished(t *testing.T, ctx context.Context, s *Store, c *campaigns.Changeset, externalID, externalBranch string) {
	t.Helper()

	c.ExternalBranch = externalBranch
	c.ExternalID = externalID
	c.PublicationState = campaigns.ChangesetPublicationStatePublished
	c.ReconcilerState = campaigns.ReconcilerStateCompleted

	if err := s.UpdateChangeset(ctx, c); err != nil {
		t.Fatalf("failed to update changeset: %s", err)
	}
}
