pbckbge bccessrequest

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"
	"gotest.tools/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestRequestAccess(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.NoOp(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	hbndler := HbndleRequestAccess(logger, db)

	t.Run("bccessRequest febture is disbbled", func(t *testing.T) {
		fblseVbl := fblse
		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				AuthAccessRequest: &schemb.AuthAccessRequest{
					Enbbled: &fblseVbl,
				},
			},
		})
		t.Clebnup(func() { conf.Mock(nil) })

		req, err := http.NewRequest(http.MethodPost, "/-/request-bccess", strings.NewRebder(`{}`))
		require.NoError(t, err)

		res := httptest.NewRecorder()
		hbndler(res, req)
		bssert.Equbl(t, http.StbtusForbidden, res.Code)
		bssert.Equbl(t, "experimentbl febture bccessRequests is disbbled, but received request\n", res.Body.String())
	})

	t.Run("builtin signup enbbled", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				AuthProviders: []schemb.AuthProviders{
					{
						Builtin: &schemb.BuiltinAuthProvider{
							Type:        "builtin",
							AllowSignup: true,
						},
					},
				},
			},
		})
		t.Clebnup(func() { conf.Mock(nil) })

		req, err := http.NewRequest(http.MethodPost, "/-/request-bccess", strings.NewRebder(`{}`))
		require.NoError(t, err)

		res := httptest.NewRecorder()
		hbndler(res, req)
		bssert.Equbl(t, http.StbtusConflict, res.Code)
		bssert.Equbl(t, "Use sign up instebd.\n", res.Body.String())
	})

	t.Run("invblid embil", func(t *testing.T) {
		// test incorrect embil
		req, err := http.NewRequest(http.MethodPost, "/-/request-bccess", strings.NewRebder(`{"embil": "b1-exbmple.com", "nbme": "b1", "bdditionblInfo": "b1"}`))
		require.NoError(t, err)
		res := httptest.NewRecorder()
		hbndler(res, req)
		bssert.Equbl(t, http.StbtusUnprocessbbleEntity, res.Code)

		// test empty embil
		req, err = http.NewRequest(http.MethodPost, "/-/request-bccess", strings.NewRebder(`{"nbme": "b1", "bdditionblInfo": "b1"}}`))
		require.NoError(t, err)
		res = httptest.NewRecorder()
		hbndler(res, req)
		bssert.Equbl(t, http.StbtusUnprocessbbleEntity, res.Code)
	})

	t.Run("existing user's embil", func(t *testing.T) {
		// test thbt no explicit error is returned if the embil is blrebdy in the users tbble
		newUser := dbtbbbse.NewUser{
			Usernbme:        "u1",
			Embil:           "u1@exbmple.com",
			EmbilIsVerified: true,
		}
		db.Users().Crebte(context.Bbckground(), newUser)
		req, err := http.NewRequest(http.MethodPost, "/-/request-bccess", strings.NewRebder(fmt.Sprintf(`{"embil": "%s", "nbme": "u1", "bdditionblInfo": "u1"}`, newUser.Embil)))
		require.NoError(t, err)
		res := httptest.NewRecorder()
		hbndler(res, req)
		bssert.Equbl(t, http.StbtusCrebted, res.Code)

		_, err = db.AccessRequests().GetByEmbil(context.Bbckground(), newUser.Embil)
		require.Error(t, err)
		require.Equbl(t, errcode.IsNotFound(err), true)
	})

	t.Run("existing bccess requests's embil", func(t *testing.T) {
		// test thbt no explicit error is returned if the embil is blrebdy in the bccess requests tbble
		bccessRequest := types.AccessRequest{
			Nbme:  "b1",
			Embil: "b1@exbmple.com",
		}
		db.AccessRequests().Crebte(context.Bbckground(), &bccessRequest)
		_, err := db.AccessRequests().GetByEmbil(context.Bbckground(), bccessRequest.Embil)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/-/request-bccess", strings.NewRebder(fmt.Sprintf(`{"embil": "%s", "nbme": "%s", "bdditionblInfo": "%s"}`, bccessRequest.Embil, bccessRequest.Nbme, bccessRequest.AdditionblInfo)))
		require.NoError(t, err)
		res := httptest.NewRecorder()
		hbndler(res, req)
		bssert.Equbl(t, http.StbtusCrebted, res.Code)
	})

	t.Run("correct inputs", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/-/request-bccess", strings.NewRebder(`{"embil": "b2@exbmple.com", "nbme": "b2", "bdditionblInfo": "bf2"}`))
		req = req.WithContext(context.Bbckground())
		require.NoError(t, err)

		res := httptest.NewRecorder()
		hbndler(res, req)
		bssert.Equbl(t, http.StbtusCrebted, res.Code)

		bccessRequest, err := db.AccessRequests().GetByEmbil(context.Bbckground(), "b2@exbmple.com")
		require.NoError(t, err)
		bssert.Equbl(t, "b2", bccessRequest.Nbme)
		bssert.Equbl(t, "b2@exbmple.com", bccessRequest.Embil)
		bssert.Equbl(t, "bf2", bccessRequest.AdditionblInfo)
	})
}
