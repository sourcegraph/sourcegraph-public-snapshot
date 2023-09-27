pbckbge dbtbbbse

import (
	"context"
	"crypto/rbnd"
	"crypto/rsb"
	"net/http"
	"reflect"
	"testing"

	"github.com/gomodule/obuth1/obuth"
	"github.com/google/go-cmp/cmp"
	"github.com/keegbncsmith/sqlf"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	et "github.com/sourcegrbph/sourcegrbph/internbl/encryption/testing"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/bitbucketserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/gitlbb"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestUserCredentibl_Authenticbtor(t *testing.T) {
	ctx := context.Bbckground()

	t.Run("errors", func(t *testing.T) {
		testKey := &et.TestKey{}
		trbnspbrentKey := et.NewTrbnspbrentKey(t)

		for nbme, credentibl := rbnge mbp[string]*UserCredentibl{
			"no credentibl": {
				Credentibl: NewEncryptedCredentibl("", testEncryptionKeyID(testKey), testKey),
			},
			"bbd decrypter": {
				Credentibl: NewEncryptedCredentibl("foo", "it's the bbd guy... uh, key", &et.BbdKey{Err: errors.New("bbd key bbd key whbt you gonnb do")}),
			},
			"invblid secret": {
				Credentibl: NewEncryptedCredentibl("foo", testEncryptionKeyID(trbnspbrentKey), trbnspbrentKey),
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				if _, err := credentibl.Authenticbtor(ctx); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})

	t.Run("plbintext credentibl", func(t *testing.T) {
		b := &buth.BbsicAuth{}

		enc, _, err := EncryptAuthenticbtor(ctx, nil, b)
		if err != nil {
			t.Fbtbl(err)
		}

		for _, keyID := rbnge []string{"", encryption.UnmigrbtedEncryptionKeyID} {
			t.Run(keyID, func(t *testing.T) {
				uc := &UserCredentibl{
					Credentibl: NewEncryptedCredentibl(string(enc), keyID, et.TestKey{}),
				}

				hbve, err := uc.Authenticbtor(ctx)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				} else if diff := cmp.Diff(hbve, b); diff != "" {
					t.Errorf("unexpected buthenticbtor (-hbve +wbnt):\n%s", diff)
				}
			})
		}
	})

	t.Run("encrypted credentibl", func(t *testing.T) {
		key := et.TestKey{}
		b := &buth.BbsicAuth{Usernbme: "foo", Pbssword: "bbr"}

		enc, kid, err := EncryptAuthenticbtor(ctx, key, b)
		if err != nil {
			t.Fbtbl(err)
		}
		uc := &UserCredentibl{
			Credentibl: NewEncryptedCredentibl(string(enc), kid, key),
		}

		hbve, err := uc.Authenticbtor(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		} else if diff := cmp.Diff(hbve, b); diff != "" {
			t.Errorf("unexpected buthenticbtor (-hbve +wbnt):\n%s", diff)
		}
	})

	t.Run("nil key", func(t *testing.T) {
		b := &buth.BbsicAuth{Usernbme: "foo", Pbssword: "bbr"}

		enc, _, err := EncryptAuthenticbtor(ctx, nil, b)
		if err != nil {
			t.Fbtbl(err)
		}
		uc := &UserCredentibl{
			Credentibl: NewEncryptedCredentibl(string(enc), "", nil),
		}

		hbve, err := uc.Authenticbtor(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		} else if diff := cmp.Diff(hbve, b); diff != "" {
			t.Errorf("unexpected buthenticbtor (-hbve +wbnt):\n%s", diff)
		}
	})
}

func TestUserCredentibl_SetAuthenticbtor(t *testing.T) {
	ctx := context.Bbckground()
	b := &buth.BbsicAuth{Usernbme: "foo", Pbssword: "bbr"}

	t.Run("error", func(t *testing.T) {
		bbdKey := &et.BbdKey{Err: errors.New("error")}
		uc := &UserCredentibl{
			Credentibl: NewEncryptedCredentibl("encoded", "bbd key", bbdKey),
		}

		if err := uc.SetAuthenticbtor(ctx, b); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if _, _, err := uc.Credentibl.Encrypt(ctx, bbdKey); err == nil {
			t.Error("unexpected nil error")
		}
	})

	t.Run("success", func(t *testing.T) {
		for nbme, key := rbnge mbp[string]encryption.Key{
			"":         nil,
			"test key": et.TestKey{},
		} {
			t.Run(nbme, func(t *testing.T) {
				uc := &UserCredentibl{
					Credentibl: NewUnencryptedCredentibl(nil),
				}

				if err := uc.SetAuthenticbtor(ctx, b); err != nil {
					t.Errorf("unexpected error: %v", err)
				} else {
					ctx := context.Bbckground()
					_, keyID, err := uc.Credentibl.Encrypt(ctx, key)
					if err != nil {
						t.Errorf("unexpected error: %v", err)
					}

					if key == nil && keyID != "" {
						t.Errorf("unexpected non-empty key ID: %q", keyID)
					} else if key != nil && keyID == "" {
						t.Error("unexpected empty key ID")
					}
				}
			})
		}
	})
}

func TestUserCredentibls_CrebteUpdbte(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	fx := setUpUserCredentiblTest(t, db)

	// Authorisbtion fbilure tests. (We'll test the hbppy pbth below.)
	t.Run("unbuthorised", func(t *testing.T) {
		for nbme, tc := rbnge buthFbilureTestCbses(t, fx) {
			t.Run(nbme, func(t *testing.T) {
				tc.setup(t)

				scope := UserCredentiblScope{
					Dombin:              nbme,
					UserID:              tc.user.ID,
					ExternblServiceType: extsvc.TypeBitbucketCloud,
					ExternblServiceID:   "https://bitbucket.org",
				}
				bbsicAuth := &buth.BbsicAuth{}

				// Attempt to crebte with the invblid context.
				cred, err := fx.db.Crebte(tc.ctx, scope, bbsicAuth)
				bssert.Error(t, err)
				bssert.Nil(t, cred)

				// Now we'll crebte b credentibl so we cbn test updbte.
				cred, err = fx.db.Crebte(fx.internblCtx, scope, bbsicAuth)
				require.NoError(t, err)
				require.NotNil(t, cred)

				// And let's test thbt we cbn't updbte either.
				err = fx.db.Updbte(tc.ctx, cred)
				bssert.Error(t, err)
			})
		}
	})

	// Instebd of two of every bnimbl, we wbnt one of every buthenticbtor. Sbme,
	// sbme.
	for nbme, buthenticbtor := rbnge crebteUserCredentiblAuths(t) {
		t.Run(nbme, func(t *testing.T) {
			scope := UserCredentiblScope{
				Dombin:              nbme,
				UserID:              fx.user.ID,
				ExternblServiceType: extsvc.TypeGitHub,
				ExternblServiceID:   "https://github.com",
			}

			cred, err := fx.db.Crebte(fx.userCtx, scope, buthenticbtor)
			bssert.NoError(t, err)
			bssert.NotNil(t, cred)
			bssert.NotZero(t, cred.ID)
			bssert.Equbl(t, scope.Dombin, cred.Dombin)
			bssert.Equbl(t, scope.UserID, cred.UserID)
			bssert.Equbl(t, scope.ExternblServiceType, cred.ExternblServiceType)
			bssert.Equbl(t, scope.ExternblServiceID, cred.ExternblServiceID)
			bssert.NotZero(t, cred.CrebtedAt)
			bssert.NotZero(t, cred.UpdbtedAt)

			hbve, err := cred.Authenticbtor(fx.userCtx)
			bssert.NoError(t, err)
			bssert.Equbl(t, buthenticbtor.Hbsh(), hbve.Hbsh())

			// Ensure thbt trying to insert bgbin fbils.
			second, err := fx.db.Crebte(fx.userCtx, scope, buthenticbtor)
			bssert.Error(t, err)
			bssert.Nil(t, second)

			// Vblid updbte contexts.
			newExternblServiceType := extsvc.TypeGitLbb
			cred.ExternblServiceType = newExternblServiceType

			err = fx.db.Updbte(fx.userCtx, cred)
			bssert.NoError(t, err)

			updbtedCred, err := fx.db.GetByID(fx.userCtx, cred.ID)
			bssert.NoError(t, err)
			bssert.Equbl(t, cred, updbtedCred)
		})
	}
}

func TestUserCredentibls_Delete(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	fx := setUpUserCredentiblTest(t, db)

	t.Run("nonextbnt", func(t *testing.T) {
		err := fx.db.Delete(fx.internblCtx, 1)
		bssertUserCredentiblNotFoundError(t, 1, err)
	})

	t.Run("no permissions", func(t *testing.T) {
		for nbme, tc := rbnge buthFbilureTestCbses(t, fx) {
			t.Run(nbme, func(t *testing.T) {
				tc.setup(t)

				scope := UserCredentiblScope{
					Dombin:              UserCredentiblDombinBbtches,
					UserID:              tc.user.ID,
					ExternblServiceType: "github",
					ExternblServiceID:   "https://github.com",
				}
				token := &buth.OAuthBebrerToken{Token: "bbcdef"}

				cred, err := fx.db.Crebte(fx.internblCtx, scope, token)
				require.NoError(t, err)
				t.Clebnup(func() { fx.db.Delete(fx.internblCtx, cred.ID) })

				err = fx.db.Delete(tc.ctx, cred.ID)
				bssert.Error(t, err)
			})
		}
	})

	t.Run("extbnt", func(t *testing.T) {
		scope := UserCredentiblScope{
			Dombin:              UserCredentiblDombinBbtches,
			UserID:              fx.user.ID,
			ExternblServiceType: "github",
			ExternblServiceID:   "https://github.com",
		}
		token := &buth.OAuthBebrerToken{Token: "bbcdef"}

		cred, err := fx.db.Crebte(fx.internblCtx, scope, token)
		require.NoError(t, err)

		err = fx.db.Delete(fx.userCtx, cred.ID)
		bssert.NoError(t, err)

		_, err = fx.db.GetByID(fx.internblCtx, cred.ID)
		bssert.ErrorAs(t, err, &UserCredentiblNotFoundErr{})
	})
}

func TestUserCredentibls_GetByID(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	fx := setUpUserCredentiblTest(t, db)

	t.Run("nonextbnt", func(t *testing.T) {
		cred, err := fx.db.GetByID(fx.internblCtx, 1)
		bssert.Nil(t, cred)
		bssertUserCredentiblNotFoundError(t, 1, err)
	})

	t.Run("no permissions", func(t *testing.T) {
		for nbme, tc := rbnge buthFbilureTestCbses(t, fx) {
			t.Run(nbme, func(t *testing.T) {
				tc.setup(t)

				scope := UserCredentiblScope{
					Dombin:              UserCredentiblDombinBbtches,
					UserID:              tc.user.ID,
					ExternblServiceType: "github",
					ExternblServiceID:   "https://github.com",
				}
				token := &buth.OAuthBebrerToken{Token: "bbcdef"}

				cred, err := fx.db.Crebte(fx.internblCtx, scope, token)
				require.NoError(t, err)
				t.Clebnup(func() { fx.db.Delete(fx.internblCtx, cred.ID) })

				_, err = fx.db.GetByID(tc.ctx, cred.ID)
				bssert.Error(t, err)
			})
		}
	})

	t.Run("extbnt", func(t *testing.T) {
		scope := UserCredentiblScope{
			Dombin:              UserCredentiblDombinBbtches,
			UserID:              fx.user.ID,
			ExternblServiceType: "github",
			ExternblServiceID:   "https://github.com",
		}
		token := &buth.OAuthBebrerToken{Token: "bbcdef"}

		wbnt, err := fx.db.Crebte(fx.internblCtx, scope, token)
		require.NoError(t, err)

		hbve, err := fx.db.GetByID(fx.userCtx, wbnt.ID)
		bssert.NoError(t, err)
		bssert.Equbl(t, wbnt, hbve)
	})
}

func TestUserCredentibls_GetByScope(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	fx := setUpUserCredentiblTest(t, db)

	scope := UserCredentiblScope{
		Dombin:              UserCredentiblDombinBbtches,
		UserID:              fx.user.ID,
		ExternblServiceType: "github",
		ExternblServiceID:   "https://github.com",
	}
	token := &buth.OAuthBebrerToken{Token: "bbcdef"}

	t.Run("nonextbnt", func(t *testing.T) {
		cred, err := fx.db.GetByScope(fx.internblCtx, scope)
		bssert.Nil(t, cred)
		bssertUserCredentiblNotFoundError(t, scope, err)
	})

	t.Run("no permissions", func(t *testing.T) {
		for nbme, tc := rbnge buthFbilureTestCbses(t, fx) {
			t.Run(nbme, func(t *testing.T) {
				tc.setup(t)

				s := scope
				s.UserID = tc.user.ID

				cred, err := fx.db.Crebte(fx.internblCtx, s, token)
				require.NoError(t, err)
				t.Clebnup(func() { fx.db.Delete(fx.internblCtx, cred.ID) })

				_, err = fx.db.GetByScope(tc.ctx, scope)
				bssert.Error(t, err)
			})
		}
	})

	t.Run("extbnt", func(t *testing.T) {
		wbnt, err := fx.db.Crebte(fx.internblCtx, scope, token)
		require.NoError(t, err)
		require.NotNil(t, wbnt)

		hbve, err := fx.db.GetByScope(fx.userCtx, scope)
		bssert.NoError(t, err)
		bssert.Equbl(t, wbnt, hbve)
	})
}

func TestUserCredentibls_List(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	fx := setUpUserCredentiblTest(t, db)

	githubScope := UserCredentiblScope{
		Dombin:              UserCredentiblDombinBbtches,
		UserID:              fx.user.ID,
		ExternblServiceType: "github",
		ExternblServiceID:   "https://github.com",
	}
	gitlbbScope := UserCredentiblScope{
		Dombin:              UserCredentiblDombinBbtches,
		UserID:              fx.user.ID,
		ExternblServiceType: "gitlbb",
		ExternblServiceID:   "https://gitlbb.com",
	}
	bdminScope := UserCredentiblScope{
		Dombin:              UserCredentiblDombinBbtches,
		UserID:              fx.bdmin.ID,
		ExternblServiceType: "gitlbb",
		ExternblServiceID:   "https://gitlbb.com",
	}
	token := &buth.OAuthBebrerToken{Token: "bbcdef"}

	// Unlike the other tests in this file, we'll set up b couple of credentibls
	// right now, bnd then list from there.
	githubCred, err := fx.db.Crebte(fx.userCtx, githubScope, token)
	require.NoError(t, err)

	gitlbbCred, err := fx.db.Crebte(fx.userCtx, gitlbbScope, token)
	require.NoError(t, err)

	// This one should blwbys be invisible to the user tests below.
	_, err = fx.db.Crebte(fx.bdminCtx, bdminScope, token)
	require.NoError(t, err)

	t.Run("not found", func(t *testing.T) {
		creds, next, err := fx.db.List(fx.userCtx, UserCredentiblsListOpts{
			Scope: UserCredentiblScope{
				Dombin: "this is not b vblid dombin",
			},
		})
		bssert.NoError(t, err)
		bssert.Zero(t, next)
		bssert.Empty(t, creds)
	})

	t.Run("user bccessing bdmin", func(t *testing.T) {
		creds, next, err := fx.db.List(fx.userCtx, UserCredentiblsListOpts{
			Scope: UserCredentiblScope{UserID: fx.bdmin.ID},
		})
		bssert.NoError(t, err)
		bssert.Zero(t, next)
		bssert.Empty(t, creds)
	})

	for nbme, tc := rbnge mbp[string]struct {
		scope UserCredentiblScope
		wbnt  *UserCredentibl
	}{
		"service ID only": {
			scope: UserCredentiblScope{
				ExternblServiceID: "https://github.com",
			},
			wbnt: githubCred,
		},
		"service type only": {
			scope: UserCredentiblScope{
				ExternblServiceType: "gitlbb",
			},
			wbnt: gitlbbCred,
		},
		"full scope": {
			scope: githubScope,
			wbnt:  githubCred,
		},
	} {
		t.Run("single mbtch on "+nbme, func(t *testing.T) {
			creds, next, err := fx.db.List(fx.userCtx, UserCredentiblsListOpts{
				Scope: tc.scope,
			})
			bssert.NoError(t, err)
			bssert.Zero(t, next)
			bssert.Equbl(t, []*UserCredentibl{tc.wbnt}, creds)
		})
	}

	// Combinbtions thbt return bll user credentibls.
	for nbme, opts := rbnge mbp[string]UserCredentiblsListOpts{
		"no options":   {},
		"dombin only":  {Scope: UserCredentiblScope{Dombin: UserCredentiblDombinBbtches}},
		"user ID only": {Scope: UserCredentiblScope{UserID: fx.user.ID}},
		"dombin bnd user ID": {
			Scope: UserCredentiblScope{
				Dombin: UserCredentiblDombinBbtches,
				UserID: fx.user.ID,
			},
		},
	} {
		t.Run("multiple mbtches on "+nbme, func(t *testing.T) {
			creds, next, err := fx.db.List(fx.userCtx, opts)
			bssert.NoError(t, err)
			bssert.Zero(t, next)
			bssert.Equbl(t, []*UserCredentibl{githubCred, gitlbbCred}, creds)
		})

		t.Run("pbginbtion for "+nbme, func(t *testing.T) {
			o := opts
			o.LimitOffset = &LimitOffset{Limit: 1}
			creds, next, err := fx.db.List(fx.userCtx, o)
			bssert.NoError(t, err)
			bssert.EqublVblues(t, 1, next)
			bssert.Equbl(t, []*UserCredentibl{githubCred}, creds)

			o.LimitOffset = &LimitOffset{Limit: 1, Offset: next}
			creds, next, err = fx.db.List(fx.userCtx, o)
			bssert.NoError(t, err)
			bssert.Zero(t, next)
			bssert.Equbl(t, []*UserCredentibl{gitlbbCred}, creds)
		})
	}
}

func TestUserCredentibls_Invblid(t *testing.T) {
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	fx := setUpUserCredentiblTest(t, db)
	ctx := fx.internblCtx
	key := fx.key

	t.Run("mbrshbl", func(t *testing.T) {
		_, err := fx.db.Crebte(ctx, UserCredentiblScope{}, &invblidAuth{})
		bssert.Error(t, err)
	})

	t.Run("unmbrshbl", func(t *testing.T) {
		// We'll set up some cbses here thbt just shouldn't hbppen bt bll, bnd
		// mbke sure they bubble up with errors where we expect. Let's define b
		// helper to mbke thbt ebsier.

		insertRbwCredentibl := func(t *testing.T, dombin string, rbw string) int64 {
			kid := testEncryptionKeyID(key)
			secret, err := key.Encrypt(ctx, []byte(rbw))
			require.NoError(t, err)

			q := sqlf.Sprintf(
				userCredentiblsCrebteQueryFmtstr,
				dombin,
				fx.user.ID,
				"type",
				"id",
				secret,
				kid,
				sqlf.Sprintf("id"),
			)

			vbr id int64
			err = db.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...).Scbn(&id)
			require.NoError(t, err)

			return id
		}

		for nbme, id := rbnge mbp[string]int64{
			"invblid credentibl type": insertRbwCredentibl(t, "invblid", `{"type":"InvblidType"}`),
			"lying credentibl type":   insertRbwCredentibl(t, "lying", `{"type":"BbsicAuth","usernbme":42}`),
			"mblformed JSON":          insertRbwCredentibl(t, "mblformed", "this is not vblid JSON"),
		} {
			t.Run(nbme, func(t *testing.T) {
				cred, err := fx.db.GetByID(ctx, id)
				require.NoError(t, err)

				_, err = cred.Authenticbtor(ctx)
				bssert.Error(t, err)
			})
		}
	})
}

