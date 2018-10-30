package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	bk "github.com/sourcegraph/sourcegraph/pkg/buildkite"
)

func init() {
	bk.Plugins["gopath-checkout#v1.0.1"] = map[string]string{
		"import": "github.com/sourcegraph/sourcegraph",
	}
}

func main() {
	pipeline := &bk.Pipeline{}

	branch := os.Getenv("BUILDKITE_BRANCH")
	version := os.Getenv("BUILDKITE_TAG")
	taggedRelease := true // true if this is a semver tagged release
	if !strings.HasPrefix(version, "v") {
		taggedRelease = false
		commit := os.Getenv("BUILDKITE_COMMIT")
		if commit == "" {
			commit = "1234567890123456789012345678901234567890" // for testing
		}
		buildNum, _ := strconv.Atoi(os.Getenv("BUILDKITE_BUILD_NUMBER"))
		version = fmt.Sprintf("%05d_%s_%.7s", buildNum, time.Now().Format("2006-01-02"), commit)
	} else {
		// The Git tag "v1.2.3" should map to the Docker image "1.2.3" (without v prefix).
		version = strings.TrimPrefix(version, "v")
	}
	releaseBranch := regexp.MustCompile(`^[0-9]+\.[0-9]+$`).MatchString(branch)

	bk.OnEveryStepOpts = append(bk.OnEveryStepOpts,
		bk.Env("GO111MODULE", "on"),
		bk.Env("ENTERPRISE", "1"),
		bk.Cmd("pushd enterprise"),
	)

	pipeline.AddStep(":white_check_mark:",
		bk.Cmd("popd"),
		bk.Cmd("./dev/check/all.sh"),
		bk.Cmd("pushd enterprise"),
		bk.Cmd("./dev/check/all.sh"),
		bk.Cmd("popd"),
	)

	pipeline.AddStep(":lipstick:",
		bk.Cmd("popd"),
		bk.Cmd("yarn --frozen-lockfile"),
		bk.Cmd("yarn run prettier"))

	pipeline.AddStep(":typescript:",
		bk.Cmd("popd"),
		bk.Env("PUPPETEER_SKIP_CHROMIUM_DOWNLOAD", "true"),
		bk.Env("FORCE_COLOR", "1"),
		bk.Cmd("yarn --frozen-lockfile"),
		bk.Cmd("yarn run tslint"))

	pipeline.AddStep(":stylelint:",
		bk.Cmd("popd"),
		bk.Env("PUPPETEER_SKIP_CHROMIUM_DOWNLOAD", "true"),
		bk.Env("FORCE_COLOR", "1"),
		bk.Cmd("yarn --frozen-lockfile"),
		bk.Cmd("yarn run stylelint --quiet"))

	pipeline.AddStep(":webpack:",
		bk.Cmd("popd"),
		bk.Env("PUPPETEER_SKIP_CHROMIUM_DOWNLOAD", "true"),
		bk.Env("FORCE_COLOR", "1"),
		bk.Cmd("yarn --frozen-lockfile"),
		bk.Cmd("yarn run browserslist"),
		bk.Cmd("NODE_ENV=production yarn run build --color"),
		bk.Cmd("GITHUB_TOKEN= yarn run bundlesize"))

	// There are no tests yet
	// pipeline.AddStep(":mocha:",
	// 	bk.Env("PUPPETEER_SKIP_CHROMIUM_DOWNLOAD", "true"),
	// 	bk.Env("FORCE_COLOR", "1"),
	// 	bk.Cmd("yarn --frozen-lockfile"),
	// 	bk.Cmd("yarn run cover"),
	// 	bk.Cmd("node_modules/.bin/nyc report -r json"),
	// 	bk.ArtifactPaths("coverage/coverage-final.json"))

	// addDockerImageStep adds a build step for a given app.
	// If the app is not in the cmd directory, it is assumed to be from the open source repo.
	addDockerImageStep := func(app string, insiders bool) {
		cmdDir := "cmd/" + app
		var pkgPath string
		if _, err := os.Stat(cmdDir); err != nil {
			fmt.Fprintf(os.Stderr, "github.com/sourcegraph/sourcegraph/enterprise/cmd/%s does not exist so building github.com/sourcegraph/sourcegraph/cmd/%s instead\n", app, app)
			pkgPath = "github.com/sourcegraph/sourcegraph/cmd/" + app
		} else {
			pkgPath = "github.com/sourcegraph/sourcegraph/enterprise/cmd/" + app
		}

		cmds := []bk.StepOpt{
			bk.Cmd(fmt.Sprintf(`echo "Building %s..."`, app)),
		}

		preBuildScript := cmdDir + "/pre-build.sh"
		if _, err := os.Stat(preBuildScript); err == nil {
			cmds = append(cmds, bk.Cmd(preBuildScript))
		}

		image := "sourcegraph/" + app
		buildScript := cmdDir + "/build.sh"
		if _, err := os.Stat(buildScript); err == nil {
			cmds = append(cmds,
				bk.Env("IMAGE", image+":"+version),
				bk.Env("VERSION", version),
				bk.Cmd(buildScript),
			)
		} else {
			cmds = append(cmds,
				bk.Cmd("go build github.com/sourcegraph/godockerize"),
				bk.Cmd(fmt.Sprintf("./godockerize build -t %s:%s --go-build-flags='-ldflags' --go-build-flags='-X github.com/sourcegraph/sourcegraph/pkg/version.version=%s' --env VERSION=%s %s", image, version, version, version, pkgPath)),
			)
		}

		if app != "server" || taggedRelease {
			cmds = append(cmds,
				bk.Cmd(fmt.Sprintf("docker push %s:%s", image, version)),
			)
		}

		if app == "server" && releaseBranch {
			cmds = append(cmds,
				bk.Cmd(fmt.Sprintf("docker tag %s:%s %s:%s-insiders", image, version, image, branch)),
				bk.Cmd(fmt.Sprintf("docker push %s:%s-insiders", image, branch)),
			)
		}

		if insiders {
			cmds = append(cmds,
				bk.Cmd(fmt.Sprintf("docker tag %s:%s %s:insiders", image, version, image)),
				bk.Cmd(fmt.Sprintf("docker push %s:insiders", image)),
			)
		}
		pipeline.AddStep(":docker:", cmds...)
	}

	pipeline.AddStep(":go:", bk.Cmd("go install ./cmd/..."))
	pipeline.AddStep(":go:",
		bk.Cmd("pushd .."),
		bk.Cmd("go generate ./cmd/..."),
		bk.Cmd("popd"),
		bk.Cmd("go generate ./cmd/..."),
		bk.Cmd("go install -tags dist ./cmd/..."),
	)

	if strings.HasPrefix(branch, "docker-images-patch-notest/") {
		version = version + "_patch"
		addDockerImageStep(branch[27:], false)
		_, err := pipeline.WriteTo(os.Stdout)
		if err != nil {
			panic(err)
		}
		return
	}

	pipeline.AddStep(":go:",
		bk.Cmd("go test -coverprofile=coverage.txt -covermode=atomic -race ./..."),
		bk.ArtifactPaths("coverage.txt"))

	pipeline.AddWait()

	pipeline.AddStep(":codecov:",
		bk.Cmd("buildkite-agent artifact download 'coverage.txt' . || true"), // ignore error when no report exists
		bk.Cmd("buildkite-agent artifact download '*/coverage-final.json' . || true"),
		bk.Cmd("bash <(curl -s https://codecov.io/bash) -X gcov -X coveragepy -X xcode"))

	fetchClusterCredentials := func(name, zone, project string) bk.StepOpt {
		return bk.Cmd(fmt.Sprintf("gcloud container clusters get-credentials %s --zone %s --project %s", name, zone, project))
	}

	addDeploySteps := func() {
		// Deploy to dogfood
		pipeline.AddStep(":dog:",
			// Protect against concurrent/out-of-order deploys
			bk.ConcurrencyGroup("deploy"),
			bk.Concurrency(1),
			bk.Env("VERSION", version),
			bk.Env("CONTEXT", "gke_sourcegraph-dev_us-central1-a_dogfood-cluster-7"),
			bk.Env("NAMESPACE", "default"),
			fetchClusterCredentials("dogfood-cluster-7", "us-central1-a", "sourcegraph-dev"),
			bk.Cmd("./dev/ci/deploy-dogfood.sh"))
		pipeline.AddWait()

		// Run e2e tests against dogfood
		pipeline.AddStep(":chromium:",
			// Protect against deploys while tests are running
			bk.ConcurrencyGroup("deploy"),
			bk.Concurrency(1),
			bk.Env("SOURCEGRAPH_BASE_URL", "https://sourcegraph.sgdev.org"),
			bk.Env("FORCE_COLOR", "1"),
			bk.Cmd("yarn --frozen-lockfile"),
			bk.Cmd("yarn run test-e2e-sgdev --retries 5"),
			bk.ArtifactPaths("./puppeteer/*.png"))
		pipeline.AddWait()

		// Deploy to prod
		pipeline.AddStep(":rocket:",
			bk.Env("VERSION", version),
			bk.Cmd("./dev/ci/deploy-prod.sh"))
	}

	switch {
	case taggedRelease:
		allDockerImages := []string{
			"frontend",
			"github-proxy",
			"gitserver",
			"query-runner",
			"repo-updater",
			"searcher",
			"server",
			"symbols",
		}

		for _, dockerImage := range allDockerImages {
			addDockerImageStep(dockerImage, false)
		}
		pipeline.AddWait()

	case releaseBranch:
		addDockerImageStep("server", false)
		pipeline.AddWait()

	case branch == "master":
		addDockerImageStep("frontend", true)
		addDockerImageStep("server", true)
		pipeline.AddWait()
		addDeploySteps()

	case strings.HasPrefix(branch, "master-dry-run/"): // replicates `master` build but does not deploy
		addDockerImageStep("frontend", true)
		addDockerImageStep("server", true)
		pipeline.AddWait()

	case strings.HasPrefix(branch, "docker-images-patch/"):
		version = version + "_patch"
		addDockerImageStep(branch[20:], false)

	case strings.HasPrefix(branch, "docker-images/"):
		addDockerImageStep(branch[14:], true)
		pipeline.AddWait()
		// Only deploy images that aren't auto deployed from master.
		if branch != "docker-images/server" && branch != "docker-images/frontend" {
			addDeploySteps()
		}
	}

	_, err := pipeline.WriteTo(os.Stdout)
	if err != nil {
		panic(err)
	}
}
