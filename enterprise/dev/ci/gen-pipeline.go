// gen-pipeline.go generates a Buildkite YAML file that tests the entire
// Sourcegraph application and writes it to stdout.
package main

import (
	"fmt"
	"os"
	"os/exec"
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

	defer func() {
		_, err := pipeline.WriteTo(os.Stdout)
		if err != nil {
			panic(err)
		}
	}()

	branch := os.Getenv("BUILDKITE_BRANCH")
	version := os.Getenv("BUILDKITE_TAG")
	commit := os.Getenv("BUILDKITE_COMMIT")
	if commit == "" {
		commit = "1234567890123456789012345678901234567890" // for testing
	}
	taggedRelease := true // true if this is a semver tagged release
	now := time.Now()
	if !strings.HasPrefix(version, "v") {
		taggedRelease = false
		buildNum, _ := strconv.Atoi(os.Getenv("BUILDKITE_BUILD_NUMBER"))
		version = fmt.Sprintf("%05d_%s_%.7s", buildNum, now.Format("2006-01-02"), commit)
	} else {
		// The Git tag "v1.2.3" should map to the Docker image "1.2.3" (without v prefix).
		version = strings.TrimPrefix(version, "v")
	}
	releaseBranch := regexp.MustCompile(`^[0-9]+\.[0-9]+$`).MatchString(branch)

	isBextReleaseBranch := branch == "bext/release"

	bk.OnEveryStepOpts = append(bk.OnEveryStepOpts,
		bk.Env("GO111MODULE", "on"),
		bk.Env("PUPPETEER_SKIP_CHROMIUM_DOWNLOAD", "true"),
		bk.Env("FORCE_COLOR", "1"),
		bk.Env("ENTERPRISE", "1"),
		bk.Env("COMMIT_SHA", commit),
		bk.Env("DATE", now.Format(time.RFC3339)),
	)

	isPR := !isBextReleaseBranch &&
		!releaseBranch &&
		!taggedRelease &&
		branch != "master" &&
		!strings.HasPrefix(branch, "master-dry-run/") &&
		!strings.HasPrefix(branch, "docker-images-patch/")
	if isPR {
		output, err := exec.Command("git", "diff", "--name-only", "origin/master...").Output()
		if err != nil {
			panic(err)
		}

		onlyDocsChange := true
		for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
			if !strings.HasPrefix(line, "doc") && line != "CHANGELOG.md" {
				onlyDocsChange = false
				break
			}
		}

		if onlyDocsChange {
			pipeline.AddStep(":memo:",
				bk.Cmd("./dev/ci/yarn-run.sh prettier-check"),
				bk.Cmd("./dev/check/docsite.sh"))
			return
		}
	}

	pipeline.AddStep(":chromium:",
		// Avoid crashing the sourcegraph/server containers. See
		// https://github.com/sourcegraph/sourcegraph/issues/2657
		bk.ConcurrencyGroup("e2e"),
		bk.Concurrency(1),

		bk.Env("IMAGE", "sourcegraph/server:"+version+"_candidate"),
		bk.Env("VERSION", version),
		bk.Env("PUPPETEER_SKIP_CHROMIUM_DOWNLOAD", ""),
		bk.Cmd("./dev/ci/e2e.sh"),
		bk.ArtifactPaths("./puppeteer/*.png;./web/e2e.mp4;./web/ffmpeg.log"))
}
