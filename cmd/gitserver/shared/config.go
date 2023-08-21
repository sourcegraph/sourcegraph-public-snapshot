package shared

import (
	"net"
	"path/filepath"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/hostname"
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

	// ExternalAddress is the name of this gitserver as it would appear in
	// SRC_GIT_SERVERS.
	//
	// Note: we can't just rely on the listen address since more than likely
	// gitserver is behind a k8s service.
	ExternalAddress string

	ListenAddress string

	SyncRepoStateInterval          time.Duration
	SyncRepoStateBatchSize         int
	SyncRepoStateUpdatePerSecond   int
	BatchLogGlobalConcurrencyLimit int

	RateLimitSyncerLimitPerSecond int

	JanitorReposDesiredPercentFree int
	JanitorInterval                time.Duration
	// EnableGCAuto is a temporary flag that allows us to control whether or not
	// `git gc --auto` is invoked during janitorial activities. This flag will
	// likely evolve into some form of site config value in the future.
	EnableGitGCAuto bool
	// sg maintenance and git gc must not be enabled at the same time. However, both
	// might be disabled at the same time, hence we need both SRC_ENABLE_GC_AUTO and
	// SRC_ENABLE_SG_MAINTENANCE.
	EnableSGMaintenance bool
	// The limit of 50 mirrors Git's gc_auto_pack_limit
	GitAutoPackLimit int
	// Our original Git gc job used 1 as limit, while git's default is 6700. We
	// don't want to be too aggressive to avoid unnecessary IO, hence we choose a
	// value somewhere in the middle. https://gitlab.com/gitlab-org/gitaly uses a
	// limit of 1024, which corresponds to an average of 4 loose objects per folder.
	// We can tune this parameter once we gain more experience.
	GitLooseObjectsLimit int
	// A failed sg maintenance run will place a log file in the git directory.
	// Subsequent sg maintenance runs are skipped unless the log file is old.
	//
	// Based on how https://github.com/git/git handles the gc.log file.
	SGMLogExpiry time.Duration
	// Each failed sg maintenance run increments a counter in the sgmLog file.
	// We reclone the repository if the number of retries exceeds sgmRetries.
	// Setting SRC_SGM_RETRIES to -1 disables recloning due to sgm failures.
	// Default value is 3 (reclone after 3 failed sgm runs).
	//
	// We mention this ENV variable in the header message of the sgmLog files. Make
	// sure that changes here are reflected in sgmLogHeader, too.
	SGMRetries int
	// The limit of repos cloned on the wrong shard to delete in one janitor run - value <=0 disables delete.
	JanitorWrongShardReposDeleteLimit int
	// Controls if gitserver cleanup tries to remove repos from disk which are not defined in the DB. Defaults to false.
	JanitorRemoveNonExistingRepos bool
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

	// First we check for it being explicitly set. This should only be
	// happening in environments were we run gitserver on localhost.
	// Otherwise we assume we can reach gitserver via its hostname / its
	// hostname is a prefix of the reachable address (see hostnameMatch).
	c.ExternalAddress = c.Get("GITSERVER_EXTERNAL_ADDR", hostname.Get(), "The name of this gitserver as it would appear in SRC_GIT_SERVERS.")

	c.ListenAddress = c.GetOptional("GITSERVER_ADDR", "The address under which the gitserver API listens. Can include a port.")
	// Fall back to a reasonable default.
	if c.ListenAddress == "" {
		port := "3178"
		host := ""
		if env.InsecureDev {
			host = "127.0.0.1"
		}
		c.ListenAddress = net.JoinHostPort(host, port)
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
	c.EnableGitGCAuto = c.GetBool("SRC_ENABLE_GC_AUTO", "true", "Use git-gc during janitorial cleanup phases")
	c.EnableSGMaintenance = c.GetBool("SRC_ENABLE_SG_MAINTENANCE", "false", "Use sg maintenance during janitorial cleanup phases")
	c.GitAutoPackLimit = c.GetInt("SRC_GIT_AUTO_PACK_LIMIT", "50", "the maximum number of pack files we tolerate before we trigger a repack")
	c.GitLooseObjectsLimit = c.GetInt("SRC_GIT_LOOSE_OBJECTS_LIMIT", "1024", "the maximum number of loose objects we tolerate before we trigger a repack")
	c.SGMLogExpiry = c.GetInterval("SRC_GIT_LOG_FILE_EXPIRY", "24h", "the number of hours after which sg maintenance runs even if a log file is present")
	c.SGMRetries = c.GetInt("SRC_SGM_RETRIES", "3", "the maximum number of times we retry sg maintenance before triggering a reclone.")
	c.JanitorWrongShardReposDeleteLimit = c.GetInt("SRC_WRONG_SHARD_DELETE_LIMIT", "10", "the maximum number of repos not assigned to this shard we delete in one run")
	c.JanitorRemoveNonExistingRepos = c.GetBool("SRC_REMOVE_NON_EXISTING_REPOS", "false", "controls if gitserver cleanup tries to remove repos from disk which are not defined in the DB")

}
