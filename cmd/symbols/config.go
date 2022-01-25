package main

import (
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	ctags types.CtagsConfig

	sanityCheck       bool
	cacheDir          string
	cacheSizeMB       int
	numCtagsProcesses int
	requestBufferSize int
	processingTimeout time.Duration
	useRockskip       bool
	// Only for Rockskip
	maxRepos int

	// The maximum sum of lengths of all paths in a single call to git archive. Without this limit, we
	// could hit the error "argument list too long" by exceeding the limit on the number of arguments to
	// a command enforced by the OS.
	//
	// Mac  : getconf ARG_MAX returns 1,048,576
	// Linux: getconf ARG_MAX returns 2,097,152
	//
	// We want to remain well under that limit, so defaulting to 100,000 seems safe (see the
	// MAX_TOTAL_PATHS_LENGTH environment variable below).
	maxTotalPathsLength int
}

var config = &Config{}

// Load reads from the environment and stores the transformed data on the config object for later retrieval.
func (c *Config) Load() {
	c.ctags.Command = c.Get("CTAGS_COMMAND", "universal-ctags", "ctags command (should point to universal-ctags executable compiled with JSON and seccomp support)")
	c.ctags.PatternLengthLimit = c.GetInt("CTAGS_PATTERN_LENGTH_LIMIT", "250", "the maximum length of the patterns output by ctags")
	logCtagsErrorsDefault := "false"
	if os.Getenv("DEPLOY_TYPE") == "dev" {
		logCtagsErrorsDefault = "true"
	}
	c.ctags.LogErrors = c.GetBool("LOG_CTAGS_ERRORS", logCtagsErrorsDefault, "log ctags errors")
	c.ctags.DebugLogs = false

	c.sanityCheck = c.GetBool("SANITY_CHECK", "false", "check that go-sqlite3 works then exit 0 if it's ok or 1 if not")
	c.cacheDir = c.Get("CACHE_DIR", "/tmp/symbols-cache", "directory in which to store cached symbols")
	c.cacheSizeMB = c.GetInt("SYMBOLS_CACHE_SIZE_MB", "100000", "maximum size of the disk cache (in megabytes)")
	c.numCtagsProcesses = c.GetInt("CTAGS_PROCESSES", strconv.Itoa(runtime.GOMAXPROCS(0)), "number of concurrent parser processes to run")
	c.requestBufferSize = c.GetInt("REQUEST_BUFFER_SIZE", "8192", "maximum size of buffered parser request channel")
	c.processingTimeout = c.GetInterval("PROCESSING_TIMEOUT", "2h", "maximum time to spend processing a repository")
	c.useRockskip = c.GetBool("USE_ROCKSKIP", "false", "use Rockskip and Postgres instead of SQLite")
	c.maxRepos = c.GetInt("MAX_REPOS", "1000", "maximum number of repositories for Rockskip to store in Postgres, with LRU eviction")
	c.maxTotalPathsLength = c.GetInt("MAX_TOTAL_PATHS_LENGTH", "100000", "maximum sum of lengths of all paths in a single call to git archive")
}
