// Package ci is responsible for generating a Buildkite pipeline configuration. It is invoked by the
// gen-pipeline.go command.
package ci

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/google/go-github/v41/github"
	"github.com/slack-go/slack"

	"github.com/sourcegraph/sourcegraph/dev/team"
	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/images"
	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/buildkite"
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

		// Additional flags
		"GO111MODULE": "on",
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

	// Test upgrades from mininum upgradeable Sourcegraph version - updated by release tool
	const minimumUpgradeableVersion = "3.36.0"

	// Set up operations that add steps to a pipeline.
	ops := operations.NewSet()

	// This statement outlines the pipeline steps for each CI case.
	//
	// PERF: Try to order steps such that slower steps are first.
	switch c.RunType {
	case PullRequest:
		if c.Diff.Has(changed.Client) {
			// triggers a slow pipeline, currently only affects web. It's optional so we
			// set it up separately from CoreTestOperations
			ops.Merge(operations.NewNamedSet(operations.PipelineSetupSetName,
				triggerAsync(buildOptions)))
		}
		ops.Merge(CoreTestOperations(c.Diff, CoreTestOperationsOptions{MinimumUpgradeableVersion: minimumUpgradeableVersion}))

	case BackendIntegrationTests:
		ops.Append(
			buildCandidateDockerImage("server", c.Version, c.candidateImageTag()),
			backendIntegrationTests(c.candidateImageTag()))

		// Run default set of PR checks as well
		ops.Merge(CoreTestOperations(c.Diff, CoreTestOperationsOptions{MinimumUpgradeableVersion: minimumUpgradeableVersion}))

	case BextReleaseBranch:
		// If this is a browser extension release branch, run the browser-extension tests and
		// builds.
		ops = operations.NewSet(
			addTsLint,
			addBrowserExt,
			frontendTests,
			wait,
			addBrowserExtensionReleaseSteps)

	case BextNightly:
		// If this is a browser extension nightly build, run the browser-extension tests and
		// e2e tests.
		ops = operations.NewSet(
			addTsLint,
			addBrowserExt,
			frontendTests,
			wait,
			addBrowserExtensionE2ESteps)

	case ImagePatch:
		// only build candidate image for the specified image in the branch name
		// see https://handbook.sourcegraph.com/engineering/deployments#building-docker-images-for-a-specific-branch
		patchImage := c.Branch[20:]
		if !contains(images.SourcegraphDockerImages, patchImage) {
			panic(fmt.Sprintf("no image %q found", patchImage))
		}
		ops = operations.NewSet(
			buildCandidateDockerImage(patchImage, c.Version, c.candidateImageTag()))

		// Trivy security scans
		ops.Append(trivyScanCandidateImage(patchImage, c.candidateImageTag()))
		// Test images
		ops.Merge(CoreTestOperations(changed.All, CoreTestOperationsOptions{MinimumUpgradeableVersion: minimumUpgradeableVersion}))
		// Publish images after everything is done
		ops.Append(
			wait,
			publishFinalDockerImage(c, patchImage))

	case ImagePatchNoTest:
		// If this is a no-test branch, then run only the Docker build. No tests are run.
		app := c.Branch[27:]
		ops = operations.NewSet(
			buildCandidateDockerImage(app, c.Version, c.candidateImageTag()),
			wait,
			publishFinalDockerImage(c, app))

	case CandidatesNoTest:
		for _, dockerImage := range images.SourcegraphDockerImages {
			ops.Append(
				buildCandidateDockerImage(dockerImage, c.Version, c.candidateImageTag()))
		}

	case ExecutorPatchNoTest:
		ops = operations.NewSet(
			buildExecutor(c.Version, c.MessageFlags.SkipHashCompare),
			publishExecutor(c.Version, c.MessageFlags.SkipHashCompare),
			buildExecutorDockerMirror(c.Version),
			publishExecutorDockerMirror(c.Version))

	default:
		// Slow async pipeline
		ops.Merge(operations.NewNamedSet(operations.PipelineSetupSetName,
			triggerAsync(buildOptions)))

		// Slow image builds
		imageBuildOps := operations.NewNamedSet("Image builds")
		for _, dockerImage := range images.SourcegraphDockerImages {
			imageBuildOps.Append(buildCandidateDockerImage(dockerImage, c.Version, c.candidateImageTag()))
		}
		// Executor VM image
		skipHashCompare := c.MessageFlags.SkipHashCompare || c.RunType.Is(ReleaseBranch)
		if c.RunType.Is(MainDryRun, MainBranch, ReleaseBranch) {
			imageBuildOps.Append(buildExecutor(c.Version, skipHashCompare))
			if c.RunType.Is(ReleaseBranch) || c.Diff.Has(changed.ExecutorDockerRegistryMirror) {
				imageBuildOps.Append(buildExecutorDockerMirror(c.Version))
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
			ChromaticShouldAutoAccept: c.RunType.Is(MainBranch),
			MinimumUpgradeableVersion: minimumUpgradeableVersion,
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
			// Flaky deployment. See https://github.com/sourcegraph/sourcegraph/issues/25977
			// clusterQA(c.candidateImageTag()),
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
		if c.RunType.Is(MainBranch, ReleaseBranch) {
			publishOps.Append(publishExecutor(c.Version, skipHashCompare))
			if c.RunType.Is(ReleaseBranch) || c.Diff.Has(changed.ExecutorDockerRegistryMirror) {
				publishOps.Append(publishExecutorDockerMirror(c.Version))
			}
		}
		ops.Merge(publishOps)
	}

	ops.Append(
		wait,                    // wait for all steps to pass
		uploadBuildeventTrace(), // upload the final buildevent trace if the build succeeded.
	)

	// Construct pipeline
	pipeline := &bk.Pipeline{Env: env}
	ops.Apply(pipeline)

	// Validate generated pipeline have unique keys
	if err := ensureUniqueKeys(pipeline); err != nil {
		return nil, err
	}

	// Add a notify block
	if c.RunType.Is(MainBranch) {
		ctx := context.Background()

		// Slack client for retriving Slack profile data, not for making the request - for
		// more details, see the config.Notify docstring.
		slc := slack.New(c.Notify.SlackToken)

		// For now, we use an unauthenticated GitHub client because `sourcegraph/sourcegraph`
		// is a public repository.
		ghc := github.NewClient(http.DefaultClient)

		// Get teammate based on GitHub author of commit
		teammates := team.NewTeammateResolver(ghc, slc)
		tm, err := teammates.ResolveByCommitAuthor(ctx, "sourcegraph", "sourcegraph", c.Commit)
		if err != nil {
			pipeline.AddFailureSlackNotify(c.Notify.Channel, "", errors.Newf("failed to get Slack user: %w", err))
		} else {
			pipeline.AddFailureSlackNotify(c.Notify.Channel, tm.SlackID, nil)
		}
	}

	return pipeline, nil
}

func ensureUniqueKeys(pipeline *bk.Pipeline) error {
	occurences := map[string]int{}
	for _, step := range pipeline.Steps {
		if s, ok := step.(*buildkite.Step); ok {
			if s.Key == "" {
				return errors.Newf("empty key on step with label %q", s.Label)
			}
			occurences[s.Key] += 1
		}
	}
	for k, count := range occurences {
		if count > 1 {
			return errors.Newf("non unique key on step with key %q", k)
		}
	}
	return nil
}
