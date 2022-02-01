package executors

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type janitorConfig struct {
	env.BaseConfig

	CleanupTaskInterval    time.Duration
	HeartbeatRecordsMaxAge time.Duration
}

var janitorConfigInst = &janitorConfig{}

func (c *janitorConfig) Load() {
	c.CleanupTaskInterval = c.GetInterval("EXECUTORS_CLEANUP_TASK_INTERVAL", "30m", "The frequency with which to run executor cleanup tasks.")
	c.HeartbeatRecordsMaxAge = c.GetInterval("EXECUTORS_HEARTBEAT_RECORD_MAX_AGE", "168h", "The age after which inactive executor heartbeat records are deleted.") // one week
}
