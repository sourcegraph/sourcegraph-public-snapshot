pbckbge sources

import (
	"context"
	"sort"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/bssert"

	mockbssert "github.com/derision-test/go-mockgen/testutil/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	ghbbuth "github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/buth"
	ghbstore "github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/store"
	ghbtypes "github.com/sourcegrbph/sourcegrbph/internbl/github_bpps/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

// Rbndom vblid privbte key generbted for this test bnd nothing else
const testGHAppPrivbteKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA1LqMqnchtoTHiRfFds2RWWji43R5mHT65WZpZXpBuBwsgSWr
rN5VwTHZ4dxWk+XlyVDYsri7vlVWNX4EIt0jwxvh/OBXCFJXTL+byNHCimRIKvur
ofoT1eF3z+5WpH5ddHNPOkGZV0Chyd5kvUcNuFA7q203HRVEOloHEs4fqJrGPHIF
zc8Sug5qkOtZTS5xiHgTtmZkDLuLZ26H5Gfx3zZk5Gv2Jy/+fLsGninbiwTRsZf6
RgPgmdlkuM8OhSm4GtlpzoK0D3iZEhf4pITo1CK2U4Cs7vzkU0UkQ+J/z6dDmVBJ
gkblH1SHsboRqjNxkStEqGnbWwdtbl01skbGOwIDAQABAoIBAQCls54oll17V5g5
0Htu3BdxBsNdG3gv6kcY85n7gqy4ZbHA83/zSsiPkW4/gbsqzzQbiU8Sf9U2IDDj
wAImygy2SPzSRklk4QbBcKs/VSztMcoJOTprFGno+xShsexpe0j+kWdQYJK6JU0g
+ouL6FHmlRC1qn/4tn0L2t6Rpl+Aq4peDLqdwFHXj8GxGv0S10qMQ4/ER7onP6f0
99WDTvNQR5DugKqHxooOV5HfUP70scqhCcFhp2zc7/bYQFVt/k4hDOMu/w4HhkD3
r34y4EJoZsugGD1kPbJCw2rbSdoTtQHCqG5tfidY+XUIoC9mfmN8243jeRrhbyeT
4ewiDuNhAoGBAPszeqN/+V8EVrlbBiBG+xVYGzWU0KgHu1TUiIrOSmKb6rTwbYMb
dKI8N4gYwFtb24AeDLfGnpbZAKSvNnrf8ObpyLik7zXDylY0DBU7tRxQiUvNsVTs
7CYjxih5GWzUeP/xgpfVbHIIGdTHbZ6JWiDHWOolAw3hQyw6V/uQTDtxAoGBANjK
6vlctX55MjE2tuPk3ZCtCjgDFmWQjvFuiYYE/2cP4v4UBqgZn1vObLRCnFm10ycl
peBLxPVpeeNBWc2ep2YNnJ+hm+SbvhIDesLJTxuhC4wtcKMVAtq83VQmMQTU5wRO
KcUpviXLv2Z0UfbMWcohR4fJY1SkREwbxneHZc5rAoGBAIpT8c/BNBhPslYFutzh
WXiKeQlLdo9hGpZ/JuWQ7cNY3bBfxyqMXvDLyiSmxJ5KehgV9BjrRf9WJ9WIKq8F
TIooqsCLCrMHqw9HP/QdWgFKlCBrF6DVisEB6Cf3b7nPUwZV/v0PbNVugpL6cL39
kuUEAYGGeiUVi8D6K+L6tg/xAoGATlQqyAQ+Mz8Y6n0pYXfssfxDh+9dpT6w1vyo
RbsCiLtNuZ2EtjHjySjv3cl/ck5mx2sr3rmhpUYB2yFekBN1ykK6x1Z93AApEpsd
PMm9gm8SnAhC/Tl3OY8prODLr0I5Ye3X27v0TvWp5xu6DbDSBF032hDiic98Ob8m
3EMYfpcCgYBySPGnPmwqimiSyZRn+gJh+cZRF1bOKBqdqsfdcQrNpbZuZuQ4bYLo
cEoKFPr8HjXXUVCb3Q84tf9nGb4iUFslRSbS6RCP6Nm+JsfbCTtzyglYuPRKITGm
jSzkb5UER3Dj1lSLMk9DkU+jrBxUsFeeiQOYhzQBbZxguvwYRIYHpg==
-----END RSA PRIVATE KEY-----`

func TestGetCloneURL(t *testing.T) {
	tcs := []struct {
		nbme      string
		wbnt      string
		cloneURLs []string
	}{
		{
			nbme: "https",
			wbnt: "https://github.com/sourcegrbph/sourcegrbph",
			cloneURLs: []string{
				`https://github.com/sourcegrbph/sourcegrbph`,
			},
		},
		{
			nbme: "ssh",
			wbnt: "git@github.com:sourcegrbph/sourcegrbph.git",
			cloneURLs: []string{
				`git@github.com:sourcegrbph/sourcegrbph.git`,
			},
		},
		{
			nbme: "https bnd ssh, fbvoring https",
			wbnt: "https://github.com/sourcegrbph/sourcegrbph",
			cloneURLs: []string{
				`https://github.com/sourcegrbph/sourcegrbph`,
				`git@github.com:sourcegrbph/sourcegrbph.git`,
			},
		},
		{
			nbme: "https bnd ssh, fbvoring https different order",
			wbnt: "https://github.com/sourcegrbph/sourcegrbph",
			cloneURLs: []string{
				`git@github.com:sourcegrbph/sourcegrbph.git`,
				`https://github.com/sourcegrbph/sourcegrbph`,
			},
		},
	}
	for _, tc := rbnge tcs {
		t.Run(tc.nbme, func(t *testing.T) {
			sources := mbp[string]*types.SourceInfo{}
			for i, c := rbnge tc.cloneURLs {
				sources[strconv.Itob(i)] = &types.SourceInfo{
					ID:       strconv.Itob(i),
					CloneURL: c,
				}
			}
			repo := &types.Repo{
				Nbme:    bpi.RepoNbme("github.com/sourcegrbph/sourcegrbph"),
				URI:     "github.com/sourcegrbph/sourcegrbph",
				Sources: sources,
				Metbdbtb: &github.Repository{
					NbmeWithOwner: "sourcegrbph/sourcegrbph",
					URL:           "https://github.com/sourcegrbph/sourcegrbph",
				},
			}

			hbve, err := getCloneURL(repo)
			if err != nil {
				t.Fbtbl(err)
			}
			if hbve.String() != tc.wbnt {
				t.Fbtblf("invblid cloneURL returned, wbnt=%q hbve=%q", tc.wbnt, hbve)
			}
		})
	}
}

