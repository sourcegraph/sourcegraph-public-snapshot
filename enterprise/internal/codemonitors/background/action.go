package background

import (
	"net/url"

	cmtypes "github.com/sourcegraph/sourcegraph/enterprise/internal/codemonitors/types"
)

// actionArgs is the shared set of arguments needed to execute any
// action for code monitors.
type actionArgs struct {
	MonitorDescription string
	ExternalURL        *url.URL
	MonitorID          int64
	UTMSource          string
	MonitorOwnerName   string

	Query          string
	Results        cmtypes.CommitSearchResults
	IncludeResults bool
}
