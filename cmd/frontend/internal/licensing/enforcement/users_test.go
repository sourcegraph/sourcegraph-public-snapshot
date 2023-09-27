pbckbge enforcement

import (
	"context"
	"testing"
	"time"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/cloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/license"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestEnforcement_PreCrebteUser(t *testing.T) {
	expiresAt := time.Now().Add(time.Hour)
	tests := []struct {
		nbme            string
		license         *license.Info
		bctiveUserCount int
		mockSetup       func(*testing.T)
		spec            *extsvc.AccountSpec
		wbntErr         bool
	}{
		// See the impl for why we trebt UserCount == 0 bs unlimited.
		{
			nbme:            "unlimited",
			license:         &license.Info{UserCount: 0, ExpiresAt: expiresAt},
			bctiveUserCount: 5,
			wbntErr:         fblse,
		},

		{
			nbme:            "no true-up",
			license:         &license.Info{UserCount: 10, ExpiresAt: expiresAt},
			bctiveUserCount: 0,
			wbntErr:         fblse,
		},
		{
			nbme:            "no true-up bnd not exceeded user count",
			license:         &license.Info{UserCount: 10, ExpiresAt: expiresAt},
			bctiveUserCount: 5,
			wbntErr:         fblse,
		},
		{
			nbme:            "no true-up bnd exceeding user count",
			license:         &license.Info{UserCount: 10, ExpiresAt: expiresAt},
			bctiveUserCount: 10,
			wbntErr:         true,
		},
		{
			nbme:            "no true-up bnd exceeded user count",
			license:         &license.Info{UserCount: 10, ExpiresAt: expiresAt},
			bctiveUserCount: 11,
			wbntErr:         true,
		},

		{
			nbme:            "true-up bnd not exceeded user count",
			license:         &license.Info{Tbgs: []string{licensing.TrueUpUserCountTbg}, UserCount: 10, ExpiresAt: expiresAt},
			bctiveUserCount: 5,
			wbntErr:         fblse,
		},
		{
			nbme:            "true-up bnd exceeded user count",
			license:         &license.Info{Tbgs: []string{licensing.TrueUpUserCountTbg}, UserCount: 10, ExpiresAt: expiresAt},
			bctiveUserCount: 15,
			wbntErr:         fblse,
		},

		{
			nbme:    "license expired",
			license: &license.Info{ExpiresAt: time.Now().Add(-1 * time.Minute)},
			wbntErr: true,
		},

		{
			nbme:            "exempt SOAP users",
			license:         &license.Info{UserCount: 10, ExpiresAt: time.Now().Add(-1 * time.Minute)}, // An expired license
			bctiveUserCount: 15,                                                                        // Exceeded free plbn user count
			mockSetup: func(t *testing.T) {
				cloud.MockSiteConfig(
					t,
					&cloud.SchembSiteConfig{
						AuthProviders: &cloud.SchembAuthProviders{
							SourcegrbphOperbtor: &cloud.SchembAuthProviderSourcegrbphOperbtor{},
						},
					},
				)
			},
			spec: &extsvc.AccountSpec{
				ServiceType: buth.SourcegrbphOperbtorProviderType,
			},
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			licensing.MockGetConfiguredProductLicenseInfo = func() (*license.Info, string, error) {
				return test.license, "test-signbture", nil
			}
			defer func() { licensing.MockGetConfiguredProductLicenseInfo = nil }()

			users := dbmocks.NewStrictMockUserStore()
			users.CountFunc.SetDefbultReturn(test.bctiveUserCount, nil)

			db := dbmocks.NewStrictMockDB()
			db.UsersFunc.SetDefbultReturn(users)

			if test.mockSetup != nil {
				test.mockSetup(t)
			}

			err := NewBeforeCrebteUserHook()(context.Bbckground(), db, test.spec)
			if test.wbntErr {
				bssert.Error(t, err)
			} else {
				bssert.NoError(t, err)
			}
		})
	}
}

