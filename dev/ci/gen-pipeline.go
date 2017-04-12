package main

import (
	"fmt"
	"go/build"
	"io/ioutil"
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

	// disabled because flaky: https://github.com/sourcegraph/sourcegraph/issues/5260
	// pipeline.AddStep(":php:",
	// 	Cmd("./xlang/php/test.sh"))

	pipeline.AddStep(":typescript:",
		Cmd("cd xlang/javascript-typescript/buildserver"),
		Cmd("yarn install"),
		Cmd("yarn run build"),
		Cmd("yarn run lint"),
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

	addDockerImageStep := func(app string, latest bool) {
		cmdDir := "./cmd/" + app
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

		image := "us.gcr.io/sourcegraph-dev/" + app
		buildScript := cmdDir + "/build.sh"
		if _, err := os.Stat(buildScript); err == nil {
			cmds = append(cmds,
				Env("IMAGE", image+":"+version),
				Env("VERSION", version),
				Cmd(buildScript),
			)
		} else {
			cmds = append(cmds,
				Cmd("go build sourcegraph.com/sourcegraph/sourcegraph/vendor/github.com/neelance/godockerize"),
				Cmd(fmt.Sprintf("./godockerize build -t %s:%s --env VERSION=%s sourcegraph.com/sourcegraph/sourcegraph/cmd/%s", image, version, version, app)),
			)
		}
		cmds = append(cmds,
			Cmd(fmt.Sprintf("gcloud docker -- push %s:%s", image, version)),
		)
		if latest {
			cmds = append(cmds,
				Cmd(fmt.Sprintf("docker tag %s:%s %s:latest", image, version, image)),
				Cmd(fmt.Sprintf("gcloud docker -- push %s:latest", image)),
			)
		}
		pipeline.AddStep(":docker:", cmds...)
	}

	switch {
	case branch == "master":
		pipeline.AddWait()
		addDockerImageStep("frontend", true)
		pipeline.AddWait()
		pipeline.AddStep(":rocket:",
			Env("VERSION", version),
			Cmd("./dev/ci/deploy-prod.sh"),
			Cmd("echo $VERSION | gsutil cp - gs://sourcegraph-metadata/latest-successful-build"))

	case strings.HasPrefix(branch, "staging/"):
		pipeline.AddWait()
		cmds, err := ioutil.ReadDir("./cmd")
		if err != nil {
			panic(err)
		}
		for _, cmd := range cmds {
			if cmd.Name() == "xlang-java" {
				continue // xlang-java currently does not build successfully on CI
			}
			addDockerImageStep(cmd.Name(), false)
		}
		pipeline.AddWait()
		pipeline.AddStep(":rocket:",
			Env("VERSION", version),
			Cmd("./dev/ci/deploy-staging.sh"))
		pipeline.AddWait()

	case strings.HasPrefix(branch, "docker-images/"):
		pipeline.AddWait()
		addDockerImageStep(branch[14:], true)
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
