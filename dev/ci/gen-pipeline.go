package main

import (
	"fmt"
	"go/build"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ghodss/yaml"
)

type Pipeline struct {
	Steps []interface{} `json:"steps"`
}

type Step struct {
	Label            string                 `json:"label"`
	Command          string                 `json:"command"`
	Env              map[string]string      `json:"env"`
	Plugins          map[string]interface{} `json:"plugins"`
	ArtifactPaths    string                 `json:"artifact_paths,omitempty"`
	ConcurrencyGroup string                 `json:"concurrency_group,omitempty"`
	Concurrency      int                    `json:"concurrency,omitempty"`
}

func (p *Pipeline) AddStep(label string, opts ...StepOpt) {
	step := &Step{
		Label:   label,
		Env:     make(map[string]string),
		Plugins: golangPlugin,
	}
	for _, opt := range opts {
		opt(step)
	}
	p.Steps = append(p.Steps, step)
}

func (p *Pipeline) WriteTo(w io.Writer) (int64, error) {
	output, err := yaml.Marshal(p)
	if err != nil {
		return 0, err
	}

	n, err := w.Write(output)
	return int64(n), err
}

type StepOpt func(step *Step)

func Cmd(command string) StepOpt {
	return func(step *Step) {
		step.Command = strings.TrimSpace(step.Command + "\n" + command)
	}
}

func ConcurrencyGroup(group string) StepOpt {
	return func(step *Step) {
		step.ConcurrencyGroup = group
	}
}

func Concurrency(limit int) StepOpt {
	return func(step *Step) {
		step.Concurrency = limit
	}
}

func Env(name, value string) StepOpt {
	return func(step *Step) {
		step.Env[name] = value
	}
}

func ArtifactPaths(paths string) StepOpt {
	return func(step *Step) {
		step.ArtifactPaths = paths
	}
}

func (p *Pipeline) AddWait() {
	p.Steps = append(p.Steps, "wait")
}

var golangPlugin = map[string]interface{}{
	"gopath-checkout#v1.0.1": map[string]string{
		"import": "github.com/sourcegraph/sourcegraph",
	},
}

