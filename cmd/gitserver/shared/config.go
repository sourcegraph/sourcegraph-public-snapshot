package shared

import (
	"path/filepath"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func LoadConfig() *Config {
	var config Config
	config.Load()
	return &config
}

type Config struct {
	env.BaseConfig

	ReposDir         string
	CoursierCacheDir string

	SyncRepoStateInterval          time.Duration
	SyncRepoStateBatchSize         int
	SyncRepoStateUpdatePerSecond   int
	BatchLogGlobalConcurrencyLimit int

	RateLimitSyncerLimitPerSecond int

	JanitorReposDesiredPercentFree int
	JanitorInterval                time.Duration
}

func (c *Config) Load() {
	c.ReposDir = c.Get("SRC_REPOS_DIR", "/data/repos", "Root dir containing repos.")
	if c.ReposDir == "" {
		c.AddError(errors.New("SRC_REPOS_DIR is required"))
	}

	// if COURSIER_CACHE_DIR is set, try create that dir and use it. If not set, use the SRC_REPOS_DIR value (or default).
	c.CoursierCacheDir = c.GetOptional("COURSIER_CACHE_DIR", "Directory in which coursier data is cached for JVM package repos.")
	if c.CoursierCacheDir == "" && c.ReposDir != "" {
		c.CoursierCacheDir = filepath.Join(c.ReposDir, "coursier")
	}

	c.SyncRepoStateInterval = c.GetInterval("SRC_REPOS_SYNC_STATE_INTERVAL", "10m", "Interval between state syncs")
	c.SyncRepoStateBatchSize = c.GetInt("SRC_REPOS_SYNC_STATE_BATCH_SIZE", "500", "Number of updates to perform per batch")
	c.SyncRepoStateUpdatePerSecond = c.GetInt("SRC_REPOS_SYNC_STATE_UPSERT_PER_SEC", "500", "The number of updated rows allowed per second across all gitserver instances")
	c.BatchLogGlobalConcurrencyLimit = c.GetInt("SRC_BATCH_LOG_GLOBAL_CONCURRENCY_LIMIT", "256", "The maximum number of in-flight Git commands from all /batch-log requests combined")

	// 80 per second (4800 per minute) is well below our alert threshold of 30k per minute.
	c.RateLimitSyncerLimitPerSecond = c.GetInt("SRC_REPOS_SYNC_RATE_LIMIT_RATE_PER_SECOND", "80", "Rate limit applied to rate limit syncing")

	// Align these variables with the 'disk_space_remaining' alerts in monitoring
	c.JanitorReposDesiredPercentFree = c.GetInt("SRC_REPOS_DESIRED_PERCENT_FREE", "10", "Target percentage of free space on disk.")
	if c.JanitorReposDesiredPercentFree < 0 {
		c.AddError(errors.Errorf("negative value given for SRC_REPOS_DESIRED_PERCENT_FREE: %d", c.JanitorReposDesiredPercentFree))
	}
	if c.JanitorReposDesiredPercentFree > 100 {
		c.AddError(errors.Errorf("excessively high value given for SRC_REPOS_DESIRED_PERCENT_FREE: %d", c.JanitorReposDesiredPercentFree))
	}

	c.JanitorInterval = c.GetInterval("SRC_REPOS_JANITOR_INTERVAL", "1m", "Interval between cleanup runs")
}
