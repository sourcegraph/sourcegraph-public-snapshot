package ci

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

// Config is the set of configuration parameters that determine the structure of the CI build. These
// parameters are extracted from the build environment (branch name, commit hash, timestamp, etc.)
type Config struct {
	now     time.Time
	branch  string
	version string
	commit  string

	// mustIncludeCommit, if non-empty, is a list of commits at least one of which must be present
	// in the branch. If empty, then no check is enforced.
	mustIncludeCommit []string

	taggedRelease       bool
	releaseBranch       bool
	isBextReleaseBranch bool
	isRenovateBranch    bool
	patch               bool
	patchNoTest         bool
	isQuick             bool
	isMasterDryRun      bool
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

	isMasterDryRun := strings.HasPrefix(branch, "master-dry-run/")

	isQuick := strings.HasPrefix(branch, "quick/")

	var mustIncludeCommits []string
	if rawMustIncludeCommit := os.Getenv("MUST_INCLUDE_COMMIT"); rawMustIncludeCommit != "" {
		mustIncludeCommits = strings.Split(rawMustIncludeCommit, ",")
		for i := range mustIncludeCommits {
			mustIncludeCommits[i] = strings.TrimSpace(mustIncludeCommits[i])
		}
	}
	return Config{
		now:               now,
		branch:            branch,
		version:           version,
		commit:            commit,
		mustIncludeCommit: mustIncludeCommits,

		taggedRelease:       taggedRelease,
		releaseBranch:       lazyregexp.New(`^[0-9]+\.[0-9]+$`).MatchString(branch),
		isBextReleaseBranch: branch == "bext/release",
		isRenovateBranch:    strings.HasPrefix(branch, "renovate/"),
		patch:               patch,
		patchNoTest:         patchNoTest,
		isQuick:             isQuick,
		isMasterDryRun:      isMasterDryRun,
	}
}

func (c Config) ensureCommit() error {
	if len(c.mustIncludeCommit) == 0 {
		return nil
	}

	found := false
	var errs error
	for _, mustIncludeCommit := range c.mustIncludeCommit {
		output, err := exec.Command("git", "merge-base", "--is-ancestor", mustIncludeCommit, "HEAD").CombinedOutput()
		if err == nil {
			found = true
			break
		}
		errs = multierror.Append(errs, fmt.Errorf("%v | Output: %q", err, string(output)))
	}
	if !found {
		fmt.Printf("This branch %q at commit %s does not include any of these commits: %s.\n", c.branch, c.commit, strings.Join(c.mustIncludeCommit, ", "))
		fmt.Println("Rebase onto the latest master to get the latest CI fixes.")
		fmt.Printf("Errors from `git merge-base --is-ancestor $COMMIT HEAD`: %s", errs.Error())
		return errs
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
