package ci

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	bk "github.com/sourcegraph/sourcegraph/internal/buildkite"
)

var allDockerImages = []string{
	"frontend",
	"github-proxy",
	"gitserver",
	"management-console",
	"query-runner",
	"replacer",
	"repo-updater",
	"searcher",
	"server",
	"symbols",
	"lsif-server",
}

// Verifies the docs formatting and builds the `docsite` command.
func addDocs(pipeline *bk.Pipeline) {
	pipeline.AddStep(":memo:",
		bk.Cmd("./dev/ci/yarn-run.sh prettier-check"),
		bk.Cmd("./dev/check/docsite.sh"))
}

// Adds the static check test step.
func addCheck(pipeline *bk.Pipeline) {
	pipeline.AddStep(":white_check_mark:", bk.Cmd("./dev/check/all.sh"))
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
	// - tslint 45s
	// - prettier 29s
	// - stylelint 7s
	// - graphql-lint 1s
	pipeline.AddStep(":eslint:",
		bk.Cmd("dev/ci/yarn-run.sh build-ts all:eslint")) // eslint depends on build-ts
	pipeline.AddStep(":lipstick: :lint-roller: :stylelint: :graphql:",
		bk.Cmd("dev/ci/yarn-run.sh prettier-check all:tslint all:stylelint graphql-lint"))
}

// Adds steps for the OSS and Enterprise web app builds. Runs the web app tests.
func addWebApp(pipeline *bk.Pipeline) {
	// Webapp build
	pipeline.AddStep(":webpack::globe_with_meridians:",
		bk.Cmd("dev/ci/yarn-build.sh web"),
		bk.Env("NODE_ENV", "production"),
		bk.Env("ENTERPRISE", "0"))

	// Webapp enterprise build
	pipeline.AddStep(":webpack::globe_with_meridians::moneybag:",
		bk.Cmd("dev/ci/yarn-build.sh web"),
		bk.Env("NODE_ENV", "production"),
		bk.Env("ENTERPRISE", "1"))

	// Webapp tests
	pipeline.AddStep(":jest::globe_with_meridians:",
		bk.Cmd("dev/ci/yarn-test.sh web"),
		bk.ArtifactPaths("web/coverage/coverage-final.json"))
}

// Builds and tests the browser extension.
func addBrowserExt(pipeline *bk.Pipeline) {
	// Browser extension build
	pipeline.AddStep(":webpack::chrome:",
		bk.Cmd("dev/ci/yarn-build.sh browser"))

	// Browser extension tests
	pipeline.AddStep(":jest::chrome:",
		bk.Cmd("dev/ci/yarn-test.sh browser"),
		bk.ArtifactPaths("browser/coverage/coverage-final.json"))
}

// Tests the LSIF server.
func addLSIFServer(pipeline *bk.Pipeline) {
	pipeline.AddStep(":jest:",
		bk.Cmd("dev/ci/yarn-test-separate.sh lsif"),
		bk.ArtifactPaths("lsif/coverage/coverage-final.json"))
}

// Adds the shared frontend tests (shared between the web app and browser extension).
func addSharedTests(pipeline *bk.Pipeline) {
	// Shared tests
	pipeline.AddStep(":jest:",
		bk.Cmd("dev/ci/yarn-test.sh shared"),
		bk.ArtifactPaths("shared/coverage/coverage-final.json"))

	// Storybook
	pipeline.AddStep(":storybook:", bk.Cmd("dev/ci/yarn-run.sh storybook:smoke-test"))
}

// Adds PostgreSQL backcompat tests.
func addPostgresBackcompat(pipeline *bk.Pipeline) {
	pipeline.AddStep(":postgres:",
		bk.Cmd("./dev/ci/ci-db-backcompat.sh"))
}

// Adds the Go test step.
func addGoTests(pipeline *bk.Pipeline) {
	pipeline.AddStep(":go:",
		bk.Cmd("./cmd/symbols/build.sh buildLibsqlite3Pcre"), // for symbols tests
		bk.Cmd("./dev/comby-install-or-upgrade.sh"),          // for searcher and replacer tests
		bk.Cmd("go test -timeout 4m -coverprofile=coverage.txt -covermode=atomic -race ./..."),
		bk.ArtifactPaths("coverage.txt"))
}

