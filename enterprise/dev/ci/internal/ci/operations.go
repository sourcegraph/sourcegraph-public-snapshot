package ci

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/images"
	bk "github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/buildkite"
)

// Operation defines a function that adds something to a pipeline.
//
// Functions that return an Operation should never accept Config as an argument - they
// should only accept `changedFiles` or evaluated arguments.
//
// Operations should never conditionally add Steps and Operations - they should only use
// arguments to create variations of specific Operations (e.g. with different arguments).
type Operation func(*bk.Pipeline)

// CoreTestOperations is a core set of tests that should be run in most CI cases. These
// steps should generally be quite fast.
func CoreTestOperations(changedFiles ChangedFiles, buildOptions bk.BuildOptions) []Operation {
	// Default set
	operations := []Operation{
		triggerAsync(buildOptions), // triggers a slow pipeline, so do it first.
		addLint,                    // ~4.5m
		frontendTests,              // ~4.5m
		addWebApp,                  // ~3m
		addBrowserExt,              // ~2m
		addBrandedTests,            // ~1.5m
		addGoTests,                 // ~1.5m
		addCheck,                   // ~1m
		addGoBuild,                 // ~0.5m
		addDockerfileLint,          // ~0.2m
	}

	// Special-case branches provide a nil changedFiles to only run default changes.
	if len(changedFiles) == 0 {
		return append(operations, wait)
	}

	// Build special pipelines for changes that only touch a subset of code.
	switch {
	case changedFiles.onlyConfig():
		// If this PR only affects e.g. .github config files, no steps are necessary to run.
		operations = []Operation{}

	case changedFiles.onlyDocs():
		// If this is a docs-only PR, run only the steps necessary to verify the docs.
		operations = []Operation{
			addDocs,
		}

	case changedFiles.onlyGo() && !changedFiles.onlySg():
		// If this is a go-only PR, run only the steps necessary to verify the go code.
		operations = []Operation{
			addGoTests, // ~1.5m
			addCheck,   // ~1m
			addGoBuild, // ~0.5m
		}

	case changedFiles.onlySg():
		// If the changes are only in ./dev/sg then we only need to run a subset of steps.
		operations = []Operation{
			addGoTests,
			addCheck,
		}
	}

	// Add additional steps
	if changedFiles.affectsClient() {
		operations = append(operations, frontendPuppeteerAndStorybook(false))
	}

	// wait for all steps to pass
	return append(operations, wait)
}

// Verifies the docs formatting and builds the `docsite` command.
func addDocs(pipeline *bk.Pipeline) {
	pipeline.AddStep(":memo: Check and build docsite",
		bk.Cmd("./dev/ci/yarn-run.sh prettier-check"),
		bk.Cmd("./dev/check/docsite.sh"))
}

// Adds the static check test step.
func addCheck(pipeline *bk.Pipeline) {
	pipeline.AddStep(":white_check_mark: Misc Linters", bk.Cmd("./dev/check/all.sh"))
}

// Adds the lint test step.
func addLint(pipeline *bk.Pipeline) {
	// If we run all lints together it is our slow step (5m). So we split it
	// into two and try balance the runtime. yarn is a fixed cost so we always
	// pay it on a step. Aim for around 3m.
	//
	// Random sample of timings:
	//
	// - yarn 41s
	// - eslint 137s
	// - build-ts 60s
	// - prettier 29s
	// - stylelint 7s
	// - graphql-lint 1s
	pipeline.AddStep(":eslint: Lint all Typescript",
		bk.Cmd("dev/ci/yarn-run.sh build-ts all:eslint")) // eslint depends on build-ts
	pipeline.AddStep(":lipstick: :lint-roller: :stylelint: :graphql:", // TODO: Add header - Similar to the previous step
		bk.Cmd("dev/ci/yarn-run.sh prettier-check all:stylelint graphql-lint"))
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
		bk.Env("ENTERPRISE", "1"))

	// Webapp tests
	pipeline.AddStep(":jest::globe_with_meridians: Test",
		bk.Cmd("dev/ci/yarn-test.sh client/web"),
		bk.Cmd("dev/ci/codecov.sh -c -F typescript -F unit"))
}

