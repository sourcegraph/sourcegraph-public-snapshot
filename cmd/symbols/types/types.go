package types

import (
	"context"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type SqliteConfig struct {
	CacheDir                string
	CacheSizeMB             int
	NumCtagsProcesses       int
	RequestBufferSize       int
	ProcessingTimeout       time.Duration
	Ctags                   CtagsConfig
	RepositoryFetcher       RepositoryFetcherConfig
	MaxConcurrentlyIndexing int
}

func LoadSqliteConfig(baseConfig env.BaseConfig, ctags CtagsConfig, repositoryFetcher RepositoryFetcherConfig) SqliteConfig {
	// Variable was renamed to have SYMBOLS_ prefix to avoid a conflict with the same env var name
	// in searcher when running as a single binary. The old name is treated as an alias to prevent
	// customer environments from breaking if they still use it, because we have no way of migrating
	// environment variables today.
	cacheDirName := env.ChooseFallbackVariableName("SYMBOLS_CACHE_DIR", "CACHE_DIR")

	return SqliteConfig{
		Ctags:                   ctags,
		RepositoryFetcher:       repositoryFetcher,
		CacheDir:                baseConfig.Get(cacheDirName, "/tmp/symbols-cache", "directory in which to store cached symbols"),
		CacheSizeMB:             baseConfig.GetInt("SYMBOLS_CACHE_SIZE_MB", "100000", "maximum size of the disk cache (in megabytes)"),
		NumCtagsProcesses:       baseConfig.GetInt("CTAGS_PROCESSES", strconv.Itoa(runtime.GOMAXPROCS(0)), "number of concurrent parser processes to run"),
		RequestBufferSize:       baseConfig.GetInt("REQUEST_BUFFER_SIZE", "8192", "maximum size of buffered parser request channel"),
		ProcessingTimeout:       baseConfig.GetInterval("PROCESSING_TIMEOUT", "2h0m0s", "maximum time to spend processing a repository"),
		MaxConcurrentlyIndexing: baseConfig.GetInt("MAX_CONCURRENTLY_INDEXING", "10", "maximum number of repositories to index at a time"),
	}
}

type CtagsConfig struct {
	UniversalCommand   string
	ScipCommand        string
	PatternLengthLimit int
	LogErrors          bool
	DebugLogs          bool
	MaxFileSize        int
	MaxSymbols         int
}

func LoadCtagsConfig(baseConfig env.BaseConfig) CtagsConfig {
	logCtagsErrorsDefault := "false"
	if os.Getenv("DEPLOY_TYPE") == "dev" {
		logCtagsErrorsDefault = "true"
	}

	return CtagsConfig{
		UniversalCommand:   baseConfig.Get("CTAGS_COMMAND", "universal-ctags", "ctags command (should point to universal-ctags executable compiled with JSON and seccomp support)"),
		ScipCommand:        baseConfig.Get("SCIP_CTAGS_COMMAND", "scip-ctags", "scip-ctags command"),
		PatternLengthLimit: baseConfig.GetInt("CTAGS_PATTERN_LENGTH_LIMIT", "250", "the maximum length of the patterns output by ctags"),
		LogErrors:          baseConfig.GetBool("LOG_CTAGS_ERRORS", logCtagsErrorsDefault, "log ctags errors"),
		DebugLogs:          false,
		MaxFileSize:        baseConfig.GetInt("CTAGS_MAX_FILE_SIZE", "524288", "skip files larger than this size (in bytes)"),
		MaxSymbols:         baseConfig.GetInt("CTAGS_MAX_SYMBOLS", "2000", "skip files with more than this many symbols"),
	}
}

type RepositoryFetcherConfig struct {
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

	MaxFileSizeKb int
}

func LoadRepositoryFetcherConfig(baseConfig env.BaseConfig) RepositoryFetcherConfig {
	// Variable was renamed to have SYMBOLS_ prefix to avoid a conflict with the same env var name
	// in searcher when running as a single binary. The old name is treated as an alias to prevent
	// customer environments from breaking if they still use it, because we have no way of migrating
	// environment variables today.
	maxTotalPathsLengthName := env.ChooseFallbackVariableName("SYMBOLS_MAX_TOTAL_PATHS_LENGTH", "MAX_TOTAL_PATHS_LENGTH")

	return RepositoryFetcherConfig{
		MaxTotalPathsLength: baseConfig.GetInt(maxTotalPathsLengthName, "100000", "maximum sum of lengths of all paths in a single call to git archive"),
		MaxFileSizeKb:       baseConfig.GetInt("MAX_FILE_SIZE_KB", "1000", "maximum file size in KB, the contents of bigger files are ignored"),
	}
}

type SearchFunc func(ctx context.Context, args search.SymbolsParameters) (results result.Symbols, err error)
