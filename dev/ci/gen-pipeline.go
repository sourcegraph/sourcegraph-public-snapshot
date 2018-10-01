// gen-pipeline.go generates a Buildkite YAML file that tests the entire
// Sourcegraph application and writes it to stdout.
//
// The main reason to generate this file on demand is to ensure all Go packages
// are tested. Go packages are tested in separate containers - instead of all
// at once - so that Buildkite can schedule them to run across more machines
// on Sourcegraph's build farm.
//
// This script also generates a different file for deploy/tag builds vs. PR
// builds.
//
// See dev/ci/init-pipeline.yml for an example of where this script is invoked.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
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
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr,
			`Generate a Buildkite YAML file that tests the entire Sourcegraph application and write it to stdout.
`)
		flag.PrintDefaults()
	}
	flag.Parse()
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
		// The Git branch "v1.2.3" should map to the Docker image "1.2.3" (without v prefix).
		version = strings.TrimPrefix(version, "v")
	}

	// addDockerImageStep adds a build step for a given app.
	addDockerImageStep := func(app string, insiders bool) {
		appBase := app
		cmdDir := "./cmd/" + appBase
		pkgPath := "github.com/sourcegraph/sourcegraph/cmd/" + appBase

		if _, err := os.Stat(cmdDir); err != nil {
			fmt.Fprintln(os.Stderr, "app does not exist: "+app)
			os.Exit(1)
		}
		cmds := []bk.StepOpt{
			bk.Cmd(fmt.Sprintf(`echo "Building %s..."`, app)),
		}

		preBuildScript := cmdDir + "/pre-build.sh"
		if _, err := os.Stat(preBuildScript); err == nil {
			cmds = append(cmds, bk.Cmd(preBuildScript))
		}

		image := "sourcegraph/" + appBase
		buildScript := cmdDir + "/build.sh"
		if _, err := os.Stat(buildScript); err == nil {
			cmds = append(cmds,
				bk.Env("IMAGE", image+":"+version),
				bk.Env("VERSION", version),
				bk.Cmd(buildScript),
			)
		} else {
			cmds = append(cmds,
				bk.Cmd("go build github.com/sourcegraph/sourcegraph/vendor/github.com/sourcegraph/godockerize"),
				bk.Cmd(fmt.Sprintf("./godockerize build -t %s:%s --go-build-flags='-ldflags' --go-build-flags='-X github.com/sourcegraph/sourcegraph/pkg/version.version=%s' --env VERSION=%s %s", image, version, version, version, pkgPath)),
			)
		}
		cmds = append(cmds,
			bk.Cmd(fmt.Sprintf("docker push %s:%s", image, version)),
		)
		if insiders {
			tags := []string{"insiders"}

			if strings.HasPrefix(appBase, "xlang") {
				// The "latest" tag is needed for the automatic docker management logic.
				tags = append(tags, "latest")
			}

			for _, tag := range tags {
				cmds = append(cmds,
					bk.Cmd(fmt.Sprintf("docker tag %s:%s %s:%s", image, version, image, tag)),
					bk.Cmd(fmt.Sprintf("docker push %s:%s", image, tag)),
				)
			}
		}
		if taggedRelease {
			cmds = append(cmds,
				bk.Cmd(fmt.Sprintf("docker tag %s:%s %s:%s", image, version, image, version)),
				bk.Cmd(fmt.Sprintf("docker push %s:%s", image, version)),
			)
		}
		pipeline.AddStep(":docker:", cmds...)
	}

	if strings.HasPrefix(branch, "docker-images-patch-notest/") {
		version = version + "_patch"
		addDockerImageStep(branch[27:], false)
		_, err := pipeline.WriteTo(os.Stdout)
		if err != nil {
			panic(err)
		}
		return
	}

	pipeline.AddStep(":white_check_mark:",
		bk.Cmd("./dev/check/all.sh"))

	pipeline.AddStep(":lipstick:",
		bk.Cmd("yarn --frozen-lockfile"),
		bk.Cmd("yarn run prettier"))

	pipeline.AddStep(":typescript:",
		bk.Env("PUPPETEER_SKIP_CHROMIUM_DOWNLOAD", "true"),
		bk.Env("FORCE_COLOR", "1"),
		bk.Cmd("yarn --frozen-lockfile"),
		bk.Cmd("yarn run tslint"))

	pipeline.AddStep(":stylelint:",
		bk.Env("PUPPETEER_SKIP_CHROMIUM_DOWNLOAD", "true"),
		bk.Env("FORCE_COLOR", "1"),
		bk.Cmd("yarn --frozen-lockfile"),
		bk.Cmd("yarn run stylelint --quiet"))

	pipeline.AddStep(":graphql:",
		bk.Cmd("yarn --frozen-lockfile"),
		bk.Cmd("yarn run graphql-lint"))

	pipeline.AddStep(":webpack:",
		bk.Env("PUPPETEER_SKIP_CHROMIUM_DOWNLOAD", "true"),
		bk.Env("FORCE_COLOR", "1"),
		bk.Cmd("yarn --frozen-lockfile"),
		bk.Cmd("yarn run browserslist"),
		bk.Cmd("NODE_ENV=production yarn run build --color"),
		bk.Cmd("GITHUB_TOKEN= yarn run bundlesize"))

	pipeline.AddStep(":mocha:",
		bk.Env("PUPPETEER_SKIP_CHROMIUM_DOWNLOAD", "true"),
		bk.Env("FORCE_COLOR", "1"),
		bk.Cmd("yarn --frozen-lockfile"),
		bk.Cmd("yarn run cover"),
		bk.Cmd("node_modules/.bin/nyc report -r json"),
		bk.ArtifactPaths("coverage/coverage-final.json"))

	pipeline.AddStep(":docker:",
		bk.Cmd("curl -sL -o hadolint \"https://github.com/hadolint/hadolint/releases/download/v1.6.5/hadolint-$(uname -s)-$(uname -m)\" && chmod 700 hadolint"),
		bk.Cmd("git ls-files | grep -v '^vendor/' | grep Dockerfile | xargs ./hadolint"))

	pipeline.AddStep(":postgres:",
		bk.Cmd("./dev/ci/ci-db-backcompat.sh"))

	pipeline.AddStep(":go:",
		bk.Cmd("./dev/ci/reset-test-db.sh || true"),
		bk.Cmd("go test -race ./..."))

	pipeline.AddWait()

	// TODO add back coverprofile to go
	//pipeline.AddStep(":codecov:",
	//	bk.Cmd("buildkite-agent artifact download '*/coverage.txt' . || true"), // ignore error when no report exists
	//	bk.Cmd("buildkite-agent artifact download '*/coverage-final.json' . || true"),
	//	bk.Cmd("bash <(curl -s https://codecov.io/bash) -X gcov -X coveragepy -X xcode -t TEMPORARILY_REDACTED"))

	if branch == "master" {
		// Publish @sourcegraph/webapp to npm
		pipeline.AddStep(":npm:",
			bk.Cmd("yarn --frozen-lockfile"),
			bk.Cmd("yarn run dist"),
			bk.Cmd("yarn run release"),
			bk.ConcurrencyGroup("webapp-publish"),
			bk.Concurrency(1),
		)
	}

	pipeline.AddWait()

	addDeploySteps := func() {
		// Only deploy pure-OSS images dogfood/prod. Images that contain some
		// private code (server/frontend) are already deployed by the
		// enterprise CI above.
		deploy := branch != "master"

		if deploy {
			// Deploy to dogfood
			pipeline.AddStep(":dog:",
				// Protect against concurrent/out-of-order deploys
				bk.ConcurrencyGroup("deploy"),
				bk.Concurrency(1),
				bk.Env("VERSION", version),
				bk.Env("CONTEXT", "gke_sourcegraph-dev_us-central1-a_dogfood-cluster-7"),
				bk.Env("NAMESPACE", "default"),
				bk.Cmd("./dev/ci/deploy-dogfood.sh"))
			pipeline.AddWait()
		}

		// Precalculate the webapp version here because metadata cannot be
		// forwarded in a trigger step, see https://github.com/buildkite/feedback/issues/403
		// This is safe because the publish step has a concurrency group so no
		// other publish can happen between now and its execution
		out, err := exec.Command("bash", "-c", "npx --quiet semver --increment patch $(npm info @sourcegraph/webapp version)").Output()
		if err != nil {
			panic(err)
		}
		webappVersion := strings.TrimSpace(string(out))

		// Trigger an enterprise repository master branch build.
		pipeline.AddTrigger(":satellite_antenna:",
			bk.Trigger("enterprise"),
			bk.ConcurrencyGroup("trigger-enterprise"),
			bk.Build(bk.BuildOptions{
				Message: ":satellite_antenna: OSS commit " + os.Getenv("BUILDKITE_COMMIT")[0:7] + ": " + os.Getenv("BUILDKITE_MESSAGE"),
				Env: map[string]string{
					"OSS_REPO_REVISION":  os.Getenv("BUILDKITE_COMMIT"),
					"OSS_WEBAPP_VERSION": webappVersion,
					"OSS_BUILD_URL":      os.Getenv("BUILDKITE_BUILD_URL"),
					"OSS_BUILD_NUMBER":   os.Getenv("BUILDKITE_BUILD_NUMBER"),
				},
				MetaData: map[string]interface{}{
					"oss-repo-revision":  os.Getenv("BUILDKITE_COMMIT"),
					"oss-webapp-version": webappVersion,
					"oss-build-url":      os.Getenv("BUILDKITE_BUILD_URL"),
					"oss-build-number":   os.Getenv("BUILDKITE_BUILD_NUMBER"),
				},
			}),
		)
		pipeline.AddWait()

		if deploy {
			// Deploy to prod
			pipeline.AddStep(":rocket:",
				bk.Env("VERSION", version),
				bk.Cmd("./dev/ci/deploy-prod.sh"))
		}
	}

	switch {
	case taggedRelease:
		latest := branch == "master"
		allDockerImages := []string{
			"github-proxy",
			"gitserver",
			"indexer",
			"lsp-proxy",
			"query-runner",
			"repo-updater",
			"searcher",
			"symbols",
		}

		for _, dockerImage := range allDockerImages {
			addDockerImageStep(dockerImage, latest)
		}
		pipeline.AddWait()

	case branch == "master":
		addDeploySteps()

	case strings.HasPrefix(branch, "docker-images-patch/"):
		version = version + "_patch"
		addDockerImageStep(branch[20:], false)

	case strings.HasPrefix(branch, "docker-images/"):
		addDockerImageStep(branch[14:], true)
		pipeline.AddWait()
		if branch != "docker-images/server" {
			addDeploySteps()
		}
	}

	_, err := pipeline.WriteTo(os.Stdout)
	if err != nil {
		panic(err)
	}
}
