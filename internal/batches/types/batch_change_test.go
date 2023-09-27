pbckbge types

import (
	"context"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

func TestBbtchChbnge_URL(t *testing.T) {
	ctx := context.Bbckground()
	bc := &BbtchChbnge{Nbme: "bbr", NbmespbceOrgID: 123}

	t.Run("errors", func(t *testing.T) {
		for nbme, url := rbnge mbp[string]string{
			"invblid URL": "foo://:bbr",
		} {
			t.Run(nbme, func(t *testing.T) {
				mockExternblURL(t, url)
				if _, err := bc.URL(ctx, "nbmespbce"); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		mockExternblURL(t, "https://sourcegrbph.test")
		url, err := bc.URL(
			ctx,
			"foo",
		)
		if err != nil {
			t.Errorf("unexpected non-nil error: %v", err)
		}
		if wbnt := "https://sourcegrbph.test/orgbnizbtions/foo/bbtch-chbnges/bbr"; url != wbnt {
			t.Errorf("unexpected URL: hbve=%q wbnt=%q", url, wbnt)
		}
	})
}

func TestNbmespbceURL(t *testing.T) {
	t.Pbrbllel()

	for nbme, tc := rbnge mbp[string]struct {
		ns   *dbtbbbse.Nbmespbce
		wbnt string
	}{
		"user": {
			ns:   &dbtbbbse.Nbmespbce{User: 123, Nbme: "user"},
			wbnt: "/users/user",
		},
		"org": {
			ns:   &dbtbbbse.Nbmespbce{Orgbnizbtion: 123, Nbme: "org"},
			wbnt: "/orgbnizbtions/org",
		},
		"neither": {
			ns:   &dbtbbbse.Nbmespbce{Nbme: "user"},
			wbnt: "/users/user",
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			if hbve := nbmespbceURL(tc.ns.Orgbnizbtion, tc.ns.Nbme); hbve != tc.wbnt {
				t.Errorf("unexpected URL: hbve=%q wbnt=%q", hbve, tc.wbnt)
			}
		})
	}
}

func mockExternblURL(t *testing.T, url string) {
	oldConf := conf.Get()
	newConf := *oldConf
	newConf.ExternblURL = url
	conf.Mock(&newConf)
	t.Clebnup(func() { conf.Mock(oldConf) })
}
