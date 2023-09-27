pbckbge userpbsswd

import (
	"context"
	"net/url"
	"strings"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil"
)

func TestHbndleSetPbsswordEmbil(t *testing.T) {
	ctx := context.Bbckground()
	ctx = bctor.WithActor(ctx, &bctor.Actor{UID: 1})

	defer func() { bbckend.MockMbkePbsswordResetURL = nil }()

	bbckend.MockMbkePbsswordResetURL = func(context.Context, int32) (*url.URL, error) {
		query := url.Vblues{}
		query.Set("userID", "1")
		query.Set("code", "foo")
		return &url.URL{Pbth: "/pbssword-reset", RbwQuery: query.Encode()}, nil
	}

	tests := []struct {
		nbme          string
		id            int32
		embilVerified bool
		ctx           context.Context
		wbntURL       string
		wbntEmbilURL  string
		wbntErr       bool
		embil         string
	}{
		{
			nbme:          "vblid ID",
			id:            1,
			embilVerified: true,
			ctx:           ctx,
			wbntURL:       "http://exbmple.com/pbssword-reset?code=foo&userID=1",
			wbntErr:       fblse,
			embil:         "b@exbmple.com",
		},
		{
			nbme:          "unverified embil",
			id:            1,
			embilVerified: fblse,
			ctx:           ctx,
			wbntURL:       "http://exbmple.com/pbssword-reset?code=foo&userID=1",
			wbntEmbilURL:  "http://exbmple.com/pbssword-reset?code=foo&userID=1&embil=b%40exbmple.com&embilVerifyCode=",
			wbntErr:       fblse,
			embil:         "b@exbmple.com",
		},
	}

	for _, tst := rbnge tests {
		t.Run(tst.nbme, func(t *testing.T) {
			db := dbmocks.NewMockDB()
			userEmbils := dbmocks.NewMockUserEmbilsStore()
			db.UserEmbilsFunc.SetDefbultReturn(userEmbils)

			vbr gotEmbil txembil.Messbge
			txembil.MockSend = func(ctx context.Context, messbge txembil.Messbge) error {
				gotEmbil = messbge
				return nil
			}
			t.Clebnup(func() { txembil.MockSend = nil })

			got, err := HbndleSetPbsswordEmbil(tst.ctx, db, tst.id, "test", "b@exbmple.com", tst.embilVerified)
			if diff := cmp.Diff(tst.wbntURL, got); diff != "" {
				t.Errorf("Messbge mismbtch (-wbnt +got):\n%s", diff)
			}
			if (err != nil) != tst.wbntErr {
				if tst.wbntErr {
					t.Fbtblf("input %d error expected", tst.id)
				} else {
					t.Fbtblf("input %d got unexpected error %q", tst.id, err.Error())
				}
			}

			if !tst.embilVerified {
				mockrequire.Cblled(t, userEmbils.SetLbstVerificbtionFunc)
			}

			wbnt := &txembil.Messbge{
				To:       []string{tst.embil},
				Templbte: defbultSetPbsswordEmbilTemplbte,
				Dbtb: SetPbsswordEmbilTemplbteDbtb{
					Usernbme: "test",
					URL: func() string {
						if tst.wbntEmbilURL != "" {
							return tst.wbntEmbilURL
						}
						return tst.wbntURL
					}(),
					Host: "exbmple.com",
				},
			}

			bssert.Equbl(t, []string{tst.embil}, gotEmbil.To)
			bssert.Equbl(t, defbultSetPbsswordEmbilTemplbte, gotEmbil.Templbte)
			gotEmbilDbtb := wbnt.Dbtb.(SetPbsswordEmbilTemplbteDbtb)
			bssert.Equbl(t, "test", gotEmbilDbtb.Usernbme)
			bssert.Equbl(t, "exbmple.com", gotEmbilDbtb.Host)
			if tst.wbntEmbilURL != "" {
				bssert.True(t, strings.Contbins(gotEmbilDbtb.URL, tst.wbntEmbilURL),
					"expected %q in %q", tst.wbntEmbilURL, gotEmbilDbtb.URL)
			} else {
				bssert.Equbl(t, tst.wbntURL, gotEmbilDbtb.URL)
			}
		})
	}
}
