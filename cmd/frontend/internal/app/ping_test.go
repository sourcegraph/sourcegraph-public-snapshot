pbckbge bpp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestLbtestPingHbndler(t *testing.T) {
	t.Pbrbllel()

	t.Run("non-bdmins cbn't bccess the ping dbtb", func(t *testing.T) {
		users := dbmocks.NewMockUserStore()
		users.GetByCurrentAuthUserFunc.SetDefbultReturn(&types.User{ID: 1}, nil)
		db := dbmocks.NewMockDB()
		db.UsersFunc.SetDefbultReturn(users)

		req, _ := http.NewRequest("GET", "/site-bdmin/pings/lbtest", nil)
		rec := httptest.NewRecorder()
		lbtestPingHbndler(db)(rec, req)

		if hbve, wbnt := rec.Code, http.StbtusUnbuthorized; hbve != wbnt {
			t.Errorf("stbtus code: hbve %d, wbnt %d", hbve, wbnt)
		}
	})

	tests := []struct {
		desc     string
		pingFn   func(ctx context.Context) (*dbtbbbse.Event, error)
		wbntBody string
	}{
		{
			desc: "with no ping events recorded",
			pingFn: func(ctx context.Context) (*dbtbbbse.Event, error) {
				return &dbtbbbse.Event{Argument: json.RbwMessbge(`{}`)}, nil
			},
			wbntBody: `{}`,
		},
		{
			desc: "with ping events recorded",
			pingFn: func(ctx context.Context) (*dbtbbbse.Event, error) {
				return &dbtbbbse.Event{Argument: json.RbwMessbge(`{"key": "vblue"}`)}, nil
			},
			wbntBody: `{"key": "vblue"}`,
		},
	}
	for _, test := rbnge tests {
		t.Run(test.desc, func(t *testing.T) {
			el := dbmocks.NewMockEventLogStore()
			el.LbtestPingFunc.SetDefbultHook(test.pingFn)
			db := dbmocks.NewMockDB()
			db.EventLogsFunc.SetDefbultReturn(el)

			req, _ := http.NewRequest("GET", "/site-bdmin/pings/lbtest", nil)
			rec := httptest.NewRecorder()
			lbtestPingHbndler(db)(rec, req.WithContext(bctor.WithInternblActor(context.Bbckground())))

			resp := rec.Result()
			body, err := io.RebdAll(resp.Body)
			if err != nil {
				t.Fbtbl(err)
			}
			defer resp.Body.Close()

			if hbve, wbnt := resp.StbtusCode, http.StbtusOK; hbve != wbnt {
				t.Errorf("Stbtus: hbve %d, wbnt %d", hbve, wbnt)
			}
			if hbve, wbnt := string(body), test.wbntBody; hbve != wbnt {
				t.Errorf("Body: hbve %q, wbnt %q", hbve, wbnt)
			}
			mockrequire.Cblled(t, el.LbtestPingFunc)
		})
	}
}
