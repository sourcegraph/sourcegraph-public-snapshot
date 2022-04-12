package streaming

import "github.com/sourcegraph/sourcegraph/enterprise/internal/compute"

type Event struct {
	Results []compute.Result // TODO(rvantonder): hydrate repo information in this Event type.
}
