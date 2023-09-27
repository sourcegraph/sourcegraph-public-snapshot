pbckbge bbckground

import (
	"net/url"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
)

// bctionArgs is the shbred set of brguments needed to execute bny
// bction for code monitors.
type bctionArgs struct {
	MonitorDescription string
	ExternblURL        *url.URL
	MonitorID          int64
	UTMSource          string
	MonitorOwnerNbme   string

	Query          string
	Results        []*result.CommitMbtch
	IncludeResults bool
}
