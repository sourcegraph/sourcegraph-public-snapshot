// Package singleprogram contains runtime utilities for the single-program (Go static binary)
// distribution of Sourcegraph.
package singleprogram

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/internal/conf/confdefaults"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

func init() {
	// Always overwrite this, even if already set, because it is intrinsic to this deployment
	// method.
	//
	// TODO(sqs): this is not actually taking effect in all cases because the value is consulted
	// before this runs, unfortunately, so we are using go build ldflags value injection as well.
	os.Setenv("DEPLOY_TYPE", "single-program")
}

func Init() {
	// HACK TODO(sqs) see the env.HackClearEnvironCache docstring
	env.HackClearEnvironCache()

	// INDEXED_SEARCH_SERVERS is empty (but defined) so that indexed search is disabled.
	setDefaultEnv("INDEXED_SEARCH_SERVERS", "")

	// Need to set this to avoid trying to look up gitservers via k8s service discovery.
	// TODO(sqs): Make this not require the hostname.
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to determine hostname:", err)
		os.Exit(1)
	}
	setDefaultEnv("SRC_GIT_SERVERS", hostname+":3178")

	setDefaultEnv("SYMBOLS_URL", "http://127.0.0.1:3184")
	setDefaultEnv("SEARCHER_URL", "http://127.0.0.1:3181")
	setDefaultEnv("REPO_UPDATER_URL", "http://127.0.0.1:3182")

	// The syntax-highlighter might not be running, but this is a better default than an internal
	// hostname.
	setDefaultEnv("SRC_SYNTECT_SERVER", "http://localhost:9238")

	// Jaeger might not be running, but this is a better default than an internal hostname.
	//
	// TODO(sqs): this isnt taking effect
	//
	// setDefaultEnv("JAEGER_SERVER_URL", "http://localhost:16686")

	// The s3proxy blobstore does need to be running. TODO(sqs): bundle this somehow?
	setDefaultEnv("PRECISE_CODE_INTEL_UPLOAD_AWS_ENDPOINT", "http://localhost:9000")
	setDefaultEnv("PRECISE_CODE_INTEL_UPLOAD_BACKEND", "blobstore")

	// Need to override this because without a host (eg ":3080") it listens only on localhost, which
	// is not accessible from the containers
	setDefaultEnv("SRC_HTTP_ADDR", "0.0.0.0:3080")

	// This defaults to an internal hostname.
	setDefaultEnv("SRC_FRONTEND_INTERNAL", "localhost:3090")

	cacheDir, err := os.UserCacheDir()
	if err == nil {
		cacheDir = filepath.Join(cacheDir, "sourcegraph-sp")
		err = os.MkdirAll(cacheDir, 0700)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to make user cache directory:", err)
		os.Exit(1)
	}

	setDefaultEnv("SRC_REPOS_DIR", filepath.Join(cacheDir, "repos"))
	setDefaultEnv("CACHE_DIR", filepath.Join(cacheDir, "cache"))

	configDir, err := os.UserConfigDir()
	if err == nil {
		configDir = filepath.Join(configDir, "sourcegraph-sp")
		err = os.MkdirAll(configDir, 0700)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to make user config directory:", err)
		os.Exit(1)
	}

	embeddedPostgreSQLRootDir := filepath.Join(configDir, "postgresql")
	if err := initPostgreSQL(embeddedPostgreSQLRootDir); err != nil {
		fmt.Fprintln(os.Stderr, "unable to set up PostgreSQL:", err)
		os.Exit(1)
	}

	writeFileIfNotExists := func(path string, data []byte) {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			err = ioutil.WriteFile(path, data, 0600)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to write file %s: %s\n", path, err)
			os.Exit(1)
		}
	}

	siteConfigPath := filepath.Join(configDir, "site-config.json")
	setDefaultEnv("SITE_CONFIG_FILE", siteConfigPath)
	setDefaultEnv("SITE_CONFIG_ALLOW_EDITS", "true")
	writeFileIfNotExists(siteConfigPath, []byte(confdefaults.SingleProgram.Site))

	globalSettingsPath := filepath.Join(configDir, "global-settings.json")
	setDefaultEnv("GLOBAL_SETTINGS_FILE", globalSettingsPath)
	setDefaultEnv("GLOBAL_SETTINGS_ALLOW_EDITS", "true")
	writeFileIfNotExists(globalSettingsPath, []byte("{}\n"))

	// Escape hatch isn't needed in local dev since the site config can always just be a file on disk.
	setDefaultEnv("NO_SITE_CONFIG_ESCAPE_HATCH", "1")

	// TODO(sqs): Executor shouldnt need a password when running in single-program.
	setDefaultEnv("EXECUTOR_FRONTEND_URL", "http://localhost:3080")
	setDefaultEnv("EXECUTOR_FRONTEND_PASSWORD", "asdf1234asdf1234asdf1234")
	setDefaultEnv("EXECUTOR_USE_FIRECRACKER", "false")
	// TODO(sqs): Make it so we can run multiple executors in single-program mode. Right now, you
	// need to change this to "batches" to use batch changes executors.
	setDefaultEnv("EXECUTOR_QUEUE_NAME", "codeintel")

	writeFile := func(path string, data []byte, perm fs.FileMode) {
		if err := ioutil.WriteFile(path, data, perm); err != nil {
			fmt.Fprintf(os.Stderr, "unable to write file %s: %s\n", path, err)
			os.Exit(1)
		}
	}

	setDefaultEnv("CTAGS_PROCESSES", "2")
	// Write script that invokes universal-ctags via Docker.
	// TODO(sqs): this assumes that the `ctags` image is already built locally.
	ctagsPath := filepath.Join(cacheDir, "universal-ctags-dev")
	writeFile(ctagsPath, []byte(universalCtagsDevScript), 0700)
	setDefaultEnv("CTAGS_COMMAND", ctagsPath)
}

// universalCtagsDevScript is copied from cmd/symbols/universal-ctags-dev.
const universalCtagsDevScript = `#!/usr/bin/env bash

# This script is a wrapper around universal-ctags.

exec docker run --rm -i \
    -a stdin -a stdout -a stderr \
    --user guest \
    --name=universal-ctags-$$ \
    --entrypoint /usr/local/bin/universal-ctags \
    ctags "$@"
`

// setDefaultEnv will set the environment variable if it is not set.
func setDefaultEnv(k, v string) string {
	if s, ok := os.LookupEnv(k); ok {
		return s
	}
	err := os.Setenv(k, v)
	if err != nil {
		log.Fatal(err)
	}
	return v
}
