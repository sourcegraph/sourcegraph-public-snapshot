package main

import (
	"runtime"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/env"
)

type Config struct {
	env.BaseConfig

	sanityCheck    bool
	cacheDir       string
	cacheSizeMB    string
	ctagsProcesses string
}

var config = &Config{}

// Load reads from the environment and stores the transformed data on the config
// object for later retrieval.
func (c *Config) Load() {
	c.sanityCheck = c.GetBool("SANITY_CHECK", "false", "check that go-sqlite3 works then exit 0 if it's ok or 1 if not")
	c.cacheDir = c.Get("CACHE_DIR", "/tmp/symbols-cache", "directory in which to store cached symbols")
	c.cacheSizeMB = c.Get("SYMBOLS_CACHE_SIZE_MB", "100000", "maximum size of the disk cache (in megabytes)")
	c.ctagsProcesses = c.Get("CTAGS_PROCESSES", strconv.Itoa(runtime.GOMAXPROCS(0)), "number of concurrent parser processes to run")
}
