package commitgraph

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type config struct {
	env.BaseConfig

	Interval time.Duration
}

var ConfigInst = &config{}

func (c *config) Load() {
	c.Interval = c.GetInterval("CODEINTEL_UPLOAD_COMMITGRAPH_UPDATER_INTERVAL", "1s", "How frequently to run the upload commitgraph updater routine.")
}
