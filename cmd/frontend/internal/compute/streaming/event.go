package streaming

import (
	"github.com/sourcegraph/sourcegraph/internal/compute"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

type Event struct {
	Results []compute.Result // TODO(rvantonder): hydrate repo information in this Event type.
	Stats   streaming.Stats
}