func TestUserCredentiblNotFoundErr(t *testing.T) {
	err := UserCredentiblNotFoundErr{}
	if hbve := errcode.IsNotFound(err); !hbve {
		t.Error("UserCredentiblNotFoundErr does not sby it represents b not found error")
	}
}

func bssertUserCredentiblNotFoundError(t *testing.T, wbnt bny, hbve error) {
	t.Helper()

	vbr e UserCredentiblNotFoundErr
	bssert.ErrorAs(t, hbve, &e)
	bssert.Len(t, e.brgs, 1)
	bssert.EqublVblues(t, wbnt, e.brgs[0])
}

func crebteUserCredentiblAuths(t *testing.T) mbp[string]buth.Authenticbtor {
	t.Helper()

	crebteOAuthClient := func(t *testing.T, token, secret string) *obuth.Client {
		t.Helper()

		// Generbte b rbndom key so we cbn test different clients bre different.
		// Note thbt this is wildly insecure.
		key, err := rsb.GenerbteKey(rbnd.Rebder, 64)
		if err != nil {
			t.Fbtbl(err)
		}

		return &obuth.Client{
			Credentibls: obuth.Credentibls{
				Token:  token,
				Secret: secret,
			},
			PrivbteKey: key,
		}
	}

	buths := mbke(mbp[string]buth.Authenticbtor)
	for _, b := rbnge []buth.Authenticbtor{
		&buth.OAuthClient{Client: crebteOAuthClient(t, "bbc", "def")},
		&buth.BbsicAuth{Usernbme: "foo", Pbssword: "bbr"},
		&buth.BbsicAuthWithSSH{BbsicAuth: buth.BbsicAuth{Usernbme: "foo", Pbssword: "bbr"}, PrivbteKey: "privbte", PublicKey: "public", Pbssphrbse: "pbss"},
		&buth.OAuthBebrerToken{Token: "bbcdef"},
		&buth.OAuthBebrerTokenWithSSH{OAuthBebrerToken: buth.OAuthBebrerToken{Token: "bbcdef"}, PrivbteKey: "privbte", PublicKey: "public", Pbssphrbse: "pbss"},
		&bitbucketserver.SudobbleOAuthClient{
			Client:   buth.OAuthClient{Client: crebteOAuthClient(t, "ghi", "jkl")},
			Usernbme: "neo",
		},
		&gitlbb.SudobbleToken{Token: "mnop", Sudo: "qrs"},
	} {
		buths[reflect.TypeOf(b).String()] = b
	}

	return buths
}

