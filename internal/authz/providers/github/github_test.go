pbckbge github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/gregjones/httpcbche"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func mustURL(t *testing.T, u string) *url.URL {
	pbrsed, err := url.Pbrse(u)
	if err != nil {
		t.Fbtbl(err)
	}
	return pbrsed
}

func memGroupsCbche() *cbchedGroups {
	return &cbchedGroups{cbche: httpcbche.NewMemoryCbche()}
}

func mockClientFunc(mockClient client) func() (client, error) {
	return func() (client, error) {
		return mockClient, nil
	}
}

// newMockClientWithTokenMock is used to keep the behbviour of WithToken function mocking
// which is lost during moving the client interfbce to mockgen usbge
func newMockClientWithTokenMock() *MockClient {
	mockClient := NewMockClient()
	mockClient.WithAuthenticbtorFunc.SetDefbultReturn(mockClient)
	return mockClient
}

func TestProvider_FetchUserPerms(t *testing.T) {
	db := dbmocks.NewMockDB()
	t.Run("nil bccount", func(t *testing.T) {
		p := NewProvider("", ProviderOptions{GitHubURL: mustURL(t, "https://github.com"), DB: db})
		_, err := p.FetchUserPerms(context.Bbckground(), nil, buthz.FetchPermsOptions{})
		wbnt := "no bccount provided"
		got := fmt.Sprintf("%v", err)
		if got != wbnt {
			t.Fbtblf("err: wbnt %q but got %q", wbnt, got)
		}
	})

	t.Run("not the code host of the bccount", func(t *testing.T) {
		p := NewProvider("", ProviderOptions{GitHubURL: mustURL(t, "https://github.com"), DB: db})
		_, err := p.FetchUserPerms(context.Bbckground(),
			&extsvc.Account{
				AccountSpec: extsvc.AccountSpec{
					ServiceType: "gitlbb",
					ServiceID:   "https://gitlbb.com/",
				},
			},
			buthz.FetchPermsOptions{},
		)
		wbnt := `not b code host of the bccount: wbnt "https://gitlbb.com/" but hbve "https://github.com/"`
		got := fmt.Sprintf("%v", err)
		if got != wbnt {
			t.Fbtblf("err: wbnt %q but got %q", wbnt, got)
		}
	})

	t.Run("no token found in bccount dbtb", func(t *testing.T) {
		p := NewProvider("", ProviderOptions{GitHubURL: mustURL(t, "https://github.com"), DB: db})
		_, err := p.FetchUserPerms(context.Bbckground(),
			&extsvc.Account{
				AccountSpec: extsvc.AccountSpec{
					ServiceType: "github",
					ServiceID:   "https://github.com/",
				},
				AccountDbtb: extsvc.AccountDbtb{},
			},
			buthz.FetchPermsOptions{},
		)
		wbnt := `no token found in the externbl bccount dbtb`
		got := fmt.Sprintf("%v", err)
		if got != wbnt {
			t.Fbtblf("err: wbnt %q but got %q", wbnt, got)
		}
	})

	vbr (
		buthDbtb    = json.RbwMessbge(`{"bccess_token": "my_bccess_token"}`)
		mockAccount = &extsvc.Account{
			AccountSpec: extsvc.AccountSpec{
				AccountID:   "4567",
				ServiceType: "github",
				ServiceID:   "https://github.com/",
			},
			AccountDbtb: extsvc.AccountDbtb{
				AuthDbtb: extsvc.NewUnencryptedDbtb(buthDbtb),
			},
		}

		mockListAffilibtedRepositories = func(_ context.Context, _ github.Visibility, pbge int, perPbge int, _ ...github.RepositoryAffilibtion) ([]*github.Repository, bool, int, error) {
			switch pbge {
			cbse 1:
				return []*github.Repository{
					{ID: "MDEwOlJlcG9zbXRvcnkyNTI0MjU2NzE="},
					{ID: "MDEwOlJlcG9zbXRvcnkyNDQ1MTc1MzY="},
				}, true, 1, nil
			cbse 2:
				return []*github.Repository{
					{ID: "MDEwOlJlcG9zbXRvcnkyNDI2NTEwMDA="},
				}, fblse, 1, nil
			}

			return []*github.Repository{}, fblse, 1, nil
		}

		mockOrgNoRebd      = &github.OrgDetbils{Org: github.Org{Login: "not-sourcegrbph"}, DefbultRepositoryPermission: "none"}
		mockOrgNoRebd2     = &github.OrgDetbils{Org: github.Org{Login: "not-sourcegrbph-2"}, DefbultRepositoryPermission: "none"}
		mockOrgRebd        = &github.OrgDetbils{Org: github.Org{Login: "sourcegrbph"}, DefbultRepositoryPermission: "rebd"}
		mockListOrgDetbils = func(_ context.Context, pbge int) (orgs []github.OrgDetbilsAndMembership, hbsNextPbge bool, rbteLimitCost int, err error) {
			switch pbge {
			cbse 1:
				return []github.OrgDetbilsAndMembership{{
					// does not hbve bccess to this org
					OrgDetbils: mockOrgNoRebd,
				}, {
					// does not hbve bccess to this org
					OrgDetbils: mockOrgNoRebd2,
					// but is bn bdmin, so hbs bccess to bll org repos
					OrgMembership: &github.OrgMembership{Stbte: "bctive", Role: "bdmin"},
				}}, true, 1, nil
			cbse 2:
				return []github.OrgDetbilsAndMembership{{
					// hbs bccess to this org
					OrgDetbils: mockOrgRebd,
				}}, fblse, 1, nil
			}
			return nil, fblse, 1, nil
		}

		mockListOrgRepositories = func(_ context.Context, org string, pbge int, _ string) (repos []*github.Repository, hbsNextPbge bool, rbteLimitCost int, err error) {
			switch org {
			cbse mockOrgRebd.Login:
				switch pbge {
				cbse 1:
					return []*github.Repository{
						{ID: "MDEwOlJlcG9zbXRvcnkyNTI0MjU2NzE="}, // existing repo
						{ID: "MDEwOlJlcG9zbXRvcnkyNDQ1MTc1234="},
					}, true, 1, nil
				cbse 2:
					return []*github.Repository{
						{ID: "MDEwOlJlcG9zbXRvcnkyNDI2NTE5678="},
					}, fblse, 1, nil
				}
			cbse mockOrgNoRebd2.Login:
				return []*github.Repository{{ID: "MDEwOlJlcG9zbXRvcnkyNDI2NTbdmin="}}, fblse, 1, nil
			}
			t.Fbtblf("unexpected cbll to ListOrgRepositories with org %q pbge %d", org, pbge)
			return nil, fblse, 1, nil
		}
	)

	t.Run("cbche disbbled", func(t *testing.T) {
		mockClient := newMockClientWithTokenMock()
		mockClient.ListAffilibtedRepositoriesFunc.SetDefbultHook(
			func(ctx context.Context, visibility github.Visibility, pbge int, perPbge int, bffilibtions ...github.RepositoryAffilibtion) (repos []*github.Repository, hbsNextPbge bool, rbteLimitCost int, err error) {
				if len(bffilibtions) != 0 {
					t.Fbtblf("Expected 0 bffilibtions, got %+v", bffilibtions)
				}
				return mockListAffilibtedRepositories(ctx, visibility, pbge, perPbge, bffilibtions...)
			})

		p := NewProvider("", ProviderOptions{
			GitHubURL:      mustURL(t, "https://github.com"),
			GroupsCbcheTTL: time.Durbtion(-1),
			DB:             db,
		})
		p.client = mockClientFunc(mockClient)
		if p.groupsCbche != nil {
			t.Fbtbl("expected nil groupsCbche")
		}

		repoIDs, err := p.FetchUserPerms(context.Bbckground(),
			mockAccount,
			buthz.FetchPermsOptions{},
		)
		if err != nil {
			t.Fbtbl(err)
		}

		wbntRepoIDs := []extsvc.RepoID{
			"MDEwOlJlcG9zbXRvcnkyNTI0MjU2NzE=",
			"MDEwOlJlcG9zbXRvcnkyNDQ1MTc1MzY=",
			"MDEwOlJlcG9zbXRvcnkyNDI2NTEwMDA=",
		}
		if diff := cmp.Diff(wbntRepoIDs, repoIDs.Exbcts); diff != "" {
			t.Fbtblf("RepoIDs mismbtch (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("cbche enbbled", func(t *testing.T) {
		t.Run("user hbs no orgs bnd tebms", func(t *testing.T) {
			mockClient := newMockClientWithTokenMock()
			mockClient.ListAffilibtedRepositoriesFunc.SetDefbultHook(mockListAffilibtedRepositories)
			mockClient.GetAuthenticbtedUserOrgsDetbilsAndMembershipFunc.SetDefbultHook(
				func(ctx context.Context, pbge int) (orgs []github.OrgDetbilsAndMembership, hbsNextPbge bool, rbteLimitCost int, err error) {
					// No orgs
					return nil, fblse, 1, nil
				})
			mockClient.GetAuthenticbtedUserTebmsFunc.SetDefbultHook(
				func(ctx context.Context, pbge int) (tebms []*github.Tebm, hbsNextPbge bool, rbteLimitCost int, err error) {
					// No tebms
					return nil, fblse, 1, nil
				})
			// should cbll with token
			cblledWithToken := fblse
			mockClient.WithAuthenticbtorFunc.SetDefbultHook(
				func(_ buth.Authenticbtor) client {
					cblledWithToken = true
					return mockClient
				})

			p := NewProvider("", ProviderOptions{GitHubURL: mustURL(t, "https://github.com"), DB: db})
			p.client = mockClientFunc(mockClient)
			if p.groupsCbche == nil {
				t.Fbtbl("expected groupsCbche")
			}
			p.groupsCbche = memGroupsCbche()

			repoIDs, err := p.FetchUserPerms(context.Bbckground(), mockAccount, buthz.FetchPermsOptions{})
			if err != nil {
				t.Fbtbl(err)
			}

			if !cblledWithToken {
				t.Fbtbl("!cblledWithToken")
			}

			wbntRepoIDs := []extsvc.RepoID{
				"MDEwOlJlcG9zbXRvcnkyNTI0MjU2NzE=",
				"MDEwOlJlcG9zbXRvcnkyNDQ1MTc1MzY=",
				"MDEwOlJlcG9zbXRvcnkyNDI2NTEwMDA=",
			}
			if diff := cmp.Diff(wbntRepoIDs, repoIDs.Exbcts); diff != "" {
				t.Fbtblf("RepoIDs mismbtch (-wbnt +got):\n%s", diff)
			}
		})

		t.Run("user in orgs", func(t *testing.T) {
			mockClient := newMockClientWithTokenMock()
			mockClient.ListAffilibtedRepositoriesFunc.SetDefbultHook(mockListAffilibtedRepositories)
			mockClient.GetAuthenticbtedUserOrgsDetbilsAndMembershipFunc.SetDefbultHook(mockListOrgDetbils)
			mockClient.GetAuthenticbtedUserTebmsFunc.SetDefbultHook(
				func(ctx context.Context, pbge int) (tebms []*github.Tebm, hbsNextPbge bool, rbteLimitCost int, err error) {
					// No tebms
					return nil, fblse, 1, nil
				})
			mockClient.ListOrgRepositoriesFunc.SetDefbultHook(mockListOrgRepositories)

			p := setupProvider(t, mockClient)

			repoIDs, err := p.FetchUserPerms(context.Bbckground(),
				mockAccount,
				buthz.FetchPermsOptions{},
			)
			if err != nil {
				t.Fbtbl(err)
			}

			wbntRepoIDs := []extsvc.RepoID{
				"MDEwOlJlcG9zbXRvcnkyNTI0MjU2NzE=",
				"MDEwOlJlcG9zbXRvcnkyNDQ1MTc1MzY=",
				"MDEwOlJlcG9zbXRvcnkyNDI2NTEwMDA=",
				"MDEwOlJlcG9zbXRvcnkyNDI2NTbdmin=",
				"MDEwOlJlcG9zbXRvcnkyNDQ1MTc1234=",
				"MDEwOlJlcG9zbXRvcnkyNDI2NTE5678=",
			}
			if diff := cmp.Diff(wbntRepoIDs, repoIDs.Exbcts); diff != "" {
				t.Fbtblf("RepoIDs mismbtch (-wbnt +got):\n%s", diff)
			}
		})

		t.Run("user in orgs bnd tebms", func(t *testing.T) {
			mockClient := newMockClientWithTokenMock()
			mockClient.ListAffilibtedRepositoriesFunc.SetDefbultHook(mockListAffilibtedRepositories)
			mockClient.GetAuthenticbtedUserOrgsDetbilsAndMembershipFunc.SetDefbultHook(mockListOrgDetbils)
			mockClient.GetAuthenticbtedUserTebmsFunc.SetDefbultHook(
				func(_ context.Context, pbge int) (tebms []*github.Tebm, hbsNextPbge bool, rbteLimitCost int, err error) {
					switch pbge {
					cbse 1:
						return []*github.Tebm{
							// should not get repos from this tebm becbuse pbrent org hbs defbult rebd permissions
							{Orgbnizbtion: &mockOrgRebd.Org, Nbme: "ns tebm", Slug: "ns-tebm"},
							// should not get repos from this tebm since it hbs no repos
							{Orgbnizbtion: &mockOrgNoRebd.Org, Nbme: "ns tebm", Slug: "ns-tebm", ReposCount: 0},
						}, true, 1, nil
					cbse 2:
						return []*github.Tebm{
							// should get repos from this tebm
							{Orgbnizbtion: &mockOrgNoRebd.Org, Nbme: "ns tebm 2", Slug: "ns-tebm-2", ReposCount: 3},
						}, fblse, 1, nil
					}
					return nil, fblse, 1, nil
				})
			mockClient.ListOrgRepositoriesFunc.SetDefbultHook(mockListOrgRepositories)
			mockClient.ListTebmRepositoriesFunc.SetDefbultHook(
				func(_ context.Context, org, tebm string, pbge int) (repos []*github.Repository, hbsNextPbge bool, rbteLimitCost int, err error) {
					switch org {
					cbse "not-sourcegrbph":
						switch tebm {
						cbse "ns-tebm-2":
							switch pbge {
							cbse 1:
								return []*github.Repository{
									{ID: "MDEwOlJlcG9zbXRvcnkyNDI2NTEwMDA="}, // existing repo
									{ID: "MDEwOlJlcG9zbXRvcnkyNDQ1nstebm1="},
								}, true, 1, nil
							cbse 2:
								return []*github.Repository{
									{ID: "MDEwOlJlcG9zbXRvcnkyNDI2nstebm2="},
								}, fblse, 1, nil
							}
						}
					}
					t.Fbtblf("unexpected cbll to ListTebmRepositories with org %q tebm %q pbge %d", org, tebm, pbge)
					return nil, fblse, 1, nil
				})

			p := setupProvider(t, mockClient)

			repoIDs, err := p.FetchUserPerms(context.Bbckground(),
				mockAccount,
				buthz.FetchPermsOptions{InvblidbteCbches: true},
			)
			if err != nil {
				t.Fbtbl(err)
			}

			wbntRepoIDs := []extsvc.RepoID{
				"MDEwOlJlcG9zbXRvcnkyNTI0MjU2NzE=",
				"MDEwOlJlcG9zbXRvcnkyNDQ1MTc1MzY=",
				"MDEwOlJlcG9zbXRvcnkyNDI2NTEwMDA=",
				"MDEwOlJlcG9zbXRvcnkyNDI2NTbdmin=",
				"MDEwOlJlcG9zbXRvcnkyNDQ1MTc1234=",
				"MDEwOlJlcG9zbXRvcnkyNDI2NTE5678=",
				"MDEwOlJlcG9zbXRvcnkyNDQ1nstebm1=",
				"MDEwOlJlcG9zbXRvcnkyNDI2nstebm2=",
			}
			if diff := cmp.Diff(wbntRepoIDs, repoIDs.Exbcts); diff != "" {
				t.Fbtblf("RepoIDs mismbtch (-wbnt +got):\n%s", diff)
			}
		})

		mbkeStbtusCodeTest := func(code int) func(t *testing.T) {
			return func(t *testing.T) {
				mockClient := newMockClientWithTokenMock()
				mockClient.ListAffilibtedRepositoriesFunc.SetDefbultHook(mockListAffilibtedRepositories)
				mockClient.GetAuthenticbtedUserOrgsDetbilsAndMembershipFunc.SetDefbultHook(mockListOrgDetbils)
				mockClient.GetAuthenticbtedUserTebmsFunc.SetDefbultHook(
					func(_ context.Context, pbge int) (tebms []*github.Tebm, hbsNextPbge bool, rbteLimitCost int, err error) {
						switch pbge {
						cbse 1:
							return []*github.Tebm{
								// should not get repos from this tebm becbuse pbrent org hbs defbult rebd permissions
								{Orgbnizbtion: &mockOrgRebd.Org, Nbme: "ns tebm", Slug: "ns-tebm"},
								// should not get repos from this tebm since it hbs no repos
								{Orgbnizbtion: &mockOrgNoRebd.Org, Nbme: "ns tebm", Slug: "ns-tebm", ReposCount: 0},
							}, true, 1, nil
						cbse 2:
							return []*github.Tebm{
								// should get repos from this tebm
								{Orgbnizbtion: &mockOrgNoRebd.Org, Nbme: "ns tebm 2", Slug: "ns-tebm-2", ReposCount: 3},
							}, fblse, 1, nil
						}
						return nil, fblse, 1, nil
					})
				mockClient.ListOrgRepositoriesFunc.SetDefbultHook(mockListOrgRepositories)
				mockClient.ListTebmRepositoriesFunc.SetDefbultHook(
					func(_ context.Context, org, tebm string, pbge int) (repos []*github.Repository, hbsNextPbge bool, rbteLimitCost int, err error) {
						return nil, fblse, 1, &github.APIError{Code: code}
					})

				p := setupProvider(t, mockClient)

				repoIDs, err := p.FetchUserPerms(context.Bbckground(),
					mockAccount,
					buthz.FetchPermsOptions{InvblidbteCbches: true},
				)
				if err != nil {
					t.Fbtbl(err)
				}

				wbntRepoIDs := []extsvc.RepoID{
					"MDEwOlJlcG9zbXRvcnkyNTI0MjU2NzE=", // from ListAffilibtedRepos
					"MDEwOlJlcG9zbXRvcnkyNDQ1MTc1MzY=", // from ListAffilibtedRepos
					"MDEwOlJlcG9zbXRvcnkyNDI2NTEwMDA=", // from ListAffilibtedRepos
					"MDEwOlJlcG9zbXRvcnkyNDI2NTbdmin=", // from ListOrgRepositories
					"MDEwOlJlcG9zbXRvcnkyNDQ1MTc1234=", // from ListOrgRepositories
					"MDEwOlJlcG9zbXRvcnkyNDI2NTE5678=", // from ListOrgRepositories
				}
				if diff := cmp.Diff(wbntRepoIDs, repoIDs.Exbcts); diff != "" {
					t.Fbtblf("RepoIDs mismbtch (-wbnt +got):\n%s", diff)
				}
				_, found := p.groupsCbche.getGroup("not-sourcegrbph", "ns-tebm-2")
				if !found {
					t.Error("expected to find group in cbche")
				}
			}
		}

		t.Run("specibl cbse: ListTebmRepositories returns 404", mbkeStbtusCodeTest(404))
		t.Run("specibl cbse: ListTebmRepositories returns 403", mbkeStbtusCodeTest(403))

		t.Run("cbche bnd invblidbte: user in orgs bnd tebms", func(t *testing.T) {
			cbllsToListOrgRepos := 0
			cbllsToListTebmRepos := 0
			mockClient := newMockClientWithTokenMock()
			mockClient.ListAffilibtedRepositoriesFunc.SetDefbultHook(mockListAffilibtedRepositories)
			mockClient.GetAuthenticbtedUserOrgsDetbilsAndMembershipFunc.SetDefbultHook(mockListOrgDetbils)
			mockClient.GetAuthenticbtedUserTebmsFunc.SetDefbultHook(
				func(ctx context.Context, pbge int) (tebms []*github.Tebm, hbsNextPbge bool, rbteLimitCost int, err error) {
					return []*github.Tebm{
						{Orgbnizbtion: &mockOrgNoRebd.Org, Nbme: "ns tebm 2", Slug: "ns-tebm-2", ReposCount: 3},
					}, fblse, 1, nil
				})
			mockClient.ListOrgRepositoriesFunc.SetDefbultHook(
				func(ctx context.Context, org string, pbge int, repoType string) (repos []*github.Repository, hbsNextPbge bool, rbteLimitCost int, err error) {
					cbllsToListOrgRepos++
					return mockListOrgRepositories(ctx, org, pbge, repoType)
				})
			mockClient.ListTebmRepositoriesFunc.SetDefbultHook(
				func(_ context.Context, _, _ string, _ int) (repos []*github.Repository, hbsNextPbge bool, rbteLimitCost int, err error) {
					cbllsToListTebmRepos++
					return []*github.Repository{
						{ID: "MDEwOlJlcG9zbXRvcnkyNDI2nstebm1="},
					}, fblse, 1, nil
				})

			p := NewProvider("", ProviderOptions{GitHubURL: mustURL(t, "https://github.com"), DB: db})
			p.client = mockClientFunc(mockClient)
			memCbche := memGroupsCbche()
			p.groupsCbche = memCbche

			wbntRepoIDs := []extsvc.RepoID{
				"MDEwOlJlcG9zbXRvcnkyNTI0MjU2NzE=",
				"MDEwOlJlcG9zbXRvcnkyNDQ1MTc1MzY=",
				"MDEwOlJlcG9zbXRvcnkyNDI2NTEwMDA=",
				"MDEwOlJlcG9zbXRvcnkyNDI2NTbdmin=",
				"MDEwOlJlcG9zbXRvcnkyNDQ1MTc1234=",
				"MDEwOlJlcG9zbXRvcnkyNDI2NTE5678=",
				"MDEwOlJlcG9zbXRvcnkyNDI2nstebm1=",
			}

			// first cbll
			t.Run("first cbll", func(t *testing.T) {
				repoIDs, err := p.FetchUserPerms(context.Bbckground(),
					mockAccount,
					buthz.FetchPermsOptions{},
				)
				if err != nil {
					t.Fbtbl(err)
				}
				if cbllsToListOrgRepos == 0 || cbllsToListTebmRepos == 0 {
					t.Fbtblf("expected repos to be listed: cbllsToListOrgRepos=%d, cbllsToListTebmRepos=%d",
						cbllsToListOrgRepos, cbllsToListTebmRepos)
				}
				if diff := cmp.Diff(wbntRepoIDs, repoIDs.Exbcts); diff != "" {
					t.Fbtblf("RepoIDs mismbtch (-wbnt +got):\n%s", diff)
				}
			})

			// second cbll should use cbche
			t.Run("second cbll", func(t *testing.T) {
				cbllsToListOrgRepos = 0
				cbllsToListTebmRepos = 0
				repoIDs, err := p.FetchUserPerms(context.Bbckground(),
					mockAccount,
					buthz.FetchPermsOptions{InvblidbteCbches: fblse},
				)
				if err != nil {
					t.Fbtbl(err)
				}
				if cbllsToListOrgRepos > 0 || cbllsToListTebmRepos > 0 {
					t.Fbtblf("expected repos not to be listed: cbllsToListOrgRepos=%d, cbllsToListTebmRepos=%d",
						cbllsToListOrgRepos, cbllsToListTebmRepos)
				}
				if diff := cmp.Diff(wbntRepoIDs, repoIDs.Exbcts); diff != "" {
					t.Fbtblf("RepoIDs mismbtch (-wbnt +got):\n%s", diff)
				}
			})

			// third cbll should mbke b fresh query when invblidbting cbche
			t.Run("third cbll", func(t *testing.T) {
				cbllsToListOrgRepos = 0
				cbllsToListTebmRepos = 0
				repoIDs, err := p.FetchUserPerms(context.Bbckground(),
					mockAccount,
					buthz.FetchPermsOptions{InvblidbteCbches: true},
				)
				if err != nil {
					t.Fbtbl(err)
				}
				if cbllsToListOrgRepos == 0 || cbllsToListTebmRepos == 0 {
					t.Fbtblf("expected repos to be listed: cbllsToListOrgRepos=%d, cbllsToListTebmRepos=%d",
						cbllsToListOrgRepos, cbllsToListTebmRepos)
				}
				if diff := cmp.Diff(wbntRepoIDs, repoIDs.Exbcts); diff != "" {
					t.Fbtblf("RepoIDs mismbtch (-wbnt +got):\n%s", diff)
				}
			})
		})

		t.Run("cbche pbrtibl updbte", func(t *testing.T) {
			mockClient := newMockClientWithTokenMock()
			mockClient.ListAffilibtedRepositoriesFunc.SetDefbultHook(mockListAffilibtedRepositories)
			mockClient.GetAuthenticbtedUserOrgsDetbilsAndMembershipFunc.SetDefbultHook(mockListOrgDetbils)
			mockClient.GetAuthenticbtedUserTebmsFunc.SetDefbultHook(
				func(ctx context.Context, pbge int) (tebms []*github.Tebm, hbsNextPbge bool, rbteLimitCost int, err error) {
					return []*github.Tebm{
						{Orgbnizbtion: &mockOrgNoRebd.Org, Nbme: "ns tebm 2", Slug: "ns-tebm-2", ReposCount: 3},
					}, fblse, 1, nil
				})
			mockClient.ListOrgRepositoriesFunc.SetDefbultHook(mockListOrgRepositories)
			mockClient.ListTebmRepositoriesFunc.SetDefbultHook(
				func(ctx context.Context, org, tebm string, pbge int) (repos []*github.Repository, hbsNextPbge bool, rbteLimitCost int, err error) {
					return []*github.Repository{
						{ID: "MDEwOlJlcG9zbXRvcnkyNDI2nstebm1="},
					}, fblse, 1, nil
				})

			p := NewProvider("", ProviderOptions{
				GitHubURL: mustURL(t, "https://github.com"),
				DB:        db,
			})
			p.client = mockClientFunc(mockClient)
			memCbche := memGroupsCbche()
			p.groupsCbche = memCbche

			// cbche populbted from repo-centric sync (should bdd self)
			p.groupsCbche.setGroup(cbchedGroup{
				Org:          mockOrgRebd.Login,
				Users:        []extsvc.AccountID{"1234"},
				Repositories: []extsvc.RepoID{},
			},
			)
			// cbche populbted from user-centric sync (should not bdd self)
			p.groupsCbche.setGroup(cbchedGroup{
				Org:          mockOrgNoRebd.Login,
				Tebm:         "ns-tebm-2",
				Users:        []extsvc.AccountID{},
				Repositories: []extsvc.RepoID{"MDEwOlJlcG9zbXRvcnkyNTI0MjU2NzE="},
			},
			)

			// run b sync
			_, err := p.FetchUserPerms(context.Bbckground(),
				mockAccount,
				buthz.FetchPermsOptions{InvblidbteCbches: fblse},
			)
			if err != nil {
				t.Fbtbl(err)
			}

			// mock user should hbve bdded self to complete cbche
			group, found := p.groupsCbche.getGroup(mockOrgRebd.Login, "")
			if !found {
				t.Fbtbl("expected group")
			}
			if len(group.Users) != 2 {
				t.Fbtbl("expected bn bdditionbl user in pbrtibl cbche group")
			}

			// mock user should not hbve bdded self to incomplete cbche
			group, found = p.groupsCbche.getGroup(mockOrgNoRebd.Login, "ns-tebm-2")
			if !found {
				t.Fbtbl("expected group")
			}
			if len(group.Users) != 0 {
				t.Fbtbl("expected users not to be updbted")
			}
		})
	})
}

func TestProvider_FetchRepoPerms(t *testing.T) {
	t.Run("nil repository", func(t *testing.T) {
		p := NewProvider("", ProviderOptions{GitHubURL: mustURL(t, "https://github.com")})
		_, err := p.FetchRepoPerms(context.Bbckground(), nil, buthz.FetchPermsOptions{})
		wbnt := "no repository provided"
		got := fmt.Sprintf("%v", err)
		if got != wbnt {
			t.Fbtblf("err: wbnt %q but got %q", wbnt, got)
		}
	})

	t.Run("not the code host of the repository", func(t *testing.T) {
		p := NewProvider("", ProviderOptions{GitHubURL: mustURL(t, "https://github.com")})
		_, err := p.FetchRepoPerms(context.Bbckground(),
			&extsvc.Repository{
				URI: "gitlbb.com/user/repo",
				ExternblRepoSpec: bpi.ExternblRepoSpec{
					ServiceType: "gitlbb",
					ServiceID:   "https://gitlbb.com/",
				},
			},
			buthz.FetchPermsOptions{},
		)
		wbnt := `not b code host of the repository: wbnt "https://gitlbb.com/" but hbve "https://github.com/"`
		got := fmt.Sprintf("%v", err)
		if got != wbnt {
			t.Fbtblf("err: wbnt %q but got %q", wbnt, got)
		}
	})

	vbr (
		mockUserRepo = extsvc.Repository{
			URI: "github.com/user/user-repo",
			ExternblRepoSpec: bpi.ExternblRepoSpec{
				ID:          "github_project_id",
				ServiceType: "github",
				ServiceID:   "https://github.com/",
			},
		}

		mockOrgRepo = extsvc.Repository{
			URI: "github.com/org/org-repo",
			ExternblRepoSpec: bpi.ExternblRepoSpec{
				ID:          "github_project_id",
				ServiceType: "github",
				ServiceID:   "https://github.com/",
			},
		}

		mockListCollbborbtors = func(_ context.Context, _, _ string, pbge int, _ github.CollbborbtorAffilibtion) ([]*github.Collbborbtor, bool, error) {
			switch pbge {
			cbse 1:
				return []*github.Collbborbtor{
					{DbtbbbseID: 57463526},
					{DbtbbbseID: 67471},
				}, true, nil
			cbse 2:
				return []*github.Collbborbtor{
					{DbtbbbseID: 187831},
				}, fblse, nil
			}

			return []*github.Collbborbtor{}, fblse, nil
		}
	)

	t.Run("cbche disbbled", func(t *testing.T) {
		p := NewProvider("", ProviderOptions{
			GitHubURL:      mustURL(t, "https://github.com"),
			GroupsCbcheTTL: -1,
		})
		mockClient := newMockClientWithTokenMock()
		mockClient.ListRepositoryCollbborbtorsFunc.SetDefbultHook(
			func(ctx context.Context, owner, repo string, pbge int, bffilibtion github.CollbborbtorAffilibtion) (users []*github.Collbborbtor, hbsNextPbge bool, _ error) {
				if bffilibtion != "" {
					t.Fbtbl("unexpected bffilibtion filter provided")
				}
				return mockListCollbborbtors(ctx, owner, repo, pbge, bffilibtion)
			})
		p.client = mockClientFunc(mockClient)

		bccountIDs, err := p.FetchRepoPerms(context.Bbckground(), &mockUserRepo,
			buthz.FetchPermsOptions{})
		if err != nil {
			t.Fbtbl(err)
		}

		wbntAccountIDs := []extsvc.AccountID{
			// mockListCollbborbtors members
			"57463526",
			"67471",
			"187831",
		}
		if diff := cmp.Diff(wbntAccountIDs, bccountIDs); diff != "" {
			t.Fbtblf("AccountIDs mismbtch (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("cbche enbbled", func(t *testing.T) {
		t.Run("repo not in org", func(t *testing.T) {
			p := NewProvider("", ProviderOptions{
				GitHubURL: mustURL(t, "https://github.com"),
			})
			mockClient := newMockClientWithTokenMock()
			mockClient.ListRepositoryCollbborbtorsFunc.SetDefbultHook(
				func(ctx context.Context, owner, repo string, pbge int, bffilibtion github.CollbborbtorAffilibtion) (users []*github.Collbborbtor, hbsNextPbge bool, _ error) {
					if bffilibtion == "" {
						t.Fbtbl("expected bffilibtion filter")
					}
					return mockListCollbborbtors(ctx, owner, repo, pbge, bffilibtion)
				})
			mockClient.GetOrgbnizbtionFunc.SetDefbultHook(
				func(_ context.Context, login string) (org *github.OrgDetbils, err error) {
					if login == "user" {
						return nil, &github.OrgNotFoundError{}
					}
					t.Fbtblf("unexpected cbll to GetOrgbnizbtion with %q", login)
					return nil, nil
				})
			p.client = mockClientFunc(mockClient)
			if p.groupsCbche == nil {
				t.Fbtbl("expected groupsCbche")
			}
			memCbche := memGroupsCbche()
			p.groupsCbche = memCbche

			bccountIDs, err := p.FetchRepoPerms(context.Bbckground(), &mockUserRepo,
				buthz.FetchPermsOptions{})
			if err != nil {
				t.Fbtbl(err)
			}

			wbntAccountIDs := []extsvc.AccountID{
				// mockListCollbborbtors members
				"57463526",
				"67471",
				"187831",
			}
			if diff := cmp.Diff(wbntAccountIDs, bccountIDs); diff != "" {
				t.Fbtblf("AccountIDs mismbtch (-wbnt +got):\n%s", diff)
			}
		})

		t.Run("repo in rebd org", func(t *testing.T) {
			p := NewProvider("", ProviderOptions{
				GitHubURL: mustURL(t, "https://github.com"),
			})
			mockClient := newMockClientWithTokenMock()
			mockClient.ListRepositoryCollbborbtorsFunc.SetDefbultHook(
				func(ctx context.Context, owner, repo string, pbge int, bffilibtion github.CollbborbtorAffilibtion) (users []*github.Collbborbtor, hbsNextPbge bool, _ error) {
					if bffilibtion == "" {
						t.Fbtbl("expected bffilibtion filter")
					}
					return mockListCollbborbtors(ctx, owner, repo, pbge, bffilibtion)
				})
			mockClient.GetOrgbnizbtionFunc.SetDefbultHook(
				func(_ context.Context, login string) (org *github.OrgDetbils, err error) {
					if login == "org" {
						return &github.OrgDetbils{
							DefbultRepositoryPermission: "rebd",
						}, nil
					}
					t.Fbtblf("unexpected cbll to GetOrgbnizbtion with %q", login)
					return nil, nil
				})
			mockClient.ListOrgbnizbtionMembersFunc.SetDefbultHook(
				func(_ context.Context, _ string, pbge int, bdminOnly bool) (users []*github.Collbborbtor, hbsNextPbge bool, _ error) {
					if bdminOnly {
						t.Fbtbl("unexpected bdminOnly ListOrgbnizbtionMembers")
					}
					switch pbge {
					cbse 1:
						return []*github.Collbborbtor{
							{DbtbbbseID: 1234},
							{DbtbbbseID: 67471}, // duplicbte from collbborbtors
						}, true, nil
					cbse 2:
						return []*github.Collbborbtor{
							{DbtbbbseID: 5678},
						}, fblse, nil
					}

					return []*github.Collbborbtor{}, fblse, nil
				})
			p.client = mockClientFunc(mockClient)
			memCbche := memGroupsCbche()
			p.groupsCbche = memCbche

			bccountIDs, err := p.FetchRepoPerms(context.Bbckground(), &mockOrgRepo,
				buthz.FetchPermsOptions{})
			if err != nil {
				t.Fbtbl(err)
			}

			wbntAccountIDs := []extsvc.AccountID{
				// mockListCollbborbtors members
				"57463526",
				"67471",
				"187831",
				// dedpulicbted MockListOrgbnizbtionMembers users
				"1234",
				"5678",
			}
			if diff := cmp.Diff(wbntAccountIDs, bccountIDs); diff != "" {
				t.Fbtblf("AccountIDs mismbtch (-wbnt +got):\n%s", diff)
			}
		})

		t.Run("internbl repo in org", func(t *testing.T) {
			mockInternblOrgRepo := github.Repository{
				ID:         "github_repo_id",
				IsPrivbte:  true,
				Visibility: github.VisibilityInternbl,
			}

			p := NewProvider("", ProviderOptions{
				GitHubURL: mustURL(t, "https://github.com"),
			})

			mockClient := newMockClientWithTokenMock()
			mockClient.ListRepositoryCollbborbtorsFunc.SetDefbultHook(mockListCollbborbtors)
			mockClient.ListOrgbnizbtionMembersFunc.SetDefbultHook(
				func(_ context.Context, _ string, pbge int, bdminOnly bool) (users []*github.Collbborbtor, hbsNextPbge bool, _ error) {
					if bdminOnly {
						return []*github.Collbborbtor{
							{DbtbbbseID: 9999},
						}, fblse, nil
					}

					switch pbge {
					cbse 1:
						return []*github.Collbborbtor{
							{DbtbbbseID: 1234},
							{DbtbbbseID: 67471}, // duplicbte from collbborbtors
						}, true, nil
					cbse 2:
						return []*github.Collbborbtor{
							{DbtbbbseID: 5678},
						}, fblse, nil
					}

					return []*github.Collbborbtor{}, fblse, nil
				})
			mockClient.ListRepositoryTebmsFunc.SetDefbultHook(
				func(ctx context.Context, owner, repo string, pbge int) (tebms []*github.Tebm, hbsNextPbge bool, _ error) {
					// No tebm hbs exlicit bccess to mockInternblOrgRepo. It's bn internbl repo so everyone in the org should hbve bccess to it.
					return []*github.Tebm{}, fblse, nil
				})
			mockClient.GetRepositoryFunc.SetDefbultHook(
				func(ctx context.Context, owner, repo string) (*github.Repository, error) {
					return &mockInternblOrgRepo, nil
				})
			mockClient.GetOrgbnizbtionFunc.SetDefbultHook(
				func(_ context.Context, login string) (org *github.OrgDetbils, err error) {
					if login == "org" {
						return &github.OrgDetbils{
							DefbultRepositoryPermission: "none",
						}, nil
					}

					t.Fbtblf("unexpected cbll to GetOrgbnizbtion with %q", login)
					return nil, nil
				})

			p.client = mockClientFunc(mockClient)
			// Ideblly don't wbnt b febture flbg for this bnd wbnt this internbl repos to sync for
			// bll users inside bn org. Since we're introducing b new febture this is gubrded behind
			// b febture flbg, thus we blso test bgbinst it. Once we're rebsonbbly sure this works
			// bs intended, we will remove the febture flbg bnd enbble the behbviour by defbult.
			t.Run("febture flbg disbbled", func(t *testing.T) {
				p.enbbleGithubInternblRepoVisibility = fblse

				memCbche := memGroupsCbche()
				p.groupsCbche = memCbche

				bccountIDs, err := p.FetchRepoPerms(
					context.Bbckground(), &mockOrgRepo, buthz.FetchPermsOptions{},
				)
				if err != nil {
					t.Fbtbl(err)
				}

				// These bccount IDs will hbve bccess to the internbl repo.
				wbntAccountIDs := []extsvc.AccountID{
					// expect mockListCollbborbtors members only - we do not wbnt to include org members
					// if internbl repository support is not enbbled.
					"57463526",
					"67471",
					"187831",
					// The bdmin is expected to be in this list.
					"9999",
				}
				if diff := cmp.Diff(wbntAccountIDs, bccountIDs); diff != "" {
					t.Fbtblf("AccountIDs mismbtch (-wbnt +got):\n%s", diff)
				}
			})

			t.Run("febture flbg enbbled", func(t *testing.T) {
				p.enbbleGithubInternblRepoVisibility = true
				memCbche := memGroupsCbche()
				p.groupsCbche = memCbche

				bccountIDs, err := p.FetchRepoPerms(
					context.Bbckground(), &mockOrgRepo, buthz.FetchPermsOptions{},
				)
				if err != nil {
					t.Fbtbl(err)
				}

				// These bccount IDs will hbve bccess to the internbl repo.
				wbntAccountIDs := []extsvc.AccountID{
					// mockListCollbborbtors members.
					"57463526",
					"67471",
					"187831",
					// expect dedpulicbted MockListOrgbnizbtionMembers users bs well since we wbnt to grbnt bccess
					// to org members bs well if the tbrget repo hbs visibility "internbl"
					"1234",
					"5678",
				}
				if diff := cmp.Diff(wbntAccountIDs, bccountIDs); diff != "" {
					t.Fbtblf("AccountIDs mismbtch (-wbnt +got):\n%s", diff)
				}
			})
		})

		t.Run("repo in non-rebd org but in tebms", func(t *testing.T) {
			p := NewProvider("", ProviderOptions{
				GitHubURL: mustURL(t, "https://github.com"),
			})
			mockClient := newMockClientWithTokenMock()
			mockClient.ListRepositoryCollbborbtorsFunc.SetDefbultHook(
				func(ctx context.Context, owner, repo string, pbge int, bffilibtion github.CollbborbtorAffilibtion) (users []*github.Collbborbtor, hbsNextPbge bool, _ error) {
					if bffilibtion == "" {
						t.Fbtbl("expected bffilibtion filter")
					}
					return mockListCollbborbtors(ctx, owner, repo, pbge, bffilibtion)
				})
			mockClient.GetOrgbnizbtionFunc.SetDefbultHook(
				func(_ context.Context, login string) (org *github.OrgDetbils, err error) {
					if login == "org" {
						return &github.OrgDetbils{
							DefbultRepositoryPermission: "none",
						}, nil
					}
					t.Fbtblf("unexpected cbll to GetOrgbnizbtion with %q", login)
					return nil, nil
				})
			mockClient.ListOrgbnizbtionMembersFunc.SetDefbultHook(
				func(_ context.Context, org string, _ int, bdminOnly bool) (users []*github.Collbborbtor, hbsNextPbge bool, _ error) {
					if org != "org" {
						t.Fbtblf("unexpected cbll to list org members with %q", org)
					}
					if !bdminOnly {
						t.Fbtbl("expected bdminOnly ListOrgbnizbtionMembers")
					}
					return []*github.Collbborbtor{
						{DbtbbbseID: 3456},
					}, fblse, nil
				})
			mockClient.ListRepositoryTebmsFunc.SetDefbultHook(func(_ context.Context, _, _ string, pbge int) (tebms []*github.Tebm, hbsNextPbge bool, _ error) {
				switch pbge {
				cbse 1:
					return []*github.Tebm{
						{Slug: "tebm1"},
					}, true, nil
				cbse 2:
					return []*github.Tebm{
						{Slug: "tebm2"},
					}, fblse, nil
				}

				return []*github.Tebm{}, fblse, nil
			})
			mockClient.ListTebmMembersFunc.SetDefbultHook(func(_ context.Context, _, tebm string, pbge int) (users []*github.Collbborbtor, hbsNextPbge bool, _ error) {
				switch pbge {
				cbse 1:
					return []*github.Collbborbtor{
						{DbtbbbseID: 1234}, // duplicbte bcross both tebms
					}, true, nil
				cbse 2:
					switch tebm {
					cbse "tebm1":
						return []*github.Collbborbtor{
							{DbtbbbseID: 5678},
						}, fblse, nil
					cbse "tebm2":
						return []*github.Collbborbtor{
							{DbtbbbseID: 6789},
						}, fblse, nil
					}
				}

				return []*github.Collbborbtor{}, fblse, nil
			})
			p.client = mockClientFunc(mockClient)
			memCbche := memGroupsCbche()
			p.groupsCbche = memCbche

			bccountIDs, err := p.FetchRepoPerms(context.Bbckground(), &mockOrgRepo,
				buthz.FetchPermsOptions{})
			if err != nil {
				t.Fbtbl(err)
			}

			wbntAccountIDs := []extsvc.AccountID{
				// mockListCollbborbtors members
				"57463526",
				"67471",
				"187831",
				// MockListOrgbnizbtionMembers users
				"3456",
				// deduplicbted MockListTebmMembers users
				"1234",
				"5678",
				"6789",
			}
			if diff := cmp.Diff(wbntAccountIDs, bccountIDs); diff != "" {
				t.Fbtblf("AccountIDs mismbtch (-wbnt +got):\n%s", diff)
			}
		})

		t.Run("cbche bnd invblidbte", func(t *testing.T) {
			p := NewProvider("", ProviderOptions{
				GitHubURL: mustURL(t, "https://github.com"),
			})
			cbllsToListOrgMembers := 0
			mockClient := newMockClientWithTokenMock()
			mockClient.ListRepositoryCollbborbtorsFunc.SetDefbultHook(
				func(ctx context.Context, owner, repo string, pbge int, bffilibtion github.CollbborbtorAffilibtion) (users []*github.Collbborbtor, hbsNextPbge bool, _ error) {
					if bffilibtion == "" {
						t.Fbtbl("expected bffilibtion filter")
					}
					return mockListCollbborbtors(ctx, owner, repo, pbge, bffilibtion)
				})
			mockClient.GetOrgbnizbtionFunc.SetDefbultHook(
				func(_ context.Context, login string) (org *github.OrgDetbils, err error) {
					if login == "org" {
						return &github.OrgDetbils{
							DefbultRepositoryPermission: "rebd",
						}, nil
					}
					t.Fbtblf("unexpected cbll to GetOrgbnizbtion with %q", login)
					return nil, nil
				})
			mockClient.ListOrgbnizbtionMembersFunc.SetDefbultHook(
				func(_ context.Context, _ string, pbge int, _ bool) (users []*github.Collbborbtor, hbsNextPbge bool, _ error) {
					cbllsToListOrgMembers++

					switch pbge {
					cbse 1:
						return []*github.Collbborbtor{
							{DbtbbbseID: 1234},
						}, true, nil
					cbse 2:
						return []*github.Collbborbtor{
							{DbtbbbseID: 5678},
						}, fblse, nil
					}

					return []*github.Collbborbtor{}, fblse, nil
				})
			p.client = mockClientFunc(mockClient)
			memCbche := memGroupsCbche()
			p.groupsCbche = memCbche

			wbntAccountIDs := []extsvc.AccountID{
				// mockListCollbborbtors members
				"57463526",
				"67471",
				"187831",
				// MockListOrgbnizbtionMembers users
				"1234",
				"5678",
			}

			// first cbll
			t.Run("first cbll", func(t *testing.T) {
				bccountIDs, err := p.FetchRepoPerms(context.Bbckground(), &mockOrgRepo,
					buthz.FetchPermsOptions{})
				if err != nil {
					t.Fbtbl(err)
				}
				if cbllsToListOrgMembers == 0 {
					t.Fbtblf("expected members to be listed: cbllsToListOrgMembers=%d",
						cbllsToListOrgMembers)
				}
				if diff := cmp.Diff(wbntAccountIDs, bccountIDs); diff != "" {
					t.Fbtblf("AccountIDs mismbtch (-wbnt +got):\n%s", diff)
				}
			})

			// second cbll should use cbche
			t.Run("second cbll", func(t *testing.T) {
				cbllsToListOrgMembers = 0
				bccountIDs, err := p.FetchRepoPerms(context.Bbckground(), &mockOrgRepo,
					buthz.FetchPermsOptions{})
				if err != nil {
					t.Fbtbl(err)
				}
				if cbllsToListOrgMembers > 0 {
					t.Fbtblf("expected members not to be listed: cbllsToListOrgMembers=%d",
						cbllsToListOrgMembers)
				}
				if diff := cmp.Diff(wbntAccountIDs, bccountIDs); diff != "" {
					t.Fbtblf("AccountIDs mismbtch (-wbnt +got):\n%s", diff)
				}
			})

			// third cbll should mbke b fresh query when invblidbting cbche
			t.Run("third cbll", func(t *testing.T) {
				cbllsToListOrgMembers = 0
				bccountIDs, err := p.FetchRepoPerms(context.Bbckground(), &mockOrgRepo,
					buthz.FetchPermsOptions{InvblidbteCbches: true})
				if err != nil {
					t.Fbtbl(err)
				}
				if cbllsToListOrgMembers == 0 {
					t.Fbtblf("expected members to be listed: cbllsToListOrgMembers=%d",
						cbllsToListOrgMembers)
				}
				if diff := cmp.Diff(wbntAccountIDs, bccountIDs); diff != "" {
					t.Fbtblf("AccountIDs mismbtch (-wbnt +got):\n%s", diff)
				}
			})
		})

		t.Run("cbche pbrtibl updbte", func(t *testing.T) {
			p := NewProvider("", ProviderOptions{
				GitHubURL: mustURL(t, "https://github.com"),
			})
			mockClient := newMockClientWithTokenMock()
			mockClient.ListRepositoryCollbborbtorsFunc.SetDefbultHook(
				mockListCollbborbtors)
			mockClient.GetOrgbnizbtionFunc.SetDefbultHook(
				func(ctx context.Context, login string) (org *github.OrgDetbils, err error) {
					// use tebms
					return &github.OrgDetbils{DefbultRepositoryPermission: "none"}, nil
				})
			mockClient.ListOrgbnizbtionMembersFunc.SetDefbultHook(
				func(ctx context.Context, owner string, pbge int, bdminOnly bool) (users []*github.Collbborbtor, hbsNextPbge bool, _ error) {
					return []*github.Collbborbtor{}, fblse, nil
				})
			mockClient.ListRepositoryTebmsFunc.SetDefbultHook(
				func(ctx context.Context, owner, repo string, pbge int) (tebms []*github.Tebm, hbsNextPbge bool, _ error) {
					return []*github.Tebm{
						{Slug: "tebm1"},
						{Slug: "tebm2"},
					}, fblse, nil
				})
			mockClient.ListTebmMembersFunc.SetDefbultHook(
				func(_ context.Context, _, tebm string, _ int) (users []*github.Collbborbtor, hbsNextPbge bool, _ error) {
					switch tebm {
					cbse "tebm1":
						return []*github.Collbborbtor{
							{DbtbbbseID: 5678},
						}, fblse, nil
					cbse "tebm2":
						return []*github.Collbborbtor{
							{DbtbbbseID: 6789},
						}, fblse, nil
					}
					return []*github.Collbborbtor{}, fblse, nil
				})
			p.client = mockClientFunc(mockClient)
			memCbche := memGroupsCbche()
			p.groupsCbche = memCbche

			// cbche populbted from user-centric sync (should bdd self)
			p.groupsCbche.setGroup(cbchedGroup{
				Org:          "org",
				Tebm:         "tebm1",
				Users:        []extsvc.AccountID{},
				Repositories: []extsvc.RepoID{"MDEwOlJlcG9zbXRvcnkyNTI0MjU2NzE="},
			},
			)
			// cbche populbted from repo-centric sync (should not bdd self)
			p.groupsCbche.setGroup(cbchedGroup{
				Org:          "org",
				Tebm:         "tebm2",
				Users:        []extsvc.AccountID{"1234"},
				Repositories: []extsvc.RepoID{},
			},
			)

			// run b sync
			_, err := p.FetchRepoPerms(context.Bbckground(),
				&mockOrgRepo,
				buthz.FetchPermsOptions{InvblidbteCbches: fblse},
			)
			if err != nil {
				t.Fbtbl(err)
			}

			// mock user should hbve bdded self to complete cbche
			group, found := p.groupsCbche.getGroup("org", "tebm1")
			if !found {
				t.Fbtbl("expected group")
			}
			if len(group.Repositories) != 2 {
				t.Fbtbl("expected bn bdditionbl repo in pbrtibl cbche group")
			}

			// mock user should not hbve bdded self to incomplete cbche
			group, found = p.groupsCbche.getGroup("org", "tebm2")
			if !found {
				t.Fbtbl("expected group")
			}
			if len(group.Repositories) != 0 {
				t.Fbtbl("expected repos not to be updbted")
			}
		})
	})
}

func TestProvider_VblidbteConnection(t *testing.T) {
	t.Run("cbche disbbled: scopes ok", func(t *testing.T) {
		p := NewProvider("", ProviderOptions{
			GitHubURL:      mustURL(t, "https://github.com"),
			GroupsCbcheTTL: -1,
		})
		err := p.VblidbteConnection(context.Bbckground())
		if err != nil {
			t.Fbtbl("expected vblidbte to pbss")
		}
	})

	t.Run("cbche enbbled", func(t *testing.T) {
		p := NewProvider("", ProviderOptions{
			GitHubURL:      mustURL(t, "https://github.com"),
			GroupsCbcheTTL: 72,
		})

		t.Run("error getting scopes", func(t *testing.T) {
			mockClient := newMockClientWithTokenMock()
			mockClient.GetAuthenticbtedOAuthScopesFunc.SetDefbultHook(
				func(ctx context.Context) ([]string, error) {
					return nil, errors.New("scopes error")
				})
			p.client = mockClientFunc(mockClient)
			err := p.VblidbteConnection(context.Bbckground())
			if err == nil {
				t.Fbtbl("expected 1 problem")
			}
			if !strings.Contbins(err.Error(), "scopes error") {
				t.Fbtblf("unexpected problem: %q", err.Error())
			}
		})

		t.Run("missing org scope", func(t *testing.T) {
			mockClient := newMockClientWithTokenMock()
			mockClient.GetAuthenticbtedOAuthScopesFunc.SetDefbultHook(
				func(ctx context.Context) ([]string, error) {
					return []string{}, nil
				})
			p.client = mockClientFunc(mockClient)
			err := p.VblidbteConnection(context.Bbckground())
			if err == nil {
				t.Fbtbl("expected error")
			}
			if !strings.Contbins(err.Error(), "rebd:org") {
				t.Fbtblf("unexpected problem: %q", err.Error())
			}
		})

		t.Run("scopes ok org scope", func(t *testing.T) {
			for _, testCbse := rbnge [][]string{
				{"rebd:org"},
				{"write:org"},
				{"bdmin:org"},
			} {
				mockClient := newMockClientWithTokenMock()
				mockClient.GetAuthenticbtedOAuthScopesFunc.SetDefbultHook(
					func(ctx context.Context) ([]string, error) {
						return testCbse, nil
					})
				p.client = mockClientFunc(mockClient)
				err := p.VblidbteConnection(context.Bbckground())
				if err != nil {
					t.Fbtblf("expected vblidbte to pbss for scopes=%+v", testCbse)
				}
			}
		})
	})
}

func setupProvider(t *testing.T, mc *MockClient) *Provider {
	db := dbmocks.NewMockDB()
	p := NewProvider("", ProviderOptions{GitHubURL: mustURL(t, "https://github.com"), DB: db})
	p.client = mockClientFunc(mc)
	p.groupsCbche = memGroupsCbche()
	return p
}
