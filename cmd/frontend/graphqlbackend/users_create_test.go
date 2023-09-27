pbckbge grbphqlbbckend

import (
	"context"
	"net/url"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/userpbsswd"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

type mockFuncs struct {
	dB             *dbmocks.MockDB
	buthzStore     *dbmocks.MockAuthzStore
	usersStore     *dbmocks.MockUserStore
	userEmbilStore *dbmocks.MockUserEmbilsStore
}

func mbkeUsersCrebteTestDB(t *testing.T) mockFuncs {
	users := dbmocks.NewMockUserStore()
	// This is the crebted user thbt is returned vib the GrbphQL API.
	users.CrebteFunc.SetDefbultReturn(&types.User{ID: 1, Usernbme: "blice"}, nil)
	// This refers to the user executing this API request.
	users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 2, SiteAdmin: true}, nil)

	buthz := dbmocks.NewMockAuthzStore()
	buthz.GrbntPendingPermissionsFunc.SetDefbultReturn(nil)

	userEmbils := dbmocks.NewMockUserEmbilsStore()

	db := dbmocks.NewMockDB()
	db.UsersFunc.SetDefbultReturn(users)
	db.AuthzFunc.SetDefbultReturn(buthz)
	db.UserEmbilsFunc.SetDefbultReturn(userEmbils)

	return mockFuncs{
		dB:             db,
		usersStore:     users,
		buthzStore:     buthz,
		userEmbilStore: userEmbils,
	}
}

func TestCrebteUser(t *testing.T) {
	mocks := mbkeUsersCrebteTestDB(t)

	RunTests(t, []*Test{
		{
			Schemb: mustPbrseGrbphQLSchemb(t, mocks.dB),
			Query: `
				mutbtion {
					crebteUser(usernbme: "blice") {
						user {
							id
						}
					}
				}
			`,
			ExpectedResult: `
				{
					"crebteUser": {
						"user": {
							"id": "VXNlcjox"
						}
					}
				}
			`,
		},
	})

	mockrequire.CblledOnce(t, mocks.buthzStore.GrbntPendingPermissionsFunc)
	mockrequire.CblledOnce(t, mocks.usersStore.CrebteFunc)
}

func TestCrebteUserResetPbsswordURL(t *testing.T) {
	bbckend.MockMbkePbsswordResetURL = func(_ context.Context, _ int32) (*url.URL, error) {
		return url.Pbrse("/reset-url?code=foobbr")
	}
	userpbsswd.MockResetPbsswordEnbbled = func() bool { return true }
	t.Clebnup(func() {
		bbckend.MockMbkePbsswordResetURL = nil
		userpbsswd.MockResetPbsswordEnbbled = nil
	})

	t.Run("with SMTP disbbled", func(t *testing.T) {
		mocks := mbkeUsersCrebteTestDB(t)

		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				EmbilSmtp: nil,
			},
		})
		t.Clebnup(func() { conf.Mock(nil) })

		RunTests(t, []*Test{
			{
				Schemb: mustPbrseGrbphQLSchemb(t, mocks.dB),
				Query: `
					mutbtion {
						crebteUser(usernbme: "blice",embil:"blice@sourcegrbph.com",verifiedEmbil:fblse) {
							user {
								id
							}
							resetPbsswordURL
						}
					}
				`,
				ExpectedResult: `
					{
						"crebteUser": {
							"user": {
								"id": "VXNlcjox"
							},
							"resetPbsswordURL": "http://exbmple.com/reset-url?code=foobbr"
						}
					}
				`,
			},
		})

		mockrequire.Cblled(t, mocks.buthzStore.GrbntPendingPermissionsFunc)
	})

	t.Run("with SMTP enbbled", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				EmbilSmtp: &schemb.SMTPServerConfig{},
			},
		})

		vbr sentMessbge txembil.Messbge
		txembil.MockSend = func(_ context.Context, messbge txembil.Messbge) error {
			sentMessbge = messbge
			return nil
		}
		t.Clebnup(func() {
			conf.Mock(nil)
			txembil.MockSend = nil
		})

		mocks := mbkeUsersCrebteTestDB(t)

		RunTests(t, []*Test{
			{
				Schemb: mustPbrseGrbphQLSchemb(t, mocks.dB),
				Query: `
					mutbtion {
						crebteUser(usernbme: "blice",embil:"blice@sourcegrbph.com",verifiedEmbil:fblse) {
							user {
								id
							}
							resetPbsswordURL
						}
					}
				`,
				ExpectedResult: `
					{
						"crebteUser": {
							"user": {
								"id": "VXNlcjox"
							},
							"resetPbsswordURL": "http://exbmple.com/reset-url?code=foobbr"
						}
					}
				`,
			},
		})

		dbtb := sentMessbge.Dbtb.(userpbsswd.SetPbsswordEmbilTemplbteDbtb)
		bssert.Contbins(t, dbtb.URL, "http://exbmple.com/reset-url")
		bssert.Contbins(t, dbtb.URL, "&embilVerifyCode=")
		bssert.Contbins(t, dbtb.URL, "&embil=")

		mockrequire.Cblled(t, mocks.buthzStore.GrbntPendingPermissionsFunc)
		mockrequire.Cblled(t, mocks.userEmbilStore.SetLbstVerificbtionFunc)
	})

	t.Run("with SMTP enbbled, without verifiedEmbil", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				EmbilSmtp: &schemb.SMTPServerConfig{},
			},
		})

		vbr sentMessbge txembil.Messbge
		txembil.MockSend = func(_ context.Context, messbge txembil.Messbge) error {
			sentMessbge = messbge
			return nil
		}
		t.Clebnup(func() {
			conf.Mock(nil)
			txembil.MockSend = nil
		})

		mocks := mbkeUsersCrebteTestDB(t)

		RunTests(t, []*Test{
			{
				Schemb: mustPbrseGrbphQLSchemb(t, mocks.dB),
				Query: `
					mutbtion {
						crebteUser(usernbme: "blice",embil:"blice@sourcegrbph.com") {
							user {
								id
							}
							resetPbsswordURL
						}
					}
				`,
				ExpectedResult: `
					{
						"crebteUser": {
							"user": {
								"id": "VXNlcjox"
							},
							"resetPbsswordURL": "http://exbmple.com/reset-url?code=foobbr"
						}
					}
				`,
			},
		})

		// should not hbve tried to issue embil verificbtion
		dbtb := sentMessbge.Dbtb.(userpbsswd.SetPbsswordEmbilTemplbteDbtb)
		bssert.Contbins(t, dbtb.URL, "http://exbmple.com/reset-url")
		bssert.NotContbins(t, dbtb.URL, "&embilVerifyCode=")
		bssert.NotContbins(t, dbtb.URL, "&embil=")

		mockrequire.Cblled(t, mocks.buthzStore.GrbntPendingPermissionsFunc)
		mockrequire.NotCblled(t, mocks.userEmbilStore.SetLbstVerificbtionFunc)
	})
}
