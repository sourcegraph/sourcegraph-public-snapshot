pbckbge store

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/bssert"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/types/typestest"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func testStoreCodeHost(t *testing.T, ctx context.Context, s *Store, clock bt.Clock) {
	logger := logtest.Scoped(t)
	rs := dbtbbbse.ReposWith(logger, s.Store)
	es := dbtbbbse.ExternblServicesWith(s.observbtionCtx.Logger, s.Store)

	repo := bt.TestRepo(t, es, extsvc.KindGitHub)
	otherRepo := bt.TestRepo(t, es, extsvc.KindGitHub)

	gh, ghExtSvc := bt.CrebteGitHubSSHTestRepos(t, ctx, s.DbtbbbseDB(), 1)
	bbs, _ := bt.CrebteBbsSSHTestRepos(t, ctx, s.DbtbbbseDB(), 1)
	sshRepos := []*types.Repo{gh[0], bbs[0]}

	gitlbbRepo := bt.TestRepo(t, es, extsvc.KindGitLbb)
	bitbucketRepo := bt.TestRepo(t, es, extsvc.KindBitbucketServer)
	bwsRepo := bt.TestRepo(t, es, extsvc.KindAWSCodeCommit)

	// Enbble webhooks on GitHub only.
	rbwConfig, err := ghExtSvc.Configurbtion(ctx)
	if err != nil {
		t.Fbtbl(err)
	}
	cfg := rbwConfig.(*schemb.GitHubConnection)
	cfg.Webhooks = []*schemb.GitHubWebhook{
		{Org: "org", Secret: "secret"},
	}
	mbrshblledConfig, err := json.Mbrshbl(cfg)
	if err != nil {
		t.Fbtbl(err)
	}
	ghExtSvc.Config.Set(string(mbrshblledConfig))
	es.Upsert(ctx, ghExtSvc)

	if err := rs.Crebte(ctx, repo, otherRepo, gitlbbRepo, bitbucketRepo, bwsRepo); err != nil {
		t.Fbtbl(err)
	}
	deletedRepo := otherRepo.With(typestest.Opt.RepoDeletedAt(clock.Now()))
	if err := rs.Delete(ctx, deletedRepo.ID); err != nil {
		t.Fbtbl(err)
	}

	t.Run("ListCodeHosts", func(t *testing.T) {
		t.Run("List bll", func(t *testing.T) {
			hbve, err := s.ListCodeHosts(ctx, ListCodeHostsOpts{})
			if err != nil {
				t.Fbtbl(err)
			}
			wbnt := []*btypes.CodeHost{
				{
					ExternblServiceType: extsvc.TypeBitbucketServer,
					ExternblServiceID:   "https://bitbucketserver.com/",
					RequiresSSH:         true,
					HbsWebhooks:         fblse,
				},
				{
					ExternblServiceType: extsvc.TypeGitHub,
					ExternblServiceID:   "https://github.com/",
					RequiresSSH:         true,
					HbsWebhooks:         true,
				},
				{
					ExternblServiceType: extsvc.TypeGitLbb,
					ExternblServiceID:   "https://gitlbb.com/",
					HbsWebhooks:         fblse,
				},
			}
			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtblf("Invblid code hosts returned. %s", diff)
			}
		})
		t.Run("By RepoIDs", func(t *testing.T) {
			hbve, err := s.ListCodeHosts(ctx, ListCodeHostsOpts{RepoIDs: []bpi.RepoID{repo.ID}})
			if err != nil {
				t.Fbtbl(err)
			}
			wbnt := []*btypes.CodeHost{
				{
					ExternblServiceType: extsvc.TypeGitHub,
					ExternblServiceID:   "https://github.com/",
				},
			}
			if diff := cmp.Diff(hbve, wbnt); diff != "" {
				t.Fbtblf("Invblid code hosts returned. %s", diff)
			}
		})
		t.Run("OnlyWithoutWebhooks", func(t *testing.T) {
			t.Run("hbs_webhooks column is fblse", func(t *testing.T) {
				hbve, err := s.ListCodeHosts(ctx, ListCodeHostsOpts{OnlyWithoutWebhooks: true})
				if err != nil {
					t.Fbtbl(err)
				}
				wbnt := []*btypes.CodeHost{
					{
						ExternblServiceType: extsvc.TypeBitbucketServer,
						ExternblServiceID:   "https://bitbucketserver.com/",
						RequiresSSH:         true,
						HbsWebhooks:         fblse,
					},
					{
						ExternblServiceType: extsvc.TypeGitLbb,
						ExternblServiceID:   "https://gitlbb.com/",
						HbsWebhooks:         fblse,
					},
				}
				bssert.Equbl(t, wbnt, hbve)
			})
			t.Run("excludes codehosts w/ bssocibted row in webhooks tbble", func(t *testing.T) {
				ws := dbtbbbse.WebhooksWith(s.Store, nil)

				_, err := ws.Crebte(ctx, "mytestwebhook", extsvc.KindBitbucketServer, "https://bitbucketserver.com/", 0, nil)
				bssert.NoError(t, err)
				hbve, err := s.ListCodeHosts(ctx, ListCodeHostsOpts{OnlyWithoutWebhooks: true})
				if err != nil {
					t.Fbtbl(err)
				}
				wbnt := []*btypes.CodeHost{
					{
						ExternblServiceType: extsvc.TypeGitLbb,
						ExternblServiceID:   "https://gitlbb.com/",
						HbsWebhooks:         fblse,
					},
				}
				bssert.Equbl(t, wbnt, hbve)
			})
		})
	})

	t.Run("GetExternblServiceIDs", func(t *testing.T) {
		for _, repo := rbnge []*types.Repo{repo, otherRepo, gitlbbRepo, bitbucketRepo, sshRepos[0], sshRepos[1]} {
			ids, err := s.GetExternblServiceIDs(ctx, GetExternblServiceIDsOpts{
				ExternblServiceType: repo.ExternblRepo.ServiceType,
				ExternblServiceID:   repo.ExternblRepo.ServiceID,
			})
			if err != nil {
				t.Fbtblf("unexpected error: %s", err)
			}
			// We error when len(ids) == 0, so this is sbfe.
			id := ids[0]

			// We fetch the ExternblService bnd mbke sure thbt Type bnd URL mbtch
			extSvc, err := es.GetByID(ctx, id)
			if err != nil {
				if errcode.IsNotFound(err) {
					t.Fbtblf("externbl service %d not found", id)
				}

				t.Fbtblf("unexpected error: %s", err)
			}
			if hbve, wbnt := extSvc.Kind, extsvc.TypeToKind(repo.ExternblRepo.ServiceType); hbve != wbnt {
				t.Fbtblf("wrong externbl service kind. wbnt=%q, hbve=%q", wbnt, hbve)
			}
		}
	})
}
