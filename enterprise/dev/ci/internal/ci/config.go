package ci

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	bk "github.com/buildkite/go-buildkite/v3/buildkite"
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
func NewConfig(now time.Time) (Config, error) {
	var (
		commit = os.Getenv("BUILDKITE_COMMIT")
		branch = os.Getenv("BUILDKITE_BRANCH")
		tag    = os.Getenv("BUILDKITE_TAG")
		token  = os.Getenv("BUILDKITE_API_TOKEN")
		// defaults to 0
		buildNumber, _ = strconv.Atoi(os.Getenv("BUILDKITE_BUILD_NUMBER"))
	)

	var bkClient *bk.Client
	if token != "" {
		bkConfig, err := bk.NewTokenConfig(token, false)
		if err != nil {
			return Config{}, err
		}

		bkClient = bk.NewClient(bkConfig.Client())
	}

	var mustIncludeCommits []string
	if rawMustIncludeCommit := os.Getenv("MUST_INCLUDE_COMMIT"); rawMustIncludeCommit != "" {
		mustIncludeCommits = strings.Split(rawMustIncludeCommit, ",")
		for i := range mustIncludeCommits {
			mustIncludeCommits[i] = strings.TrimSpace(mustIncludeCommits[i])
		}
	}

	// detect changed files
	changedFiles, commit, err := getChangedFiles(bkClient, branch, commit)
	if err != nil {
		return Config{}, err
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
	}, nil
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

func getChangedFiles(bkClient *buildkite.Client, branch, commit string) ([]string, string, error) {
	var changedFiles []string

	diffCommand, commit, err := buildDiffCommand(bkClient, branch, commit)
	if err != nil {
		return nil, commit, err
	}

	debugLog("Running git %s\n", strings.Join(diffCommand, " "))
	cmd := exec.Command("git", diffCommand...)
	cmd.Stderr = os.Stderr
	if output, err := cmd.Output(); err != nil {
		return nil, "", err
	} else {
		changedFiles = strings.Split(strings.TrimSpace(string(output)), "\n")
	}

	return changedFiles, commit, nil
}

func buildDiffCommand(bkClient *buildkite.Client, branch, commit string) (args []string, newCommit string, err error) {
	diffCommand := []string{"diff", "--name-only"}
	if commit == "" {
		fmt.Fprintln(os.Stderr, "No commit. Comparing with main")
		diffCommand = append(diffCommand, "origin/main...")
		// for testing
		commit = "1234567890123456789012345678901234567890"
		return diffCommand, commit, nil
	}

	if bkClient == nil {
		// if there is no builkite API client, run a diff with main
		debugLog("BUILDKITE_API_TOKEN env var not found, comparing with main...")
		return append(diffCommand, "origin/main..."+commit), commit, nil
	}

	// get the latest successful build for this branch and run a diff against that commit
	builds, _, err := bkClient.Builds.ListByPipeline("sourcegraph", "sourcegraph", &buildkite.BuildsListOptions{
		Branch: branch,
		State:  []string{"passed"},
		ListOptions: buildkite.ListOptions{
			PerPage: 1,
		},
	})
	if err != nil {
		return
	}

	// if there are no previous builds diff with main
	if len(builds) == 0 || builds[0].State == nil || *(builds[0].State) != "passed" {
		fmt.Fprintln(os.Stderr, "No previous passing build. Comparing with main...")
		return append(diffCommand, "origin/main..."+commit), commit, nil
	}

	build := builds[0]

	// diff with the previous build commit
	// after making sure the commit is in still in that branch
	// (the branch may have been rebased)

	// fetch the current branch
	if out, err := exec.Command("git", "fetch", "origin", branch).CombinedOutput(); err != nil {
		return nil, "", fmt.Errorf("error while fetching the current branch, err: %w, output: %q", err, string(out))
	}
	// checkout the current branch
	if out, err := exec.Command("git", "checkout", branch).CombinedOutput(); err != nil {
		return nil, "", fmt.Errorf("error while checking out the current branch, err: %w, output: %q", err, string(out))
	}

	defer func() {
		// go back on the previous commit
		if out, er := exec.Command("git", "checkout", "-").CombinedOutput(); er != nil {
			err = multierror.Append(err, fmt.Errorf("error while checking out the current commit, err: %w, output: %q", err, string(out)))
		}
	}()

	var buf bytes.Buffer
	cmd := exec.Command("git", "branch", "--contains", *build.Commit)
	cmd.Stderr = os.Stderr
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return nil, "", err
	}
	debugLog("Previous build commit %q found in the following branches: %v\n", *build.Commit, buf.String())

	var found bool
	for _, b := range strings.Split(buf.String(), "\n") {
		if strings.TrimSpace(strings.TrimLeft(b, "*")) == branch {
			found = true
			break
		}
	}

	if !found {
		debugLog("Previous build commit %q not found in current branch. Comparing with main...\n", *build.Commit)
		return append(diffCommand, "origin/main..."+commit), commit, nil
	}

	fmt.Fprintln(os.Stderr, "Comparing with "+*build.Commit)
	diffCommand = append(diffCommand, *build.Commit+"..."+branch)

	return diffCommand, commit, nil
}

// debugLog logs on os.Stderr using fmt.Fprintf.
// Node: os.Stdout cannot be used for logging, as it is read by
// builkite to generate the pipeline.
func debugLog(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}