// Builds the OSS and Enterprise Go commands.
func addGoBuild(pipeline *bk.Pipeline) {
	pipeline.AddStep(":go:",
		bk.Cmd("go generate ./..."),
		bk.Cmd("go install -tags dist ./cmd/... ./enterprise/cmd/..."),
	)
}

// Lints the Dockerfiles.
func addDockerfileLint(pipeline *bk.Pipeline) {
	pipeline.AddStep(":docker:",
		bk.Cmd("curl -sL -o hadolint \"https://github.com/hadolint/hadolint/releases/download/v1.15.0/hadolint-$(uname -s)-$(uname -m)\" && chmod 700 hadolint"),
		bk.Cmd("git ls-files | grep Dockerfile | xargs ./hadolint"))
}

// Code coverage.
func addCodeCov(pipeline *bk.Pipeline) {
	pipeline.AddStep(":codecov:",
		bk.Cmd("buildkite-agent artifact download 'coverage.txt' . || true"), // ignore error when no report exists
		bk.Cmd("buildkite-agent artifact download '*/coverage-final.json' . || true"),
		bk.Cmd("bash <(curl -s https://codecov.io/bash) -X gcov -X coveragepy -X xcode"))
}

// Release the browser extension.
func addBrowserExtensionReleaseSteps(pipeline *bk.Pipeline) {
	for _, browser := range []string{"chrome" /* , "firefox" */} {
		// Run e2e tests
		pipeline.AddStep(fmt.Sprintf(":%s:", browser),
			bk.Env("PUPPETEER_SKIP_CHROMIUM_DOWNLOAD", ""),
			bk.Env("EXTENSION_PERMISSIONS_ALL_URLS", "true"),
			bk.Env("E2E_BROWSER", browser),
			bk.Cmd("yarn --frozen-lockfile --network-timeout 60000"),
			bk.Cmd("pushd browser"),
			bk.Cmd("yarn -s run build"),
			bk.Cmd("yarn -s jest --runInBand ./e2e/github.test"),
			bk.Cmd("popd"),
			bk.ArtifactPaths("./puppeteer/*.png"))
	}

	pipeline.AddWait()

	// Release to the Chrome Webstore
	pipeline.AddStep(":rocket::chrome:",
		bk.Cmd("yarn --frozen-lockfile --network-timeout 60000"),
		bk.Cmd("pushd browser"),
		bk.Cmd("yarn -s run build"),
		bk.Cmd("yarn release:chrome"),
		bk.Cmd("popd"))

	// Build and self sign the FF extension and upload it to ...
	pipeline.AddStep(":rocket::firefox:",
		bk.Cmd("yarn --frozen-lockfile --network-timeout 60000"),
		bk.Cmd("pushd browser"),
		bk.Cmd("yarn release:ff"),
		bk.Cmd("popd"))

	// Release to npm
	pipeline.AddStep(":rocket::npm:",
		bk.Cmd("yarn --frozen-lockfile --network-timeout 60000"),
		bk.Cmd("pushd browser"),
		bk.Cmd("yarn -s run build"),
		bk.Cmd("yarn release:npm"),
		bk.Cmd("popd"))
}

// Adds a Buildkite pipeline "Wait".
func wait(pipeline *bk.Pipeline) {
	pipeline.AddWait()
}

func triggerE2E(c Config) func(*bk.Pipeline) {
	// hardFail if we publish docker images
	hardFail := c.branch == "master" || c.isRenovateBranch || c.taggedRelease || c.isBextReleaseBranch || c.patch

	return func(pipeline *bk.Pipeline) {
		pipeline.AddTrigger(":chromium:",
			bk.Trigger("sourcegraph-e2e"),
			bk.Async(!hardFail),
			bk.Build(bk.BuildOptions{
				Message: fmt.Sprintf("Test"),
				Commit:  c.commit,
				Branch:  c.branch,
				Env: copyEnv(
					"BUILDKITE_PULL_REQUEST",
					"BUILDKITE_PULL_REQUEST_BASE_BRANCH",
					"BUILDKITE_PULL_REQUEST_REPO"),
			}))
	}
}

