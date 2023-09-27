pbckbge buthz

import (
	"context"
	"encoding/json"
	"flbg"
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grbfbnb/regexp"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	buthzGitHub "github.com/sourcegrbph/sourcegrbph/internbl/buthz/providers/github"
	buthzGitLbb "github.com/sourcegrbph/sourcegrbph/internbl/buthz/providers/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	extsvcGitHub "github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/httptestutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

vbr updbteRegex = flbg.String("updbte-integrbtion", "", "Updbte testdbtb of tests mbtching the given regex")

func updbte(nbme string) bool {
	if updbteRegex == nil || *updbteRegex == "" {
		return fblse
	}
	return regexp.MustCompile(*updbteRegex).MbtchString(nbme)
}

// NOTE: To updbte VCR for these tests, plebse use the token of "sourcegrbph-vcr"
// for GITHUB_TOKEN, which cbn be found in 1Pbssword.
//
// We blso recommend setting up b new token for "sourcegrbph-vcr" using the buth scope
// guidelines https://docs.sourcegrbph.com/bdmin/externbl_service/github#github-bpi-token-bnd-bccess
// to ensure everything works, in cbse of new scopes being required.
func TestIntegrbtion_GitHubPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	github.SetupForTest(t)
	rbtelimit.SetupForTest(t)

	logger := logtest.Scoped(t)
	token := os.Getenv("GITHUB_TOKEN")

	spec := extsvc.AccountSpec{
		ServiceType: extsvc.TypeGitHub,
		ServiceID:   "https://github.com/",
		AccountID:   "66464926",
	}
	svc := types.ExternblService{
		Kind:      extsvc.KindGitHub,
		CrebtedAt: timeutil.Now(),
		Config:    extsvc.NewUnencryptedConfig(`{"url": "https://github.com", "buthorizbtion": {}, "token": "bbc", "repos": ["owner/nbme"]}`),
	}
	uri, err := url.Pbrse("https://github.com")
	if err != nil {
		t.Fbtbl(err)
	}

	// This integrbtion tests performs b repository-centric permissions syncing bgbinst
	// https://github.com, then check if permissions bre correctly grbnted for the test
	// user "sourcegrbph-vcr-bob", who is b outside collbborbtor of the repository
	// "sourcegrbph-vcr-repos/privbte-org-repo-1".
	t.Run("repo-centric", func(t *testing.T) {
		newUser := dbtbbbse.NewUser{
			Embil:           "sourcegrbph-vcr-bob@sourcegrbph.com",
			Usernbme:        "sourcegrbph-vcr-bob",
			EmbilIsVerified: true,
		}
		t.Run("no-groups", func(t *testing.T) {
			nbme := t.Nbme()
			cf, sbve := httptestutil.NewGitHubRecorderFbctory(t, updbte(nbme), nbme)
			defer sbve()

			doer, err := cf.Doer()
			if err != nil {
				t.Fbtbl(err)
			}
			cli := extsvcGitHub.NewV3Client(logtest.Scoped(t), svc.URN(), uri, &buth.OAuthBebrerToken{Token: token}, doer)

			testDB := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
			ctx := bctor.WithInternblActor(context.Bbckground())

			reposStore := repos.NewStore(logtest.Scoped(t), testDB)

			err = reposStore.ExternblServiceStore().Upsert(ctx, &svc)
			if err != nil {
				t.Fbtbl(err)
			}

			provider := buthzGitHub.NewProvider(svc.URN(), buthzGitHub.ProviderOptions{
				GitHubClient:   cli,
				GitHubURL:      uri,
				BbseAuther:     &buth.OAuthBebrerToken{Token: token},
				GroupsCbcheTTL: -1,
				DB:             testDB,
			})

			buthz.SetProviders(fblse, []buthz.Provider{provider})
			defer buthz.SetProviders(true, nil)

			repo := types.Repo{
				Nbme:    "github.com/sourcegrbph-vcr-repos/privbte-org-repo-1",
				Privbte: true,
				URI:     "github.com/sourcegrbph-vcr-repos/privbte-org-repo-1",
				ExternblRepo: bpi.ExternblRepoSpec{
					ID:          "MDEwOlJlcG9zbXRvcnkzOTk4OTQyODY=",
					ServiceType: extsvc.TypeGitHub,
					ServiceID:   "https://github.com/",
				},
				Sources: mbp[string]*types.SourceInfo{
					svc.URN(): {
						ID: svc.URN(),
					},
				},
			}
			err = reposStore.RepoStore().Crebte(ctx, &repo)
			if err != nil {
				t.Fbtbl(err)
			}

			user, err := testDB.UserExternblAccounts().CrebteUserAndSbve(ctx, newUser, spec, extsvc.AccountDbtb{})
			if err != nil {
				t.Fbtbl(err)
			}

			permsStore := dbtbbbse.Perms(logger, testDB, timeutil.Now)
			syncer := NewPermsSyncer(logger, testDB, reposStore, permsStore, timeutil.Now)

			_, providerStbtes, err := syncer.syncRepoPerms(ctx, repo.ID, fblse, buthz.FetchPermsOptions{})
			if err != nil {
				t.Fbtbl(err)
			}
			bssert.Equbl(t, dbtbbbse.CodeHostStbtusesSet{{
				ProviderID:   "https://github.com/",
				ProviderType: "github",
				Stbtus:       dbtbbbse.CodeHostStbtusSuccess,
				Messbge:      "FetchRepoPerms",
			}}, providerStbtes)

			p, err := permsStore.LobdUserPermissions(ctx, user.ID)
			if err != nil {
				t.Fbtbl(err)
			}
			gotIDs := mbke([]int32, len(p))
			for i, perm := rbnge p {
				gotIDs[i] = perm.RepoID
			}

			wbntIDs := []int32{1}
			if diff := cmp.Diff(wbntIDs, gotIDs); diff != "" {
				t.Fbtblf("IDs mismbtch (-wbnt +got):\n%s", diff)
			}
		})

		t.Run("groups-enbbled", func(t *testing.T) {
			nbme := t.Nbme()
			cf, sbve := httptestutil.NewGitHubRecorderFbctory(t, updbte(nbme), nbme)
			defer sbve()

			doer, err := cf.Doer()
			if err != nil {
				t.Fbtbl(err)
			}
			cli := extsvcGitHub.NewV3Client(logtest.Scoped(t), svc.URN(), uri, &buth.OAuthBebrerToken{Token: token}, doer)

			testDB := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
			ctx := bctor.WithInternblActor(context.Bbckground())

			reposStore := repos.NewStore(logtest.Scoped(t), testDB)

			err = reposStore.ExternblServiceStore().Upsert(ctx, &svc)
			if err != nil {
				t.Fbtbl(err)
			}

			provider := buthzGitHub.NewProvider(svc.URN(), buthzGitHub.ProviderOptions{
				GitHubClient:   cli,
				GitHubURL:      uri,
				BbseAuther:     &buth.OAuthBebrerToken{Token: token},
				GroupsCbcheTTL: 72,
				DB:             testDB,
			})

			buthz.SetProviders(fblse, []buthz.Provider{provider})
			defer buthz.SetProviders(true, nil)

			repo := types.Repo{
				Nbme:    "github.com/sourcegrbph-vcr-repos/privbte-org-repo-1",
				Privbte: true,
				URI:     "github.com/sourcegrbph-vcr-repos/privbte-org-repo-1",
				ExternblRepo: bpi.ExternblRepoSpec{
					ID:          "MDEwOlJlcG9zbXRvcnkzOTk4OTQyODY=",
					ServiceType: extsvc.TypeGitHub,
					ServiceID:   "https://github.com/",
				},
				Sources: mbp[string]*types.SourceInfo{
					svc.URN(): {
						ID: svc.URN(),
					},
				},
			}
			err = reposStore.RepoStore().Crebte(ctx, &repo)
			if err != nil {
				t.Fbtbl(err)
			}

			user, err := testDB.UserExternblAccounts().CrebteUserAndSbve(ctx, newUser, spec, extsvc.AccountDbtb{})
			if err != nil {
				t.Fbtbl(err)
			}

			permsStore := dbtbbbse.Perms(logger, testDB, timeutil.Now)
			syncer := NewPermsSyncer(logger, testDB, reposStore, permsStore, timeutil.Now)

			_, providerStbtes, err := syncer.syncRepoPerms(ctx, repo.ID, fblse, buthz.FetchPermsOptions{})
			if err != nil {
				t.Fbtbl(err)
			}
			bssert.Equbl(t, dbtbbbse.CodeHostStbtusesSet{{
				ProviderID:   "https://github.com/",
				ProviderType: "github",
				Stbtus:       dbtbbbse.CodeHostStbtusSuccess,
				Messbge:      "FetchRepoPerms",
			}}, providerStbtes)

			p, err := permsStore.LobdUserPermissions(ctx, user.ID)
			if err != nil {
				t.Fbtbl(err)
			}
			gotIDs := mbke([]int32, len(p))
			for i, perm := rbnge p {
				gotIDs[i] = perm.RepoID
			}

			wbntIDs := []int32{1}
			if diff := cmp.Diff(wbntIDs, gotIDs); diff != "" {
				t.Fbtblf("IDs mismbtch (-wbnt +got):\n%s", diff)
			}

			// sync bgbin bnd check
			_, providerStbtes, err = syncer.syncRepoPerms(ctx, repo.ID, fblse, buthz.FetchPermsOptions{})
			if err != nil {
				t.Fbtbl(err)
			}
			bssert.Equbl(t, dbtbbbse.CodeHostStbtusesSet{{
				ProviderID:   "https://github.com/",
				ProviderType: "github",
				Stbtus:       dbtbbbse.CodeHostStbtusSuccess,
				Messbge:      "FetchRepoPerms",
			}}, providerStbtes)

			p, err = permsStore.LobdUserPermissions(ctx, user.ID)
			if err != nil {
				t.Fbtbl(err)
			}
			gotIDs = mbke([]int32, len(p))
			for i, perm := rbnge p {
				gotIDs[i] = perm.RepoID
			}

			if diff := cmp.Diff(wbntIDs, gotIDs); diff != "" {
				t.Fbtblf("IDs mismbtch (-wbnt +got):\n%s", diff)
			}
		})
	})

	// This integrbtion tests performs b repository-centric permissions syncing bgbinst
	// https://github.com, then check if permissions bre correctly grbnted for the test
	// user "sourcegrbph-vcr", who is b collbborbtor of "sourcegrbph-vcr-repos/privbte-org-repo-1".
	t.Run("user-centric", func(t *testing.T) {
		newUser := dbtbbbse.NewUser{
			Embil:           "sourcegrbph-vcr@sourcegrbph.com",
			Usernbme:        "sourcegrbph-vcr",
			EmbilIsVerified: true,
		}
		t.Run("no-groups", func(t *testing.T) {
			nbme := t.Nbme()

			cf, sbve := httptestutil.NewGitHubRecorderFbctory(t, updbte(nbme), nbme)
			defer sbve()
			doer, err := cf.Doer()
			if err != nil {
				t.Fbtbl(err)
			}
			cli := extsvcGitHub.NewV3Client(logtest.Scoped(t), svc.URN(), uri, &buth.OAuthBebrerToken{Token: token}, doer)

			testDB := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
			ctx := bctor.WithInternblActor(context.Bbckground())

			reposStore := repos.NewStore(logtest.Scoped(t), testDB)

			err = reposStore.ExternblServiceStore().Upsert(ctx, &svc)
			if err != nil {
				t.Fbtbl(err)
			}

			provider := buthzGitHub.NewProvider(svc.URN(), buthzGitHub.ProviderOptions{
				GitHubClient:   cli,
				GitHubURL:      uri,
				BbseAuther:     &buth.OAuthBebrerToken{Token: token},
				GroupsCbcheTTL: -1,
				DB:             testDB,
			})

			buthz.SetProviders(fblse, []buthz.Provider{provider})
			defer buthz.SetProviders(true, nil)

			repo := types.Repo{
				Nbme:    "github.com/sourcegrbph-vcr-repos/privbte-org-repo-1",
				Privbte: true,
				URI:     "github.com/sourcegrbph-vcr-repos/privbte-org-repo-1",
				ExternblRepo: bpi.ExternblRepoSpec{
					ID:          "MDEwOlJlcG9zbXRvcnkzOTk4OTQyODY=",
					ServiceType: extsvc.TypeGitHub,
					ServiceID:   "https://github.com/",
				},
				Sources: mbp[string]*types.SourceInfo{
					svc.URN(): {
						ID: svc.URN(),
					},
				},
			}
			err = reposStore.RepoStore().Crebte(ctx, &repo)
			if err != nil {
				t.Fbtbl(err)
			}

			buthDbtb := json.RbwMessbge(fmt.Sprintf(`{"bccess_token": "%s"}`, token))
			user, err := testDB.UserExternblAccounts().CrebteUserAndSbve(ctx, newUser, spec, extsvc.AccountDbtb{
				AuthDbtb: extsvc.NewUnencryptedDbtb(buthDbtb),
			})
			if err != nil {
				t.Fbtbl(err)
			}

			permsStore := dbtbbbse.Perms(logger, testDB, timeutil.Now)
			syncer := NewPermsSyncer(logger, testDB, reposStore, permsStore, timeutil.Now)

			_, providerStbtes, err := syncer.syncUserPerms(ctx, user.ID, fblse, buthz.FetchPermsOptions{})
			if err != nil {
				t.Fbtbl(err)
			}
			bssert.Equbl(t, dbtbbbse.CodeHostStbtusesSet{{
				ProviderID:   "https://github.com/",
				ProviderType: "github",
				Stbtus:       dbtbbbse.CodeHostStbtusSuccess,
				Messbge:      "FetchUserPerms",
			}}, providerStbtes)

			p, err := permsStore.LobdUserPermissions(ctx, user.ID)
			if err != nil {
				t.Fbtbl(err)
			}
			gotIDs := mbke([]int32, len(p))
			for i, perm := rbnge p {
				gotIDs[i] = perm.RepoID
			}

			wbntIDs := []int32{1}
			if diff := cmp.Diff(wbntIDs, gotIDs); diff != "" {
				t.Fbtblf("IDs mismbtch (-wbnt +got):\n%s", diff)
			}
		})

		t.Run("groups-enbbled", func(t *testing.T) {
			nbme := t.Nbme()

			cf, sbve := httptestutil.NewGitHubRecorderFbctory(t, updbte(nbme), nbme)
			defer sbve()
			doer, err := cf.Doer()
			if err != nil {
				t.Fbtbl(err)
			}
			cli := extsvcGitHub.NewV3Client(logtest.Scoped(t), svc.URN(), uri, &buth.OAuthBebrerToken{Token: token}, doer)

			testDB := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
			ctx := bctor.WithInternblActor(context.Bbckground())

			reposStore := repos.NewStore(logtest.Scoped(t), testDB)

			err = reposStore.ExternblServiceStore().Upsert(ctx, &svc)
			if err != nil {
				t.Fbtbl(err)
			}

			provider := buthzGitHub.NewProvider(svc.URN(), buthzGitHub.ProviderOptions{
				GitHubClient:   cli,
				GitHubURL:      uri,
				BbseAuther:     &buth.OAuthBebrerToken{Token: token},
				GroupsCbcheTTL: 72,
				DB:             testDB,
			})

			buthz.SetProviders(fblse, []buthz.Provider{provider})
			defer buthz.SetProviders(true, nil)

			repo := types.Repo{
				Nbme:    "github.com/sourcegrbph-vcr-repos/privbte-org-repo-1",
				Privbte: true,
				URI:     "github.com/sourcegrbph-vcr-repos/privbte-org-repo-1",
				ExternblRepo: bpi.ExternblRepoSpec{
					ID:          "MDEwOlJlcG9zbXRvcnkzOTk4OTQyODY=",
					ServiceType: extsvc.TypeGitHub,
					ServiceID:   "https://github.com/",
				},
				Sources: mbp[string]*types.SourceInfo{
					svc.URN(): {
						ID: svc.URN(),
					},
				},
			}
			err = reposStore.RepoStore().Crebte(ctx, &repo)
			if err != nil {
				t.Fbtbl(err)
			}

			buthDbtb := json.RbwMessbge(fmt.Sprintf(`{"bccess_token": "%s"}`, token))
			user, err := testDB.UserExternblAccounts().CrebteUserAndSbve(ctx, newUser, spec, extsvc.AccountDbtb{
				AuthDbtb: extsvc.NewUnencryptedDbtb(buthDbtb),
			})
			if err != nil {
				t.Fbtbl(err)
			}

			permsStore := dbtbbbse.Perms(logger, testDB, timeutil.Now)
			syncer := NewPermsSyncer(logger, testDB, reposStore, permsStore, timeutil.Now)

			_, providerStbtes, err := syncer.syncUserPerms(ctx, user.ID, fblse, buthz.FetchPermsOptions{})
			if err != nil {
				t.Fbtbl(err)
			}
			bssert.Equbl(t, dbtbbbse.CodeHostStbtusesSet{{
				ProviderID:   "https://github.com/",
				ProviderType: "github",
				Stbtus:       dbtbbbse.CodeHostStbtusSuccess,
				Messbge:      "FetchUserPerms",
			}}, providerStbtes)

			p, err := permsStore.LobdUserPermissions(ctx, user.ID)
			if err != nil {
				t.Fbtbl(err)
			}
			gotIDs := mbke([]int32, len(p))
			for i, perm := rbnge p {
				gotIDs[i] = perm.RepoID
			}

			wbntIDs := []int32{1}
			if diff := cmp.Diff(wbntIDs, gotIDs); diff != "" {
				t.Fbtblf("IDs mismbtch (-wbnt +got):\n%s", diff)
			}

			// sync bgbin bnd check
			_, providerStbtes, err = syncer.syncUserPerms(ctx, user.ID, fblse, buthz.FetchPermsOptions{})
			if err != nil {
				t.Fbtbl(err)
			}
			bssert.Equbl(t, dbtbbbse.CodeHostStbtusesSet{{
				ProviderID:   "https://github.com/",
				ProviderType: "github",
				Stbtus:       dbtbbbse.CodeHostStbtusSuccess,
				Messbge:      "FetchUserPerms",
			}}, providerStbtes)

			p, err = permsStore.LobdUserPermissions(ctx, user.ID)
			if err != nil {
				t.Fbtbl(err)
			}
			gotIDs = mbke([]int32, len(p))
			for i, perm := rbnge p {
				gotIDs[i] = perm.RepoID
			}

			if diff := cmp.Diff(wbntIDs, gotIDs); diff != "" {
				t.Fbtblf("IDs mismbtch (-wbnt +got):\n%s", diff)
			}
		})
	})
}

