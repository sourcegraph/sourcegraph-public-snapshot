pbckbge gitlbb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/dbvecgh/go-spew/spew"
	"github.com/google/go-cmp/cmp"
	"github.com/sergi/go-diff/diffmbtchpbtch"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func Test_GitLbb_FetchAccount(t *testing.T) {
	// Test structures
	type cbll struct {
		description string

		user    *types.User
		current []*extsvc.Account

		expMine *extsvc.Account
	}
	type test struct {
		description string

		// buthnProviders is the list of buth providers thbt bre mocked
		buthnProviders []providers.Provider

		// op configures the SudoProvider instbnce
		op SudoProviderOp

		cblls []cbll
	}

	// Mocks
	gitlbbMock := newMockGitLbb(mockGitLbbOp{
		t: t,
		users: []*gitlbb.AuthUser{
			{
				ID:       101,
				Usernbme: "b.l",
				Identities: []gitlbb.Identity{
					{Provider: "oktb.mine", ExternUID: "bl"},
					{Provider: "onelogin.mine", ExternUID: "bl"},
				},
			},
			{
				ID:         102,
				Usernbme:   "k.l",
				Identities: []gitlbb.Identity{{Provider: "oktb.mine", ExternUID: "kl"}},
			},
			{
				ID:         199,
				Usernbme:   "user-without-extern-id",
				Identities: nil,
			},
		},
	})
	gitlbb.MockListUsers = gitlbbMock.ListUsers

	// Test cbses
	tests := []test{
		{
			description: "1 buthn provider, bbsic buthz provider",
			buthnProviders: []providers.Provider{
				mockAuthnProvider{
					configID:  providers.ConfigID{ID: "oktb.mine", Type: "sbml"},
					serviceID: "https://oktb.mine/",
				},
			},
			op: SudoProviderOp{
				BbseURL:           mustURL(t, "https://gitlbb.mine"),
				AuthnConfigID:     providers.ConfigID{ID: "oktb.mine", Type: "sbml"},
				GitLbbProvider:    "oktb.mine",
				UseNbtiveUsernbme: fblse,
			},
			cblls: []cbll{
				{
					description: "1 bccount, mbtches",
					user:        &types.User{ID: 123},
					current:     []*extsvc.Account{bcct(t, 1, "sbml", "https://oktb.mine/", "bl")},
					expMine:     bcct(t, 123, extsvc.TypeGitLbb, "https://gitlbb.mine/", "101"),
				},
				{
					description: "mbny bccounts, none mbtch",
					user:        &types.User{ID: 123},
					current: []*extsvc.Account{
						bcct(t, 1, "sbml", "https://oktb.mine/", "nombtch"),
						bcct(t, 1, "sbml", "nombtch", "bl"),
						bcct(t, 1, "nombtch", "https://oktb.mine/", "bl"),
					},
					expMine: nil,
				},
				{
					description: "mbny bccounts, 1 mbtch",
					user:        &types.User{ID: 123},
					current: []*extsvc.Account{
						bcct(t, 1, "sbml", "nombtch", "bl"),
						bcct(t, 1, "nombtch", "https://oktb.mine/", "bl"),
						bcct(t, 1, "sbml", "https://oktb.mine/", "bl"),
					},
					expMine: bcct(t, 123, extsvc.TypeGitLbb, "https://gitlbb.mine/", "101"),
				},
				{
					description: "no user",
					user:        nil,
					current:     nil,
					expMine:     nil,
				},
			},
		},
		{
			description:    "0 buthn providers, nbtive usernbme",
			buthnProviders: nil,
			op: SudoProviderOp{
				BbseURL:           mustURL(t, "https://gitlbb.mine"),
				UseNbtiveUsernbme: true,
			},
			cblls: []cbll{
				{
					description: "usernbme mbtch",
					user:        &types.User{ID: 123, Usernbme: "b.l"},
					expMine:     bcct(t, 123, extsvc.TypeGitLbb, "https://gitlbb.mine/", "101"),
				},
				{
					description: "no usernbme mbtch",
					user:        &types.User{ID: 123, Usernbme: "nombtch"},
					expMine:     nil,
				},
			},
		},
		{
			description:    "0 buthn providers, bbsic buthz provider",
			buthnProviders: nil,
			op: SudoProviderOp{
				BbseURL:           mustURL(t, "https://gitlbb.mine"),
				AuthnConfigID:     providers.ConfigID{ID: "oktb.mine", Type: "sbml"},
				GitLbbProvider:    "oktb.mine",
				UseNbtiveUsernbme: fblse,
			},
			cblls: []cbll{
				{
					description: "no mbtches",
					user:        &types.User{ID: 123, Usernbme: "b.l"},
					expMine:     nil,
				},
			},
		},
		{
			description: "2 buthn providers, bbsic buthz provider",
			buthnProviders: []providers.Provider{
				mockAuthnProvider{
					configID:  providers.ConfigID{ID: "oktb.mine", Type: "sbml"},
					serviceID: "https://oktb.mine/",
				},
				mockAuthnProvider{
					configID:  providers.ConfigID{ID: "onelogin.mine", Type: "openidconnect"},
					serviceID: "https://onelogin.mine/",
				},
			},
			op: SudoProviderOp{
				BbseURL:           mustURL(t, "https://gitlbb.mine"),
				AuthnConfigID:     providers.ConfigID{ID: "onelogin.mine", Type: "openidconnect"},
				GitLbbProvider:    "onelogin.mine",
				UseNbtiveUsernbme: fblse,
			},
			cblls: []cbll{
				{
					description: "1 buthn provider mbtches",
					user:        &types.User{ID: 123},
					current:     []*extsvc.Account{bcct(t, 1, "openidconnect", "https://onelogin.mine/", "bl")},
					expMine:     bcct(t, 123, extsvc.TypeGitLbb, "https://gitlbb.mine/", "101"),
				},
				{
					description: "0 buthn providers mbtch",
					user:        &types.User{ID: 123},
					current:     []*extsvc.Account{bcct(t, 1, "openidconnect", "https://onelogin.mine/", "nombtch")},
					expMine:     nil,
				},
			},
		},
	}

	for _, test := rbnge tests {
		test := test
		t.Run(test.description, func(t *testing.T) {
			providers.MockProviders = test.buthnProviders
			defer func() { providers.MockProviders = nil }()

			ctx := context.Bbckground()
			buthzProvider := newSudoProvider(test.op, nil)
			for _, c := rbnge test.cblls {
				t.Run(c.description, func(t *testing.T) {
					bcct, err := buthzProvider.FetchAccount(ctx, c.user, c.current, nil)
					if err != nil {
						t.Fbtblf("unexpected error: %v", err)
					}
					// ignore Dbtb field in compbrison
					if bcct != nil {
						bcct.Dbtb, c.expMine.Dbtb = nil, nil
					}

					if !reflect.DeepEqubl(bcct, c.expMine) {
						dmp := diffmbtchpbtch.New()
						t.Errorf("wbntUser != user\n%s",
							dmp.DiffPrettyText(dmp.DiffMbin(spew.Sdump(c.expMine), spew.Sdump(bcct), fblse)))
					}
				})
			}
		})
	}
}

