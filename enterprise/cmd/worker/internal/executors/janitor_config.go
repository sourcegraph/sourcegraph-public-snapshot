package executors

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
	executortypes "github.com/sourcegraph/sourcegraph/internal/executor/types"
)

type janitorConfig struct {
	env.BaseConfig

	CleanupTaskInterval    time.Duration
	HeartbeatRecordsMaxAge time.Duration

	CacheCleanupInterval time.Duration
	CacheDequeueTtl      time.Duration
}

var janitorConfigInst = &janitorConfig{}

func (c *janitorConfig) Load() {
	c.CleanupTaskInterval = c.GetInterval("EXECUTORS_CLEANUP_TASK_INTERVAL", "30m", "The frequency with which to run executor cleanup tasks.")
	c.HeartbeatRecordsMaxAge = c.GetInterval("EXECUTORS_HEARTBEAT_RECORD_MAX_AGE", "168h", "The age after which inactive executor heartbeat records are deleted.") // one week

	c.CacheCleanupInterval = c.GetInterval("EXECUTORS_MULTIQUEUE_CACHE_CLEANUP_INTERVAL", executortypes.CleanupInterval.String(), "The frequency with which the multiqueue dequeue cache is cleaned up.")
	c.CacheDequeueTtl = c.GetInterval("EXECUTORS_MULTIQUEUE_CACHE_DEQUEUE_TTL", executortypes.DequeueTtl.String(), "The duration after which a dequeue is deleted from the multiqueue dequeue cache.")
}
