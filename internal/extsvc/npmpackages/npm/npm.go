// Code for interfacing with Javascript and Typescript package registries such
// as NPMJS.com.
package npm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/schema"
)

// Sourced from https://docs.npmjs.com/cli/v8/using-npm/config

const (
	npmConfigCacheEnvVar        = "NPM_CONFIG_CACHE"
	npmConfigFetchTimeoutEnvVar = "NPM_CONFIG_FETCH_TIMEOUT"
	npmConfigRegistryEnvVar     = "NPM_CONFIG_REGISTRY"
)

var (
	NPMBinary = "npm"

	// Register all npm config variables we use for testing.
	npmUsedConfigVars = []string{
		npmConfigCacheEnvVar,
		npmConfigFetchTimeoutEnvVar,
		npmConfigRegistryEnvVar,
	}
	npmCacheDir        string
	observationContext *observation.Context
	operations         *Operations
	invocTimeout, _    = time.ParseDuration(
		env.Get("SRC_NPM_TIMEOUT", "30s", "Time limit per NPM invocation, which is used to resolve NPM dependencies."))
	incorrectTarballNameRegex = regexp.MustCompile("^@" + reposource.NPMScopeRegexString + "/")
)

func init() {
	observationContext = &observation.Context{
		Logger:     log15.Root(),
		Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
		Registerer: prometheus.DefaultRegisterer,
	}
	operations = NewOperations(observationContext)

	// Should only be set for gitserver for persistence, repo-updater will use ephemeral storage.
	// repo-updater only performs existence checks which doesnt involve downloading any JARs (except for JDK),
	// only POM files which are much lighter.
	if reposDir := os.Getenv("SRC_REPOS_DIR"); reposDir != "" {
		npmCacheDir = filepath.Join(reposDir, "npm")
		if err := os.MkdirAll(npmCacheDir, os.ModePerm); err != nil {
			log.Fatalf("failed to create npm cache dir in %s: %s", npmCacheDir, err)
		}
	}
}

func FetchSources(ctx context.Context, config *schema.NPMPackagesConnection, dependency reposource.NPMDependency) (filename string, err error) {
	ctx, endObservation := operations.fetchSources.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("dependency", dependency.PackageManagerSyntax()),
	}})
	defer endObservation(1, observation.Args{})

	npmJsonOutput, err := runNPMCommand(ctx, config, "pack", dependency.PackageManagerSyntax(), "--json")
	if err != nil {
		return "", errors.Wrapf(err, "failed to fetch sources for %s", dependency.PackageManagerSyntax())
	}
	// [ { "id": "packageName@version", ..., "filename": "tarball.tgz", ... } ]
	var parsedJson interface{}
	err = json.Unmarshal([]byte(npmJsonOutput), &parsedJson)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse npm pack's output '%s'", npmJsonOutput)
	}
	output, err := parseNPMPackOutput(npmJsonOutput)
	return output, err
}

func Exists(ctx context.Context, config *schema.NPMPackagesConnection, dependency reposource.NPMDependency) (err error) {
	if operations == nil {
		fmt.Println("operations is nil!")
	}
	ctx, endObservation := operations.exists.With(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("dependency", dependency.PackageManagerSyntax()),
	}})
	defer endObservation(1, observation.Args{})

	out, err := runNPMCommand(
		ctx,
		config,
		"view",
		dependency.PackageManagerSyntax())
	if err != nil {
		return errors.Wrapf(err, "tried to check if npm package %s exists but failed", dependency.PackageManagerSyntax())
	}
	// Ideally, checking the exit code would be enough, but it's not, due to an
	// NPM bug https://github.com/npm/cli/issues/3184#issuecomment-963387099 😔
	if len(out) == 0 || (strings.TrimSpace(out) == "") {
		return errors.Newf("npm package %s does not exist", dependency.PackageManagerSyntax())
	}
	return nil
}

func runNPMCommand(ctx context.Context, config *schema.NPMPackagesConnection, args ...string) (output string, err error) {
	ctx, cancel := context.WithTimeout(ctx, invocTimeout)
	defer cancel()

	ctx, trace, endObservation := operations.runCommand.WithAndLogger(ctx, &err, observation.Args{LogFields: []otlog.Field{
		otlog.String("registry", config.Registry),
		otlog.String("args", strings.Join(args, ", ")),
	}})
	defer endObservation(1, observation.Args{})

	cmd := exec.CommandContext(ctx, NPMBinary, args...)
	// Make sure node is usable, but don't copy the full environment.
	// See also: forwardedHostEnvVars
	cmd.Env = []string{}
	cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", os.Getenv("HOME")))
	cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s", os.Getenv("PATH")))
	// TODO: [npm-package-support-credentials] Unlike Coursier where credentials are passed directly,
	// npm's API involves login/logout commands. My instinct is that doing a login+logout for every command
	// is a bad idea. Maybe we can hoist out the login and logout operations? Say something like:
	//     npmLoginToken, err := npm.Login(...)
	//     if err != nil { ... }
	//     defer func(){ errLogout := npmLoginToken.Logout(); if err == nil { err = errLogout } }()
	//     npmLoginToken.doStuff(...)

	registry := config.Registry
	if len(registry) != 0 {
		cmd.Env = append(
			cmd.Env, fmt.Sprintf("%s=%s", npmConfigRegistryEnvVar, registry))
	}

	if npmCacheDir != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", npmConfigCacheEnvVar, npmCacheDir))
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", errors.Wrapf(err, "npm command %q failed with stderr %q and stdout %q", cmd, stderr, &stdout)
	}
	trace.Log(otlog.String("stdout", stdout.String()), otlog.String("stderr", stderr.String()))

	return stdout.String(), nil
}

type npmPackOutput = []struct {
	Filename string `json:"filename"`
}

func parseNPMPackOutput(output string) (filename string, err error) {
	var parsedJson npmPackOutput
	var errInfo string
	if err = json.Unmarshal([]byte(output), &parsedJson); err != nil {
		errInfo = err.Error()
	} else if len(parsedJson) != 1 {
		errInfo = "expected output array to have 1 element"
	} else if parsedJson[0].Filename == "" {
		errInfo = "expected non-empty filename field in first object"
	} else {
		filename = parsedJson[0].Filename
		// [NOTE: npm-tarball-filename-workaround]
		// For scoped packages, npm gives the wrong output
		// (tested with 7.20.1 and 8.1.2). The actual file will
		// be saved at scope-package-version.tgz, but the
		// filename field is @scope/package-version.tgz
		//
		// See https://github.com/npm/cli/issues/3405
		if filename[0] == '@' {
			filename = incorrectTarballNameRegex.ReplaceAllString(filename, "$1-")
		}
		return filename, nil
	}
	return "", errors.Errorf("failed to parse npm pack's JSON output (%s): %s", errInfo, output)
}
