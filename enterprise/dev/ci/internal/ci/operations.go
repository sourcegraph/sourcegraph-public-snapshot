package ci

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/images"
	bk "github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/buildkite"
	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/ci/changed"
	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/ci/operations"
)

// CoreTestOperationsOptions should be used ONLY to adjust the behaviour of specific steps,
// e.g. by adding flags, and not as a condition for adding steps or commands.
type CoreTestOperationsOptions struct {
	// for clientChromaticTests
	ChromaticShouldAutoAccept bool
}

// CoreTestOperations is a core set of tests that should be run in most CI cases. More
// notably, this is what is used to define operations that run on PRs. Please read the
// following notes:
//
// - changedFiles can be nil to run all tests.
// - opts should be used ONLY to adjust the behaviour of specific steps, e.g. by adding flags,
// and not as a condition for adding steps or commands.
// - be careful not to add duplicate steps.
//
// If the conditions for the addition of an operation cannot be expressed using the above
// arguments, please add it to the switch case within `GeneratePipeline` instead.
func CoreTestOperations(changedFiles changed.Files, opts CoreTestOperationsOptions) *operations.Set {
	// Various RunTypes can provide a nil changedFiles to run all checks.
	runAll := len(changedFiles) == 0

	// Base set
	ops := operations.NewSet([]operations.Operation{
		// lightweight check that works over a lot of stuff - we are okay with running
		// these on all PRs
		addPrettier,
		addCheck,
	})

	if runAll || changedFiles.AffectsClient() || changedFiles.AffectsGraphQL() {
		// If there are any Graphql changes, they are impacting the client as well.
		ops.Append(
			clientIntegrationTests,
			clientChromaticTests(opts.ChromaticShouldAutoAccept),
			frontendTests,   // ~4.5m
			addWebApp,       // ~3m
			addBrowserExt,   // ~2m
			addBrandedTests, // ~1.5m
			addTsLint,
		)
	}

	if runAll || changedFiles.AffectsGo() || changedFiles.AffectsGraphQL() {
		// If there are any Graphql changes, they are impacting the backend as well.
		ops.Append(
			addGoTests,
		)

		// If the changes are only in ./dev/sg then we skip the build
		if runAll || !changedFiles.AffectsSg() {
			ops.Append(
				addGoBuild, // ~0.5m
			)
		}
	}

	if runAll || changedFiles.AffectsGraphQL() {
		ops.Append(addGraphQLLint)
	}

	if runAll || changedFiles.AffectsDockerfiles() {
		ops.Append(addDockerfileLint)
	}

	if runAll || changedFiles.AffectsDocs() {
		ops.Append(addDocs)
	}

	// wait for all steps to pass
	ops.Append(wait)
	return &ops
}

// Verifies the docs formatting and builds the `docsite` command.
func addDocs(pipeline *bk.Pipeline) {
	pipeline.AddStep(":memo: Check and build docsite",
		bk.Cmd("./dev/check/docsite.sh"))
}

// Adds the static check test step.
func addCheck(pipeline *bk.Pipeline) {
	pipeline.AddStep(":clipboard: Misc Linters",
		bk.Cmd("./dev/check/all.sh"))
}

// yarn ~41s + ~30s
func addPrettier(pipeline *bk.Pipeline) {
	pipeline.AddStep(":lipstick: Prettier",
		bk.Cmd("dev/ci/yarn-run.sh prettier-check"))
}

// yarn ~41s + ~1s
func addGraphQLLint(pipeline *bk.Pipeline) {
	pipeline.AddStep(":lipstick: :graphql:",
		bk.Cmd("dev/ci/yarn-run.sh graphql-lint"))
}

// Adds Typescript linting. (2x ~41s) + ~60s + ~137s + 7s
func addTsLint(pipeline *bk.Pipeline) {
	// - yarn 41s (required on all steps)
	// - build-ts 60s
	// - eslint 137s
	// - stylelint 7s
	pipeline.AddStep(":eslint: Typescript eslint",
		bk.Cmd("dev/ci/yarn-run.sh build-ts all:eslint")) // eslint depends on build-ts
	pipeline.AddStep(":stylelint: Stylelint",
		bk.Cmd("dev/ci/yarn-run.sh all:stylelint"))
}

