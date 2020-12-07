package campaigns

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ct "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/testing"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	gitprotocol "github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestExecutor_ExecutePlan_PublishedChangesetDuplicateBranch(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()
	dbtesting.SetupGlobalTestDB(t)

	store := NewStore(dbconn.Global)

	rs, _ := ct.CreateTestRepos(t, ctx, dbconn.Global, 1)

	commonHeadRef := "refs/heads/collision"

	// Create a published changeset.
	createChangeset(t, ctx, store, testChangesetOpts{
		repo:             rs[0].ID,
		publicationState: campaigns.ChangesetPublicationStatePublished,
		externalBranch:   commonHeadRef,
		externalID:       "123",
	})

	// Build a changeset that would be pushed on the same HeadRef/ExternalBranch.
	spec := buildChangesetSpec(t, testSpecOpts{
		repo:      rs[0].ID,
		headRef:   commonHeadRef,
		published: true,
	})
	changeset := buildChangeset(testChangesetOpts{repo: rs[0].ID})

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

	store := NewStore(dbconn.Global)

	admin := createTestUser(t, true)
	user := createTestUser(t, false)

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

	store := NewStore(dbconn.Global)

	admin := createTestUser(t, true)
	user := createTestUser(t, false)

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
		repo           *repos.Repo
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
				RemoteURL: "https://my-secret-github-token@github.com/" + gitHubRepo.Name,
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
				RemoteURL: "https://git:my-secret-gitlab-token@gitlab.com/" + gitLabRepo.Name,
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
				RemoteURL: "https://fredwoard%20johnssen:my-secret-bbs-token@bitbucket.sourcegraph.com/scm/" + bbsRepo.Name,
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
				spec: buildChangesetSpec(t, testSpecOpts{
					headRef:    "refs/heads/my-branch",
					published:  true,
					commitDiff: "testdiff",
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
