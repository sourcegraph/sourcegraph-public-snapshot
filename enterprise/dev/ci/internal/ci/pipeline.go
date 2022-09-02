// Package ci is responsible for generating a Buildkite pipeline configuration. It is invoked by the
// gen-pipeline.go command.
package ci

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/ci/runtype"
	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/images"
	bk "github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/buildkite"
	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/ci/changed"
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

		// Go flags
		"GO111MODULE": "on",

		// Additional flags
		"FORCE_COLOR": "3",
		"ENTERPRISE":  "1",
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

		// HoneyComb dataset that stores build traces.
		"CI_BUILDEVENT_DATASET": "buildkite",
	}
	bk.FeatureFlags.ApplyEnv(env)

	// On release branches Percy must compare to the previous commit of the release branch, not main.
	if c.RunType.Is(runtype.ReleaseBranch, runtype.TaggedRelease) {
		env["PERCY_TARGET_BRANCH"] = c.Branch
		// When we are building a release, we do not want to cache the client bundle.
		//
		// This is a defensive measure, as caching the client bundle is tricky when it comes to invalidating it.
		// This makes sure that we're running integration tests on a fresh bundle and, the image
		// that 99% of our customers are using is exactly the same as the other deployments.
		env["SERVER_NO_CLIENT_BUNDLE_CACHE"] = "true"
	}

	// Build options for pipeline operations that spawn more build steps
	buildOptions := bk.BuildOptions{
		Message: os.Getenv("BUILDKITE_MESSAGE"),
		Commit:  c.Commit,
		Branch:  c.Branch,
		Env:     env,
	}

	// Test upgrades from mininum upgradeable Sourcegraph version - updated by release tool
	const minimumUpgradeableVersion = "3.43.0"

	// Set up operations that add steps to a pipeline.
	ops := operations.NewSet()

	if op, err := exposeBuildMetadata(c); err == nil {
		ops.Merge(operations.NewNamedSet("Metadata", op))
	}

	// This statement outlines the pipeline steps for each CI case.
	//
	// PERF: Try to order steps such that slower steps are first.
	switch c.RunType {
	case runtype.PullRequest:
		// First, we set up core test operations that apply both to PRs and to other run
		// types such as main.
		ops.Merge(CoreTestOperations(c.Diff, CoreTestOperationsOptions{
			MinimumUpgradeableVersion: minimumUpgradeableVersion,
			ForceReadyForReview:       c.MessageFlags.ForceReadyForReview,
			// TODO: (@umpox, @valerybugakov) Figure out if we can reliably enable this in PRs.
			ClientLintOnlyChangedFiles: false,
		}))

		// Now we set up conditional operations that only apply to pull requests.
		if c.Diff.Has(changed.Client) {
			// triggers a slow pipeline, currently only affects web. It's optional so we
			// set it up separately from CoreTestOperations
			ops.Merge(operations.NewNamedSet(operations.PipelineSetupSetName,
				triggerAsync(buildOptions)))

			// Do not create client PR preview if Go or GraphQL is changed to avoid confusing
			// preview behavior, because only Client code is used to deploy application preview.
			if !c.Diff.Has(changed.Go) && !c.Diff.Has(changed.GraphQL) {
				ops.Append(prPreview())
			}
		}
		if c.Diff.Has(changed.DockerImages) {
			// Build and scan docker images
			testBuilds := operations.NewNamedSet("Test builds")
			scanBuilds := operations.NewNamedSet("Scan test builds")
			for _, image := range images.SourcegraphDockerImages {
				testBuilds.Append(buildCandidateDockerImage(image, c.Version, c.candidateImageTag(), false))
				scanBuilds.Append(trivyScanCandidateImage(image, c.candidateImageTag()))
			}
			ops.Merge(testBuilds)
			ops.Merge(scanBuilds)
		}

	case runtype.ReleaseNightly:
		ops.Append(triggerReleaseBranchHealthchecks(minimumUpgradeableVersion))

	case runtype.BackendIntegrationTests:
		ops.Append(
			buildCandidateDockerImage("server", c.Version, c.candidateImageTag(), false),
			backendIntegrationTests(c.candidateImageTag()))

		// always include very backend-oriented changes in this set of tests
		testDiff := c.Diff | changed.DatabaseSchema | changed.Go
		ops.Merge(CoreTestOperations(
			testDiff,
			CoreTestOperationsOptions{MinimumUpgradeableVersion: minimumUpgradeableVersion},
		))

	case runtype.BextReleaseBranch:
		// If this is a browser extension release branch, run the browser-extension tests and
		// builds.
		ops = operations.NewSet(
			addClientLintersForAllFiles,
			addBrowserExtensionUnitTests,
			addBrowserExtensionIntegrationTests(0), // we pass 0 here as we don't have other pipeline steps to contribute to the resulting Percy build
			frontendTests,
			wait,
			addBrowserExtensionReleaseSteps)

	case runtype.VsceReleaseBranch:
		// If this is a vs code extension release branch, run the vscode-extension tests and release
		ops = operations.NewSet(
			addClientLintersForAllFiles,
			addVsceIntegrationTests,
			wait,
			addVsceReleaseSteps)

	case runtype.BextNightly:
		// If this is a browser extension nightly build, run the browser-extension tests and
		// e2e tests.
		ops = operations.NewSet(
			addClientLintersForAllFiles,
			addBrowserExtensionUnitTests,
			recordBrowserExtensionIntegrationTests,
			frontendTests,
			wait,
			addBrowserExtensionE2ESteps)

	case runtype.VsceNightly:
		// If this is a VS Code extension nightly build, run the vsce-extension integration tests
		ops = operations.NewSet(
			addClientLintersForAllFiles,
			// TODO: fix integrations tests and re-enable: https://github.com/sourcegraph/sourcegraph/issues/40891
			// addVsceIntegrationTests,
		)

	case runtype.ImagePatch:
		// only build image for the specified image in the branch name
		// see https://handbook.sourcegraph.com/engineering/deployments#building-docker-images-for-a-specific-branch
		patchImage, err := c.RunType.Matcher().ExtractBranchArgument(c.Branch)
		if err != nil {
			panic(fmt.Sprintf("ExtractBranchArgument: %s", err))
		}
		if !contains(images.SourcegraphDockerImages, patchImage) {
			panic(fmt.Sprintf("no image %q found", patchImage))
		}

		ops = operations.NewSet(
			buildCandidateDockerImage(patchImage, c.Version, c.candidateImageTag(), false),
			trivyScanCandidateImage(patchImage, c.candidateImageTag()))
		// Test images
		ops.Merge(CoreTestOperations(changed.All, CoreTestOperationsOptions{
			MinimumUpgradeableVersion: minimumUpgradeableVersion,
		}))
		// Publish images after everything is done
		ops.Append(
			wait,
			publishFinalDockerImage(c, patchImage))

	case runtype.ImagePatchNoTest:
		// If this is a no-test branch, then run only the Docker build. No tests are run.
		patchImage, err := c.RunType.Matcher().ExtractBranchArgument(c.Branch)
		if err != nil {
			panic(fmt.Sprintf("ExtractBranchArgument: %s", err))
		}
		if !contains(images.SourcegraphDockerImages, patchImage) {
			panic(fmt.Sprintf("no image %q found", patchImage))
		}
		ops = operations.NewSet(
			buildCandidateDockerImage(patchImage, c.Version, c.candidateImageTag(), false),
			wait,
			publishFinalDockerImage(c, patchImage))

	case runtype.CandidatesNoTest:
		for _, dockerImage := range images.SourcegraphDockerImages {
			ops.Append(
				buildCandidateDockerImage(dockerImage, c.Version, c.candidateImageTag(), false))
		}

	case runtype.ExecutorPatchNoTest:
		executorVMImage := "executor-vm"
		ops = operations.NewSet(
			buildCandidateDockerImage(executorVMImage, c.Version, c.candidateImageTag(), false),
			trivyScanCandidateImage(executorVMImage, c.candidateImageTag()),
			buildExecutor(c, true),
			buildExecutorDockerMirror(c),
			wait,
			publishFinalDockerImage(c, executorVMImage),
			publishExecutor(c, true),
			publishExecutorDockerMirror(c),
		)

	default:
		// Slow async pipeline
		ops.Merge(operations.NewNamedSet(operations.PipelineSetupSetName,
			triggerAsync(buildOptions)))

		// Slow image builds
		imageBuildOps := operations.NewNamedSet("Image builds")
		for _, dockerImage := range images.SourcegraphDockerImages {
			// Only upload sourcemaps for the "frontend" image, on the Main branch build
			uploadSourcemaps := false
			if c.RunType.Is(runtype.MainBranch) && dockerImage == "frontend" {
				uploadSourcemaps = true
			}
			imageBuildOps.Append(buildCandidateDockerImage(dockerImage, c.Version, c.candidateImageTag(), uploadSourcemaps))
		}
		// Executor VM image
		skipHashCompare := c.MessageFlags.SkipHashCompare || c.RunType.Is(runtype.ReleaseBranch, runtype.TaggedRelease)
		if c.RunType.Is(runtype.MainDryRun, runtype.MainBranch, runtype.ReleaseBranch, runtype.TaggedRelease) {
			imageBuildOps.Append(buildExecutor(c, skipHashCompare))
			if c.RunType.Is(runtype.ReleaseBranch, runtype.TaggedRelease) || c.Diff.Has(changed.ExecutorDockerRegistryMirror) {
				imageBuildOps.Append(buildExecutorDockerMirror(c))
			}
		}
		ops.Merge(imageBuildOps)

		// Trivy security scans
		imageScanOps := operations.NewNamedSet("Image security scans")
		for _, dockerImage := range images.SourcegraphDockerImages {
			imageScanOps.Append(trivyScanCandidateImage(dockerImage, c.candidateImageTag()))
		}
		ops.Merge(imageScanOps)

		// Core tests
		ops.Merge(CoreTestOperations(changed.All, CoreTestOperationsOptions{
			ChromaticShouldAutoAccept: c.RunType.Is(runtype.MainBranch),
			MinimumUpgradeableVersion: minimumUpgradeableVersion,
			ForceReadyForReview:       c.MessageFlags.ForceReadyForReview,
		}))

		// Integration tests
		ops.Merge(operations.NewNamedSet("Integration tests",
			backendIntegrationTests(c.candidateImageTag()),
			codeIntelQA(c.candidateImageTag()),
		))
		// End-to-end tests
		ops.Merge(operations.NewNamedSet("End-to-end tests",
			serverE2E(c.candidateImageTag()),
			serverQA(c.candidateImageTag()),
			clusterQA(c.candidateImageTag()),
			testUpgrade(c.candidateImageTag(), minimumUpgradeableVersion),
		))

		// All operations before this point are required
		ops.Append(wait)

		// Add final artifacts
		publishOps := operations.NewNamedSet("Publish images")
		for _, dockerImage := range images.SourcegraphDockerImages {
			publishOps.Append(publishFinalDockerImage(c, dockerImage))
		}
		// Executor VM image
		if c.RunType.Is(runtype.MainBranch, runtype.TaggedRelease) {
			publishOps.Append(publishExecutor(c, skipHashCompare))
			if c.RunType.Is(runtype.TaggedRelease) || c.Diff.Has(changed.ExecutorDockerRegistryMirror) {
				publishOps.Append(publishExecutorDockerMirror(c))
			}
		}
		ops.Merge(publishOps)
	}

	ops.Append(
		wait,                    // wait for all steps to pass
		uploadBuildeventTrace(), // upload the final buildevent trace if the build succeeded.
	)

	// Construct pipeline
	pipeline := &bk.Pipeline{
		Env: env,
		AfterEveryStepOpts: []bk.StepOpt{
			withDefaultTimeout,
			withAgentQueueDefaults,
			withAgentLostRetries,
		},
	}
	// Toggle profiling of each step
	if c.MessageFlags.ProfilingEnabled {
		pipeline.AfterEveryStepOpts = append(pipeline.AfterEveryStepOpts, withProfiling)
	}

	// Apply operations on pipeline
	ops.Apply(pipeline)

	// Validate generated pipeline have unique keys
	if err := pipeline.EnsureUniqueKeys(make(map[string]int)); err != nil {
		return nil, err
	}

	return pipeline, nil
}