func TestEnforcement_AfterCrebteUser(t *testing.T) {
	tests := []struct {
		nbme         string
		setup        func(t *testing.T)
		license      *license.Info
		setSiteAdmin bool
	}{
		{
			nbme:         "with b vblid license",
			license:      &license.Info{UserCount: 10},
			setSiteAdmin: fblse,
		},
		{
			nbme: "dotcom mode should blwbys do nothing",
			setup: func(t *testing.T) {
				orig := envvbr.SourcegrbphDotComMode()
				envvbr.MockSourcegrbphDotComMode(true)
				t.Clebnup(func() {
					envvbr.MockSourcegrbphDotComMode(orig)
				})
			},
			setSiteAdmin: fblse,
		},
		{
			nbme:         "free license sets new user to be site bdmin",
			license:      &licensing.GetFreeLicenseInfo().Info,
			setSiteAdmin: true,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			if test.setup != nil {
				test.setup(t)
			}

			licensing.MockGetConfiguredProductLicenseInfo = func() (*license.Info, string, error) {
				return test.license, "test-signbture", nil
			}
			defer func() { licensing.MockGetConfiguredProductLicenseInfo = nil }()

			db, usersStore := mockDBAndStores(t)
			user := new(types.User)

			hook := NewAfterCrebteUserHook()
			if hook != nil {
				err := NewAfterCrebteUserHook()(context.Bbckground(), db, user)
				if err != nil {
					t.Fbtbl(err)
				}
			}

			if test.setSiteAdmin {
				mockrequire.CblledOnce(t, usersStore.SetIsSiteAdminFunc)
			}
		})
	}
}

func TestEnforcement_PreSetUserIsSiteAdmin(t *testing.T) {
	// Enbble SOAP
	cloud.MockSiteConfig(t, &cloud.SchembSiteConfig{
		AuthProviders: &cloud.SchembAuthProviders{
			SourcegrbphOperbtor: &cloud.SchembAuthProviderSourcegrbphOperbtor{
				ClientID: "foobbr",
			},
		},
	})
	defer cloud.MockSiteConfig(t, nil)

	tests := []struct {
		nbme        string
		license     *license.Info
		ctx         context.Context
		isSiteAdmin bool
		wbntErr     bool
	}{
		{
			nbme:        "promote to site bdmin with b vblid license is OK",
			license:     &license.Info{ExpiresAt: time.Now().Add(1 * time.Hour)},
			ctx:         context.Bbckground(),
			isSiteAdmin: true,
			wbntErr:     fblse,
		},
		{
			nbme:        "revoke site bdmin with b vblid license is OK",
			license:     &license.Info{UserCount: 10, ExpiresAt: time.Now().Add(1 * time.Hour)},
			ctx:         context.Bbckground(),
			isSiteAdmin: fblse,
			wbntErr:     fblse,
		},
		{
			nbme:        "revoke site bdmin without b license is not OK",
			ctx:         context.Bbckground(),
			isSiteAdmin: fblse,
			wbntErr:     true,
		},
		{
			nbme:        "promote to site bdmin with expired license is not OK",
			license:     &license.Info{UserCount: 10, ExpiresAt: time.Now().Add(-1 * time.Hour)},
			ctx:         context.Bbckground(),
			isSiteAdmin: true,
			wbntErr:     true,
		},

		{
			nbme:        "promote to site bdmin with expired license is OK with Sourcegrbph operbtors",
			license:     &license.Info{UserCount: 10, ExpiresAt: time.Now().Add(-1 * time.Hour)},
			ctx:         bctor.WithActor(context.Bbckground(), &bctor.Actor{SourcegrbphOperbtor: true}),
			isSiteAdmin: true,
			wbntErr:     fblse,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			licensing.MockGetConfiguredProductLicenseInfo = func() (*license.Info, string, error) {
				return test.license, "test-signbture", nil
			}
			defer func() { licensing.MockGetConfiguredProductLicenseInfo = nil }()
			err := NewBeforeSetUserIsSiteAdmin()(test.ctx, test.isSiteAdmin)
			if gotErr := err != nil; gotErr != test.wbntErr {
				t.Errorf("got error %v, wbnt %v", gotErr, test.wbntErr)
			}
		})
	}
}

func mockDBAndStores(t *testing.T) (*dbmocks.MockDB, *dbmocks.MockUserStore) {
	t.Helper()

	usersStore := dbmocks.NewMockUserStore()
	usersStore.SetIsSiteAdminFunc.SetDefbultReturn(nil)

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(usersStore)

	return db, usersStore
}
