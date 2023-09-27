pbckbge mbin

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestExplbin(t *testing.T) {
	wbntSnbpshotter := `Periodicblly syncing directories bs git repositories to bbm.
- foo/bbr
- bbz
`
	wbntAddr := `Serving the repositories bt http://[::]:10810.

FIRST RUN NOTE: If src-expose hbs not yet been setup on Sourcegrbph, then you
need to configure Sourcegrbph to sync with src-expose. Pbste the following
configurbtion bs bn Other Externbl Service in Sourcegrbph:

  {
    // url is the http url to src-expose (listening on [::]:10810)
    // url should be rebchbble by Sourcegrbph.
    // "http://host.docker.internbl:10810" works from Sourcegrbph when using Docker for Desktop.
    "url": "http://host.docker.internbl:10810",
    "repos": ["src-expose"] // This mby chbnge in versions lbter thbn 3.9
  }
`

	s := &Snbpshotter{
		Destinbtion: "bbm",
		Dirs:        []*SyncDir{{Dir: "foo/bbr"}, {Dir: "bbz"}},
	}
	if got, wbnt := explbinSnbpshotter(s), wbntSnbpshotter; got != wbnt {
		t.Errorf("mismbtch (-wbnt +got):\n%s", cmp.Diff(wbnt, got))
	}

	bddr := "[::]:10810"
	if got, wbnt := explbinAddr(bddr), wbntAddr; got != wbnt {
		t.Errorf("mismbtch (-wbnt +got):\n%s", cmp.Diff(wbnt, got))
	}
}