// withDefaultTimeout makes all command steps timeout after 60 minutes in case a buildkite
// agent got stuck / died.
func withDefaultTimeout(s *bk.Step) {
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
}

// withAgentQueueDefaults ensures all agents target a specific queue, and ensures they
// steps are configured appropriately to run on the queue
func withAgentQueueDefaults(s *bk.Step) {
	if len(s.Agents) == 0 || s.Agents["queue"] == "" {
		s.Agents["queue"] = bk.AgentQueueStateless
	}
}

// withProfiling wraps "time -v" around each command for CPU/RAM utilization information
func withProfiling(s *bk.Step) {
	var prefixed []string
	for _, cmd := range s.Command {
		prefixed = append(prefixed, fmt.Sprintf("env time -v %s", cmd))
	}
	s.Command = prefixed
}

// withAgentLostRetries insert automatic retries when the job has failed because it lost its agent.
//
// If the step has been marked as not retryable, the retry will be skipped.
func withAgentLostRetries(s *bk.Step) {
	if s.Retry != nil && s.Retry.Manual != nil && !s.Retry.Manual.Allowed {
		return
	}
	if s.Retry == nil {
		s.Retry = &bk.RetryOptions{}
	}
	if s.Retry.Automatic == nil {
		s.Retry.Automatic = []bk.AutomaticRetryOptions{}
	}
	s.Retry.Automatic = append(s.Retry.Automatic, bk.AutomaticRetryOptions{
		Limit:      1,
		ExitStatus: -1,
	})
}