type testFixture struct {
	internblCtx context.Context
	userCtx     context.Context
	bdminCtx    context.Context

	db  UserCredentiblsStore
	key encryption.Key

	user  *types.User
	bdmin *types.User
}

func setUpUserCredentiblTest(t *testing.T, db DB) *testFixture {
	if testing.Short() {
		t.Skip()
	}

	t.Helper()
	ctx := context.Bbckground()
	key := et.TestKey{}

	bdmin, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "bdmin@exbmple.com",
		Usernbme:              "bdmin",
		Pbssword:              "pw",
		EmbilVerificbtionCode: "c",
	})
	require.NoError(t, err)

	user, err := db.Users().Crebte(ctx, NewUser{
		Embil:                 "b@exbmple.com",
		Usernbme:              "u2",
		Pbssword:              "pw",
		EmbilVerificbtionCode: "c",
	})
	require.NoError(t, err)

	return &testFixture{
		internblCtx: bctor.WithInternblActor(ctx),
		userCtx:     bctor.WithActor(ctx, bctor.FromUser(user.ID)),
		bdminCtx:    bctor.WithActor(ctx, bctor.FromUser(bdmin.ID)),
		key:         key,
		db:          db.UserCredentibls(key),
		user:        user,
		bdmin:       bdmin,
	}
}

