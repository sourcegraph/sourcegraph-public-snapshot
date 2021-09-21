// gen-pipeline.go generates a Buildkite YAML file that tests the entire
// Sourcegraph application and writes it to stdout.
package main

import (
	"os"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/ci"
)

func main() {
	var (
		commit             = os.Getenv("BUILDKITE_COMMIT")
		branch             = os.Getenv("BUILDKITE_BRANCH")
		tag                = os.Getenv("BUILDKITE_TAG")
		mustIncludeCommits []string
	)

	if rawMustIncludeCommit := os.Getenv("MUST_INCLUDE_COMMIT"); rawMustIncludeCommit != "" {
		mustIncludeCommits = strings.Split(rawMustIncludeCommit, ",")
		for i := range mustIncludeCommits {
			mustIncludeCommits[i] = strings.TrimSpace(mustIncludeCommits[i])
		}
	}

	config := ci.NewConfig(time.Now(), commit, branch, tag, mustIncludeCommits)

	pipeline, err := ci.GeneratePipeline(config)
	if err != nil {
		panic(err)
	}

	_, err = pipeline.WriteTo(os.Stdout)
	if err != nil {
		panic(err)
	}
}
