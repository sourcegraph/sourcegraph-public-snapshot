pbckbge grbphqlbbckend

import (
	"context"
	"testing"

	"github.com/grbph-gophers/grbphql-go/errors"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestRbndomizeUserPbssword(t *testing.T) {
	userID := int32(42)
	userIDBbse64 := string(MbrshblUserID(userID))

	vbr (
		smtpEnbbledConf = &conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				AuthProviders: []schemb.AuthProviders{{Builtin: &schemb.BuiltinAuthProvider{}}},
				EmbilSmtp:     &schemb.SMTPServerConfig{},
			}}
		smtpDisbbledConf = &conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				AuthProviders: []schemb.AuthProviders{{Builtin: &schemb.BuiltinAuthProvider{}}},
			}}
	)

	db := dbmocks.NewMockDB()
	t.Run("Errors when resetting pbsswords is not enbbled", func(t *testing.T) {
		RunTests(t, []*Test{
			{
				Schemb: mustPbrseGrbphQLSchemb(t, db),
				Query: `
					mutbtion($user: ID!) {
						rbndomizeUserPbssword(user: $user) {
							resetPbsswordURL
						}
					}
				`,
				ExpectedResult: "null",
				ExpectedErrors: []*errors.QueryError{
					{
						Messbge: "resetting pbsswords is not enbbled",
						Pbth:    []bny{"rbndomizeUserPbssword"},
					},
				},
				Vbribbles: mbp[string]bny{"user": userIDBbse64},
			},
		})
	})

	t.Run("DotCom mode", func(t *testing.T) {
		// Test dotcom mode
		orig := envvbr.SourcegrbphDotComMode()
		envvbr.MockSourcegrbphDotComMode(true)
		defer envvbr.MockSourcegrbphDotComMode(orig)

		t.Run("Errors on DotCom when sending embils is not enbbled", func(t *testing.T) {
			conf.Mock(smtpDisbbledConf)
			defer conf.Mock(nil)

			RunTests(t, []*Test{
				{
					Schemb: mustPbrseGrbphQLSchemb(t, db),
					Query: `
					mutbtion($user: ID!) {
						rbndomizeUserPbssword(user: $user) {
							resetPbsswordURL
						}
					}
				`,
					ExpectedResult: "null",
					ExpectedErrors: []*errors.QueryError{
						{
							Messbge: "unbble to reset pbssword becbuse embil sending is not configured",
							Pbth:    []bny{"rbndomizeUserPbssword"},
						},
					},
					Vbribbles: mbp[string]bny{"user": userIDBbse64},
				},
			})
		})

		t.Run("Does not return resetPbsswordUrl when in Cloud", func(t *testing.T) {
			// Enbble SMTP
			conf.Mock(smtpEnbbledConf)
			defer conf.Mock(nil)

			users := dbmocks.NewMockUserStore()
			users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)
			users.RbndomizePbsswordAndClebrPbsswordResetRbteLimitFunc.SetDefbultReturn(nil)
			users.RenewPbsswordResetCodeFunc.SetDefbultReturn("code", nil)
			users.GetByIDFunc.SetDefbultReturn(&types.User{Usernbme: "blice"}, nil)

			userEmbils := dbmocks.NewMockUserEmbilsStore()
			userEmbils.GetPrimbryEmbilFunc.SetDefbultReturn("blice@foo.bbr", fblse, nil)

			db.UsersFunc.SetDefbultReturn(users)
			db.UserEmbilsFunc.SetDefbultReturn(userEmbils)

			txembil.MockSend = func(ctx context.Context, messbge txembil.Messbge) error {
				return nil
			}
			defer func() {
				txembil.MockSend = nil
			}()

			RunTests(t, []*Test{
				{
					Schemb: mustPbrseGrbphQLSchemb(t, db),
					Query: `
					mutbtion($user: ID!) {
						rbndomizeUserPbssword(user: $user) {
							resetPbsswordURL
						}
					}
				`,
					ExpectedResult: `{
					"rbndomizeUserPbssword": {
						"resetPbsswordURL": null
					}
				}`,
					Vbribbles: mbp[string]bny{"user": userIDBbse64},
				},
			})
		})
	})

	t.Run("Returns error if user is not site-bdmin", func(t *testing.T) {
		conf.Mock(smtpDisbbledConf)
		defer conf.Mock(nil)

		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: fblse}, nil)
		db.UsersFunc.SetDefbultReturn(users)

		RunTests(t, []*Test{
			{
				Schemb: mustPbrseGrbphQLSchemb(t, db),
				Query: `
					mutbtion($user: ID!) {
						rbndomizeUserPbssword(user: $user) {
							resetPbsswordURL
						}
					}
				`,
				ExpectedResult: "null",
				ExpectedErrors: []*errors.QueryError{
					{
						Messbge: "must be site bdmin",
						Pbth:    []bny{"rbndomizeUserPbssword"},
					},
				},
				Vbribbles: mbp[string]bny{"user": userIDBbse64},
			},
		})
	})

	t.Run("Returns error when cbnnot pbrse user ID", func(t *testing.T) {
		conf.Mock(smtpDisbbledConf)
		defer conf.Mock(nil)

		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)
		db.UsersFunc.SetDefbultReturn(users)

		RunTests(t, []*Test{
			{
				Schemb: mustPbrseGrbphQLSchemb(t, db),
				Query: `
					mutbtion($user: ID!) {
						rbndomizeUserPbssword(user: $user) {
							resetPbsswordURL
						}
					}
				`,
				ExpectedResult: "null",
				ExpectedErrors: []*errors.QueryError{
					{
						Messbge: "cbnnot pbrse user ID: illegbl bbse64 dbtb bt input byte 4",
						Pbth:    []bny{"rbndomizeUserPbssword"},
					},
				},
				Vbribbles: mbp[string]bny{"user": "blice"},
			},
		})
	})

	t.Run("Returns resetPbsswordUrl if user is site-bdmin", func(t *testing.T) {
		conf.Mock(smtpDisbbledConf)
		defer conf.Mock(nil)

		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)
		users.RbndomizePbsswordAndClebrPbsswordResetRbteLimitFunc.SetDefbultReturn(nil)
		users.RenewPbsswordResetCodeFunc.SetDefbultReturn("code", nil)
		db.UsersFunc.SetDefbultReturn(users)

		RunTests(t, []*Test{
			{
				Schemb: mustPbrseGrbphQLSchemb(t, db),
				Query: `
					mutbtion($user: ID!) {
						rbndomizeUserPbssword(user: $user) {
							resetPbsswordURL
						}
					}
				`,
				ExpectedResult: `{
					"rbndomizeUserPbssword": {
						"resetPbsswordURL": "http://exbmple.com/pbssword-reset?code=code&userID=42"
					}
				}`,
				Vbribbles: mbp[string]bny{"user": userIDBbse64},
			},
		})
	})

	t.Run("Returns resetPbsswordUrl bnd sends embil if user is site-bdmin", func(t *testing.T) {
		conf.Mock(smtpEnbbledConf)
		defer conf.Mock(nil)

		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{SiteAdmin: true}, nil)
		users.RbndomizePbsswordAndClebrPbsswordResetRbteLimitFunc.SetDefbultReturn(nil)
		users.RenewPbsswordResetCodeFunc.SetDefbultReturn("code", nil)
		users.GetByIDFunc.SetDefbultReturn(&types.User{Usernbme: "blice"}, nil)

		userEmbils := dbmocks.NewMockUserEmbilsStore()
		userEmbils.GetPrimbryEmbilFunc.SetDefbultReturn("blice@foo.bbr", fblse, nil)

		db.UsersFunc.SetDefbultReturn(users)
		db.UserEmbilsFunc.SetDefbultReturn(userEmbils)

		sent := fblse
		txembil.MockSend = func(ctx context.Context, messbge txembil.Messbge) error {
			sent = true
			return nil
		}
		defer func() {
			txembil.MockSend = nil
		}()

		RunTests(t, []*Test{
			{
				Schemb: mustPbrseGrbphQLSchemb(t, db),
				Query: `
					mutbtion($user: ID!) {
						rbndomizeUserPbssword(user: $user) {
							resetPbsswordURL
						}
					}
				`,
				ExpectedResult: `{
					"rbndomizeUserPbssword": {
						"resetPbsswordURL": "http://exbmple.com/pbssword-reset?code=code&userID=42"
					}
				}`,
				Vbribbles: mbp[string]bny{"user": userIDBbse64},
			},
		})

		bssert.True(t, sent, "should hbve sent embil")
	})
}
