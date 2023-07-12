package store

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	bt "github.com/sourcegraph/sourcegraph/internal/batches/testing"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/types/typestest"
	"github.com/sourcegraph/sourcegraph/schema"
)

func testStoreCodeHost(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
	logger := logtest.Scoped(t)
	rs := database.ReposWith(logger, s.Store)
	es := database.ExternalServicesWith(s.observationCtx.Logger, s.Store)

	repo := bt.TestRepo(t, es, extsvc.KindGitHub)
	otherRepo := bt.TestRepo(t, es, extsvc.KindGitHub)

	gh, ghExtSvc := bt.CreateGitHubSSHTestRepos(t, ctx, s.DatabaseDB(), 1)
	bbs, _ := bt.CreateBbsSSHTestRepos(t, ctx, s.DatabaseDB(), 1)
	sshRepos := []*types.Repo{gh[0], bbs[0]}

	gitlabRepo := bt.TestRepo(t, es, extsvc.KindGitLab)
	bitbucketRepo := bt.TestRepo(t, es, extsvc.KindBitbucketServer)
	awsRepo := bt.TestRepo(t, es, extsvc.KindAWSCodeCommit)

	// Enable webhooks on GitHub only.
	rawConfig, err := ghExtSvc.Configuration(ctx)
	if err != nil {
		t.Fatal(err)
	}
	cfg := rawConfig.(*schema.GitHubConnection)
	cfg.Webhooks = []*schema.GitHubWebhook{
		{Org: "org", Secret: "secret"},
	}
	marshalledConfig, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ghExtSvc.Config.Set(string(marshalledConfig))
	es.Upsert(ctx, ghExtSvc)

	if err := rs.Create(ctx, repo, otherRepo, gitlabRepo, bitbucketRepo, awsRepo); err != nil {
		t.Fatal(err)
	}
	deletedRepo := otherRepo.With(typestest.Opt.RepoDeletedAt(clock.Now()))
	if err := rs.Delete(ctx, deletedRepo.ID); err != nil {
		t.Fatal(err)
	}

	t.Run("ListCodeHosts", func(t *testing.T) {
		t.Run("List all", func(t *testing.T) {
			have, err := s.ListCodeHosts(ctx, ListCodeHostsOpts{})
			if err != nil {
				t.Fatal(err)
			}
			want := []*btypes.CodeHost{
				{
					ExternalServiceType: extsvc.TypeBitbucketServer,
					ExternalServiceID:   "https://bitbucketserver.com/",
					RequiresSSH:         true,
					HasWebhooks:         false,
				},
				{
					ExternalServiceType: extsvc.TypeGitHub,
					ExternalServiceID:   "https://github.com/",
					RequiresSSH:         true,
					HasWebhooks:         true,
				},
				{
					ExternalServiceType: extsvc.TypeGitLab,
					ExternalServiceID:   "https://gitlab.com/",
					HasWebhooks:         false,
				},
			}
			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatalf("Invalid code hosts returned. %s", diff)
			}
		})
		t.Run("By RepoIDs", func(t *testing.T) {
			have, err := s.ListCodeHosts(ctx, ListCodeHostsOpts{RepoIDs: []api.RepoID{repo.ID}})
			if err != nil {
				t.Fatal(err)
			}
			want := []*btypes.CodeHost{
				{
					ExternalServiceType: extsvc.TypeGitHub,
					ExternalServiceID:   "https://github.com/",
				},
			}
			if diff := cmp.Diff(have, want); diff != "" {
				t.Fatalf("Invalid code hosts returned. %s", diff)
			}
		})
		t.Run("OnlyWithoutWebhooks", func(t *testing.T) {
			t.Run("has_webhooks column is false", func(t *testing.T) {
				have, err := s.ListCodeHosts(ctx, ListCodeHostsOpts{OnlyWithoutWebhooks: true})
				if err != nil {
					t.Fatal(err)
				}
				want := []*btypes.CodeHost{
					{
						ExternalServiceType: extsvc.TypeBitbucketServer,
						ExternalServiceID:   "https://bitbucketserver.com/",
						RequiresSSH:         true,
						HasWebhooks:         false,
					},
					{
						ExternalServiceType: extsvc.TypeGitLab,
						ExternalServiceID:   "https://gitlab.com/",
						HasWebhooks:         false,
					},
				}
				assert.Equal(t, want, have)
			})
			t.Run("excludes codehosts w/ associated row in webhooks table", func(t *testing.T) {
				ws := database.WebhooksWith(s.Store, nil)

				_, err := ws.Create(ctx, "mytestwebhook", extsvc.KindBitbucketServer, "https://bitbucketserver.com/", 0, nil)
				assert.NoError(t, err)
				have, err := s.ListCodeHosts(ctx, ListCodeHostsOpts{OnlyWithoutWebhooks: true})
				if err != nil {
					t.Fatal(err)
				}
				want := []*btypes.CodeHost{
					{
						ExternalServiceType: extsvc.TypeGitLab,
						ExternalServiceID:   "https://gitlab.com/",
						HasWebhooks:         false,
					},
				}
				assert.Equal(t, want, have)
			})
		})
	})

	t.Run("GetExternalServiceIDs", func(t *testing.T) {
		for _, repo := range []*types.Repo{repo, otherRepo, gitlabRepo, bitbucketRepo, sshRepos[0], sshRepos[1]} {
			ids, err := s.GetExternalServiceIDs(ctx, GetExternalServiceIDsOpts{
				ExternalServiceType: repo.ExternalRepo.ServiceType,
				ExternalServiceID:   repo.ExternalRepo.ServiceID,
			})
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			// We error when len(ids) == 0, so this is safe.
			id := ids[0]

			// We fetch the ExternalService and make sure that Type and URL match
			extSvc, err := es.GetByID(ctx, id)
			if err != nil {
				if errcode.IsNotFound(err) {
					t.Fatalf("external service %d not found", id)
				}

				t.Fatalf("unexpected error: %s", err)
			}
			if have, want := extSvc.Kind, extsvc.TypeToKind(repo.ExternalRepo.ServiceType); have != want {
				t.Fatalf("wrong external service kind. want=%q, have=%q", want, have)
			}
		}
	})
}