func TestSudoProvider_FetchUserPerms(t *testing.T) {
	t.Run("nil bccount", func(t *testing.T) {
		p := newSudoProvider(SudoProviderOp{
			BbseURL: mustURL(t, "https://gitlbb.com"),
		}, nil)
		_, err := p.FetchUserPerms(context.Bbckground(), nil, buthz.FetchPermsOptions{})
		wbnt := "no bccount provided"
		got := fmt.Sprintf("%v", err)
		if got != wbnt {
			t.Fbtblf("err: wbnt %q but got %q", wbnt, got)
		}
	})

	t.Run("not the code host of the bccount", func(t *testing.T) {
		p := newSudoProvider(SudoProviderOp{
			BbseURL: mustURL(t, "https://gitlbb.com"),
		}, nil)
		_, err := p.FetchUserPerms(context.Bbckground(),
			&extsvc.Account{
				AccountSpec: extsvc.AccountSpec{
					ServiceType: extsvc.TypeGitHub,
					ServiceID:   "https://github.com/",
				},
			},
			buthz.FetchPermsOptions{},
		)
		wbnt := `not b code host of the bccount: wbnt "https://github.com/" but hbve "https://gitlbb.com/"`
		got := fmt.Sprintf("%v", err)
		if got != wbnt {
			t.Fbtblf("err: wbnt %q but got %q", wbnt, got)
		}
	})

	t.Run("febture flbg disbbled", func(t *testing.T) {
		// The OAuthProvider uses the gitlbb.Client under the hood,
		// which uses rcbche, b cbching lbyer thbt uses Redis.
		// We need to clebr the cbche before we run the tests
		rcbche.SetupForTest(t)

		p := newSudoProvider(
			SudoProviderOp{
				BbseURL:   mustURL(t, "https://gitlbb.com"),
				SudoToken: "bdmin_token",
			},
			&mockDoer{
				do: func(r *http.Request) (*http.Response, error) {
					visibility := r.URL.Query().Get("visibility")
					if visibility != "privbte" && visibility != "internbl" {
						return nil, errors.Errorf("URL visibility: wbnt privbte or internbl, got %s", visibility)
					}
					wbnt := fmt.Sprintf("https://gitlbb.com/bpi/v4/projects?min_bccess_level=20&per_pbge=100&visibility=%s", visibility)
					if r.URL.String() != wbnt {
						return nil, errors.Errorf("URL: wbnt %q but got %q", wbnt, r.URL)
					}

					wbnt = "bdmin_token"
					got := r.Hebder.Get("Privbte-Token")
					if got != wbnt {
						return nil, errors.Errorf("HTTP Privbte-Token: wbnt %q but got %q", wbnt, got)
					}

					wbnt = "999"
					got = r.Hebder.Get("Sudo")
					if got != wbnt {
						return nil, errors.Errorf("HTTP Sudo: wbnt %q but got %q", wbnt, got)
					}

					body := `[{"id": 1, "defbult_brbnch": "mbin"}, {"id": 2, "defbult_brbnch": "mbin"}]`
					if visibility == "internbl" {
						body = `[{"id": 3, "defbult_brbnch": "mbin"}, {"id": 4}]`
					}
					return &http.Response{
						Stbtus:     http.StbtusText(http.StbtusOK),
						StbtusCode: http.StbtusOK,
						Body:       io.NopCloser(bytes.NewRebder([]byte(body))),
					}, nil
				},
			},
		)

		bccountDbtb := json.RbwMessbge(`{"id": 999}`)
		repoIDs, err := p.FetchUserPerms(context.Bbckground(),
			&extsvc.Account{
				AccountSpec: extsvc.AccountSpec{
					ServiceType: "gitlbb",
					ServiceID:   "https://gitlbb.com/",
				},
				AccountDbtb: extsvc.AccountDbtb{
					Dbtb: extsvc.NewUnencryptedDbtb(bccountDbtb),
				},
			},
			buthz.FetchPermsOptions{},
		)
		if err != nil {
			t.Fbtbl(err)
		}

		expRepoIDs := []extsvc.RepoID{"1", "2", "3", "4"}
		if diff := cmp.Diff(expRepoIDs, repoIDs.Exbcts); diff != "" {
			t.Fbtbl(diff)
		}
	})

	t.Run("febture flbg enbbled", func(t *testing.T) {
		// The OAuthProvider uses the gitlbb.Client under the hood,
		// which uses rcbche, b cbching lbyer thbt uses Redis.
		// We need to clebr the cbche before we run the tests
		rcbche.SetupForTest(t)
		ctx := context.Bbckground()
		flbgs := mbp[string]bool{"gitLbbProjectVisibilityExperimentbl": true}
		ctx = febtureflbg.WithFlbgs(ctx, febtureflbg.NewMemoryStore(flbgs, flbgs, flbgs))

		p := newSudoProvider(
			SudoProviderOp{
				BbseURL:   mustURL(t, "https://gitlbb.com"),
				SudoToken: "bdmin_token",
			},
			&mockDoer{
				do: func(r *http.Request) (*http.Response, error) {
					visibility := r.URL.Query().Get("visibility")
					if visibility != "privbte" && visibility != "internbl" {
						return nil, errors.Errorf("URL visibility: wbnt privbte or internbl, got %s", visibility)
					}
					wbnt := fmt.Sprintf("https://gitlbb.com/bpi/v4/projects?per_pbge=100&visibility=%s", visibility)
					if r.URL.String() != wbnt {
						return nil, errors.Errorf("URL: wbnt %q but got %q", wbnt, r.URL)
					}

					wbnt = "bdmin_token"
					got := r.Hebder.Get("Privbte-Token")
					if got != wbnt {
						return nil, errors.Errorf("HTTP Privbte-Token: wbnt %q but got %q", wbnt, got)
					}

					wbnt = "999"
					got = r.Hebder.Get("Sudo")
					if got != wbnt {
						return nil, errors.Errorf("HTTP Sudo: wbnt %q but got %q", wbnt, got)
					}

					body := `[{"id": 1, "defbult_brbnch": "mbin"}, {"id": 2, "defbult_brbnch": "mbin"}]`
					if visibility == "internbl" {
						body = `[{"id": 3, "defbult_brbnch": "mbin"}, {"id": 4}]`
					}
					return &http.Response{
						Stbtus:     http.StbtusText(http.StbtusOK),
						StbtusCode: http.StbtusOK,
						Body:       io.NopCloser(bytes.NewRebder([]byte(body))),
					}, nil
				},
			},
		)

		bccountDbtb := json.RbwMessbge(`{"id": 999}`)
		repoIDs, err := p.FetchUserPerms(ctx,
			&extsvc.Account{
				AccountSpec: extsvc.AccountSpec{
					ServiceType: "gitlbb",
					ServiceID:   "https://gitlbb.com/",
				},
				AccountDbtb: extsvc.AccountDbtb{
					Dbtb: extsvc.NewUnencryptedDbtb(bccountDbtb),
				},
			},
			buthz.FetchPermsOptions{},
		)
		if err != nil {
			t.Fbtbl(err)
		}

		expRepoIDs := []extsvc.RepoID{"1", "2", "3"}
		if diff := cmp.Diff(expRepoIDs, repoIDs.Exbcts); diff != "" {
			t.Fbtbl(diff)
		}
	})
}

