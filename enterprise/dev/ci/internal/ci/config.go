package ci

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/images"
)

// Config is the set of configuration parameters that determine the structure of the CI build. These
// parameters are extracted from the build environment (branch name, commit hash, timestamp, etc.)
type Config struct {
	// runType indicates what kind of pipeline run should be generated, based on various
	// bits of metadata
	runType RunType

	// Build metadata
	now         time.Time
	branch      string
	version     string
	commit      string
	buildNumber int

	// changedFiles is the list of files that have changed since the
	// merge-base with origin/main.
	changedFiles ChangedFiles

	// profilingEnabled, if true, tells buildkite to print timing and resource utilization information
	// for each command
	profilingEnabled bool

	// mustIncludeCommit, if non-empty, is a list of commits at least one of which must be present
	// in the branch. If empty, then no check is enforced.
	mustIncludeCommit []string
}

func ComputeConfig() Config {
	now := time.Now()
	branch := os.Getenv("BUILDKITE_BRANCH")
	tag := os.Getenv("BUILDKITE_TAG")
	commit := os.Getenv("BUILDKITE_COMMIT")
	if commit == "" {
		commit = "1234567890123456789012345678901234567890" // for testing
	}
	buildNumber, _ := strconv.Atoi(os.Getenv("BUILDKITE_BUILD_NUMBER"))
	runType := computeRunType(tag, branch)

	switch {
	case runType.is(TaggedRelease):
		// The Git tag "v1.2.3" should map to the Docker image "1.2.3" (without v prefix).
		tag = strings.TrimPrefix(tag, "v")
	default:
		tag = fmt.Sprintf("%05d_%s_%.7s", buildNumber, now.Format("2006-01-02"), commit)
	}

	if runType.is(ImagePatch, ImagePatchNoTest) {
		// Add additional patch suffix
		tag = tag + "_patch"
	}

	var mustIncludeCommits []string
	if rawMustIncludeCommit := os.Getenv("MUST_INCLUDE_COMMIT"); rawMustIncludeCommit != "" {
		mustIncludeCommits = strings.Split(rawMustIncludeCommit, ",")
		for i := range mustIncludeCommits {
			mustIncludeCommits[i] = strings.TrimSpace(mustIncludeCommits[i])
		}
	}

	var changedFiles []string
	if output, err := exec.Command("git", "diff", "--name-only", "origin/main...").Output(); err != nil {
		panic(err)
	} else {
		changedFiles = strings.Split(strings.TrimSpace(string(output)), "\n")
	}

	return Config{
		runType: runType,

		now:               now,
		branch:            branch,
		version:           tag,
		commit:            commit,
		mustIncludeCommit: mustIncludeCommits,
		changedFiles:      changedFiles,
		buildNumber:       buildNumber,

		profilingEnabled: strings.Contains(branch, "buildkite-enable-profiling"),
	}

}

func (c Config) shortCommit() string {
	// http://git-scm.com/book/en/v2/Git-Tools-Revision-Selection#Short-SHA-1
	if len(c.commit) < 12 {
		return c.commit
	}

	return c.commit[:12]
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
		errs = multierror.Append(errs, errors.Errorf("%v | Output: %q", err, string(output)))
	}
	if !found {
		fmt.Printf("This branch %q at commit %s does not include any of these commits: %s.\n", c.branch, c.commit, strings.Join(c.mustIncludeCommit, ", "))
		fmt.Println("Rebase onto the latest main to get the latest CI fixes.")
		fmt.Printf("Errors from `git merge-base --is-ancestor $COMMIT HEAD`: %s", errs)
		return errs
	}
	return nil
}

// candidateImageTag provides the tag for a candidate image built for this Buildkite run.
//
// Note that the availability of this image depends on whether a candidate gets built,
// as determined in `addDockerImages()`.
func (c Config) candidateImageTag() string {
	buildNumber := os.Getenv("BUILDKITE_BUILD_NUMBER")
	return images.CandidateImageTag(c.commit, buildNumber)
}
