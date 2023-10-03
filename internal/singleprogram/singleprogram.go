// Package singleprogram contains runtime utilities for the single-binary
// distribution of Sourcegraph.
package singleprogram

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf/confdefaults"
	"github.com/sourcegraph/sourcegraph/internal/conf/deploy"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const appDirectory = "sourcegraph"

type CleanupFunc func() error

func Init(logger log.Logger) CleanupFunc {
	if deploy.IsApp() {
		fmt.Fprintln(os.Stderr, "✱ Cody App version:", version.Version(), runtime.GOOS, runtime.GOARCH)
	} else if deploy.IsDeployTypeSingleProgram(deploy.Type()) {
		fmt.Fprintln(os.Stderr, "✱ Sourcegraph (single-program) version:", version.Version(), runtime.GOOS, runtime.GOARCH)
	}

	// TODO(sqs) TODO(single-binary): see the env.HackClearEnvironCache docstring, we should be able to remove this
	// eventually.
	env.HackClearEnvironCache()

	// INDEXED_SEARCH_SERVERS is empty (but defined) so that indexed search is disabled.
	setDefaultEnv(logger, "INDEXED_SEARCH_SERVERS", "")

	if runtime.GOOS == "windows" {
		// POSTGRES database, specifying a non-default port to avoid conflicting with developer's
		// local servers, if they happen to have PostgreSQL running on their machines.
		setDefaultEnv(logger, "PGPORT", "5434")
	}

	// GITSERVER_EXTERNAL_ADDR is used by gitserver to identify itself in the
	// list in SRC_GIT_SERVERS.
	setDefaultEnv(logger, "GITSERVER_ADDR", "127.0.0.1:3178")
	setDefaultEnv(logger, "GITSERVER_EXTERNAL_ADDR", "127.0.0.1:3178")
	setDefaultEnv(logger, "SRC_GIT_SERVERS", "127.0.0.1:3178")

	setDefaultEnv(logger, "SYMBOLS_URL", "http://127.0.0.1:3184")
	setDefaultEnv(logger, "SEARCHER_URL", "http://127.0.0.1:3181")
	setDefaultEnv(logger, "BLOBSTORE_URL", deploy.BlobstoreDefaultEndpoint())
	setDefaultEnv(logger, "EMBEDDINGS_URL", "http://127.0.0.1:9991")

	// The syntax-highlighter might not be running, but this is a better default than an internal
	// hostname.
	setDefaultEnv(logger, "SRC_SYNTECT_SERVER", "http://localhost:9238")

	// Code Insights does not run in App
	setDefaultEnv(logger, "DISABLE_CODE_INSIGHTS", "true")

	// Jaeger might not be running, but this is a better default than an internal hostname.
	//
	// TODO(sqs) TODO(single-binary): this isnt taking effect
	//
	// setDefaultEnv(logger, "JAEGER_SERVER_URL", "http://localhost:16686")

	// Use blobstore on localhost.
	setDefaultEnv(logger, "PRECISE_CODE_INTEL_UPLOAD_AWS_ENDPOINT", deploy.BlobstoreDefaultEndpoint())
	setDefaultEnv(logger, "PRECISE_CODE_INTEL_UPLOAD_BACKEND", "blobstore")
	setDefaultEnv(logger, "EMBEDDINGS_UPLOAD_AWS_ENDPOINT", deploy.BlobstoreDefaultEndpoint())

	// Need to override this because without a host (eg ":3080") it listens only on localhost, which
	// is not accessible from the containers
	setDefaultEnv(logger, "SRC_HTTP_ADDR", "0.0.0.0:3080")

	// This defaults to an internal hostname.
	setDefaultEnv(logger, "SRC_FRONTEND_INTERNAL", "localhost:3090")

	cacheDir, err := setupAppDir(os.Getenv("SRC_APP_CACHE"), os.UserCacheDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to setup cache directory. Please see log for more details")
		logger.Fatal("failed to setup cache directory", log.Error(err))
	}

	setDefaultEnv(logger, "SRC_REPOS_DIR", filepath.Join(cacheDir, "repos"))
	setDefaultEnv(logger, "BLOBSTORE_DATA_DIR", filepath.Join(cacheDir, "blobstore"))
	setDefaultEnv(logger, "SYMBOLS_CACHE_DIR", filepath.Join(cacheDir, "symbols"))
	setDefaultEnv(logger, "SEARCHER_CACHE_DIR", filepath.Join(cacheDir, "searcher"))

	configDir, err := SetupAppConfigDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to setup user config directory. Please see log for more details")
		logger.Fatal("failed to setup config directory", log.Error(err))
		os.Exit(1)
	}

	if err := removeLegacyDirs(); err != nil {
		logger.Warn("failed to remove legacy dirs", log.Error(err))
	}

	embeddedPostgreSQLRootDir := filepath.Join(configDir, "postgresql")
	postgresCleanup, err := initPostgreSQL(logger, embeddedPostgreSQLRootDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "unable to set up PostgreSQL:", err)
		os.Exit(1)
	}

	writeFileIfNotExists := func(path string, data []byte) {
		var err error
		if _, err = os.Stat(path); os.IsNotExist(err) {
			err = os.WriteFile(path, data, 0600)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to write file %s: %s\n", path, err)
			os.Exit(1)
		}
	}

	siteConfigPath := filepath.Join(configDir, "site-config.json")
	setDefaultEnv(logger, "SITE_CONFIG_FILE", siteConfigPath)
	writeFileIfNotExists(siteConfigPath, []byte(confdefaults.App.Site))

	globalSettingsPath := filepath.Join(configDir, "global-settings.json")
	setDefaultEnv(logger, "GLOBAL_SETTINGS_FILE", globalSettingsPath)
	setDefaultEnv(logger, "GLOBAL_SETTINGS_ALLOW_EDITS", "true")
	writeFileIfNotExists(globalSettingsPath, []byte("{}\n"))

	// Set configuration file path for local repositories
	setDefaultEnv(logger, "SRC_LOCAL_REPOS_CONFIG_FILE", filepath.Join(configDir, "repos.json"))

	// We disable the use of executors passwords, because executors only listen on `localhost` this
	// is safe to do.
	setDefaultEnv(logger, "EXECUTOR_FRONTEND_URL", "http://localhost:3080")
	setDefaultEnv(logger, "EXECUTOR_FRONTEND_PASSWORD", confdefaults.AppInMemoryExecutorPassword)
	// Required because we set "executors.frontendURL": "http://host.docker.internal:3080" in site
	// configuration.
	setDefaultEnv(logger, "EXECUTOR_DOCKER_ADD_HOST_GATEWAY", "true")

	// TODO(single-binary): HACK: This is a hack to workaround the fact that the 2nd time you run `sourcegraph`
	// OOB migration validation fails:
	//
	// {"SeverityText":"FATAL","Timestamp":1675128552556359000,"InstrumentationScope":"sourcegraph","Caller":"svcmain/svcmain.go:143","Function":"github.com/sourcegraph/sourcegraph/internal/service/svcmain.run.func1","Body":"failed to start service","Resource":{"service.name":"sourcegraph","service.version":"0.0.196384-snapshot+20230131-6902ad","service.instance.id":"Stephens-MacBook-Pro.local"},"Attributes":{"service":"frontend","error":"failed to validate out of band migrations: Unfinished migrations. Please revert Sourcegraph to the previous version and wait for the following migrations to complete.\n  - migration 1 expected to be at 0.00% (at 100.00%)\n  - migration 13 expected to be at 0.00% (at 100.00%)\n  - migration 14 expected to be at 0.00% (at 100.00%)\n  - migration 15 expected to be at 0.00% (at 100.00%)\n  - migration 16 expected to be at 0.00% (at 100.00%)\n  - migration 17 expected to be at 0.00% (at 100.00%)\n  - migration 18 expected to be at 0.00% (at 100.00%)\n  - migration 19 expected to be at 0.00% (at 100.00%)\n  - migration 2 expected to be at 0.00% (at 100.00%)\n  - migration 20 expected to be at 0.00% (at 100.00%)\n  - migration 4 expected to be at 0.00% (at 100.00%)\n  - migration 5 expected to be at 0.00% (at 100.00%)\n  - migration 7 expected to be at 0.00% (at 100.00%)"}}
	//
	setDefaultEnv(logger, "SRC_DISABLE_OOBMIGRATION_VALIDATION", "1")

	setDefaultEnv(logger, "EXECUTOR_USE_FIRECRACKER", "false")
	// TODO(sqs): TODO(single-binary): Make it so we can run multiple executors in app mode. Right now, you
	// need to change this to "batches" to use batch changes executors.
	setDefaultEnv(logger, "EXECUTOR_QUEUE_NAME", "codeintel")

	writeFile := func(path string, data []byte, perm fs.FileMode) {
		if err := os.WriteFile(path, data, perm); err != nil {
			fmt.Fprintf(os.Stderr, "unable to write file %s: %s\n", path, err)
			os.Exit(1)
		}
	}

	if !deploy.IsApp() {
		setDefaultEnv(logger, "CTAGS_PROCESSES", "2")

		haveDocker := isDockerAvailable()
		if !haveDocker {
			printStatusCheckError(
				"Docker is unavailable",
				"Sourcegraph is better when Docker is available; some features may not work:",
				"- Batch changes",
				"- Symbol search",
				"- Symbols overview tab (on repository pages)",
			)
		}

		if _, err := exec.LookPath("src"); err != nil {
			printStatusCheckError(
				"src-cli is unavailable",
				"Sourcegraph is better when src-cli is available; batch changes may not work.",
				"Installation: https://github.com/sourcegraph/src-cli",
			)
		}

		// generate a shell script to run a ctags Docker image
		// unless the environment is already set up to find ctags
		ctagsPath := os.Getenv("CTAGS_COMMAND")
		if stat, err := os.Stat(ctagsPath); err != nil || stat.IsDir() {
			// Write script that invokes universal-ctags via Docker, if Docker is available.
			// TODO(single-binary): stop relying on a ctags Docker image
			if haveDocker {
				ctagsPath = filepath.Join(cacheDir, "universal-ctags-dev")
				writeFile(ctagsPath, []byte(universalCtagsDevScript), 0700)
				setDefaultEnv(logger, "CTAGS_COMMAND", ctagsPath)
			}
		}
	}
	return func() error {
		return postgresCleanup()
	}
}

