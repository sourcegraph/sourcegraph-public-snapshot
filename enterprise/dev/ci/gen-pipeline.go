// gen-pipeline.go generates a Buildkite YAML file that tests the entire
// Sourcegraph application and writes it to stdout.
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/ci"
)

func main() {
	var (
		commit = os.Getenv("BUILDKITE_COMMIT")
		branch = os.Getenv("BUILDKITE_BRANCH")
		tag    = os.Getenv("BUILDKITE_TAG")
	)

	config := ci.NewConfig(time.Now(), commit, branch, tag)
	fmt.Printf("%+v\n", config)

	pipeline, err := ci.GeneratePipeline(config)
	if err != nil {
		panic(err)
	}

	_, err = pipeline.WriteTo(os.Stdout)
	if err != nil {
		panic(err)
	}
}
