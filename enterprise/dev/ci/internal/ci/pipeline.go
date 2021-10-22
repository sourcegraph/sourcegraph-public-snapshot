// Package ci is responsible for generating a Buildkite pipeline configuration. It is invoked by the
// gen-pipeline.go command.
package ci

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/images"
	bk "github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/buildkite"
	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/ci/operations"
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
		"COMMIT_SHA":                         c.Commit,
		"DATE":                               c.Time.Format(time.RFC3339),
		"VERSION":                            c.Version,

		// Additional flags
		"GO111MODULE":                      "on",
		"PUPPETEER_SKIP_CHROMIUM_DOWNLOAD": "true",
		"FORCE_COLOR":                      "3",
		"ENTERPRISE":                       "1",
		// Add debug flags for scripts to consume
		"CI_DEBUG_PROFILE": strconv.FormatBool(c.MessageFlags.ProfilingEnabled),
		// Bump Node.js memory to prevent OOM crashes
		"NODE_OPTIONS": "--max_old_space_size=8192",

		// Bundlesize configuration: https://github.com/siddharthkp/bundlesize2#build-status-and-checks-for-github
		"CI_REPO_OWNER": "sourcegraph",
		"CI_REPO_NAME":  "sourcegraph",
		"CI_COMMIT_SHA": os.Getenv("BUILDKITE_COMMIT"),
		// $ in commit messages must be escaped to not attempt interpolation which will fail.
		"CI_COMMIT_MESSAGE": strings.ReplaceAll(os.Getenv("BUILDKITE_MESSAGE"), "$", "$$"),
	}

	// On release branches Percy must compare to the previous commit of the release branch, not main.
	if c.RunType.Is(ReleaseBranch) {
		env["PERCY_TARGET_BRANCH"] = c.Branch
	}

	// Build options for pipeline operations that spawn more build steps
	buildOptions := bk.BuildOptions{
		Message: os.Getenv("BUILDKITE_MESSAGE"),
		Commit:  c.Commit,
		Branch:  c.Branch,
		Env:     env,
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

	// Toggle profiling of each step
	if c.MessageFlags.ProfilingEnabled {
		bk.AfterEveryStepOpts = append(bk.AfterEveryStepOpts, func(s *bk.Step) {
			// wrap "time -v" around each command for CPU/RAM utilization information
			var prefixed []string
			for _, cmd := range s.Command {
				prefixed = append(prefixed, fmt.Sprintf("env time -v %s", cmd))
			}
			s.Command = prefixed
		})
	}

	// Set up operations that add steps to a pipeline.
	var ops operations.Set

	// This statement outlines the pipeline steps for each CI case.
	//
	// PERF: Try to order steps such that slower steps are first.
	switch c.RunType {
	case PullRequest:
		if c.ChangedFiles.AffectsClient() {
			// triggers a slow pipeline, currently only affects web. It's optional so we
			// set it up separately from CoreTestOperations
			ops.Append(triggerAsync(buildOptions))
		}

		ops.Merge(CoreTestOperations(c.ChangedFiles, CoreTestOperationsOptions{}))

	case BackendIntegrationTests:
		ops.Append(
			buildCandidateDockerImage("server", c.Version, c.candidateImageTag()),
			backendIntegrationTests(c.candidateImageTag()))

		// Run default set of PR checks as well
		ops.Merge(CoreTestOperations(c.ChangedFiles, CoreTestOperationsOptions{}))

	case BextReleaseBranch:
		// If this is a browser extension release branch, run the browser-extension tests and
		// builds.
		ops = operations.NewSet([]operations.Operation{
			addTsLint,
			addBrowserExt,
			frontendTests,
			wait,
			addBrowserExtensionReleaseSteps,
		})

	case BextNightly:
		// If this is a browser extension nightly build, run the browser-extension tests and
		// e2e tests.
		ops = operations.NewSet([]operations.Operation{
			addTsLint,
			addBrowserExt,
			frontendTests,
			wait,
			addBrowserExtensionE2ESteps,
		})

	case ImagePatch:
		// only build candidate image for the specified image in the branch name
		// see https://about.sourcegraph.com/handbook/engineering/deployments/testing#building-docker-images-for-a-specific-branch
		patchImage := c.Branch[20:]
		if !contains(images.SourcegraphDockerImages, patchImage) {
			panic(fmt.Sprintf("no image %q found", patchImage))
		}
		ops = operations.NewSet([]operations.Operation{
			buildCandidateDockerImage(patchImage, c.Version, c.candidateImageTag()),
		})

		// Trivy security scans
		ops.Append(trivyScanCandidateImage(patchImage, c.candidateImageTag()))
		// Test images
		ops.Merge(CoreTestOperations(nil, CoreTestOperationsOptions{}))
		// Publish images
		ops.Append(publishFinalDockerImage(c, patchImage, false))

	case ImagePatchNoTest:
		// If this is a no-test branch, then run only the Docker build. No tests are run.
		app := c.Branch[27:]
		ops = operations.NewSet([]operations.Operation{
			buildCandidateDockerImage(app, c.Version, c.candidateImageTag()),
			wait,
			publishFinalDockerImage(c, app, false),
		})

	case CandidatesNoTest:
		for _, dockerImage := range images.SourcegraphDockerImages {
			ops.Append(
				buildCandidateDockerImage(dockerImage, c.Version, c.candidateImageTag()))
		}

	case ExecutorPatchNoTest:
		ops = operations.NewSet([]operations.Operation{
			buildExecutor(c.Version, c.MessageFlags.SkipHashCompare),
			publishExecutor(c.Version, c.MessageFlags.SkipHashCompare),
		})

	default:
		// Slow async pipeline
		ops.Append(triggerAsync(buildOptions))

		// Slow image builds
		for _, dockerImage := range images.SourcegraphDockerImages {
			ops.Append(buildCandidateDockerImage(dockerImage, c.Version, c.candidateImageTag()))
		}

		// Trivy security scans
		for _, dockerImage := range images.SourcegraphDockerImages {
			ops.Append(trivyScanCandidateImage(dockerImage, c.candidateImageTag()))
		}

		// Executor VM image
		skipHashCompare := c.MessageFlags.SkipHashCompare || c.RunType.Is(ReleaseBranch)
		if c.RunType.Is(MainDryRun, MainBranch) {
			ops.Append(buildExecutor(c.Version, skipHashCompare))
		}

		// Slow tests
		if c.RunType.Is(MainDryRun, MainBranch) {
			ops.Append(backendIntegrationTests(c.candidateImageTag()))
		}

		// Core tests
		ops.Merge(CoreTestOperations(nil, CoreTestOperationsOptions{
			ChromaticShouldAutoAccept: c.RunType.Is(MainBranch),
		}))

		// Trigger e2e late so that it can leverage candidate images
		ops.Append(triggerE2EandQA(e2eAndQAOptions{
			candidateImage: c.candidateImageTag(),
			buildOptions:   buildOptions,
			async:          c.RunType.Is(MainBranch),
		}))

		// Add final artifacts
		for _, dockerImage := range images.SourcegraphDockerImages {
			ops.Append(publishFinalDockerImage(c, dockerImage, c.RunType.Is(MainBranch)))
		}
		// Executor VM image
		if c.RunType.Is(MainBranch) {
			ops.Append(publishExecutor(c.Version, skipHashCompare))
		}

		// Propagate changes elsewhere
		if c.RunType.Is(MainBranch) {
			ops.Append(
				// wait for all steps to pass
				wait,
				triggerUpdaterPipeline)
		}
	}

	// Construct pipeline
	pipeline := &bk.Pipeline{
		Env: env,
	}
	ops.Apply(pipeline)
	return pipeline, nil
}