func copyEnv(keys ...string) map[string]string {
	m := map[string]string{}
	for _, k := range keys {
		if v, ok := os.LookupEnv(k); ok {
			m[k] = v
		}
	}
	return m
}

// Build all relevant Docker images for Sourcegraph, given the current CI case (e.g., "tagged
// release", "release branch", "master branch", etc.)
func addDockerImages(c Config) func(*bk.Pipeline) {
	return func(pipeline *bk.Pipeline) {
		switch {
		case c.taggedRelease:
			for _, dockerImage := range allDockerImages {
				addDockerImage(c, dockerImage, false)(pipeline)
			}
			pipeline.AddWait()
		case c.releaseBranch:
			addDockerImage(c, "server", false)(pipeline)
			pipeline.AddWait()
		case strings.HasPrefix(c.branch, "master-dry-run/"): // replicates `master` build but does not deploy
			fallthrough
		case c.branch == "master":
			for _, dockerImage := range allDockerImages {
				addDockerImage(c, dockerImage, true)(pipeline)
			}
			pipeline.AddWait()

		case strings.HasPrefix(c.branch, "docker-images-patch/"):
			addDockerImage(c, c.branch[20:], false)(pipeline)
			pipeline.AddWait()
		}
	}
}

// Build Docker image for the service defined by `app`. The Sourcegraph Server Docker image is
// special-cased, because it is built in another step as a candidate image, so we just need to tag
// the candidate instead of rebuilding the image.
func addDockerImage(c Config, app string, insiders bool) func(*bk.Pipeline) {
	return func(pipeline *bk.Pipeline) {
		cmds := []bk.StepOpt{
			bk.Cmd(fmt.Sprintf(`echo "Building %s..."`, app)),
			bk.Env("DOCKER_BUILDKIT", "1"),
		}

		cmdDir := func() string {
			cmdDirByApp := map[string]string{
				"lsif-server": "lsif",
			}
			if cmdDir, ok := cmdDirByApp[app]; ok {
				return cmdDir
			}
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

		baseImage := "sourcegraph/" + app

		getBuildScript := func() string {
			buildScriptByApp := map[string]string{
				"symbols": "env BUILD_TYPE=dist ./cmd/symbols/build.sh buildSymbolsDockerImage",

				// The server image was built prior to e2e tests in a previous step.
				"server": fmt.Sprintf("docker tag %s:%s_candidate %s:%s", baseImage, c.version, baseImage, c.version),
			}
			if buildScript, ok := buildScriptByApp[app]; ok {
				return buildScript
			}
			return cmdDir + "/build.sh"
		}

		cmds = append(cmds,
			bk.Env("IMAGE", baseImage+":"+c.version),
			bk.Env("VERSION", c.version),
			bk.Cmd(getBuildScript()),
		)

		cmds = append(cmds, bk.Cmd("yes | gcloud auth configure-docker"))

		dockerHubImage := fmt.Sprintf("index.docker.io/%s", baseImage)
		gcrImage := fmt.Sprintf("us.gcr.io/sourcegraph-dev/%s", strings.TrimPrefix(baseImage, "sourcegraph/"))

		for _, image := range []string{dockerHubImage, gcrImage} {
			if app != "server" || c.taggedRelease || c.patch || c.patchNoTest {
				cmds = append(cmds,
					bk.Cmd(fmt.Sprintf("docker tag %s:%s %s:%s", baseImage, c.version, image, c.version)),
					bk.Cmd(fmt.Sprintf("docker push %s:%s", image, c.version)),
				)
			}

			if app == "server" && c.releaseBranch {
				cmds = append(cmds,
					bk.Cmd(fmt.Sprintf("docker tag %s:%s %s:%s-insiders", baseImage, c.version, image, c.branch)),
					bk.Cmd(fmt.Sprintf("docker push %s:%s-insiders", image, c.branch)),
				)
			}

			if insiders {
				cmds = append(cmds,
					bk.Cmd(fmt.Sprintf("docker tag %s:%s %s:insiders", baseImage, c.version, image)),
					bk.Cmd(fmt.Sprintf("docker push %s:insiders", image)),
				)
			}
		}

		pipeline.AddStep(":docker:", cmds...)
	}
}
