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

	var rs []*repos.Repo
	for i := 0; i < 4; i++ {
		rs = append(rs, testRepo(i, extsvc.TypeGitHub))
	}

	reposStore := repos.NewDBStore(dbconn.Global, sql.TxOptions{})
	err := reposStore.UpsertRepos(ctx, rs...)
	if err != nil {
		t.Fatal(err)
	}

	createTestData := func(t *testing.T, s *Store, svc *Service, author int32) (*campaigns.Campaign, *campaigns.Changeset, *campaigns.CampaignSpec) {
		campaign := testCampaign(author)
		if err = s.CreateCampaign(ctx, campaign); err != nil {
			t.Fatal(err)
		}

		changeset := testChangeset(rs[0].ID, campaign.ID, campaign.ID, campaigns.ChangesetExternalStateOpen)
		if err = s.CreateChangesets(ctx, changeset); err != nil {
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
				err = svc.EnqueueChangesetSync(currentUserCtx, changeset.ID)
				tc.assertFunc(t, err)
			})

			t.Run("CloseCampaign", func(t *testing.T) {
				_, err = svc.CloseCampaign(currentUserCtx, campaign.ID, false)
				tc.assertFunc(t, err)
			})

			t.Run("DeleteCampaign", func(t *testing.T) {
				err = svc.DeleteCampaign(currentUserCtx, campaign.ID)
				tc.assertFunc(t, err)
			})

			t.Run("MoveCampaign", func(t *testing.T) {
				_, err = svc.MoveCampaign(currentUserCtx, MoveCampaignOpts{
					CampaignID: campaign.ID,
					NewName:    "foobar2",
				})
				tc.assertFunc(t, err)
			})

			t.Run("ApplyCampaign", func(t *testing.T) {
				_, err = svc.ApplyCampaign(currentUserCtx, ApplyCampaignOpts{
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

		changeset := testChangeset(rs[0].ID, campaign.ID, 0, campaigns.ChangesetExternalStateOpen)
		if err = store.CreateChangesets(ctx, changeset); err != nil {
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
		changeset1 := testChangeset(rs[0].ID, 0, 121314, campaigns.ChangesetExternalStateOpen)
		changeset2 := testChangeset(rs[1].ID, 0, 141516, campaigns.ChangesetExternalStateOpen)
		if err = store.CreateChangesets(ctx, changeset1, changeset2); err != nil {
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

			var wantFields campaigns.ChangesetSpecDescription
			if err := json.Unmarshal([]byte(spec.RawSpec), &wantFields); err != nil {
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
		svc := NewServiceWithClock(store, cf, clock)

		createCampaignSpec := func(t *testing.T, name string, userID int32) *campaigns.CampaignSpec {
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

		t.Run("new campaign", func(t *testing.T) {
			campaignSpec := createCampaignSpec(t, "campaign-name", admin.ID)
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
			campaignSpec := createCampaignSpec(t, "campaign-name", admin.ID)
			campaign, err := svc.ApplyCampaign(ctx, ApplyCampaignOpts{
				CampaignSpecRandID: campaignSpec.RandID,
			})
			if err != nil {
				t.Fatal(err)
			}

			if campaign.ID == 0 {
				t.Fatalf("campaign ID is 0")
			}

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
				campaignSpec2 := createCampaignSpec(t, "campaign-name", admin.ID)
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
				campaignSpec2 := createCampaignSpec(t, "campaign-name", user2.ID)

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
				campaignSpec2 := createCampaignSpec(t, "campaign-name", admin.ID)

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
				campaignSpec2 := createCampaignSpec(t, "campaign-name", admin.ID)

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
		Name:            "Testing Campaign",
		Description:     "Testing Campaign",
		AuthorID:        user,
		NamespaceUserID: user,
	}

	return c
}

func testChangeset(repoID api.RepoID, campaign int64, changesetJob int64, extState campaigns.ChangesetExternalState) *campaigns.Changeset {
	changeset := &campaigns.Changeset{
		RepoID:              repoID,
		ExternalServiceType: extsvc.TypeGitHub,
		ExternalID:          fmt.Sprintf("ext-id-%d", changesetJob),
		Metadata:            &github.PullRequest{State: string(extState)},
		ExternalState:       extState,
	}

	if campaign != 0 {
		changeset.CampaignIDs = []int64{campaign}
	}

	return changeset
}
