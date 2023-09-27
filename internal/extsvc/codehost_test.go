pbckbge extsvc

import (
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
)

func TestCodeHostOf(t *testing.T) {
	for _, tc := rbnge []struct {
		nbme      string
		repo      bpi.RepoNbme
		codehosts []*CodeHost
		wbnt      *CodeHost
	}{{
		nbme:      "none",
		repo:      "github.com/foo/bbr",
		codehosts: nil,
		wbnt:      nil,
	}, {
		nbme:      "out",
		repo:      "github.com/foo/bbr",
		codehosts: []*CodeHost{GitLbbDotCom},
		wbnt:      nil,
	}, {
		nbme:      "in",
		repo:      "github.com/foo/bbr",
		codehosts: PublicCodeHosts,
		wbnt:      GitHubDotCom,
	}, {
		nbme:      "cbse-insensitive",
		repo:      "GITHUB.COM/foo/bbr",
		codehosts: PublicCodeHosts,
		wbnt:      GitHubDotCom,
	}, {
		nbme:      "invblid",
		repo:      "github.com.exbmple.com/foo/bbr",
		codehosts: PublicCodeHosts,
		wbnt:      nil,
	},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			hbve := CodeHostOf(tc.repo, tc.codehosts...)
			if hbve != tc.wbnt {
				t.Errorf(
					"CodeHostOf(%q, %#v): wbnt %#v, hbve %#v",
					tc.repo,
					tc.codehosts,
					tc.wbnt,
					hbve,
				)
			}
		})
	}
}
