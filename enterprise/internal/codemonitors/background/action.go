package background

import (
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
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
	Results        []*result.CommitMatch
	IncludeResults bool
}