// Adds steps for the OSS and Enterprise web app builds. Runs the web app tests.
func addWebApp(pipeline *bk.Pipeline) {
	// Webapp build
	pipeline.AddStep(":webpack::globe_with_meridians: Build",
		bk.Cmd("dev/ci/yarn-build.sh client/web"),
		bk.Env("NODE_ENV", "production"),
		bk.Env("ENTERPRISE", ""))

	// Webapp enterprise build
	pipeline.AddStep(":webpack::globe_with_meridians::moneybag: Enterprise build",
		bk.Cmd("dev/ci/yarn-build.sh client/web"),
		bk.Env("NODE_ENV", "production"),
		bk.Env("ENTERPRISE", "1"),
		bk.Env("CHECK_BUNDLESIZE", "1"),
		// To ensure the Bundlesize output can be diffed to the baseline on main
		bk.Env("WEBPACK_USE_NAMED_CHUNKS", "true"))

	// Webapp tests
	pipeline.AddStep(":jest::globe_with_meridians: Test",
		bk.Cmd("dev/ci/yarn-test.sh client/web"),
		bk.Cmd("dev/ci/codecov.sh -c -F typescript -F unit"))
}

// We provide our own Chromium instance that is installed through the `download-puppeteer-browser` script
var percyBrowserExecutableEnv = bk.Env("PERCY_BROWSER_EXECUTABLE", "node_modules/puppeteer/.local-chromium/linux-901812/chrome-linux/chrome")

// Builds and tests the browser extension.
func addBrowserExt(pipeline *bk.Pipeline) {
	// Browser extension integration tests
	for _, browser := range []string{"chrome"} {
		pipeline.AddStep(
			fmt.Sprintf(":%s: Puppeteer tests for %s extension", browser, browser),
			percyBrowserExecutableEnv,
			bk.Env("EXTENSION_PERMISSIONS_ALL_URLS", "true"),
			bk.Env("BROWSER", browser),
			bk.Env("LOG_BROWSER_CONSOLE", "true"),
			bk.Env("SOURCEGRAPH_BASE_URL", "https://sourcegraph.com"),
			bk.Env("RECORD", "false"), // ensure that we use existing recordings
			bk.Cmd("yarn --frozen-lockfile --network-timeout 60000"),
			bk.Cmd("yarn --cwd client/shared run download-puppeteer-browser"),
			bk.Cmd("yarn --cwd client/browser -s run build"),
			bk.Cmd("yarn run cover-browser-integration"),
			bk.Cmd("yarn nyc report -r json"),
			bk.Cmd("dev/ci/codecov.sh -c -F typescript -F integration"),
			bk.ArtifactPaths("./puppeteer/*.png"),
		)
	}

	// Browser extension unit tests
	pipeline.AddStep(":jest::chrome: Test browser extension",
		bk.Cmd("dev/ci/yarn-test.sh client/browser"),
		bk.Cmd("dev/ci/codecov.sh -c -F typescript -F unit"))
}

func clientIntegrationTests(pipeline *bk.Pipeline) {
	chunkSize := 3
	prepStepKey := "puppeteer:prep"
	skipGitCloneStep := bk.Plugin("uber-workflow/run-without-clone", "")

	// Build web application used for integration tests to share it between multiple parallel steps.
	pipeline.AddStep(":puppeteer::electric_plug: Puppeteer tests prep",
		bk.Key(prepStepKey),
		bk.Env("ENTERPRISE", "1"),
		bk.Cmd("COVERAGE_INSTRUMENT=true dev/ci/yarn-build.sh client/web"),
		bk.Cmd("dev/ci/create-client-artifact.sh"))

	// Chunk web integration tests to save time via parallel execution.
	chunkedTestFiles := getChunkedWebIntegrationFileNames(chunkSize)
	// Percy finalize step should be executed after all integration tests.
	puppeteerFinalizeDependencies := make([]bk.StepOpt, len(chunkedTestFiles))

	// Add pipeline step for each chunk of web integrations files.
	for i, chunkTestFiles := range chunkedTestFiles {
		stepLabel := fmt.Sprintf(":puppeteer::electric_plug: Puppeteer tests chunk #%s", fmt.Sprint(i+1))

		stepKey := fmt.Sprintf("puppeteer:chunk:%s", fmt.Sprint(i+1))
		puppeteerFinalizeDependencies[i] = bk.DependsOn(stepKey)

		pipeline.AddStep(stepLabel,
			bk.Key(stepKey),
			bk.DependsOn(prepStepKey),
			bk.DisableManualRetry("The Percy build is finalized even if one of the concurrent agents fails. To retry correctly, restart the entire pipeline."),
			percyBrowserExecutableEnv,
			bk.Env("PERCY_ON", "true"),
			bk.Cmd(fmt.Sprintf(`dev/ci/yarn-web-integration.sh "%s"`, chunkTestFiles)),
			bk.ArtifactPaths("./puppeteer/*.png"))
	}

	finalizeSteps := []bk.StepOpt{
		// Allow to teardown the Percy build even if there was a failure in the earlier Percy steps.
		bk.AllowDependencyFailure(),
		// Percy service often fails for obscure reasons. The step is pretty fast, so we
		// just retry a few times.
		bk.AutomaticRetry(3),
		// Finalize just uses a remote package.
		skipGitCloneStep,
		bk.Cmd("npx @percy/cli build:finalize"),
	}

	pipeline.AddStep(":puppeteer::electric_plug: Puppeteer tests finalize",
		append(finalizeSteps, puppeteerFinalizeDependencies...)...)
}

