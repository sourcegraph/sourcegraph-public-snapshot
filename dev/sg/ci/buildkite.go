package ci

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/gen2brain/beeep"
	sgrun "github.com/sourcegraph/run"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/ci/runtype"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/bk"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

const (
	ciLogsOutTerminal = "terminal"
	ciLogsOutSimple   = "simple"
	ciLogsOutJSON     = "json"
)

type buildTargetType string

const (
	buildTargetTypeBranch      buildTargetType = "branch"
	buildTargetTypeBuildNumber buildTargetType = "build"
	buildTargetTypeCommit      buildTargetType = "commit"
)

type targetBuild struct {
	targetType buildTargetType
	// target identifier - could br a branch or a build
	target string
	// buildkite pipeline to query
	pipeline string

	// Whether or not the target is set from a flag
	fromFlag bool
}

// getBuildTarget returns a targetBuild that can be used to retrieve details about a
// Buildkite build.
//
// Requires ciBranchFlag and ciBuildFlag to be registered on the command.
func getBuildTarget(cmd *cli.Context) (target targetBuild, err error) {
	target.pipeline = ciPipelineFlag.Get(cmd)
	if target.pipeline == "" {
		target.pipeline = "sourcegraph"
	}

	var (
		branch = ciBranchFlag.Get(cmd)
		build  = ciBuildFlag.Get(cmd)
		commit = ciCommitFlag.Get(cmd)
	)
	if branch != "" && build != "" {
		return target, errors.New("branch and build cannot both be set")
	}

	target.fromFlag = true
	switch {
	case branch != "":
		target.target = branch
		target.targetType = buildTargetTypeBranch

	case build != "":
		target.target = build
		target.targetType = buildTargetTypeBuildNumber

	case commit != "":
		// get the full commit
		target.target, err = root.Run(sgrun.Cmd(cmd.Context, "git rev-parse", commit)).String()
		if err != nil {
			return
		}
		target.targetType = buildTargetTypeCommit

	default:
		target.target, err = run.TrimResult(run.GitCmd("branch", "--show-current"))
		target.fromFlag = false
		target.targetType = buildTargetTypeBranch
	}
	return
}

func (t targetBuild) GetBuild(ctx context.Context, client *bk.Client) (build *buildkite.Build, err error) {
	switch t.targetType {
	case buildTargetTypeBranch:
		build, err = client.GetMostRecentBuild(ctx, t.pipeline, t.target)
		if err != nil {
			return nil, errors.Newf("failed to get most recent build for branch %q: %w", t.target, err)
		}
	case buildTargetTypeBuildNumber:
		build, err = client.GetBuildByNumber(ctx, t.pipeline, t.target)
		if err != nil {
			return nil, errors.Newf("failed to find build number %q: %w", t.target, err)
		}
	case buildTargetTypeCommit:
		build, err = client.GetBuildByCommit(ctx, t.pipeline, t.target)
		if err != nil {
			return nil, errors.Newf("failed to find build number %q: %w", t.target, err)
		}
	default:
		panic("bad target type " + t.targetType)
	}
	return
}

func getAllowedBuildTypeArgs() []string {
	var results []string
	for _, rt := range runtype.RunTypes() {
		if rt.Matcher().IsBranchPrefixMatcher() {
			display := fmt.Sprintf("%s - %s", strings.TrimSuffix(rt.Matcher().Branch, "/"), rt.String())
			results = append(results, display)
		}
	}
	return results
}

func printBuildOverview(build *buildkite.Build) {
	std.Out.WriteLine(output.Styledf(output.StyleBold, "Most recent build: %s", *build.WebURL))
	std.Out.Writef("Commit:\t\t%s", *build.Commit)
	std.Out.Writef("Message:\t%s", *build.Message)
	if build.Author != nil {
		std.Out.Writef("Author:\t\t%s <%s>", build.Author.Name, build.Author.Email)
	}
	if build.PullRequest != nil {
		std.Out.Writef("PR:\t\thttps://github.com/sourcegraph/sourcegraph/pull/%s", *build.PullRequest.ID)
	}
}

func agentKind(job *buildkite.Job) string {
	for _, rule := range job.AgentQueryRules {
		if strings.Contains(rule, "bazel") {
			return "bazel"
		}
	}
	return "stateless"
}

func formatBuildResult(result string) (string, output.Style) {
	var style output.Style
	var emoji string

	switch result {
	case "passed":
		style = output.StyleSuccess
		emoji = output.EmojiSuccess
	case "waiting", "blocked", "scheduled":
		style = output.StyleSuggestion
	case "skipped", "not_run", "broken":
		style = output.StyleReset
		emoji = output.EmojiOk
	case "running":
		style = output.StylePending
		emoji = output.EmojiInfo
	case "failed":
		emoji = output.EmojiFailure
		style = output.StyleFailure
	case "soft failed":
		emoji = output.EmojiOk
		style = output.StyleSearchLink
	default:
		style = output.StyleWarning
	}

	return emoji, style
}