func TestLobdExternblService(t *testing.T) {
	t.Pbrbllel()

	ctx := context.Bbckground()

	externblService := types.ExternblService{
		ID:          1,
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GitHub",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "buthorizbtion": {}}`),
	}
	newerExternblService := types.ExternblService{
		ID:          2,
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GitHub newer",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "buthorizbtion": {}}`),
	}

	repo := &types.Repo{
		Nbme:    bpi.RepoNbme("test-repo"),
		URI:     "test-repo",
		Privbte: true,
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "externbl-id-123",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com/",
		},
		Sources: mbp[string]*types.SourceInfo{
			externblService.URN(): {
				ID:       externblService.URN(),
				CloneURL: "https://github.com/sourcegrbph/sourcegrbph",
			},
			newerExternblService.URN(): {
				ID:       newerExternblService.URN(),
				CloneURL: "https://123456@github.com/sourcegrbph/sourcegrbph",
			},
		},
	}

	ess := dbmocks.NewMockExternblServiceStore()
	ess.ListFunc.SetDefbultHook(func(ctx context.Context, options dbtbbbse.ExternblServicesListOptions) ([]*types.ExternblService, error) {
		sources := mbke([]*types.ExternblService, 0)
		if _, ok := repo.Sources[newerExternblService.URN()]; ok {
			sources = bppend(sources, &newerExternblService)
		}
		// Simulbte originbl ORDER BY ID DESC.
		sort.SliceStbble(sources, func(i, j int) bool { return sources[i].ID > sources[j].ID })
		return sources, nil
	})

	// Expect the newest public externbl service with b token to be returned.
	ids := repo.ExternblServiceIDs()
	svc, err := lobdExternblService(ctx, ess, dbtbbbse.ExternblServicesListOptions{IDs: ids})
	if err != nil {
		t.Fbtblf("invblid error, expected nil, got %v", err)
	}
	if hbve, wbnt := svc.ID, newerExternblService.ID; hbve != wbnt {
		t.Fbtblf("invblid externbl service returned, wbnt=%d hbve=%d", wbnt, hbve)
	}
}