func printStatusCheckError(title, description string, details ...string) {
	pad := func(s string, n int) string {
		spaces := n - len(s)
		if spaces < 0 {
			spaces = 0
		}
		return s + strings.Repeat(" ", spaces)
	}

	newLine := "\033[0m\n"
	titleRed := color.New(color.FgRed, color.BgYellow, color.Bold)
	titleRed.Fprintf(os.Stderr, "|------------------------------------------------------------------------------|"+newLine)
	titleRed.Fprintf(os.Stderr, "| %s |"+newLine, pad(title, 76))
	titleRed.Fprintf(os.Stderr, "|------------------------------------------------------------------------------|"+newLine)

	subline := func(s string) string {
		return color.RedString("%s %s %s"+newLine, titleRed.Sprint("|"), pad(s, 76), titleRed.Sprint("|"))
	}
	msg := subline(description)
	msg += subline("")
	for _, detail := range details {
		msg += subline(detail)
	}
	msg += subline("")
	fmt.Fprintf(os.Stderr, "%s", msg)
	titleRed.Fprintf(os.Stderr, "|------------------------------------------------------------------------------|"+newLine)
}

func isDockerAvailable() bool {
	if _, err := exec.LookPath("docker"); err != nil {
		return false
	}

	cmd := exec.Command("docker", "stats", "--no-stream")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// universalCtagsDevScript is copied from cmd/symbols/universal-ctags-dev.
const universalCtagsDevScript = `#!/usr/bin/env bash

# This script is a wrapper around universal-ctags.

exec docker run --rm -i \
    -a stdin -a stdout -a stderr \
    --user guest \
    --platform=linux/amd64 \
    --name=universal-ctags-$$ \
    --entrypoint /usr/local/bin/universal-ctags \
    slimsag/ctags:latest@sha256:dd21503a3ae51524ab96edd5c0d0b8326d4baaf99b4238dfe8ec0232050af3c7 "$@"
`

func SetupAppConfigDir() (string, error) {
	return setupAppDir(os.Getenv("SRC_APP_CONFIG"), os.UserConfigDir)
}

func setupAppDir(root string, defaultDirFn func() (string, error)) (string, error) {
	var base = root
	var dir = ""
	var err error
	if base == "" {
		dir = appDirectory
		if version.IsDev(version.Version()) {
			dir = fmt.Sprintf("%s-dev", dir)
		}
		base, err = defaultDirFn()
	}
	if err != nil {
		return "", err
	}

	path := filepath.Join(base, dir)
	return path, os.MkdirAll(path, 0700)
}

// Effectively runs:
//
// rm -rf $HOME/.cache/sourcegraph-sp
// rm -rf $HOME/.config/sourcegraph-sp
// rm -rf $HOME/Library/Application\ Support/sourcegraph-sp
// rm -rf $HOME/Library/Caches/sourcegraph-sp
//
// This deletes data from old Cody app directories, which came from before we switched to
// Tauri - so that users don't have to. In theory, these directories have no impact and can't conflict,
// but just for our own sanity we get rid of them.
func removeLegacyDirs() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return errors.Wrap(err, "UserConfigDir")
	}
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return errors.Wrap(err, "UserCacheDir")
	}
	if err := os.RemoveAll(filepath.Join(cacheDir, "sourcegraph-sp")); err != nil {
		return errors.Wrap(err, "RemoveAll cacheDir")
	}
	if err := os.RemoveAll(filepath.Join(configDir, "sourcegraph-sp")); err != nil {
		return errors.Wrap(err, "RemoveAll configDir")
	}
	return nil
}

// setDefaultEnv will set the environment variable if it is not set.
func setDefaultEnv(logger log.Logger, k, v string) {
	if _, ok := os.LookupEnv(k); ok {
		return
	}
	err := os.Setenv(k, v)
	if err != nil {
		logger.Fatal("setting default env variable", log.Error(err))
	}
}