func printBuildResults(build *buildkite.Build, annotations bk.JobAnnotations, notify bool) (failed bool) {
	std.Out.Writef("Started:\t%s", build.StartedAt)
	if build.FinishedAt != nil {
		std.Out.Writef("Finished:\t%s (elapsed: %s)", build.FinishedAt, build.FinishedAt.Sub(build.StartedAt.Time))
	}

	var statelessDuration time.Duration
	var bazelDuration time.Duration
	var totalDuration time.Duration

	// Check build state
	// Valid states: running, scheduled, passed, failed, blocked, canceled, canceling, skipped, not_run, waiting
	// https://buildkite.com/docs/apis/rest-api/builds
	emoji, style := formatBuildResult(*build.State)
	block := std.Out.Block(output.Styledf(style, "Status:\t\t%s %s", emoji, *build.State))

	// Inspect jobs individually.
	failedSummary := []string{"Failed jobs:"}
	for _, job := range build.Jobs {
		var elapsed time.Duration
		if job.State == nil || job.Name == nil {
			continue
		}
		if *job.State == "failed" && job.SoftFailed {
			*job.State = "soft failed"
		}

		_, style := formatBuildResult(*job.State)
		// Check job state.
		switch *job.State {
		case "passed":
			elapsed = job.FinishedAt.Sub(job.StartedAt.Time)
		case "waiting", "blocked", "scheduled", "assigned":
		case "broken":
			// State 'broken' happens when a conditional is not met, namely the 'if' block
			// on a job. Why is it 'broken' and not 'skipped'? We don't think it be like
			// this, but it do. Anyway, we pretend it was skipped and treat it as such.
			// https://buildkite.com/docs/pipelines/conditionals#conditionals-and-the-broken-state
			*job.State = "skipped"
			fallthrough
		case "skipped", "not_run":
		case "running":
			elapsed = time.Since(job.StartedAt.Time)
		case "failed":
			elapsed = job.FinishedAt.Sub(job.StartedAt.Time)
			failedSummary = append(failedSummary, fmt.Sprintf("- %s", *job.Name))
			failed = true
		default:
			style = output.StyleWarning
		}

		if elapsed > 0 {
			block.WriteLine(output.Styledf(style, "- [%s] %s (%s)", *job.State, *job.Name, elapsed))
		} else {
			block.WriteLine(output.Styledf(style, "- [%s] %s", *job.State, *job.Name))
		}

		totalDuration += elapsed
		if agentKind(job) == "bazel" {
			bazelDuration += elapsed
		} else {
			statelessDuration += elapsed
		}
		if annotation, exist := annotations[*job.ID]; exist {
			block.WriteMarkdown(annotation.Content, output.MarkdownNoMargin, output.MarkdownIndent(2))
		}
	}

	block.Close()

	if build.FinishedAt != nil {
		statusStr := fmt.Sprintf("Status:\t\t%s %s\n", emoji, *build.State)
		std.Out.Write(strings.Repeat("-", len(statusStr)+8*2)) // 2 * \t
		std.Out.WriteLine(output.Linef(emoji, output.StyleReset, statusStr))
		std.Out.WriteLine(output.Linef("", output.StyleReset, "Finished at: %s", build.FinishedAt))
		std.Out.WriteLine(output.Linef("", output.StyleReset, "- ⏲️  Wall-clock time: %s", build.FinishedAt.Sub(build.StartedAt.Time)))
		std.Out.WriteLine(output.Linef("", output.StyleReset, "- 🗒️ CI agents time:  %s", totalDuration))
		std.Out.WriteLine(output.Linef("", output.StyleReset, "  - Bazel: %s", bazelDuration))
		std.Out.WriteLine(output.Linef("", output.StyleReset, "  - Stateless: %s", statelessDuration))
	}

	if notify {
		if failed {
			beeep.Alert(fmt.Sprintf("❌ Build failed (%s)", *build.Branch), strings.Join(failedSummary, "\n"), "")
		} else {
			beeep.Notify(fmt.Sprintf("✅ Build passed (%s)", *build.Branch), fmt.Sprintf("%d jobs passed in %s", len(build.Jobs), build.FinishedAt.Sub(build.StartedAt.Time)), "")
		}
	}

	return failed
}

func statusTicker(ctx context.Context, f func() (bool, error)) error {
	// Start immediately
	ok, err := f()
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	// Not finished, start ticking ...
	ticker := time.NewTicker(20 * time.Second)
	for {
		select {
		case <-ticker.C:
			ok, err := f()
			if err != nil {
				return err
			}
			if ok {
				return nil
			}
		case <-time.After(30 * time.Minute):
			return errors.Newf("polling timeout reached")
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func fetchJobs(ctx context.Context, client *bk.Client, buildPtr **buildkite.Build, pending output.Pending) func() (bool, error) {
	return func() (bool, error) {
		build, err := client.GetBuildByNumber(ctx, "sourcegraph", strconv.Itoa(*((*buildPtr).Number)))
		if err != nil {
			return false, errors.Newf("failed to get most recent build for branch %q: %w", *build.Branch, err)
		}

		// Update the original build reference with the refreshed one.
		*buildPtr = build

		// Check if all jobs are finished
		finishedJobs := 0
		for _, job := range build.Jobs {
			if job.State != nil {
				if *job.State == "failed" && !job.SoftFailed {
					// If a job has failed, return immediately, we don't have to wait until all
					// steps are completed.
					return true, nil
				}
				if *job.State == "passed" || job.SoftFailed {
					finishedJobs++
				}
			}
		}

		// once started, poll for status
		if build.StartedAt != nil {
			pending.Updatef("Waiting for %d out of %d jobs... (elapsed: %v)",
				len(build.Jobs)-finishedJobs, len(build.Jobs), time.Since(build.StartedAt.Time))
		}

		if build.FinishedAt == nil {
			// No failure yet, we can keep waiting.
			return false, nil
		}
		return true, nil
	}
}
