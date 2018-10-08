package main

import (
	"fmt"
	"go/build"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	bk "github.com/sourcegraph/sourcegraph/pkg/buildkite"
)

func init() {
	bk.Plugins["gopath-checkout#v1.0.1"] = map[string]string{
		"import": "github.com/sourcegraph/enterprise",
	}
}

func pkgs() []string {
	pkgs := []string{"cmd/xlang-go/internal/server"} // put slow tests first
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}
		if path == "." || !info.IsDir() {
			return nil
		}
		switch path {
		case ".git", "dev":
			return filepath.SkipDir
		}
		if filepath.Base(path) == "vendor" {
			return filepath.SkipDir
		}

		if path == "xlang" {
			return nil // already first entry
		}

		pkg, err := build.Import("github.com/sourcegraph/enterprise/"+path, "", 0)
		if err != nil {
			if _, ok := err.(*build.NoGoError); ok {
				return nil
			}
			panic(err)
		}

		if len(pkg.TestGoFiles) != 0 || len(pkg.XTestGoFiles) != 0 {
			pkgs = append(pkgs, path)
		}

		return nil
	})
	if err != nil {
		panic(err)
	}
	return pkgs
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
		// The Git branch "v1.2.3" should map to the Docker image "1.2.3" (without v prefix).
		version = strings.TrimPrefix(version, "v")
	}

	bk.OnEveryStepOpts = append(bk.OnEveryStepOpts,
		bk.Env("GO111MODULE", "on"),
		bk.Cmd("./dev/ci/ensure-go-mod.sh"),
	)

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

	pipeline.AddStep(":webpack:",
		bk.Env("PUPPETEER_SKIP_CHROMIUM_DOWNLOAD", "true"),
		bk.Env("FORCE_COLOR", "1"),
		bk.Cmd("yarn --frozen-lockfile"),
		bk.Cmd("yarn list --pattern @sourcegraph/webapp"),
		bk.Cmd("yarn run browserslist"),
		bk.Cmd("NODE_ENV=production yarn run build --color"),
		bk.Cmd("GITHUB_TOKEN= yarn run bundlesize"))

	// There are no tests yet
	// pipeline.AddStep(":mocha:",
	// 	bk.Env("PUPPETEER_SKIP_CHROMIUM_DOWNLOAD", "true"),
	// 	bk.Env("FORCE_COLOR", "1"),
	// 	bk.Cmd("yarn --frozen-lockfile"),
	//  bk.Cmd("yarn list --pattern @sourcegraph/webapp"),
	// 	bk.Cmd("yarn run cover"),
	// 	bk.Cmd("node_modules/.bin/nyc report -r json"),
	// 	bk.ArtifactPaths("coverage/coverage-final.json"))

	// addDockerImageStep adds a build step for a given app. If the app name has prefix
	// "enterprise-", that signals it is part of the enterprise distribution.
	addDockerImageStep := func(app string, insiders bool) {
		isEnterprise := strings.HasPrefix(app, "enterprise:")
		appBase := strings.TrimPrefix(app, "enterprise:")

		var cmdDir string
		var pkgPath string
		if isEnterprise {
			cmdDir = "cmd/" + appBase
			pkgPath = "github.com/sourcegraph/enterprise/cmd/" + appBase
		} else {
			log.Fatal("Only enterprise builds allowed")
		}

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
				bk.Cmd("go build github.com/sourcegraph/godockerize"),
				bk.Cmd(fmt.Sprintf("./godockerize build -t %s:%s --go-build-flags='-ldflags' --go-build-flags='-X github.com/sourcegraph/sourcegraph/pkg/version.version=%s' --env VERSION=%s %s", image, version, version, version, pkgPath)),
			)
		}
		cmds = append(cmds,
			bk.Cmd(fmt.Sprintf("docker push %s:%s", image, version)),
		)
		if insiders {
			cmds = append(cmds,
				bk.Cmd(fmt.Sprintf("docker tag %s:%s %s:insiders", image, version, image)),
				bk.Cmd(fmt.Sprintf("docker push %s:insiders", image)),
			)
		}
		if taggedRelease {
			cmds = append(cmds,
				bk.Cmd(fmt.Sprintf("docker tag %s:%s %s:%s", image, version, image, version)),
				bk.Cmd(fmt.Sprintf("docker push %s:%s", image, version)),
			)
		}
		pipeline.AddStep(":docker:", cmds...)
	}

	pipeline.AddStep(":go:", bk.Cmd("go install ./cmd/..."))
	pipeline.AddStep(":go:",
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

	for _, path := range pkgs() {
		coverageFile := path + "/coverage.txt"
		stepOpts := []bk.StepOpt{
			bk.Cmd("go test ./" + path + " -v -race -i"),
			bk.Cmd("go test ./" + path + " -v -race -coverprofile=" + coverageFile + " -covermode=atomic -coverpkg=github.com/sourcegraph/enterprise/..."),
			bk.ArtifactPaths(coverageFile),
		}
		pipeline.AddStep(":go:", stepOpts...)
	}

	pipeline.AddWait()

	pipeline.AddStep(":codecov:",
		bk.Cmd("buildkite-agent artifact download '*/coverage.txt' . || true"), // ignore error when no report exists
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
			bk.Cmd("yarn list --pattern @sourcegraph/webapp"),
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
		latest := branch == "master"
		allDockerImages := []string{
			"enterprise:frontend",
			"enterprise:server",
			"xlang-go",
		}

		for _, dockerImage := range allDockerImages {
			addDockerImageStep(dockerImage, latest)
		}
		pipeline.AddWait()

	case branch == "master":
		addDockerImageStep("enterprise:frontend", true)
		addDockerImageStep("enterprise:server", true)
		pipeline.AddWait()
		addDeploySteps()

	case strings.HasPrefix(branch, "master-dry-run/"): // replicates `master` build but does not deploy
		addDockerImageStep("enterprise:frontend", true)
		addDockerImageStep("enterprise:server", true)
		pipeline.AddWait()

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
