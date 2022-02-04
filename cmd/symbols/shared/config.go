package shared

import (
	"os"
	"runtime"
	"strconv"
	"time"

	sharedtypes "github.com/sourcegraph/sourcegraph/cmd/symbols/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	Ctags sharedtypes.CtagsConfig

	SanityCheck       bool
	CacheDir          string
	CacheSizeMB       int
	NumCtagsProcesses int
	RequestBufferSize int
	ProcessingTimeout time.Duration
	UseRockskip       bool
	// Only for Rockskip
	MaxRepos int

	// The maximum sum of lengths of all paths in a single call to git archive. Without this limit, we
	// could hit the error "argument list too long" by exceeding the limit on the number of arguments to
	// a command enforced by the OS.
	//
	// Mac  : getconf ARG_MAX returns 1,048,576
	// Linux: getconf ARG_MAX returns 2,097,152
	//
	// We want to remain well under that limit, so defaulting to 100,000 seems safe (see the
	// MAX_TOTAL_PATHS_LENGTH environment variable below).
	MaxTotalPathsLength int
}

var config = &Config{}

// Load reads from the environment and stores the transformed data on the config object for later retrieval.
func (c *Config) Load() {
	c.Ctags.Command = c.Get("CTAGS_COMMAND", "universal-ctags", "ctags command (should point to universal-ctags executable compiled with JSON and seccomp support)")
	c.Ctags.PatternLengthLimit = c.GetInt("CTAGS_PATTERN_LENGTH_LIMIT", "250", "the maximum length of the patterns output by ctags")
	logCtagsErrorsDefault := "false"
	if os.Getenv("DEPLOY_TYPE") == "dev" {
		logCtagsErrorsDefault = "true"
	}
	c.Ctags.LogErrors = c.GetBool("LOG_CTAGS_ERRORS", logCtagsErrorsDefault, "log ctags errors")
	c.Ctags.DebugLogs = false

	c.SanityCheck = c.GetBool("SANITY_CHECK", "false", "check that go-sqlite3 works then exit 0 if it's ok or 1 if not")
	c.CacheDir = c.Get("CACHE_DIR", "/tmp/symbols-cache", "directory in which to store cached symbols")
	c.CacheSizeMB = c.GetInt("SYMBOLS_CACHE_SIZE_MB", "100000", "maximum size of the disk cache (in megabytes)")
	c.NumCtagsProcesses = c.GetInt("CTAGS_PROCESSES", strconv.Itoa(runtime.GOMAXPROCS(0)), "number of concurrent parser processes to run")
	c.RequestBufferSize = c.GetInt("REQUEST_BUFFER_SIZE", "8192", "maximum size of buffered parser request channel")
	c.ProcessingTimeout = c.GetInterval("PROCESSING_TIMEOUT", "2h", "maximum time to spend processing a repository")

	// TODO  - move to enterprise/
	c.UseRockskip = c.GetBool("USE_ROCKSKIP", "false", "use Rockskip and Postgres instead of SQLite")
	c.MaxRepos = c.GetInt("MAX_REPOS", "1000", "maximum number of repositories for Rockskip to store in Postgres, with LRU eviction")
	c.MaxTotalPathsLength = c.GetInt("MAX_TOTAL_PATHS_LENGTH", "100000", "maximum sum of lengths of all paths in a single call to git archive")
}