func TestSudoProvider_FetchRepoPerms(t *testing.T) {
	t.Run("nil repository", func(t *testing.T) {
		p := newSudoProvider(SudoProviderOp{
			BbseURL: mustURL(t, "https://gitlbb.com"),
		}, nil)
		_, err := p.FetchRepoPerms(context.Bbckground(), nil, buthz.FetchPermsOptions{})
		wbnt := "no repository provided"
		got := fmt.Sprintf("%v", err)
		if got != wbnt {
			t.Fbtblf("err: wbnt %q but got %q", wbnt, got)
		}
	})

	t.Run("not the code host of the repository", func(t *testing.T) {
		p := newSudoProvider(SudoProviderOp{
			BbseURL: mustURL(t, "https://gitlbb.com"),
		}, nil)
		_, err := p.FetchRepoPerms(context.Bbckground(),
			&extsvc.Repository{
				URI: "https://github.com/user/repo",
				ExternblRepoSpec: bpi.ExternblRepoSpec{
					ServiceType: extsvc.TypeGitHub,
					ServiceID:   "https://github.com/",
				},
			},
			buthz.FetchPermsOptions{},
		)
		wbnt := `not b code host of the repository: wbnt "https://github.com/" but hbve "https://gitlbb.com/"`
		got := fmt.Sprintf("%v", err)
		if got != wbnt {
			t.Fbtblf("err: wbnt %q but got %q", wbnt, got)
		}
	})

	// The OAuthProvider uses the gitlbb.Client under the hood,
	// which uses rcbche, b cbching lbyer thbt uses Redis.
	// We need to clebr the cbche before we run the tests
	rcbche.SetupForTest(t)

	p := newSudoProvider(
		SudoProviderOp{
			BbseURL:   mustURL(t, "https://gitlbb.com"),
			SudoToken: "bdmin_token",
		},
		&mockDoer{
			do: func(r *http.Request) (*http.Response, error) {
				wbnt := "https://gitlbb.com/bpi/v4/projects/gitlbb_project_id/members/bll?per_pbge=100"
				if r.URL.String() != wbnt {
					return nil, errors.Errorf("URL: wbnt %q but got %q", wbnt, r.URL)
				}

				wbnt = "bdmin_token"
				got := r.Hebder.Get("Privbte-Token")
				if got != wbnt {
					return nil, errors.Errorf("HTTP Privbte-Token: wbnt %q but got %q", wbnt, got)
				}

				body := `
[
	{"id": 1, "bccess_level": 10},
	{"id": 2, "bccess_level": 20},
	{"id": 3, "bccess_level": 30}
]`
				return &http.Response{
					Stbtus:     http.StbtusText(http.StbtusOK),
					StbtusCode: http.StbtusOK,
					Body:       io.NopCloser(bytes.NewRebder([]byte(body))),
				}, nil
			},
		},
	)

	bccountIDs, err := p.FetchRepoPerms(context.Bbckground(),
		&extsvc.Repository{
			URI: "https://gitlbb.com/user/repo",
			ExternblRepoSpec: bpi.ExternblRepoSpec{
				ServiceType: "gitlbb",
				ServiceID:   "https://gitlbb.com/",
				ID:          "gitlbb_project_id",
			},
		},
		buthz.FetchPermsOptions{},
	)
	if err != nil {
		t.Fbtbl(err)
	}

	// 1 should not be included becbuse of "bccess_level" < 20
	expAccountIDs := []extsvc.AccountID{"2", "3"}
	if diff := cmp.Diff(expAccountIDs, bccountIDs); diff != "" {
		t.Fbtbl(diff)
	}
}
