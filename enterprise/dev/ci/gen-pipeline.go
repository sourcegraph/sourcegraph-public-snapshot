// gen-pipeline.go generates a Buildkite YAML file that tests the entire
// Sourcegraph application and writes it to stdout.
package main

import (
	"os"

	bk "github.com/sourcegraph/sourcegraph/pkg/buildkite"
)

func init() {
	bk.Plugins["gopath-checkout#v1.0.1"] = map[string]string{
		"import": "github.com/sourcegraph/sourcegraph",
	}
}

func main() {
	pipeline := &bk.Pipeline{}

	defer func() {
		_, err := pipeline.WriteTo(os.Stdout)
		if err != nil {
			panic(err)
		}
	}()

	// Run e2e tests
	pipeline.AddStep(":chromium:",
		bk.Env("FORCE_COLOR", "1"),
		bk.Env("PUPPETEER_SKIP_CHROMIUM_DOWNLOAD", ""),
		bk.Env("DISPLAY", ":99"),
		bk.Cmd("Xvfb :99 &"),
		bk.Cmd("yarn --frozen-lockfile --network-timeout 60000"),
		bk.Cmd("pushd web"),
		bk.Cmd("yarn run test-e2e -t 'theme'"), // only run the theme switcher test for speed of debugging
		bk.Cmd("popd"),
		bk.ArtifactPaths("./puppeteer/*.png"),
	)
}
