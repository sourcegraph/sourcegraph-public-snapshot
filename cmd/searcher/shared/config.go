package shared

import (
	"net"
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

	ListenAddress                   string
	CacheDir                        string
	CacheSizeMB                     int
	BackgroundTimeout               time.Duration
	MaxTotalGitArchivePathsLength   int
	DisableHybridSearch             bool
	ExhaustiveRequestLoggingEnabled bool
}

func (c *Config) Load() {
	c.ListenAddress = c.GetOptional("SEARCHER_ADDR", "The address under which the searcher API listens. Can include a port.")
	// Fall back to a reasonable default.
	if c.ListenAddress == "" {
		port := "3181"
		host := ""
		if env.InsecureDev {
			host = "127.0.0.1"
		}
		c.ListenAddress = net.JoinHostPort(host, port)
	}

	c.CacheDir = c.Get(env.ChooseFallbackVariableName("SEARCHER_CACHE_DIR", "CACHE_DIR"), "/tmp", "Directory to store cached archives in.")
	c.CacheSizeMB = c.GetInt("SEARCHER_CACHE_SIZE_MB", "100000", "Maximum size of the on disk cache in megabytes when cached archives get evicted.")
	if c.CacheSizeMB < 0 {
		c.AddError(errors.New("SEARCHER_CACHE_SIZE_MB must be >= 0"))
	}
	// Same environment variable name (and default value) used by symbols.
	c.BackgroundTimeout = c.GetInterval("PROCESSING_TIMEOUT", "2h", "Maximum time to spend processing a repository.")
	c.MaxTotalGitArchivePathsLength = c.GetInt("MAX_TOTAL_PATHS_LENGTH", "100000", "Maximum number of paths passed in a single call to git archive.")
	if c.MaxTotalGitArchivePathsLength < 0 {
		c.AddError(errors.New("MAX_TOTAL_PATHS_LENGTH must be >= 0"))
	}
	c.DisableHybridSearch = c.GetBool("DISABLE_HYBRID_SEARCH", "false", "If true, unindexed search will not consult indexed search to speed up searches.")
	c.ExhaustiveRequestLoggingEnabled = c.GetBool("SRC_SEARCHER_EXHAUSTIVE_LOGGING_ENABLED", "false", "Enable exhaustive request logging in searcher")
}
