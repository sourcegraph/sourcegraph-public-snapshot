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
	"github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/ci/changed"
)

// Config is the set of configuration parameters that determine the structure of the CI build. These
// parameters are extracted from the build environment (branch name, commit hash, timestamp, etc.)
type Config struct {
	// RunType indicates what kind of pipeline run should be generated, based on various
	// bits of metadata
	RunType RunType

	// Build metadata
	Time        time.Time
	Branch      string
	Version     string
	Commit      string
	BuildNumber int

	// ChangedFiles is the list of files that have changed since the
	// merge-base with origin/main.
	ChangedFiles changed.Files

	// MustIncludeCommit, if non-empty, is a list of commits at least one of which must be present
	// in the branch. If empty, then no check is enforced.
	MustIncludeCommit []string

	// MessageFlags contains flags parsed from commit messages.
	MessageFlags MessageFlags
}

// NewConfig computes configuration for the pipeline generator based on Buildkite environment
// variables.
func NewConfig(now time.Time) Config {
	var (
		commit = os.Getenv("BUILDKITE_COMMIT")
		branch = os.Getenv("BUILDKITE_BRANCH")
		tag    = os.Getenv("BUILDKITE_TAG")
		// defaults to 0
		buildNumber, _ = strconv.Atoi(os.Getenv("BUILDKITE_BUILD_NUMBER"))
	)

	var mustIncludeCommits []string
	if rawMustIncludeCommit := os.Getenv("MUST_INCLUDE_COMMIT"); rawMustIncludeCommit != "" {
		mustIncludeCommits = strings.Split(rawMustIncludeCommit, ",")
		for i := range mustIncludeCommits {
			mustIncludeCommits[i] = strings.TrimSpace(mustIncludeCommits[i])
		}
	}

	// detect changed files
	var changedFiles []string
	diffCommand := []string{"diff", "--name-only"}
	if commit != "" {
		diffCommand = append(diffCommand, "origin/main..."+commit)
	} else {
		diffCommand = append(diffCommand, "origin/main...")
		// for testing
		commit = "1234567890123456789012345678901234567890"
	}
	if output, err := exec.Command("git", diffCommand...).Output(); err != nil {
		panic(err)
	} else {
		changedFiles = strings.Split(strings.TrimSpace(string(output)), "\n")
	}

	// evaluates what type of pipeline run this is
	runType := computeRunType(tag, branch)

	// special tag adjustments based on run type
	switch {
	case runType.Is(TaggedRelease):
		// The Git tag "v1.2.3" should map to the Docker image "1.2.3" (without v prefix).
		tag = strings.TrimPrefix(tag, "v")
	default:
		tag = fmt.Sprintf("%05d_%s_%.7s", buildNumber, now.Format("2006-01-02"), commit)
	}
	if runType.Is(ImagePatch, ImagePatchNoTest, ExecutorPatchNoTest) {
		// Add additional patch suffix
		tag = tag + "_patch"
	}

	return Config{
		RunType: runType,

		Time:              now,
		Branch:            branch,
		Version:           tag,
		Commit:            commit,
		MustIncludeCommit: mustIncludeCommits,
		ChangedFiles:      changedFiles,
		BuildNumber:       buildNumber,

		// get flags from commit message
		MessageFlags: parseMessageFlags(os.Getenv("BUILDKITE_MESSAGE")),
	}
}

func (c Config) shortCommit() string {
	// http://git-scm.com/book/en/v2/Git-Tools-Revision-Selection#Short-SHA-1
	if len(c.Commit) < 12 {
		return c.Commit
	}

	return c.Commit[:12]
}

func (c Config) ensureCommit() error {
	if len(c.MustIncludeCommit) == 0 {
		return nil
	}

	found := false
	var errs error
	for _, mustIncludeCommit := range c.MustIncludeCommit {
		output, err := exec.Command("git", "merge-base", "--is-ancestor", mustIncludeCommit, "HEAD").CombinedOutput()
		if err == nil {
			found = true
			break
		}
		errs = multierror.Append(errs, errors.Errorf("%v | Output: %q", err, string(output)))
	}
	if !found {
		fmt.Printf("This branch %q at commit %s does not include any of these commits: %s.\n", c.Branch, c.Commit, strings.Join(c.MustIncludeCommit, ", "))
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
	return images.CandidateImageTag(c.Commit, strconv.Itoa(c.BuildNumber))
}

// MessageFlags indicates flags that can be parsed out of commit messages to change
// pipeline behaviour. Use sparingly! If you are generating a new pipeline, please use
// RunType instead.
type MessageFlags struct {
	// ProfilingEnabled, if true, tells buildkite to print timing and resource utilization information
	// for each command
	ProfilingEnabled bool

	// SkipHashCompare, if true, tells buildkite to disable skipping of steps that compare
	// hash output.
	SkipHashCompare bool
}

// parseMessageFlags gets MessageFlags from the given commit message.
func parseMessageFlags(msg string) MessageFlags {
	return MessageFlags{
		ProfilingEnabled: strings.Contains(msg, "[buildkite-enable-profiling]"),
		SkipHashCompare:  strings.Contains(msg, "[skip-hash-compare]"),
	}
}
