// Package ci is responsible for generating a Buildkite pipeline configuration. It is invoked by the
// gen-pipeline.go command.
package ci

import (
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
	bk.OnEveryStepOpts = append(bk.OnEveryStepOpts,
		bk.Env("GO111MODULE", "on"),
		bk.Env("PUPPETEER_SKIP_CHROMIUM_DOWNLOAD", "true"),
		bk.Env("FORCE_COLOR", "1"),
		bk.Env("ENTERPRISE", "1"),
		bk.Env("COMMIT_SHA", c.commit),
		bk.Env("DATE", c.now.Format(time.RFC3339)),
	)

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
		pipelineOperations = append(pipelineOperations, addDockerImage(c, app, false))

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
			triggerE2E(c),
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
			wait,
			addCodeCov,
			addDockerImages(c),
		}
	}

	// Construct pipeline
	pipeline := &bk.Pipeline{}
	for _, p := range pipelineOperations {
		p(pipeline)
	}
	return pipeline, nil
}
