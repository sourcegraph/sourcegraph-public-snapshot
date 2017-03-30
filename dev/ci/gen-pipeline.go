package main

import (
	"fmt"
	"go/build"
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
	Label   string                 `json:"label"`
	Command string                 `json:"command"`
	Env     map[string]string      `json:"env"`
	Plugins map[string]interface{} `json:"plugins"`
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

type StepOpt func(step *Step)

func Cmd(command string) StepOpt {
	return func(step *Step) {
		step.Command = strings.TrimSpace(step.Command + "\n" + command)
	}
}

func Env(name, value string) StepOpt {
	return func(step *Step) {
		step.Env[name] = value
	}
}

func (p *Pipeline) AddWait() {
	p.Steps = append(p.Steps, "wait")
}

var golangPlugin = map[string]interface{}{
	"golang#v0.3": map[string]string{
		"import": "sourcegraph.com/sourcegraph/sourcegraph",
	},
}

func main() {
	xlang := "sourcegraph.com/sourcegraph/sourcegraph/xlang"
	pkgs := []string{xlang} // put slow xlang test first
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}
		if path == "." || !info.IsDir() {
			return nil
		}
		if path == ".git" || path == "ui" || path == "vendor" {
			return filepath.SkipDir
		}

		importPath := "sourcegraph.com/sourcegraph/sourcegraph/" + path
		if importPath == xlang {
			return nil // already first entry
		}

		pkg, err := build.Import(importPath, "", 0)
		if err != nil {
			if _, ok := err.(*build.NoGoError); ok {
				return nil
			}
			panic(err)
		}

		if len(pkg.TestGoFiles) == 0 && len(pkg.XTestGoFiles) == 0 {
			return nil
		}

		pkgs = append(pkgs, importPath)

		return nil
	})
	if err != nil {
		panic(err)
	}

	pipeline := &Pipeline{}

	pipeline.AddStep(":white_check_mark:",
		Cmd("./dev/check/all.sh"))

	pipeline.AddStep(":desktop_computer:",
		Cmd("cd ui"),
		Cmd("yarn install"),
		Cmd("yarn run test"))

	pipeline.AddStep(":chrome:",
		Cmd("cd client/browser-ext"),
		Cmd("yarn install"),
		Cmd("yarn run build"))

	pipeline.AddStep(":php:",
		Cmd("./xlang/php/test.sh"))

	pipeline.AddStep(":typescript:",
		Cmd("cd xlang/javascript-typescript/buildserver"),
		Cmd("yarn install"),
		Cmd("yarn run build"),
		Cmd("yarn run fmt-check"),
		Cmd("yarn test"))

	for _, pkg := range pkgs {
		pipeline.AddStep(":go:",
			Cmd("go test -race -v "+pkg))
	}

	branch := os.Getenv("BUILDKITE_BRANCH")
	commit := os.Getenv("BUILDKITE_COMMIT")
	if commit == "" {
		commit = "1234567890123456789012345678901234567890" // for testing
	}
	buildNum, _ := strconv.Atoi(os.Getenv("BUILDKITE_BUILD_NUMBER"))
	version := fmt.Sprintf("%05d_%s_%.7s", buildNum, time.Now().Format("2006-01-02"), commit)

	switch {
	case branch == "master":
		pipeline.AddWait()
		pipeline.AddStep(":docker:",
			Env("TAG", version),
			Cmd("./cmd/frontend/build.sh"),
			Cmd("docker tag us.gcr.io/sourcegraph-dev/sourcegraph:"+version+" us.gcr.io/sourcegraph-dev/sourcegraph:latest"),
			Cmd("gcloud docker -- push us.gcr.io/sourcegraph-dev/sourcegraph:"+version),
			Cmd("gcloud docker -- push us.gcr.io/sourcegraph-dev/sourcegraph:latest"),
			Env("VERSION", version),
			Cmd("./dev/ci/deploy-prod.sh"))
		pipeline.AddWait()
		pipeline.AddStep(":rocket:",
			Env("VERSION", version),
			Cmd("./dev/ci/deploy-prod.sh"),
			Cmd("echo $VERSION | gsutil cp - gs://sourcegraph-metadata/latest-successful-build"))

	case strings.HasPrefix(branch,
		"staging/"):
		stagingName := strings.Replace(strings.TrimPrefix(branch, "staging/"), "/", "-", -1)
		pipeline.AddWait()
		pipeline.AddStep(":docker:",
			Env("TAG", version),
			Cmd("./cmd/frontend/build.sh"),
			Cmd("gcloud docker -- push us.gcr.io/sourcegraph-dev/sourcegraph:"+version))
		pipeline.AddWait()
		pipeline.AddStep(":rocket:",
			Env("VERSION", version),
			Cmd("./dev/ci/deploy-staging.sh"))
		pipeline.AddWait()
		pipeline.AddStep(":selenium:",
			Cmd("cd test/e2e2"),
			Cmd("pip install virtualenv"),
			Cmd("virtualenv .env"),
			Env("VERSION", version),
			Env("STAGING_NAME", stagingName),
			Cmd("../../dev/ci/wait-for-deploy.sh"),
			Env("NOVNC", "1"),
			Env("SOURCEGRAPH_URL", fmt.Sprintf("http://%s.staging.sgdev.org", stagingName)),
			Cmd("make ci"),
		)

	case strings.HasPrefix(branch,
		"docker-images/"):
		pipeline.AddWait()
		pipeline.AddStep(":docker:",
			Env("TAG", version),
			Cmd("./dev/ci/docker-images.sh "+branch[14:]))
		pipeline.AddWait()
		pipeline.AddStep(":rocket:",
			Env("VERSION", version),
			Cmd("./dev/ci/deploy-prod.sh"))

	}

	output, err := yaml.Marshal(pipeline)
	if err != nil {
		panic(err)
	}

	os.Stdout.Write(output)
}