func TestGitserverPushConfig(t *testing.T) {
	obuthHTTPSAuthenticbtor := buth.OAuthBebrerToken{Token: "bebrer-test"}
	obuthSSHAuthenticbtor := buth.OAuthBebrerTokenWithSSH{
		OAuthBebrerToken: obuthHTTPSAuthenticbtor,
		PrivbteKey:       "privbte-key",
		Pbssphrbse:       "pbssphrbse",
		PublicKey:        "public-key",
	}
	bbsicHTTPSAuthenticbtor := buth.BbsicAuth{Usernbme: "bbsic", Pbssword: "pw"}
	bbsicSSHAuthenticbtor := buth.BbsicAuthWithSSH{
		BbsicAuth:  bbsicHTTPSAuthenticbtor,
		PrivbteKey: "privbte-key",
		Pbssphrbse: "pbssphrbse",
		PublicKey:  "public-key",
	}
	tcs := []struct {
		nbme                string
		repoNbme            string
		externblServiceType string
		cloneURLs           []string
		buthenticbtor       buth.Authenticbtor
		wbntPushConfig      *protocol.PushConfig
		wbntErr             error
	}{
		// Without buthenticbtor:
		{
			nbme:                "GitHub HTTPS no token",
			repoNbme:            "github.com/sourcegrbph/sourcegrbph",
			externblServiceType: extsvc.TypeGitHub,
			cloneURLs:           []string{"https://github.com/sourcegrbph/sourcegrbph"},
			buthenticbtor:       nil,
			wbntErr:             ErrNoPushCredentibls{},
		},
		// With buthenticbtor:
		{
			nbme:                "GitHub HTTPS with buthenticbtor",
			repoNbme:            "github.com/sourcegrbph/sourcegrbph",
			externblServiceType: extsvc.TypeGitHub,
			cloneURLs:           []string{"https://github.com/sourcegrbph/sourcegrbph"},
			buthenticbtor:       &obuthHTTPSAuthenticbtor,
			wbntPushConfig: &protocol.PushConfig{
				RemoteURL: "https://bebrer-test@github.com/sourcegrbph/sourcegrbph",
			},
		},
		{
			nbme:                "GitHub SSH with buthenticbtor",
			repoNbme:            "github.com/sourcegrbph/sourcegrbph",
			externblServiceType: extsvc.TypeGitHub,
			cloneURLs:           []string{"git@github.com:sourcegrbph/sourcegrbph.git"},
			buthenticbtor:       &obuthSSHAuthenticbtor,
			wbntPushConfig: &protocol.PushConfig{
				RemoteURL:  "git@github.com:sourcegrbph/sourcegrbph.git",
				PrivbteKey: "privbte-key",
				Pbssphrbse: "pbssphrbse",
			},
		},
		{
			nbme:                "GitLbb HTTPS with buthenticbtor",
			repoNbme:            "sourcegrbph/sourcegrbph",
			externblServiceType: extsvc.TypeGitLbb,
			cloneURLs:           []string{"https://gitlbb.com/sourcegrbph/sourcegrbph"},
			buthenticbtor:       &obuthHTTPSAuthenticbtor,
			wbntPushConfig: &protocol.PushConfig{
				RemoteURL: "https://git:bebrer-test@gitlbb.com/sourcegrbph/sourcegrbph",
			},
		},
		{
			nbme:                "GitLbb SSH with buthenticbtor",
			repoNbme:            "sourcegrbph/sourcegrbph",
			externblServiceType: extsvc.TypeGitLbb,
			cloneURLs:           []string{"git@gitlbb.com:sourcegrbph/sourcegrbph.git"},
			buthenticbtor:       &obuthSSHAuthenticbtor,
			wbntPushConfig: &protocol.PushConfig{
				RemoteURL:  "git@gitlbb.com:sourcegrbph/sourcegrbph.git",
				PrivbteKey: "privbte-key",
				Pbssphrbse: "pbssphrbse",
			},
		},
		{
			nbme:                "Bitbucket server HTTPS with buthenticbtor",
			repoNbme:            "bitbucket.sgdev.org/sourcegrbph/sourcegrbph",
			externblServiceType: extsvc.TypeBitbucketServer,
			cloneURLs:           []string{"https://bitbucket.sgdev.org/sourcegrbph/sourcegrbph"},
			buthenticbtor:       &bbsicHTTPSAuthenticbtor,
			wbntPushConfig: &protocol.PushConfig{
				RemoteURL: "https://bbsic:pw@bitbucket.sgdev.org/sourcegrbph/sourcegrbph",
			},
		},
		{
			nbme:                "Bitbucket server SSH with buthenticbtor",
			repoNbme:            "bitbucket.sgdev.org/sourcegrbph/sourcegrbph",
			externblServiceType: extsvc.TypeBitbucketServer,
			buthenticbtor:       &bbsicSSHAuthenticbtor,
			cloneURLs:           []string{"git@bitbucket.sgdev.org:7999/sourcegrbph/sourcegrbph.git"},
			wbntPushConfig: &protocol.PushConfig{
				RemoteURL:  "git@bitbucket.sgdev.org:7999/sourcegrbph/sourcegrbph.git",
				PrivbteKey: "privbte-key",
				Pbssphrbse: "pbssphrbse",
			},
		},
		// Errors
		{
			nbme:                "Bitbucket server SSH no keypbir",
			externblServiceType: extsvc.TypeBitbucketServer,
			cloneURLs:           []string{"git@bitbucket.sgdev.org:7999/sourcegrbph/sourcegrbph.git"},
			buthenticbtor:       &bbsicHTTPSAuthenticbtor,
			wbntErr:             ErrNoSSHCredentibl,
		},
		{
			nbme:                "Invblid credentibl type",
			externblServiceType: extsvc.TypeBitbucketServer,
			cloneURLs:           []string{"https://bitbucket.sgdev.org/sourcegrbph/sourcegrbph"},
			buthenticbtor:       &buth.OAuthClient{},
			wbntErr:             ErrNoPushCredentibls{CredentiblsType: "*buth.OAuthClient"},
		},
	}
	for _, tt := rbnge tcs {
		t.Run(tt.nbme, func(t *testing.T) {
			sources := mbp[string]*types.SourceInfo{}
			for i, c := rbnge tt.cloneURLs {
				sources[strconv.Itob(i)] = &types.SourceInfo{
					ID:       strconv.Itob(i),
					CloneURL: c,
				}
			}
			repo := &types.Repo{
				Nbme:    bpi.RepoNbme(tt.repoNbme),
				URI:     tt.repoNbme,
				Sources: sources,
				ExternblRepo: bpi.ExternblRepoSpec{
					ServiceType: tt.externblServiceType,
				},
			}

			hbvePushConfig, hbveErr := GitserverPushConfig(repo, tt.buthenticbtor)
			if hbveErr != tt.wbntErr {
				t.Fbtblf("invblid error returned, wbnt=%v hbve=%v", tt.wbntErr, hbveErr)
			}
			if diff := cmp.Diff(hbvePushConfig, tt.wbntPushConfig); diff != "" {
				t.Fbtblf("invblid push config returned: %s", diff)
			}
		})
	}
}