func clientChromaticTests(autoAcceptChanges bool) operations.Operation {
	return func(pipeline *bk.Pipeline) {
		// Upload storybook to Chromatic
		chromaticCommand := "yarn chromatic --exit-zero-on-changes --exit-once-uploaded"

		if autoAcceptChanges {
			chromaticCommand += " --auto-accept-changes"
		}

		pipeline.AddStep(":chromatic: Upload Storybook to Chromatic",
			bk.AutomaticRetry(3),
			bk.Cmd("yarn --mutex network --frozen-lockfile --network-timeout 60000"),
			bk.Cmd("yarn gulp generate"),
			bk.Env("MINIFY", "1"),
			bk.Cmd(chromaticCommand))
	}
}

// Adds the shared frontend tests (shared between the web app and browser extension).
func frontendTests(pipeline *bk.Pipeline) {
	// Shared tests
	pipeline.AddStep(":jest: Test shared client code",
		bk.Cmd("dev/ci/yarn-test.sh client/shared"),
		bk.Cmd("dev/ci/codecov.sh -c -F typescript -F unit"))

	// Wildcard tests
	pipeline.AddStep(":jest: Test wildcard client code",
		bk.Cmd("dev/ci/yarn-test.sh client/wildcard"),
		bk.Cmd("dev/ci/codecov.sh -c -F typescript -F unit"))
}

func addBrandedTests(pipeline *bk.Pipeline) {
	pipeline.AddStep(":jest: Test branded client code",
		bk.Cmd("dev/ci/yarn-test.sh client/branded"),
		bk.Cmd("dev/ci/codecov.sh -c -F typescript -F unit"))
}

// Adds the Go test step.
func addGoTests(pipeline *bk.Pipeline) {
	pipeline.AddStep(":go: Test",
		// Temporary fix until we fix the code-intel flake
		bk.AutomaticRetry(3),
		bk.Cmd("./dev/ci/go-test.sh"),
		bk.Cmd("dev/ci/codecov.sh -c -F go"))
}

// Builds the OSS and Enterprise Go commands.
func addGoBuild(pipeline *bk.Pipeline) {
	pipeline.AddStep(":go: Build",
		bk.Cmd("./dev/ci/go-build.sh"),
	)
}

// Lints the Dockerfiles.
func addDockerfileLint(pipeline *bk.Pipeline) {
	pipeline.AddStep(":docker: Lint",
		bk.Cmd("./dev/ci/docker-lint.sh"))
}

// Adds backend integration tests step.
//
// Runtime: ~11m
func backendIntegrationTests(candidateImageTag string) operations.Operation {
	return func(pipeline *bk.Pipeline) {
		pipeline.AddStep(":chains: Backend integration tests",
			// Run tests against the candidate server image
			bk.DependsOn(candidateImageStepKey("server")),
			bk.Env("IMAGE",
				images.DevRegistryImage("server", candidateImageTag)),
			bk.Cmd("./dev/ci/backend-integration.sh"),
			bk.ArtifactPaths("./*.log"))
	}
}

