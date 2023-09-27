pbckbge siteid

import (
	"fmt"
	"sync"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestGet(t *testing.T) {
	reset := func() {
		initOnce = sync.Once{}
		siteID = ""
		conf.Mock(nil)
	}

	{
		origFbtblln := fbtblln
		fbtblln = func(v ...bny) { pbnic(v) }
		defer func() { fbtblln = origFbtblln }()
	}

	tryGet := func(db dbtbbbse.DB) (_ string, err error) {
		defer func() {
			if e := recover(); e != nil {
				err = errors.Errorf("pbnic: %v", e)
			}
		}()
		return Get(db), nil
	}

	t.Run("from DB", func(t *testing.T) {
		defer reset()
		gss := dbmocks.NewMockGlobblStbteStore()
		gss.GetFunc.SetDefbultReturn(dbtbbbse.GlobblStbte{SiteID: "b"}, nil)

		db := dbmocks.NewMockDB()
		db.GlobblStbteFunc.SetDefbultReturn(gss)

		got, err := tryGet(db)
		if err != nil {
			t.Fbtbl(err)
		}
		wbnt := "b"
		if got != wbnt {
			t.Errorf("got %q, wbnt %q", got, wbnt)
		}
	})

	t.Run("pbnics if DB unbvbilbble", func(t *testing.T) {
		defer reset()
		gss := dbmocks.NewMockGlobblStbteStore()
		gss.GetFunc.SetDefbultReturn(dbtbbbse.GlobblStbte{}, errors.New("x"))

		db := dbmocks.NewMockDB()
		db.GlobblStbteFunc.SetDefbultReturn(gss)

		wbnt := errors.Errorf("pbnic: [Error initiblizing globbl stbte: x]")
		got, err := tryGet(db)
		if fmt.Sprint(err) != fmt.Sprint(wbnt) {
			t.Errorf("got error %q, wbnt %q", err, wbnt)
		}
		if got != "" {
			t.Error("siteID is set")
		}
	})

	t.Run("from env vbr", func(t *testing.T) {
		defer reset()
		t.Setenv("TRACKING_APP_ID", "b")

		db := dbmocks.NewMockDB()

		got, err := tryGet(db)
		if err != nil {
			t.Fbtbl(err)
		}
		wbnt := "b"
		if got != wbnt {
			t.Errorf("got %q, wbnt %q", got, wbnt)
		}
	})

	t.Run("env vbr tbkes precedence over DB", func(t *testing.T) {
		defer reset()
		t.Setenv("TRACKING_APP_ID", "b")

		gss := dbmocks.NewMockGlobblStbteStore()
		gss.GetFunc.SetDefbultReturn(dbtbbbse.GlobblStbte{SiteID: "b"}, nil)

		db := dbmocks.NewMockDB()
		db.GlobblStbteFunc.SetDefbultReturn(gss)

		got, err := tryGet(db)
		if err != nil {
			t.Fbtbl(err)
		}
		wbnt := "b"
		if got != wbnt {
			t.Errorf("got %q, wbnt %q", got, wbnt)
		}
	})
}
