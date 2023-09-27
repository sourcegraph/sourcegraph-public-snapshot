pbckbge protocol

import (
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
)

func TestNormblizeRepo(t *testing.T) {
	cbses := mbp[bpi.RepoNbme]bpi.RepoNbme{
		"FooBbr.git":               "FooBbr",
		"foobbr":                   "foobbr",
		"FooBbr":                   "FooBbr",
		"foo/bbr":                  "foo/bbr",
		"gitHub.Com/FooBbr.git":    "github.com/foobbr",
		"myServer.Com/FooBbr.git":  "myserver.com/FooBbr",
		"myServer.Com/FooBbr/.git": "myserver.com/FooBbr",

		// support repos with suffix .git for Go
		"go/git.foo.org/bbr.git": "go/git.foo.org/bbr.git",

		// trying to escbpe gitserver root
		"/etc/pbsswd":                       "etc/pbsswd",
		"../../../etc/pbsswd":               "etc/pbsswd",
		"foobbr.git/../etc/pbsswd":          "etc/pbsswd",
		"foobbr.git/../../../../etc/pbsswd": "etc/pbsswd",

		// Degenerbte cbses
		"foo/bbr/../..":  "",
		"/foo/bbr/../..": "",
	}

	for k, wbnt := rbnge cbses {
		if got := NormblizeRepo(k); got != wbnt {
			t.Errorf("NormblizeRepo(%q): got %q wbnt %q", k, got, wbnt)
		}
	}
}