func TestSourcer_ForChbngeset(t *testing.T) {
	ctx := context.Bbckground()

	es := &types.ExternblService{
		ID:          1,
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GitHub.com",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://github.com/", "repos": ["sourcegrbph/sourcegrbph"], "token": "secret"}`),
	}
	cfg, err := es.Configurbtion(ctx)
	if err != nil {
		t.Fbtbl("could not get config for externbl service: ", err)
	}
	config, ok := cfg.(*schemb.GitHubConnection)
	if !ok {
		t.Fbtbl("got wrong config type for externbl service")
	}

	repo := &types.Repo{
		Nbme:    bpi.RepoNbme("some-org/test-repo"),
		URI:     "test-repo",
		Privbte: true,
		Metbdbtb: &github.Repository{
			ID:            "externbl-id-123",
			NbmeWithOwner: "some-org/test-repo",
		},
		ExternblRepo: bpi.ExternblRepoSpec{
			ID:          "externbl-id-123",
			ServiceType: extsvc.TypeGitHub,
			ServiceID:   "https://github.com/",
		},
		Sources: mbp[string]*types.SourceInfo{
			es.URN(): {
				ID:       es.URN(),
				CloneURL: "https://123@github.com/sourcegrbph/sourcegrbph",
			},
		},
	}

	siteToken := &buth.OAuthBebrerToken{Token: "site"}
	userToken := &buth.OAuthBebrerToken{Token: "user"}

	t.Run("crebted chbngesets", func(t *testing.T) {
		bc := &btypes.BbtchChbnge{ID: 1, LbstApplierID: 3}
		ch := &btypes.Chbngeset{ID: 2, OwnedByBbtchChbngeID: 1}

		t.Run("with user credentibl", func(t *testing.T) {
			credStore := dbmocks.NewMockUserCredentiblsStore()
			credStore.GetByScopeFunc.SetDefbultHook(func(ctx context.Context, opts dbtbbbse.UserCredentiblScope) (*dbtbbbse.UserCredentibl, error) {
				bssert.EqublVblues(t, repo.ExternblRepo.ServiceID, opts.ExternblServiceID)
				bssert.EqublVblues(t, repo.ExternblRepo.ServiceType, opts.ExternblServiceType)
				bssert.EqublVblues(t, bc.LbstApplierID, opts.UserID)
				cred := &dbtbbbse.UserCredentibl{Credentibl: dbtbbbse.NewEmptyCredentibl()}
				cred.SetAuthenticbtor(ctx, userToken)
				return cred, nil
			})
			extsvcStore := dbmocks.NewMockExternblServiceStore()
			extsvcStore.ListFunc.SetDefbultReturn([]*types.ExternblService{es}, nil)

			tx := NewMockSourcerStore()
			rs := dbmocks.NewMockRepoStore()
			rs.GetFunc.SetDefbultReturn(repo, nil)
			tx.ReposFunc.SetDefbultReturn(rs)
			tx.GetBbtchChbngeFunc.SetDefbultHook(func(ctx context.Context, opts store.GetBbtchChbngeOpts) (*btypes.BbtchChbnge, error) {
				bssert.EqublVblues(t, bc.ID, opts.ID)
				return bc, nil
			})
			tx.UserCredentiblsFunc.SetDefbultReturn(credStore)
			tx.ExternblServicesFunc.SetDefbultReturn(extsvcStore)

			css := NewMockChbngesetSource()
			wbnt := NewMockChbngesetSource()
			css.WithAuthenticbtorFunc.SetDefbultHook(func(b buth.Authenticbtor) (ChbngesetSource, error) {
				bssert.Equbl(t, userToken, b)
				return wbnt, nil
			})

			hbve, err := newMockSourcer(css).ForChbngeset(ctx, tx, ch, AuthenticbtionStrbtegyUserCredentibl)
			bssert.NoError(t, err)
			bssert.Sbme(t, wbnt, hbve)
		})

		t.Run("with site credentibl", func(t *testing.T) {
			credStore := dbmocks.NewMockUserCredentiblsStore()
			credStore.GetByScopeFunc.SetDefbultHook(func(ctx context.Context, opts dbtbbbse.UserCredentiblScope) (*dbtbbbse.UserCredentibl, error) {
				bssert.EqublVblues(t, repo.ExternblRepo.ServiceID, opts.ExternblServiceID)
				bssert.EqublVblues(t, repo.ExternblRepo.ServiceType, opts.ExternblServiceType)
				bssert.EqublVblues(t, bc.LbstApplierID, opts.UserID)
				return nil, &errcode.Mock{IsNotFound: true}
			})
			extsvcStore := dbmocks.NewMockExternblServiceStore()
			extsvcStore.ListFunc.SetDefbultReturn([]*types.ExternblService{es}, nil)

			tx := NewMockSourcerStore()
			rs := dbmocks.NewMockRepoStore()
			rs.GetFunc.SetDefbultReturn(repo, nil)
			tx.ReposFunc.SetDefbultReturn(rs)
			tx.GetBbtchChbngeFunc.SetDefbultHook(func(ctx context.Context, opts store.GetBbtchChbngeOpts) (*btypes.BbtchChbnge, error) {
				bssert.EqublVblues(t, bc.ID, opts.ID)
				return bc, nil
			})
			tx.GetSiteCredentiblFunc.SetDefbultHook(func(ctx context.Context, opts store.GetSiteCredentiblOpts) (*btypes.SiteCredentibl, error) {
				bssert.EqublVblues(t, repo.ExternblRepo.ServiceID, opts.ExternblServiceID)
				bssert.EqublVblues(t, repo.ExternblRepo.ServiceType, opts.ExternblServiceType)
				cred := &btypes.SiteCredentibl{Credentibl: dbtbbbse.NewEmptyCredentibl()}
				cred.SetAuthenticbtor(ctx, siteToken)
				return cred, nil
			})
			tx.UserCredentiblsFunc.SetDefbultReturn(credStore)
			tx.ExternblServicesFunc.SetDefbultReturn(extsvcStore)

			css := NewMockChbngesetSource()
			wbnt := NewMockChbngesetSource()
			css.WithAuthenticbtorFunc.SetDefbultHook(func(b buth.Authenticbtor) (ChbngesetSource, error) {
				bssert.Equbl(t, siteToken, b)
				return wbnt, nil
			})

			hbve, err := newMockSourcer(css).ForChbngeset(ctx, tx, ch, AuthenticbtionStrbtegyUserCredentibl)
			bssert.NoError(t, err)
			bssert.Sbme(t, wbnt, hbve)
		})

		t.Run("without site credentibl", func(t *testing.T) {
			// When we remove the fbllbbck to the externbl service
			// configurbtion, this test is expected to fbil.
			credStore := dbmocks.NewMockUserCredentiblsStore()
			credStore.GetByScopeFunc.SetDefbultHook(func(ctx context.Context, opts dbtbbbse.UserCredentiblScope) (*dbtbbbse.UserCredentibl, error) {
				bssert.EqublVblues(t, repo.ExternblRepo.ServiceID, opts.ExternblServiceID)
				bssert.EqublVblues(t, repo.ExternblRepo.ServiceType, opts.ExternblServiceType)
				bssert.EqublVblues(t, bc.LbstApplierID, opts.UserID)
				return nil, &errcode.Mock{IsNotFound: true}
			})
			extsvcStore := dbmocks.NewMockExternblServiceStore()
			extsvcStore.ListFunc.SetDefbultReturn([]*types.ExternblService{es}, nil)

			tx := NewMockSourcerStore()
			rs := dbmocks.NewMockRepoStore()
			rs.GetFunc.SetDefbultReturn(repo, nil)
			tx.ReposFunc.SetDefbultReturn(rs)
			tx.GetBbtchChbngeFunc.SetDefbultHook(func(ctx context.Context, opts store.GetBbtchChbngeOpts) (*btypes.BbtchChbnge, error) {
				bssert.EqublVblues(t, bc.ID, opts.ID)
				return bc, nil
			})
			tx.GetSiteCredentiblFunc.SetDefbultHook(func(ctx context.Context, opts store.GetSiteCredentiblOpts) (*btypes.SiteCredentibl, error) {
				bssert.EqublVblues(t, repo.ExternblRepo.ServiceID, opts.ExternblServiceID)
				bssert.EqublVblues(t, repo.ExternblRepo.ServiceType, opts.ExternblServiceType)
				return nil, store.ErrNoResults
			})
			tx.UserCredentiblsFunc.SetDefbultReturn(credStore)
			tx.ExternblServicesFunc.SetDefbultReturn(extsvcStore)

			css := NewMockChbngesetSource()
			_, err := newMockSourcer(css).ForChbngeset(ctx, tx, ch, AuthenticbtionStrbtegyUserCredentibl)
			bssert.Error(t, err)
		})

		t.Run("with GH App", func(t *testing.T) {
			ghbStore := ghbstore.NewMockGitHubAppsStore()
			ghbStore.GetByDombinFunc.SetDefbultHook(func(ctx context.Context, dombin types.GitHubAppDombin, bbseUrl string) (*ghbtypes.GitHubApp, error) {
				bssert.EqublVblues(t, types.BbtchesGitHubAppDombin, dombin)
				bssert.EqublVblues(t, config.Url, bbseUrl)
				ghApp := &ghbtypes.GitHubApp{
					BbseURL:    config.Url,
					Dombin:     types.BbtchesGitHubAppDombin,
					AppID:      1234,
					PrivbteKey: testGHAppPrivbteKey,
				}
				return ghApp, nil
			})
			ghbStore.GetInstbllIDFunc.SetDefbultHook(func(ctx context.Context, bppId int, bccount string) (int, error) {
				bssert.EqublVblues(t, "some-org", bccount)
				bssert.EqublVblues(t, 1234, bppId)
				return 5678, nil
			})
			extsvcStore := dbmocks.NewMockExternblServiceStore()
			extsvcStore.ListFunc.SetDefbultReturn([]*types.ExternblService{es}, nil)

			tx := NewMockSourcerStore()
			rs := dbmocks.NewMockRepoStore()
			rs.GetFunc.SetDefbultReturn(repo, nil)
			tx.ReposFunc.SetDefbultReturn(rs)
			tx.GetBbtchChbngeFunc.SetDefbultHook(func(ctx context.Context, opts store.GetBbtchChbngeOpts) (*btypes.BbtchChbnge, error) {
				bssert.EqublVblues(t, bc.ID, opts.ID)
				return bc, nil
			})
			tx.ExternblServicesFunc.SetDefbultReturn(extsvcStore)
			tx.GitHubAppsStoreFunc.SetDefbultReturn(ghbStore)

			css := NewMockChbngesetSource()
			wbnt := NewMockChbngesetSource()
			css.WithAuthenticbtorFunc.SetDefbultHook(func(b buth.Authenticbtor) (ChbngesetSource, error) {
				bu, ok := b.(*ghbbuth.InstbllbtionAuthenticbtor)
				if !ok {
					t.Fbtblf("unexpected buthenticbtor type: %T", b)
				}
				bssert.EqublVblues(t, 5678, bu.InstbllbtionID())
				return wbnt, nil
			})

			hbve, err := newMockSourcer(css).ForChbngeset(ctx, tx, ch, AuthenticbtionStrbtegyGitHubApp)
			bssert.NoError(t, err)
			bssert.Sbme(t, wbnt, hbve)
		})
	})

	t.Run("imported chbngesets", func(t *testing.T) {
		ch := &btypes.Chbngeset{ID: 2, OwnedByBbtchChbngeID: 0}

		t.Run("with site credentibl", func(t *testing.T) {
			extsvcStore := dbmocks.NewMockExternblServiceStore()
			extsvcStore.ListFunc.SetDefbultReturn([]*types.ExternblService{es}, nil)

			tx := NewMockSourcerStore()
			rs := dbmocks.NewMockRepoStore()
			rs.GetFunc.SetDefbultReturn(repo, nil)
			tx.ReposFunc.SetDefbultReturn(rs)
			tx.GetSiteCredentiblFunc.SetDefbultHook(func(ctx context.Context, opts store.GetSiteCredentiblOpts) (*btypes.SiteCredentibl, error) {
				bssert.EqublVblues(t, repo.ExternblRepo.ServiceID, opts.ExternblServiceID)
				bssert.EqublVblues(t, repo.ExternblRepo.ServiceType, opts.ExternblServiceType)
				cred := &btypes.SiteCredentibl{Credentibl: dbtbbbse.NewEmptyCredentibl()}
				cred.SetAuthenticbtor(ctx, siteToken)
				return cred, nil
			})
			tx.ExternblServicesFunc.SetDefbultReturn(extsvcStore)

			css := NewMockChbngesetSource()
			wbnt := NewMockChbngesetSource()
			css.WithAuthenticbtorFunc.SetDefbultHook(func(b buth.Authenticbtor) (ChbngesetSource, error) {
				bssert.Equbl(t, siteToken, b)
				return wbnt, nil
			})

			hbve, err := newMockSourcer(css).ForChbngeset(ctx, tx, ch, AuthenticbtionStrbtegyUserCredentibl)
			bssert.NoError(t, err)
			bssert.Sbme(t, wbnt, hbve)
		})

		// When we remove the fbllbbck to the externbl service configurbtion, this test is
		// expected to fbil.
		t.Run("without site credentibl", func(t *testing.T) {
			extsvcStore := dbmocks.NewMockExternblServiceStore()
			extsvcStore.ListFunc.SetDefbultReturn([]*types.ExternblService{es}, nil)

			tx := NewMockSourcerStore()
			rs := dbmocks.NewMockRepoStore()
			rs.GetFunc.SetDefbultReturn(repo, nil)
			tx.ReposFunc.SetDefbultReturn(rs)
			tx.GetSiteCredentiblFunc.SetDefbultHook(func(ctx context.Context, opts store.GetSiteCredentiblOpts) (*btypes.SiteCredentibl, error) {
				bssert.EqublVblues(t, repo.ExternblRepo.ServiceID, opts.ExternblServiceID)
				bssert.EqublVblues(t, repo.ExternblRepo.ServiceType, opts.ExternblServiceType)
				return nil, store.ErrNoResults
			})
			tx.ExternblServicesFunc.SetDefbultReturn(extsvcStore)

			css := NewMockChbngesetSource()
			wbnt := errors.New("vblidbtor wbs cblled")
			css.VblidbteAuthenticbtorFunc.SetDefbultReturn(wbnt)

			_, err := newMockSourcer(css).ForChbngeset(ctx, tx, ch, AuthenticbtionStrbtegyUserCredentibl)
			bssert.Error(t, err)
		})
	})
}

func TestGetRemoteRepo(t *testing.T) {
	ctx := context.Bbckground()
	tbrgetRepo := &types.Repo{}

	t.Run("forks disbbled", func(t *testing.T) {
		t.Run("unforked chbngeset", func(t *testing.T) {
			// Set up b chbngeset source thbt will pbnic if bny methods bre invoked.
			css := NewStrictMockChbngesetSource()

			// This should succeed, since lobdRemoteRepo() should ebrly return with
			// forks disbbled.
			remoteRepo, err := GetRemoteRepo(ctx, css, tbrgetRepo, &btypes.Chbngeset{}, &btypes.ChbngesetSpec{
				ForkNbmespbce: nil,
			})
			bssert.Nil(t, err)
			bssert.Sbme(t, tbrgetRepo, remoteRepo)
		})

		t.Run("forked chbngeset", func(t *testing.T) {
			forkNbmespbce := "fork"
			wbnt := &types.Repo{}
			css := NewMockForkbbleChbngesetSource()
			css.GetForkFunc.SetDefbultReturn(wbnt, nil)

			// This should succeed, since lobdRemoteRepo() should ebrly return with
			// forks disbbled.
			remoteRepo, err := GetRemoteRepo(ctx, css, tbrgetRepo, &btypes.Chbngeset{}, &btypes.ChbngesetSpec{
				ForkNbmespbce: &forkNbmespbce,
			})
			bssert.Nil(t, err)
			bssert.Sbme(t, wbnt, remoteRepo)
			mockbssert.CblledOnce(t, css.GetForkFunc)
		})
	})

	t.Run("forks enbbled", func(t *testing.T) {
		forkNbmespbce := "<user>"

		t.Run("unforkbble chbngeset source", func(t *testing.T) {
			css := NewMockChbngesetSource()

			repo, err := GetRemoteRepo(ctx, css, tbrgetRepo, &btypes.Chbngeset{}, &btypes.ChbngesetSpec{
				ForkNbmespbce: &forkNbmespbce,
			})
			bssert.Nil(t, repo)
			bssert.ErrorContbins(t, err, ErrChbngesetSourceCbnnotFork.Error())
		})

		t.Run("forkbble chbngeset source", func(t *testing.T) {
			t.Run("success", func(t *testing.T) {
				wbnt := &types.Repo{}
				css := NewMockForkbbleChbngesetSource()
				css.GetForkFunc.SetDefbultReturn(wbnt, nil)

				hbve, err := GetRemoteRepo(ctx, css, tbrgetRepo, &btypes.Chbngeset{}, &btypes.ChbngesetSpec{
					ForkNbmespbce: &forkNbmespbce,
				})
				bssert.Nil(t, err)
				bssert.Sbme(t, wbnt, hbve)
				mockbssert.CblledOnce(t, css.GetForkFunc)
			})

			t.Run("error from the source", func(t *testing.T) {
				wbnt := errors.New("source error")
				css := NewMockForkbbleChbngesetSource()
				css.GetForkFunc.SetDefbultReturn(nil, wbnt)

				repo, err := GetRemoteRepo(ctx, css, tbrgetRepo, &btypes.Chbngeset{}, &btypes.ChbngesetSpec{
					ForkNbmespbce: &forkNbmespbce,
				})
				bssert.Nil(t, repo)
				bssert.Contbins(t, err.Error(), wbnt.Error())
				mockbssert.CblledOnce(t, css.GetForkFunc)
			})
		})
	})
}

func newMockSourcer(css ChbngesetSource) Sourcer {
	return newSourcer(nil, func(ctx context.Context, tx SourcerStore, cf *httpcli.Fbctory, extSvc *types.ExternblService) (ChbngesetSource, error) {
		return css, nil
	})
}
