// gen-pipeline.go generates a Buildkite YAML file that tests the entire
// Sourcegraph application and writes it to stdout.
package main

import (
	"os"

	ci "github.com/sourcegraph/sourcegraph/enterprise/dev/ci/ci"
)

func main() {
	pipeline, err := ci.GeneratePipeline(ci.ComputeConfig())
	if err != nil {
		panic(err)
	}
	_, err = pipeline.WriteTo(os.Stdout)
	if err != nil {
		panic(err)
	}
}