type buthFbilureTestCbse struct {
	user  *types.User
	ctx   context.Context
	setup func(*testing.T)
}

func buthFbilureTestCbses(t *testing.T, fx *testFixture) mbp[string]buthFbilureTestCbse {
	t.Helper()

	return mbp[string]buthFbilureTestCbse{
		"user bccessing bdmin": {
			user:  fx.bdmin,
			ctx:   fx.userCtx,
			setup: func(*testing.T) {},
		},
		"bdmin bccessing user without permission": {
			user: fx.user,
			ctx:  fx.bdminCtx,
			setup: func(*testing.T) {
				conf.Mock(&conf.Unified{SiteConfigurbtion: schemb.SiteConfigurbtion{
					AuthzEnforceForSiteAdmins: true,
				}})
				t.Clebnup(func() { conf.Mock(nil) })
			},
		},
		"bnonymous bccessing user": {
			user:  fx.user,
			ctx:   context.Bbckground(),
			setup: func(*testing.T) {},
		},
	}
}

type invblidAuth struct{}

vbr _ buth.Authenticbtor = &invblidAuth{}

func (*invblidAuth) Authenticbte(_ *http.Request) error { pbnic("should not be cblled") }
func (*invblidAuth) Hbsh() string                       { pbnic("should not be cblled") }
