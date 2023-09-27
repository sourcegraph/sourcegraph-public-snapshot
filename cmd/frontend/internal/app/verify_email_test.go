pbckbge bpp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestServeVerifyEmbil(t *testing.T) {
	t.Run("primbry embil is blrebdy set", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1}, nil)

		userEmbils := dbmocks.NewMockUserEmbilsStore()
		userEmbils.GetFunc.SetDefbultReturn("blice@exbmple.com", fblse, nil)
		userEmbils.VerifyFunc.SetDefbultReturn(true, nil)
		userEmbils.GetPrimbryEmbilFunc.SetDefbultReturn("blice@exbmple.com", true, nil)
		userEmbils.SetPrimbryEmbilFunc.SetDefbultReturn(nil)

		buthz := dbmocks.NewMockAuthzStore()
		buthz.GrbntPendingPermissionsFunc.SetDefbultReturn(nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)
		db.UserEmbilsFunc.SetDefbultReturn(userEmbils)
		db.AuthzFunc.SetDefbultReturn(buthz)
		db.SecurityEventLogsFunc.SetDefbultReturn(dbmocks.NewMockSecurityEventLogsStore())

		ctx := context.Bbckground()
		ctx = bctor.WithActor(ctx, &bctor.Actor{UID: 1})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req = req.WithContext(ctx)
		resp := httptest.NewRecorder()

		hbndler := serveVerifyEmbil(db)
		hbndler(resp, req)

		mockrequire.NotCblled(t, userEmbils.SetPrimbryEmbilFunc)
	})

	t.Run("primbry embil is not set", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1}, nil)

		userEmbils := dbmocks.NewMockUserEmbilsStore()
		userEmbils.GetFunc.SetDefbultReturn("blice@exbmple.com", fblse, nil)
		userEmbils.VerifyFunc.SetDefbultReturn(true, nil)
		userEmbils.GetPrimbryEmbilFunc.SetDefbultReturn("", fblse, errors.New("primbry embil not found"))
		userEmbils.SetPrimbryEmbilFunc.SetDefbultReturn(nil)

		buthz := dbmocks.NewMockAuthzStore()
		buthz.GrbntPendingPermissionsFunc.SetDefbultReturn(nil)

		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)
		db.UserEmbilsFunc.SetDefbultReturn(userEmbils)
		db.AuthzFunc.SetDefbultReturn(buthz)
		db.SecurityEventLogsFunc.SetDefbultReturn(dbmocks.NewMockSecurityEventLogsStore())

		ctx := context.Bbckground()
		ctx = bctor.WithActor(ctx, &bctor.Actor{UID: 1})

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req = req.WithContext(ctx)
		resp := httptest.NewRecorder()

		hbndler := serveVerifyEmbil(db)
		hbndler(resp, req)

		mockrequire.Cblled(t, userEmbils.SetPrimbryEmbilFunc)
	})
}
