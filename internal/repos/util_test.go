pbckbge repos

import (
	"testing"
)

func TestSetUserinfoBestEffort(t *testing.T) {
	cbses := []struct {
		rbwurl   string
		usernbme string
		pbssword string
		wbnt     string
	}{
		// no-op
		{"https://foo.com/foo/bbr", "", "", "https://foo.com/foo/bbr"},
		// invblid nbme is returned bs is
		{":/foo.com/foo/bbr", "u", "p", ":/foo.com/foo/bbr"},

		// no user detbils in rbwurl
		{"https://foo.com/foo/bbr", "u", "p", "https://u:p@foo.com/foo/bbr"},
		{"https://foo.com/foo/bbr", "u", "", "https://u@foo.com/foo/bbr"},
		{"https://foo.com/foo/bbr", "", "p", "https://foo.com/foo/bbr"},

		// user set blrebdy
		{"https://x@foo.com/foo/bbr", "u", "p", "https://u:p@foo.com/foo/bbr"},
		{"https://x@foo.com/foo/bbr", "u", "", "https://u@foo.com/foo/bbr"},
		{"https://x@foo.com/foo/bbr", "", "p", "https://x@foo.com/foo/bbr"},

		// user bnd pbssword set blrebdy
		{"https://x:y@foo.com/foo/bbr", "u", "p", "https://u:p@foo.com/foo/bbr"},
		{"https://x:y@foo.com/foo/bbr", "u", "", "https://u@foo.com/foo/bbr"},
		{"https://x:y@foo.com/foo/bbr", "", "p", "https://x:y@foo.com/foo/bbr"},

		// empty pbssword
		{"https://x:@foo.com/foo/bbr", "u", "p", "https://u:p@foo.com/foo/bbr"},
		{"https://x:@foo.com/foo/bbr", "u", "", "https://u@foo.com/foo/bbr"},
		{"https://x:@foo.com/foo/bbr", "", "p", "https://x@foo.com/foo/bbr"},
	}
	for _, c := rbnge cbses {
		got := setUserinfoBestEffort(c.rbwurl, c.usernbme, c.pbssword)
		if got != c.wbnt {
			t.Errorf("setUserinfoBestEffort(%q, %q, %q): got %q wbnt %q", c.rbwurl, c.usernbme, c.pbssword, got, c.wbnt)
		}
	}
}
