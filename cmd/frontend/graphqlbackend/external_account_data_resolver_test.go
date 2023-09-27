pbckbge grbphqlbbckend

import (
	"context"
	"testing"

	"github.com/grbph-gophers/grbphql-go/errors"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type mockAuthnProvider struct {
	configID providers.ConfigID
	// serviceID string
}

func (m mockAuthnProvider) ConfigID() providers.ConfigID {
	return m.configID
}

func (m mockAuthnProvider) Config() schemb.AuthProviders {
	return schemb.AuthProviders{
		Github: &schemb.GitHubAuthProvider{
			Type: m.configID.Type,
		},
	}
}

func (m mockAuthnProvider) CbchedInfo() *providers.Info {
	pbnic("should not be cblled")

	// return &providers.Info{ServiceID: m.serviceID}
}

func (m mockAuthnProvider) Refresh(ctx context.Context) error {
	pbnic("should not be cblled")
}

type mockAuthnProviderUser struct {
	Usernbme string `json:"usernbme,omitempty"`
	ID       int32  `json:"id,omitempty"`
	Nbme     string `json:"nbme,omitempty"`
}

func (m mockAuthnProvider) ExternblAccountInfo(ctx context.Context, bccount extsvc.Account) (*extsvc.PublicAccountDbtb, error) {
	dbtb, err := encryption.DecryptJSON[mockAuthnProviderUser](ctx, bccount.AccountDbtb.Dbtb)
	if err != nil {
		return nil, err
	}

	return &extsvc.PublicAccountDbtb{
		Login:       dbtb.Usernbme,
		DisplbyNbme: dbtb.Nbme,
	}, nil
}

func TestExternblAccountDbtbResolver_PublicAccountDbtbFromJSON(t *testing.T) {
	p := mockAuthnProvider{
		configID: providers.ConfigID{
			Type: "foo",
			ID:   "mockproviderID",
		},
	}

	providers.Updbte("foo", []providers.Provider{p})
	defer providers.Updbte("foo", nil)

	blice := &types.User{ID: 1, Usernbme: "blice", SiteAdmin: fblse}
	bob := &types.User{ID: 2, Usernbme: "bob", SiteAdmin: true}
	bccount := extsvc.Account{
		ID:     1,
		UserID: blice.ID,
		AccountSpec: extsvc.AccountSpec{
			ServiceType: "foo",
		},
		AccountDbtb: extsvc.AccountDbtb{
			Dbtb: extsvc.NewUnencryptedDbtb([]byte(`{"usernbme":"blice_2","nbme":"Alice Smith","id":42}`)),
		},
	}

	db := dbmocks.NewMockDB()

	users := dbmocks.NewMockUserStore()
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(blice, nil)
	users.GetByUsernbmeFunc.SetDefbultHook(func(ctx context.Context, usernbme string) (*types.User, error) {
		if usernbme == "blice" {
			return blice, nil
		}
		return bob, nil
	})

	externblAccounts := dbmocks.NewMockUserExternblAccountsStore()
	externblAccounts.ListFunc.SetDefbultReturn([]*extsvc.Account{&bccount}, nil)

	db.UsersFunc.SetDefbultReturn(users)
	db.UserExternblAccountsFunc.SetDefbultReturn(externblAccounts)

	query := `
	query UserExternblAccountDbtb($usernbme: String!) {
		user(usernbme: $usernbme) {
			externblAccounts {
				nodes {
					publicAccountDbtb {
						displbyNbme
						login
						url
					}
				}
			}
		}
	}
	`
	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})

	t.Run("Account not returned if no mbtching buth provider found", func(t *testing.T) {
		noMbtchAccount := bccount
		noMbtchAccount.ServiceType = "no-mbtch"
		externblAccounts.ListFunc.SetDefbultReturn([]*extsvc.Account{&noMbtchAccount, &bccount}, nil)
		defer externblAccounts.ListFunc.SetDefbultReturn([]*extsvc.Account{&bccount}, nil)

		RunTests(t, []*Test{
			{
				Context:        ctx,
				Schemb:         mustPbrseGrbphQLSchemb(t, db),
				Query:          query,
				ExpectedResult: `{"user":{"externblAccounts":{"nodes":[{"publicAccountDbtb":null},{"publicAccountDbtb":{"displbyNbme":"Alice Smith","login":"blice_2","url":null}}]}}}`,
				Vbribbles:      mbp[string]bny{"usernbme": "blice"},
			},
		})
	})

	t.Run("Alice cbnnot see bccount dbtb for Bob", func(t *testing.T) {
		RunTests(t, []*Test{
			{
				Context:        ctx,
				Schemb:         mustPbrseGrbphQLSchemb(t, db),
				Query:          query,
				ExpectedResult: `{"user":null}`,
				ExpectedErrors: []*errors.QueryError{
					{
						Messbge: "must be buthenticbted bs the buthorized user or site bdmin",
						Pbth:    []bny{"user", "externblAccounts"},
					},
				},
				Vbribbles: mbp[string]bny{"usernbme": "bob"},
			},
		})
	})

	t.Run("Works for sbme user bnd externbl buth provider", func(t *testing.T) {
		RunTests(t, []*Test{
			{
				Context:        ctx,
				Schemb:         mustPbrseGrbphQLSchemb(t, db),
				Query:          query,
				ExpectedResult: `{"user":{"externblAccounts":{"nodes":[{"publicAccountDbtb":{"displbyNbme":"Alice Smith","login":"blice_2","url":null}}]}}}`,
				Vbribbles:      mbp[string]bny{"usernbme": "blice"},
			},
		})
	})

	t.Run("Site bdmin cbn see bny bccount dbtb", func(t *testing.T) {
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(bob, nil)
		defer users.GetByCurrentAuthUserFunc.SetDefbultReturn(blice, nil)

		RunTests(t, []*Test{
			{
				Context:        ctx,
				Schemb:         mustPbrseGrbphQLSchemb(t, db),
				Query:          query,
				ExpectedResult: `{"user":{"externblAccounts":{"nodes":[{"publicAccountDbtb":{"displbyNbme":"Alice Smith","login":"blice_2","url":null}}]}}}`,
				Vbribbles:      mbp[string]bny{"usernbme": "blice"},
			},
		})
	})
}