// Builds and tests the browser extension.
func addBrowserExt(pipeline *bk.Pipeline) {
	// Browser extension integration tests
	for _, browser := range []string{"chrome"} {
		pipeline.AddStep(
			fmt.Sprintf(":%s: Puppeteer tests for %s extension", browser, browser),
			bk.Env("PUPPETEER_SKIP_CHROMIUM_DOWNLOAD", "true"), // Don't download browser, we use "download-puppeteer-browser" script instead
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

func frontendPuppeteerAndStorybook(autoAcceptChanges bool) Operation {
	return func(pipeline *bk.Pipeline) {
		// Client integration tests
		pipeline.AddStep(":puppeteer::electric_plug: Puppeteer tests",
			bk.Env("PUPPETEER_SKIP_CHROMIUM_DOWNLOAD", "true"), // Don't download browser, we use "download-puppeteer-browser" script instead
			bk.Env("ENTERPRISE", "1"),
			bk.Env("PERCY_ON", "true"),
			bk.Cmd("COVERAGE_INSTRUMENT=true dev/ci/yarn-build.sh client/web"),
			bk.Cmd("echo \"--- Install puppeteer\" && yarn --cwd client/shared run download-puppeteer-browser"),
			bk.Cmd("echo \"--- Run integration test suite\" && yarn percy exec -- yarn run cover-integration"),
			bk.Cmd("echo \"--- Process NYC report\" && yarn nyc report -r json"),
			bk.Cmd("echo \"--- Upload coverage report\" && dev/ci/codecov.sh -c -F typescript -F integration"),
			bk.ArtifactPaths("./puppeteer/*.png"))

		// Upload storybook to Chromatic
		chromaticCommand := "yarn chromatic --exit-zero-on-changes --exit-once-uploaded"
		if autoAcceptChanges {
			chromaticCommand += " --auto-accept-changes"
		}
		pipeline.AddStep(":chromatic: Upload storybook to Chromatic",
			bk.AutomaticRetry(5),
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
func addBackendIntegrationTests(pipeline *bk.Pipeline) {
	pipeline.AddStep(":chains: Backend integration tests",
		bk.Cmd("pushd enterprise"),
		bk.Cmd("./cmd/server/pre-build.sh"),
		bk.Cmd("./cmd/server/build.sh"),
		bk.Cmd("popd"),
		bk.Cmd("./dev/ci/backend-integration.sh"),
		bk.Cmd(`docker image rm -f "$IMAGE"`),
	)
}

func addBrowserExtensionE2ESteps(pipeline *bk.Pipeline) {
	for _, browser := range []string{"chrome"} {
		// Run e2e tests
		pipeline.AddStep(fmt.Sprintf(":%s: E2E for %s extension", browser, browser),
			bk.Env("PUPPETEER_SKIP_CHROMIUM_DOWNLOAD", "true"), // Don't download browser, we use "download-puppeteer-browser" script instead
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

// Trigger the async pipeline to run.
func triggerAsync(buildOptions bk.BuildOptions) Operation {
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

func triggerE2EandQA(opts e2eAndQAOptions) Operation {
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
		pipeline.AddTrigger(":chromium: Trigger E2E",
			bk.Trigger("sourcegraph-e2e"),
			bk.Async(opts.async),
			bk.Build(customOptions),
		)
		pipeline.AddTrigger(":chromium: Trigger QA",
			bk.Trigger("qa"),
			bk.Async(opts.async),
			bk.Build(customOptions),
		)
		pipeline.AddTrigger(":chromium: Trigger Code Intel QA",
			bk.Trigger("code-intel-qa"),
			bk.Async(opts.async),
			bk.Build(customOptions),
		)
	}
}

// Build a candidate docker image that will re-tagged with the final
// tags once the e2e tests pass.
func buildCandidateDockerImage(app, version, tag string) Operation {
	return func(pipeline *bk.Pipeline) {
		image := strings.ReplaceAll(app, "/", "-")
		localImage := "sourcegraph/" + image + ":" + version

		cmds := []bk.StepOpt{
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

		devImage := fmt.Sprintf("%s/%s", images.SourcegraphDockerDevRegistry, image)
		cmds = append(cmds,
			// Retag the local image for dev registry
			bk.Cmd(fmt.Sprintf("docker tag %s %s:%s", localImage, devImage, tag)),
			// Publish tagged image
			bk.Cmd(fmt.Sprintf("docker push %s:%s", devImage, tag)),
		)

		pipeline.AddStep(fmt.Sprintf(":docker: :construction: %s", app), cmds...)
	}
}

// Tag and push final Docker image for the service defined by `app`
// after the e2e tests pass.
//
// It requires Config as an argument because published images require a lot of metadata.
func publishFinalDockerImage(c Config, app string, insiders bool) Operation {
	return func(pipeline *bk.Pipeline) {
		image := strings.ReplaceAll(app, "/", "-")
		devImage := fmt.Sprintf("%s/%s", images.SourcegraphDockerDevRegistry, image)
		publishImage := fmt.Sprintf("%s/%s", images.SourcegraphDockerPublishRegistry, image)

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

		pipeline.AddStep(fmt.Sprintf(":docker: :white_check_mark: %s", app), bk.Cmd(cmd))
	}
}

// ~6m (building executor base VM)
func buildExecutor(timestamp time.Time, version string) Operation {
	return func(pipeline *bk.Pipeline) {
		cmds := []bk.StepOpt{
			bk.Cmd(`echo "Building executor cloud image..."`),
			bk.Env("VERSION", version),
			bk.Env("BUILD_TIMESTAMP", strconv.Itoa(int(timestamp.UTC().Unix()))),
			bk.Cmd("./enterprise/cmd/executor/build.sh"),
		}

		pipeline.AddStep(":packer: :construction: executor image", cmds...)
	}
}

func publishExecutor(timestamp time.Time, version string) Operation {
	return func(pipeline *bk.Pipeline) {
		cmds := []bk.StepOpt{
			bk.Cmd(`echo "Releasing executor cloud image..."`),
			bk.Env("VERSION", version),
			bk.Env("BUILD_TIMESTAMP", strconv.Itoa(int(timestamp.UTC().Unix()))),
			bk.Cmd("./enterprise/cmd/executor/release.sh"),
		}

		pipeline.AddStep(":packer: :white_check_mark: executor image", cmds...)
	}
}
