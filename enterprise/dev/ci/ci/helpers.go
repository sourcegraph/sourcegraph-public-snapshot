package ci

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Config is the set of configuration parameters that determine the structure of the CI build. These
// parameters are extracted from the build environment (branch name, commit hash, timestamp, etc.)
type Config struct {
	now               time.Time
	pipeline          string
	branch            string
	version           string
	commit            string
	mustIncludeCommit string

	taggedRelease       bool
	releaseBranch       bool
	isBextReleaseBranch bool
	patch               bool
	patchNoTest         bool
}

func ComputeConfig() Config {
	now := time.Now()
	branch := os.Getenv("BUILDKITE_BRANCH")
	version := os.Getenv("BUILDKITE_TAG")
	commit := os.Getenv("BUILDKITE_COMMIT")
	if commit == "" {
		commit = "1234567890123456789012345678901234567890" // for testing
	}

	taggedRelease := true // true if this is a tagged release
	switch {
	case strings.HasPrefix(version, "v"):
		// The Git tag "v1.2.3" should map to the Docker image "1.2.3" (without v prefix).
		version = strings.TrimPrefix(version, "v")
	default:
		taggedRelease = false
		buildNum, _ := strconv.Atoi(os.Getenv("BUILDKITE_BUILD_NUMBER"))
		version = fmt.Sprintf("%05d_%s_%.7s", buildNum, now.Format("2006-01-02"), commit)
	}

	patchNoTest := strings.HasPrefix(branch, "docker-images-patch-notest/")
	patch := strings.HasPrefix(branch, "docker-images-patch/")
	if patchNoTest || patch {
		version = version + "_patch"
	}

	pipeline := os.Getenv("BUILDKITE_PIPELINE")
	if pipeline == "" {
		pipeline = "sourcegraph"
	}

	return Config{
		now:               now,
		pipeline:          pipeline,
		branch:            branch,
		version:           version,
		commit:            commit,
		mustIncludeCommit: os.Getenv("MUST_INCLUDE_COMMIT"),

		taggedRelease:       taggedRelease,
		releaseBranch:       regexp.MustCompile(`^[0-9]+\.[0-9]+$`).MatchString(branch),
		isBextReleaseBranch: branch == "bext/release",
		patch:               patch,
		patchNoTest:         patchNoTest,
	}
}

func (c Config) ensureCommit() error {
	if c.mustIncludeCommit != "" {
		output, err := exec.Command("git", "merge-base", "--is-ancestor", c.mustIncludeCommit, "HEAD").CombinedOutput()
		if err != nil {
			fmt.Printf("This branch %s at commit %s does not include commit %s.\n", c.branch, c.commit, c.mustIncludeCommit)
			fmt.Println("Rebase onto the latest master to get the latest CI fixes.")
			fmt.Println(string(output))
			return err
		}
	}
	return nil
}

func (c Config) isPR() bool {
	return !c.isBextReleaseBranch &&
		!c.releaseBranch &&
		!c.taggedRelease &&
		c.branch != "master" &&
		!strings.HasPrefix(c.branch, "master-dry-run/") &&
		!strings.HasPrefix(c.branch, "docker-images-patch/")
}

func isDocsOnly() bool {
	output, err := exec.Command("git", "diff", "--name-only", "origin/master...").Output()
	if err != nil {
		panic(err)
	}
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if !strings.HasPrefix(line, "doc") && line != "CHANGELOG.md" {
			return false
		}
	}
	return true
}

func gitRevParse(rev string) string {
	commit, err := exec.Command("git", "rev-parse", rev).CombinedOutput()
	if err != nil {
		panic(err.Error() + " " + string(commit))
	}
	return string(bytes.TrimSpace(commit))
}