func addBrowserExtensionE2ESteps(pipeline *bk.Pipeline) {
	for _, browser := range []string{"chrome"} {
		// Run e2e tests
		pipeline.AddStep(fmt.Sprintf(":%s: E2E for %s extension", browser, browser),
			percyBrowserExecutableEnv,
			bk.Env("EXTENSION_PERMISSIONS_ALL_URLS", "true"),
			bk.Env("BROWSER", browser),
			bk.Env("LOG_BROWSER_CONSOLE", "true"),
			bk.Env("SOURCEGRAPH_BASE_URL", "https://sourcegraph.com"),
			bk.Cmd("yarn --frozen-lockfile --network-timeout 60000"),
			bk.Cmd("yarn --cwd client/shared run download-puppeteer-browser"),
			bk.Cmd("pushd client/browser"),
			bk.Cmd("yarn -s run build"),
			bk.Cmd("yarn -s mocha ./src/end-to-end/github.test.ts ./src/end-to-end/gitlab.test.ts"),
			bk.Cmd("popd"),
			bk.ArtifactPaths("./puppeteer/*.png"))
	}
}

// Release the browser extension.
func addBrowserExtensionReleaseSteps(pipeline *bk.Pipeline) {
	addBrowserExtensionE2ESteps(pipeline)

	pipeline.AddWait()

	// Release to the Chrome Webstore
	pipeline.AddStep(":rocket::chrome: Extension release",
		bk.Cmd("yarn --frozen-lockfile --network-timeout 60000"),
		bk.Cmd("pushd client/browser"),
		bk.Cmd("yarn -s run build"),
		bk.Cmd("yarn release:chrome"),
		bk.Cmd("popd"))

	// Build and self sign the FF add-on and upload it to a storage bucket
	pipeline.AddStep(":rocket::firefox: Extension release",
		bk.Cmd("yarn --frozen-lockfile --network-timeout 60000"),
		bk.Cmd("pushd client/browser"),
		bk.Cmd("yarn release:ff"),
		bk.Cmd("popd"))

	// Release to npm
	pipeline.AddStep(":rocket::npm: NPM Release",
		bk.Cmd("yarn --frozen-lockfile --network-timeout 60000"),
		bk.Cmd("pushd client/browser"),
		bk.Cmd("yarn -s run build"),
		bk.Cmd("yarn release:npm"),
		bk.Cmd("popd"))
}

// Adds a Buildkite pipeline "Wait".
func wait(pipeline *bk.Pipeline) {
	pipeline.AddWait()
}

// Trigger the async pipeline to run. See pipeline.async.yaml.
func triggerAsync(buildOptions bk.BuildOptions) operations.Operation {
	return func(pipeline *bk.Pipeline) {
		pipeline.AddTrigger(":snail: Trigger async",
			bk.Trigger("sourcegraph-async"),
			bk.Async(true),
			bk.Build(buildOptions),
		)
	}
}

func triggerUpdaterPipeline(pipeline *bk.Pipeline) {
	pipeline.AddStep(":github: :date: :k8s: Trigger k8s updates if current commit is tip of 'main'",
		bk.Cmd(".buildkite/updater/trigger-if-tip-of-main.sh"),
		bk.Concurrency(1),
		bk.ConcurrencyGroup("sourcegraph/sourcegraph-k8s-update-trigger"),
	)
}

// images used by cluster-qa test
func clusterDockerImages(images []string) string {
	var clusterImages []string
	imagesToRemove := map[string]bool{"server": true, "ignite-ubuntu": true}
	for _, image := range images {
		if _, exists := imagesToRemove[image]; !exists {
			clusterImages = append(clusterImages, image)
		}
	}
	return strings.Join(clusterImages, "\n")
}

type e2eAndQAOptions struct {
	candidateImage string
	buildOptions   bk.BuildOptions
	async          bool
}

// copyEnv copies a subset of env variables from the given BuildOptions
func (opts *e2eAndQAOptions) copyEnv(keys ...string) map[string]string {
	m := map[string]string{}
	for _, k := range keys {
		if v, ok := opts.buildOptions.Env[k]; ok {
			m[k] = v
		}
	}
	return m
}