func pkgs() []string {
	pkgs := []string{"xlang", "cmd/frontend/internal/db"} // put slow tests first
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}
		if path == "." || !info.IsDir() {
			return nil
		}
		switch path {
		case ".git", "dev", "ui":
			return filepath.SkipDir
		}
		if filepath.Base(path) == "vendor" {
			return filepath.SkipDir
		}

		if path == "xlang" || path == "cmd/frontend/internal/db" {
			return nil // already first entry
		}

		pkg, err := build.Import("github.com/sourcegraph/sourcegraph/"+path, "", 0)
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
	pipeline := &Pipeline{}

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

	// addDockerImageStep adds a build step for a given app. If the app name has prefix
	// "enterprise/", that signals it is part of the enterprise distribution.
	addDockerImageStep := func(app string, insiders bool) {
		isEnterprise := strings.HasPrefix(app, "enterprise/")
		appBase := strings.TrimPrefix(app, "enterprise/")

		var cmdDir string
		if isEnterprise {
			cmdDir = "./enterprise/cmd/" + appBase
		} else {
			cmdDir = "./cmd/" + appBase
		}

		if _, err := os.Stat(cmdDir); err != nil {
			fmt.Fprintln(os.Stderr, "app does not exist: "+app)
			os.Exit(1)
		}
		cmds := []StepOpt{
			Cmd(fmt.Sprintf(`echo "Building %s..."`, app)),
		}

		preBuildScript := cmdDir + "/pre-build.sh"
		if _, err := os.Stat(preBuildScript); err == nil {
			cmds = append(cmds, Cmd(preBuildScript))
		}

		image := "sourcegraph/" + app
		buildScript := cmdDir + "/build.sh"
		if _, err := os.Stat(buildScript); err == nil {
			cmds = append(cmds,
				Env("IMAGE", image+":"+version),
				Env("VERSION", version),
				Cmd(buildScript),
			)
		} else {
			cmds = append(cmds,
				Cmd("go build github.com/sourcegraph/sourcegraph/vendor/github.com/sourcegraph/godockerize"),
				Cmd(fmt.Sprintf("./godockerize build -t %s:%s --go-build-flags='-ldflags' --go-build-flags='-X github.com/sourcegraph/sourcegraph/pkg/version.version=%s' --env VERSION=%s github.com/sourcegraph/sourcegraph/cmd/%s", image, version, version, version, app)),
			)
		}
		cmds = append(cmds,
			Cmd(fmt.Sprintf("docker push %s:%s", image, version)),
		)
		if insiders {
			tags := []string{"insiders"}

			if strings.HasPrefix(appBase, "xlang") {
				// The "latest" tag is needed for the automatic docker management logic.
				tags = append(tags, "latest")
			}

			for _, tag := range tags {
				cmds = append(cmds,
					Cmd(fmt.Sprintf("docker tag %s:%s %s:%s", image, version, image, tag)),
					Cmd(fmt.Sprintf("docker push %s:%s", image, tag)),
				)
			}
		}
		if taggedRelease {
			cmds = append(cmds,
				Cmd(fmt.Sprintf("docker tag %s:%s %s:%s", image, version, image, version)),
				Cmd(fmt.Sprintf("docker push %s:%s", image, version)),
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
		Cmd("./dev/check/all.sh"))

	pipeline.AddStep(":lipstick:",
		Cmd("npm ci"),
		Cmd("npm run prettier"))

	pipeline.AddStep(":typescript:",
		Env("PUPPETEER_SKIP_CHROMIUM_DOWNLOAD", "true"),
		Env("FORCE_COLOR", "1"),
		Cmd("cd web"),
		Cmd("npm ci"),
		Cmd("npm run tslint"))

	pipeline.AddStep(":stylelint:",
		Env("PUPPETEER_SKIP_CHROMIUM_DOWNLOAD", "true"),
		Env("FORCE_COLOR", "1"),
		Cmd("cd web"),
		Cmd("npm ci"),
		Cmd("npm run stylelint -- --quiet"))

	pipeline.AddStep(":graphql:",
		Cmd("npm ci"),
		Cmd("npm run graphql-lint"))

	pipeline.AddStep(":webpack:",
		Env("PUPPETEER_SKIP_CHROMIUM_DOWNLOAD", "true"),
		Env("FORCE_COLOR", "1"),
		Cmd("npm ci"),
		Cmd("cd web"),
		Cmd("npm ci"),
		Cmd("npm run browserslist"),
		Cmd("NODE_ENV=production npm run build -- --color"),
		Cmd("GITHUB_TOKEN= npm run bundlesize"))

	pipeline.AddStep(":mocha:",
		Env("PUPPETEER_SKIP_CHROMIUM_DOWNLOAD", "true"),
		Env("FORCE_COLOR", "1"),
		Cmd("cd web"),
		Cmd("npm ci"),
		Cmd("npm run cover"),
		Cmd("node_modules/.bin/nyc report -r json"),
		ArtifactPaths("web/coverage/coverage-final.json"))

	pipeline.AddStep(":docker:",
		Cmd("curl -sL -o hadolint \"https://github.com/hadolint/hadolint/releases/download/v1.6.5/hadolint-$(uname -s)-$(uname -m)\" && chmod 700 hadolint"),
		Cmd("git ls-files | grep Dockerfile | xargs ./hadolint"))

	pipeline.AddStep(":go:",
		Cmd("dev/check/go-dep.sh"))

	pipeline.AddStep(":postgres:",
		Cmd("./dev/ci/ci-db-backcompat.sh"))

	for _, path := range pkgs() {
		coverageFile := path + "/coverage.txt"
		stepOpts := []StepOpt{
			Cmd("go test ./" + path + " -v -race -i"),
			Cmd("go test ./" + path + " -v -race -coverprofile=" + coverageFile + " -covermode=atomic -coverpkg=github.com/sourcegraph/sourcegraph/..."),
			ArtifactPaths(coverageFile),
		}
		if path == "cmd/frontend/internal/db" {
			stepOpts = append([]StepOpt{Cmd("./dev/ci/reset-test-db.sh || true")}, stepOpts...)
		}
		pipeline.AddStep(":go:", stepOpts...)
	}

	pipeline.AddWait()

	pipeline.AddStep(":codecov:",
		Cmd("buildkite-agent artifact download '*/coverage.txt' . || true"), // ignore error when no report exists
		Cmd("buildkite-agent artifact download '*/coverage-final.json' . || true"),
		Cmd("bash <(curl -s https://codecov.io/bash) -X gcov -X coveragepy -X xcode -t 89422d4b-0369-4d6c-bb5b-d709b5487a56"))

	addDeploySteps := func() {
		// Deploy to dogfood
		pipeline.AddStep(":dog:",
			ConcurrencyGroup("deploy"),
			Concurrency(1),
			Env("VERSION", version),
			Env("CONTEXT", "gke_sourcegraph-dev_us-central1-a_dogfood-cluster-7"),
			Env("NAMESPACE", "default"),
			Cmd("./dev/ci/deploy-dogfood.sh"))
		pipeline.AddWait()

		// Run e2e tests against dogfood
		pipeline.AddStep(":chromium:",
			ConcurrencyGroup("deploy"),
			Concurrency(1),
			Env("SOURCEGRAPH_BASE_URL", "https://sourcegraph.sgdev.org"),
			Env("FORCE_COLOR", "1"),
			Cmd("cd web"),
			Cmd("npm ci"),
			Cmd("npm run test-e2e -- --retries 5"),
			ArtifactPaths("web/puppeteer/*.png"))
		pipeline.AddWait()

		// Deploy to prod
		pipeline.AddStep(":rocket:",
			Env("VERSION", version),
			Cmd("./dev/ci/deploy-prod.sh"))
	}

	switch {
	case taggedRelease:
		latest := branch == "master"
		allDockerImages := []string{
			"frontend",
			"enterprise/frontend",
			"github-proxy",
			"gitserver",
			"indexer",
			"lsp-proxy",
			"query-runner",
			"repo-updater",
			"searcher",
			"enterprise/server",
			"symbols",
			"xlang-go",
		}

		for _, dockerImage := range allDockerImages {
			addDockerImageStep(dockerImage, latest)
		}
		pipeline.AddWait()

	case branch == "master":
		addDockerImageStep("frontend", true)
		addDockerImageStep("enterprise/frontend", true)
		addDockerImageStep("enterprise/server", true)
		pipeline.AddWait()
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