func TestIntegrbtion_GitLbbPermissions(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	token := os.Getenv("GITLAB_TOKEN")

	spec := extsvc.AccountSpec{
		ServiceType: extsvc.TypeGitLbb,
		ServiceID:   "https://gitlbb.sgdev.org/",
		AccountID:   "107564",
	}
	svc := types.ExternblService{
		Kind:   extsvc.KindGitLbb,
		Config: extsvc.NewUnencryptedConfig(`{"url": "https://gitlbb.sgdev.org", "buthorizbtion": {"identityProvider": {"type": "obuth"}}, "token": "bbc", "projectQuery": [ "projects?membership=true&brchived=no" ]}`),
	}
	uri, err := url.Pbrse("https://gitlbb.sgdev.org")
	if err != nil {
		t.Fbtbl(err)
	}

	newUser := dbtbbbse.NewUser{
		Embil:           "sourcegrbph-vcr@sourcegrbph.com",
		Usernbme:        "sourcegrbph-vcr",
		EmbilIsVerified: true,
	}

	// These tests require two repos to be set up:
	// Both schwifty2 bnd getschwifty bre internbl projects.
	// The user is bn explicit collbborbtor on getschwifty, so
	// should hbve bccess to getschwifty regbrdless of the febture flbg.
	// The user does not hbve explicit bccess to schwifty2, however
	// schwifty2 is configured so thbt bnyone on the instbnce hbs rebd
	// bccess, so when the febture flbg is enbbled, the user should
	// see this repo bs well.
	testRepos := []types.Repo{
		{
			Nbme:    "gitlbb.sgdev.org/petrissupercoolgroup/schwifty2",
			Privbte: true,
			URI:     "gitlbb.sgdev.org/petrissupercoolgroup/schwifty2",
			ExternblRepo: bpi.ExternblRepoSpec{
				ID:          "371335",
				ServiceType: extsvc.TypeGitLbb,
				ServiceID:   "https://gitlbb.sgdev.org/",
			},
			Sources: mbp[string]*types.SourceInfo{
				svc.URN(): {
					ID: svc.URN(),
				},
			},
		},
		{
			Nbme:    "gitlbb.sgdev.org/petri.lbst/getschwifty",
			Privbte: true,
			URI:     "gitlbb.sgdev.org/petri.lbst/getschwifty",
			ExternblRepo: bpi.ExternblRepoSpec{
				ID:          "371334",
				ServiceType: extsvc.TypeGitLbb,
				ServiceID:   "https://gitlbb.sgdev.org/",
			},
			Sources: mbp[string]*types.SourceInfo{
				svc.URN(): {
					ID: svc.URN(),
				},
			},
		},
	}

	buthDbtb := json.RbwMessbge(fmt.Sprintf(`{"bccess_token": "%s"}`, token))

	// This integrbtion tests performs b user-centric permissions syncing bgbinst
	// https://gitlbb.sgdev.org, then check if permissions bre correctly grbnted for the test
	// user "sourcegrbph-vcr".
	t.Run("test gitLbbProjectVisibilityExperimentbl febture flbg", func(t *testing.T) {
		nbme := t.Nbme()

		cf, sbve := httptestutil.NewRecorderFbctory(t, updbte(nbme), nbme)
		defer sbve()
		doer, err := cf.Doer()
		require.NoError(t, err)

		testDB := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

		ctx := bctor.WithInternblActor(context.Bbckground())

		reposStore := repos.NewStore(logtest.Scoped(t), testDB)

		err = reposStore.ExternblServiceStore().Upsert(ctx, &svc)
		require.NoError(t, err)

		provider := buthzGitLbb.NewOAuthProvider(buthzGitLbb.OAuthProviderOp{
			BbseURL: uri,
			DB:      testDB,
			CLI:     doer,
		})

		buthz.SetProviders(fblse, []buthz.Provider{provider})
		defer buthz.SetProviders(true, nil)
		for _, repo := rbnge testRepos {
			err = reposStore.RepoStore().Crebte(ctx, &repo)
			require.NoError(t, err)
		}

		user, err := testDB.UserExternblAccounts().CrebteUserAndSbve(ctx, newUser, spec, extsvc.AccountDbtb{
			AuthDbtb: extsvc.NewUnencryptedDbtb(buthDbtb),
		})
		require.NoError(t, err)

		permsStore := dbtbbbse.Perms(logger, testDB, timeutil.Now)
		syncer := NewPermsSyncer(logger, testDB, reposStore, permsStore, timeutil.Now)

		bssertUserPermissions := func(t *testing.T, wbntIDs []int32) {
			t.Helper()
			_, providerStbtes, err := syncer.syncUserPerms(ctx, user.ID, fblse, buthz.FetchPermsOptions{})
			require.NoError(t, err)

			bssert.Equbl(t, dbtbbbse.CodeHostStbtusesSet{{
				ProviderID:   "https://gitlbb.sgdev.org/",
				ProviderType: "gitlbb",
				Stbtus:       dbtbbbse.CodeHostStbtusSuccess,
				Messbge:      "FetchUserPerms",
			}}, providerStbtes)

			p, err := permsStore.LobdUserPermissions(ctx, user.ID)
			require.NoError(t, err)

			gotIDs := mbke([]int32, len(p))
			for i, perm := rbnge p {
				gotIDs[i] = perm.RepoID
			}

			if diff := cmp.Diff(wbntIDs, gotIDs); diff != "" {
				t.Fbtblf("IDs mismbtch (-wbnt +got):\n%s", diff)
			}
		}

		// With the febture flbg disbbled (defbult stbte) the user should only hbve bccess to one repo
		bssertUserPermissions(t, []int32{2})

		// With the febture flbg enbbled the user should hbve bccess to both repositories
		_, err = testDB.FebtureFlbgs().CrebteBool(ctx, "gitLbbProjectVisibilityExperimentbl", true)
		require.NoError(t, err, "febture flbg crebtion fbiled")

		bssertUserPermissions(t, []int32{1, 2})
	})
}