func triggerE2EandQA(opts e2eAndQAOptions) operations.Operation {
	customOptions := bk.BuildOptions{
		Message: opts.buildOptions.Message,
		Branch:  opts.buildOptions.Branch,
		Commit:  opts.buildOptions.Commit,
		Env: opts.copyEnv(
			"BUILDKITE_PULL_REQUEST",
			"BUILDKITE_PULL_REQUEST_BASE_BRANCH",
			"BUILDKITE_PULL_REQUEST_REPO",
			"COMMIT_SHA",
			"DATE",
			"VERSION",
			"CI_DEBUG_PROFILE",
		),
	}

	// Set variables that indicate the tag for 'us.gcr.io/sourcegraph-dev' images built
	// from this CI run's commit, and credentials to access them.
	customOptions.Env["CANDIDATE_VERSION"] = opts.candidateImage
	customOptions.Env["VAGRANT_SERVICE_ACCOUNT"] = "buildkite@sourcegraph-ci.iam.gserviceaccount.com"

	// Test upgrades from mininum upgradeable Sourcegraph version - updated by release tool
	customOptions.Env["MINIMUM_UPGRADEABLE_VERSION"] = "3.32.0"

	// Docker images used in cluster tests
	customOptions.Env["DOCKER_CLUSTER_IMAGES_TXT"] = clusterDockerImages(images.SourcegraphDockerImages)

	return func(pipeline *bk.Pipeline) {
		pipeline.AddTrigger(":chromium: Trigger QA pipeline",
			bk.Trigger("qa"),
			bk.Async(opts.async),
			bk.Build(customOptions),
		)
	}
}

// candidateImageStepKey is the key for the given app (see the `images` package). Useful for
// adding dependencies on a step.
func candidateImageStepKey(app string) string {
	return strings.ReplaceAll(app, ".", "-") + ":candidate"
}

// Build a candidate docker image that will re-tagged with the final
// tags once the e2e tests pass.
//
// Version is the actual version of the code, and
func buildCandidateDockerImage(app, version, tag string) operations.Operation {
	return func(pipeline *bk.Pipeline) {
		image := strings.ReplaceAll(app, "/", "-")
		localImage := "sourcegraph/" + image + ":" + version

		cmds := []bk.StepOpt{
			bk.Key(candidateImageStepKey(app)),
			bk.Cmd(fmt.Sprintf(`echo "Building candidate %s image..."`, app)),
			bk.Env("DOCKER_BUILDKIT", "1"),
			bk.Env("IMAGE", localImage),
			bk.Env("VERSION", version),
			bk.Cmd("yes | gcloud auth configure-docker"),
		}

		if _, err := os.Stat(filepath.Join("docker-images", app)); err == nil {
			// Building Docker image located under $REPO_ROOT/docker-images/
			cmds = append(cmds, bk.Cmd(filepath.Join("docker-images", app, "build.sh")))
		} else {
			// Building Docker images located under $REPO_ROOT/cmd/
			cmdDir := func() string {
				if _, err := os.Stat(filepath.Join("enterprise/cmd", app)); err != nil {
					fmt.Fprintf(os.Stderr, "github.com/sourcegraph/sourcegraph/enterprise/cmd/%s does not exist so building github.com/sourcegraph/sourcegraph/cmd/%s instead\n", app, app)
					return "cmd/" + app
				}
				return "enterprise/cmd/" + app
			}()
			preBuildScript := cmdDir + "/pre-build.sh"
			if _, err := os.Stat(preBuildScript); err == nil {
				cmds = append(cmds, bk.Cmd(preBuildScript))
			}
			cmds = append(cmds, bk.Cmd(cmdDir+"/build.sh"))
		}

		devImage := images.DevRegistryImage(app, tag)
		cmds = append(cmds,
			// Retag the local image for dev registry
			bk.Cmd(fmt.Sprintf("docker tag %s %s", localImage, devImage)),
			// Publish tagged image
			bk.Cmd(fmt.Sprintf("docker push %s", devImage)),
		)

		pipeline.AddStep(fmt.Sprintf(":docker: :construction: %s", app), cmds...)
	}
}

// Ask trivy, a security scanning tool, to scan the candidate image
// specified by "app" and "tag".
func trivyScanCandidateImage(app, tag string) operations.Operation {
	image := images.DevRegistryImage(app, tag)

	// This is the special exit code that we tell trivy to use
	// if it finds a vulnerability. This is also used to soft-fail
	// this step.
	vulnerabilityExitCode := 27

	return func(pipeline *bk.Pipeline) {
		cmds := []bk.StepOpt{
			bk.DependsOn(candidateImageStepKey(app)),

			bk.Cmd(fmt.Sprintf("docker pull %s", image)),

			// have trivy use a shorter name in its output
			bk.Cmd(fmt.Sprintf("docker tag %s %s", image, app)),

			bk.Env("IMAGE", app),
			bk.Env("VULNERABILITY_EXIT_CODE", fmt.Sprintf("%d", vulnerabilityExitCode)),
			bk.ArtifactPaths("./*-security-report.html"),
			bk.SoftFail(vulnerabilityExitCode),

			bk.Cmd("./dev/ci/trivy/trivy-scan-high-critical.sh"),
		}

		pipeline.AddStep(fmt.Sprintf(":trivy: :docker: 🔎 %q", app), cmds...)
	}
}

