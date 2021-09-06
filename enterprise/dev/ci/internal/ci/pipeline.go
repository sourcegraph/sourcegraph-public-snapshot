// Package ci is responsible for generating a Buildkite pipeline configuration. It is invoked by the
// gen-pipeline.go command.
package ci

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/images"
	bk "github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/buildkite"
)

// GeneratePipeline is the main pipeline generation function. It defines the build pipeline for each of the
// main CI cases, which are defined in the main switch statement in the function.
func GeneratePipeline(c Config) (*bk.Pipeline, error) {
	if err := c.ensureCommit(); err != nil {
		return nil, err
	}

	// Common build env
	env := map[string]string{
		// Build meta
		"BUILDKITE_PULL_REQUEST":             os.Getenv("BUILDKITE_PULL_REQUEST"),
		"BUILDKITE_PULL_REQUEST_BASE_BRANCH": os.Getenv("BUILDKITE_PULL_REQUEST_BASE_BRANCH"),
		"BUILDKITE_PULL_REQUEST_REPO":        os.Getenv("BUILDKITE_PULL_REQUEST_REPO"),
		"COMMIT_SHA":                         c.commit,
		"DATE":                               c.now.Format(time.RFC3339),
		"VERSION":                            c.version,

		// Additional flags
		"GO111MODULE":                      "on",
		"PUPPETEER_SKIP_CHROMIUM_DOWNLOAD": "true",
		"FORCE_COLOR":                      "3",
		"ENTERPRISE":                       "1",
		// Add debug flags for scripts to consume
		"CI_DEBUG_PROFILE": strconv.FormatBool(c.profilingEnabled),
		// Bump Node.js memory to prevent OOM crashes
		"NODE_OPTIONS": "--max_old_space_size=4096",
	}

	// Build options for pipeline operations that spawn more build steps
	buildOptions := bk.BuildOptions{
		Message: os.Getenv("BUILDKITE_MESSAGE"),
		Commit:  c.commit,
		Branch:  c.branch,
		Env:     env,
	}

	// On release branches Percy must compare to the previous commit of the release branch, not main.
	if c.runType.is(ReleaseBranch) {
		env["PERCY_TARGET_BRANCH"] = c.branch
	}

	// Make all command steps timeout after 60 minutes in case a buildkite agent
	// got stuck / died.
	bk.AfterEveryStepOpts = append(bk.AfterEveryStepOpts, func(s *bk.Step) {
		// bk.Step is a union containing fields across all the different step types.
		// However, "timeout_in_minutes" only applies to the "command" step type.
		//
		// Testing the length of the "Command" field seems to be the most reliable way
		// of differentiating "command" steps from other step types without refactoring
		// everything.
		if len(s.Command) > 0 {
			if s.TimeoutInMinutes == "" {

				// Set the default value iff someone else hasn't set a custom one.
				s.TimeoutInMinutes = "60"
			}
		}
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
	//
	// PERF: Try to order steps such that slower steps are first.
	var operations []func(*bk.Pipeline)
	switch c.runType {
	case BextReleaseBranch:
		// If this is a browser extension release branch, run the browser-extension tests and
		// builds.
		operations = []func(*bk.Pipeline){
			addLint,
			addBrowserExt,
			addSharedFrontendTests(c),
			wait,
			addBrowserExtensionReleaseSteps,
		}

	case BextNightly:
		// If this is a browser extension nightly build, run the browser-extension tests and
		// e2e tests.
		operations = []func(*bk.Pipeline){
			addLint,
			addBrowserExt,
			addSharedFrontendTests(c),
			wait,
			addBrowserExtensionE2ESteps,
		}

	case ImagePatch:
		// only build candidate image for the specified image in the branch name
		// see https://about.sourcegraph.com/handbook/engineering/deployments/testing#building-docker-images-for-a-specific-branch
		patchImage := c.branch[20:]
		if !contains(images.SourcegraphDockerImages, patchImage) {
			panic(fmt.Sprintf("no image %q found", patchImage))
		}
		operations = append([]func(*bk.Pipeline){
			buildCandidateDockerImage(patchImage, c.version, c.candidateImageTag())},
			coreTestOperations(c, buildOptions)...)
		operations = append(operations,
			publishFinalDockerImage(c, patchImage, false))

	case ImagePatchNoTest:
		// If this is a no-test branch, then run only the Docker build. No tests are run.
		app := c.branch[27:]
		operations = []func(*bk.Pipeline){
			buildCandidateDockerImage(app, c.version, c.candidateImageTag()),
			wait,
			publishFinalDockerImage(c, app, false),
		}

	case CandidatesNoTest:
		operations = []func(*bk.Pipeline){}
		for _, dockerImage := range images.SourcegraphDockerImages {
			operations = append(operations,
				buildCandidateDockerImage(dockerImage, c.version, c.candidateImageTag()))
		}

	case PullRequest:
		// Run checks for pull requests
		switch {
		case c.changedFiles.isDocsOnly():
			// If this is a docs-only PR, run only the steps necessary to verify the docs.
			operations = []func(*bk.Pipeline){
				addDocs,
			}

		case c.changedFiles.isGoOnly() && !c.changedFiles.isSgOnly():
			// If this is a go-only PR, run only the steps necessary to verify the go code.
			operations = []func(*bk.Pipeline){
				addGoTests,            // ~1.5m
				addCheck,              // ~1m
				addGoBuild,            // ~0.5m
				addPostgresBackcompat, // ~0.25m
			}

		case c.changedFiles.isSgOnly():
			// If the changes are only in ./dev/sg then we only need to run a subset of steps.
			operations = []func(*bk.Pipeline){
				addGoTests,
				addCheck,
			}

		default:
			operations = coreTestOperations(c, buildOptions)
		}

	default:
		// Slow image builds
		for _, dockerImage := range images.SourcegraphDockerImages {
			operations = append(operations,
				buildCandidateDockerImage(dockerImage, c.version, c.candidateImageTag()))
		}
		if c.runType.is(MainDryRun, MainBranch) {
			buildExecutor(c.now, c.version)
		}

		// Core tests
		operations = append(operations, coreTestOperations(c, buildOptions)...)

		// Trigger e2e late so that it can leverage candidate images
		var async bool
		if c.runType.is(MainBranch) {
			async = true
		} else {
			async = false
		}
		operations = append(operations, triggerE2EandQA(e2eAndQAOptions{
			candidateImage: c.candidateImageTag(),
			buildOptions:   buildOptions,
			async:          async,
		}))

		// Add final artifacts
		for _, dockerImage := range images.SourcegraphDockerImages {
			insiders := c.runType.is(MainBranch)
			operations = append(operations,
				publishFinalDockerImage(c, dockerImage, insiders))
		}
		if c.runType.is(MainBranch) {
			operations = append(operations,
				publishExecutor(c.now, c.version), // ~6m (building executor base VM)
			)
		}

		// Propogate changes elsewhere
		if !c.runType.is(MainDryRun) {
			operations = append(operations,
				// wait for all steps to pass
				wait,
				triggerUpdaterPipeline)
		}
	}

	// Construct pipeline
	pipeline := &bk.Pipeline{
		Env: env,
	}
	for _, p := range operations {
		p(pipeline)
	}
	return pipeline, nil
}
