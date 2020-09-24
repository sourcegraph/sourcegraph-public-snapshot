// Package ci is responsible for generating a Buildkite pipeline configuration. It is invoked by the
// gen-pipeline.go command.
package ci

import (
	"fmt"
	"os"
	"strconv"
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

		// Add debug flags for scripts to consume
		"CI_DEBUG_PROFILE": strconv.FormatBool(c.profilingEnabled),
	}

	// On release branches Percy must compare to the previous commit of the release branch, not main.
	if c.releaseBranch {
		env["PERCY_TARGET_BRANCH"] = c.branch
	}

	for k, v := range env {
		bk.BeforeEveryStepOpts = append(bk.BeforeEveryStepOpts, bk.Env(k, v))
	}

	bk.AfterEveryStepOpts = append(bk.AfterEveryStepOpts, func(s *bk.Step) {
		s.Agents["queue"] = "test"
	})

	if c.profilingEnabled {
		bk.AfterEveryStepOpts = append(bk.AfterEveryStepOpts, func(s *bk.Step) {
			// wrap "time -v" around each command for CPU/RAM utilization information

			var prefixed []string
			for _, cmd := range s.Command {
				prefixed = append(prefixed, fmt.Sprintf("env time -v %s", cmd))
			}

			s.Command = prefixed
		})
	}

	// Generate pipeline steps. This statement outlines the pipeline steps for each CI case.
	var pipelineOperations []func(*bk.Pipeline)
	pipelineOperations = []func(*bk.Pipeline){
		triggerE2E(c, env),
		addBackendIntegrationTests(c), // ~11m
		addDockerImages(c, false),     // ~8m
		addLint,                       // ~4.5m
		addSharedTests(c),             // ~4.5m
		addWebApp,                     // ~3m
		addBrowserExt,                 // ~2m
		addGoTests,                    // ~1.5m
		addCheck,                      // ~1m
		addGoBuild,                    // ~0.5m
		addPostgresBackcompat,         // ~0.25m
		addDockerfileLint,             // ~0.2m
	}

	// Construct pipeline
	pipeline := &bk.Pipeline{}
	for _, p := range pipelineOperations {
		p(pipeline)
	}
	return pipeline, nil
}
