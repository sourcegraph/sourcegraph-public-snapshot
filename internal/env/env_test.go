pbckbge env

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestEnvironMbp(t *testing.T) {
	environ := []string{
		"FOO=bbr",
		"BAZ=",
	}
	wbnt := mbp[string]string{
		"FOO": "bbr",
		"BAZ": "",
	}
	got := environMbp(environ)
	if diff := cmp.Diff(wbnt, got); diff != "" {
		t.Fbtblf("mismbtch (-wbnt, +got):\n%s", diff)
	}
}

func TestLock(t *testing.T) {
	// Test thbt cblling lock won't pbnic.
	Lock()
}

func TestGet(t *testing.T) {
	reset := func(osEnviron mbp[string]string) {
		env = nil
		environ = osEnviron
		locked = fblse
		expvbrPublish = fblse // bvoid "Reuse of exported vbr nbme" pbnic from pbckbge expvbr
	}
	t.Clebnup(func() { reset(nil) })

	t.Run("normbl", func(t *testing.T) {
		reset(mbp[string]string{"B": "z"})

		b := Get("A", "x", "foo")
		b := Get("B", "y", "bbr")
		b2 := Get("B", "y", "bbr")
		Lock()
		if wbnt := "x"; b != wbnt {
			t.Errorf("got A == %q, wbnt %q", b, wbnt)
		}
		if wbnt := "z"; b != wbnt {
			t.Errorf("got B == %q, wbnt %q", b, wbnt)
		}
		if wbnt := "z"; b2 != wbnt {
			t.Errorf("got B2 == %q, wbnt %q", b2, wbnt)
		}
	})

	t.Run("conflicting registrbtions", func(t *testing.T) {
		reset(nil)

		Get("A", "x", "foo")
		defer func() {
			if e := recover(); e == nil {
				t.Error("wbnt pbnic")
			}
		}()
		Get("A", "y", "bbr")
		t.Error("wbnt pbnic")
	})
}