// Tag and push final Docker image for the service defined by `app`
// after the e2e tests pass.
//
// It requires Config as an argument because published images require a lot of metadata.
func publishFinalDockerImage(c Config, app string, insiders bool) operations.Operation {
	return func(pipeline *bk.Pipeline) {
		devImage := images.DevRegistryImage(app, "")
		publishImage := images.PublishedRegistryImage(app, "")

		var images []string
		for _, image := range []string{publishImage, devImage} {
			if app != "server" || c.RunType.Is(TaggedRelease, ImagePatch, ImagePatchNoTest) {
				images = append(images, fmt.Sprintf("%s:%s", image, c.Version))
			}

			if app == "server" && c.RunType.Is(ReleaseBranch) {
				images = append(images, fmt.Sprintf("%s:%s-insiders", image, c.Branch))
			}

			if insiders {
				images = append(images, fmt.Sprintf("%s:insiders", image))
			}
		}

		// these tags are pushed to our dev registry, and are only
		// used internally
		for _, tag := range []string{
			c.Version,
			c.Commit,
			c.shortCommit(),
			fmt.Sprintf("%s_%s_%d", c.shortCommit(), c.Time.Format("2006-01-02"), c.BuildNumber),
			fmt.Sprintf("%s_%d", c.shortCommit(), c.BuildNumber),
			fmt.Sprintf("%s_%d", c.Commit, c.BuildNumber),
			strconv.Itoa(c.BuildNumber),
		} {
			internalImage := fmt.Sprintf("%s:%s", devImage, tag)
			images = append(images, internalImage)
		}

		candidateImage := fmt.Sprintf("%s:%s", devImage, c.candidateImageTag())
		cmd := fmt.Sprintf("./dev/ci/docker-publish.sh %s %s", candidateImage, strings.Join(images, " "))

		pipeline.AddStep(fmt.Sprintf(":docker: :truck: %s", app),
			// This step just pulls a prebuild image and pushes it to some registries. The
			// only possible failure here is a registry flake, so we retry a few times.
			bk.AutomaticRetry(3),
			bk.Cmd(cmd))
	}
}

// ~15m (building executor base VM)
func buildExecutor(version string, skipHashCompare bool) operations.Operation {
	return func(pipeline *bk.Pipeline) {
		stepOpts := []bk.StepOpt{
			bk.Key(candidateImageStepKey("executor")),
			bk.Env("VERSION", version),
		}
		if !skipHashCompare {
			compareHashScript := "./enterprise/dev/ci/scripts/compare-hash.sh"
			stepOpts = append(stepOpts,
				// Soft-fail with code 222 if nothing has changed
				bk.SoftFail(222),
				bk.Cmd(fmt.Sprintf("%s ./enterprise/cmd/executor/hash.sh", compareHashScript)))
		}
		stepOpts = append(stepOpts,
			bk.Cmd("./enterprise/cmd/executor/build.sh"))

		pipeline.AddStep(":packer: :construction: executor image", stepOpts...)
	}
}

func publishExecutor(version string, skipHashCompare bool) operations.Operation {
	return func(pipeline *bk.Pipeline) {
		candidateBuildStep := candidateImageStepKey("executor")
		stepOpts := []bk.StepOpt{
			bk.DependsOn(candidateBuildStep),
			bk.Env("VERSION", version),
		}
		if !skipHashCompare {
			// Publish iff not soft-failed on previous step
			checkDependencySoftFailScript := "./enterprise/dev/ci/scripts/check-dependency-soft-fail.sh"
			stepOpts = append(stepOpts,
				// Soft-fail with code 222 if nothing has changed
				bk.SoftFail(222),
				bk.Cmd(fmt.Sprintf("%s %s", checkDependencySoftFailScript, candidateBuildStep)))
		}
		stepOpts = append(stepOpts,
			bk.Cmd("./enterprise/cmd/executor/release.sh"))

		pipeline.AddStep(":packer: :white_check_mark: executor image", stepOpts...)
	}
}
