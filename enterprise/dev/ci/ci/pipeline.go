// Package ci is responsible for generating a Buildkite pipeline configuration. It is invoked by the
// gen-pipeline.go command.
package ci

import (
	"os"
	"time"

	bk "github.com/sourcegraph/sourcegraph/internal/buildkite"
)

// GeneratePipeline is the main pipeline generation function. It defines the build pipeline for each of the
// main CI cases, which are defined in the main switch statement in the function.
func GeneratePipeline(c Config) (*bk.Pipeline, error) {
	if err := c.ensureCommit(); err != nil {
		return nil, err
	}

	// Common build env
	env := map[string]string{
		"GO111MODULE":                      "on",
		"PUPPETEER_SKIP_CHROMIUM_DOWNLOAD": "true",
		"FORCE_COLOR":                      "3",
		"ENTERPRISE":                       "1",
		"COMMIT_SHA":                       c.commit,
		"DATE":                             c.now.Format(time.RFC3339),
		"VERSION":                          c.version,
		// For Bundlesize
		"CI_REPO_OWNER":     "sourcegraph",
		"CI_REPO_NAME":      "sourcegraph",
		"CI_COMMIT_SHA":     os.Getenv("BUILDKITE_COMMIT"),
		"CI_COMMIT_MESSAGE": os.Getenv("BUILDKITE_MESSAGE"),
	}

	for k, v := range env {
		bk.OnEveryStepOpts = append(bk.OnEveryStepOpts, bk.Env(k, v))
	}

	// Generate pipeline steps. This statement outlines the pipeline steps for each CI case.
	var pipelineOperations []func(*bk.Pipeline)
	switch {
	case c.isPR() && isDocsOnly():
		// If this is a docs-only PR, run only the steps necessary to verify the docs.
		pipelineOperations = []func(*bk.Pipeline){
			addDocs,
		}
	case c.patchNoTest:
		// If this is a no-test branch, then run only the Docker build. No tests are run.
		app := c.branch[27:]
		pipelineOperations = append(pipelineOperations,
			addCanidateDockerImage(c, app),
			addFinalDockerImage(c, app, false),
		)

	case c.isBextReleaseBranch:
		// If this is a browser extension release branch, run the browser-extension tests and
		// builds.
		pipelineOperations = []func(*bk.Pipeline){
			addLint,
			addBrowserExt,
			addSharedTests,
			wait,
			addCodeCov,
			addBrowserExtensionReleaseSteps,
		}

	case c.isQuick:
		// Run fast steps only
		pipelineOperations = []func(*bk.Pipeline){
			addCheck,
			addLint,
			addBrowserExt,
			addWebApp,
			addLSIFServer,
			addSharedTests,
			addGoTests,
			addGoBuild,
			addDockerfileLint,
			wait,
			addCodeCov,
		}

	default:
		// Otherwise, run the CI steps for the Sourcegraph web app. Specific
		// steps may be modified or skipped for certain branches; these
		// variations are defined in the functions parameterized by the
		// config.
		//
		// PERF: Try to order steps such that slower steps are first.
		pipelineOperations = []func(*bk.Pipeline){
			triggerE2E(c, env),
			addLint,    // ~5m
			addWebApp,  // ~3m
			addGoTests, // ~2m
			addGoBuild, // ~2m
			addCheck,   // ~2m
			addBrowserExt,
			addLSIFServer,
			addSharedTests,
			addPostgresBackcompat,
			addDockerfileLint,
			addDockerImages(c, false),
			wait,
			addCodeCov,
			addDockerImages(c, true),
		}
	}

	// Construct pipeline
	pipeline := &bk.Pipeline{}
	for _, p := range pipelineOperations {
		p(pipeline)
	}
	return pipeline, nil
}
