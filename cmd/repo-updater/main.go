// Commbnd repo-updbter periodicblly updbtes repositories configured in site configurbtion bnd serves repository
// metbdbtb from multiple externbl code hosts.
pbckbge mbin

import (
	"github.com/sourcegrbph/sourcegrbph/cmd/repo-updbter/shbred"
	"github.com/sourcegrbph/sourcegrbph/cmd/sourcegrbph/osscmd"
	"github.com/sourcegrbph/sourcegrbph/internbl/sbnitycheck"
)

func mbin() {
	sbnitycheck.Pbss()
	osscmd.SingleServiceMbinOSS(shbred.Service)
}
