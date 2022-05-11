package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/gen2brain/beeep"
	"github.com/grafana/regexp"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/ci/runtype"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/bk"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/loki"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/open"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

const (
	ciLogsOutTerminal = "terminal"
	ciLogsOutSimple   = "simple"
	ciLogsOutJSON     = "json"
)

var (
	ciBranch     string
	ciBranchFlag = cli.StringFlag{
		Name:        "branch",
		Aliases:     []string{"b"},
		Usage:       "Branch `name` of build to target (defaults to current branch)",
		Destination: &ciBranch,
	}

	ciBuild     string
	ciBuildFlag = cli.StringFlag{
		Name:        "build",
		Usage:       "Override branch detection with a specific build `number`",
		Destination: &ciBuild,
	}
)

// get branch from flag or git
func getCIBranch() (branch string, fromFlag bool, err error) {
	if ciBranch != "" && ciBuild != "" {
		return "", false, errors.New("branch and build cannot both be set")
	}

	fromFlag = true
	switch {
	case ciBranch != "":
		branch = ciBranch
	case ciBuild != "":
		branch = ciBuild
	default:
		branch, err = run.TrimResult(run.GitCmd("branch", "--show-current"))
		fromFlag = false
	}
	return
}

var ciCommand = &cli.Command{
	Name:  "ci",
	Usage: "Interact with Sourcegraph's continuous integration pipelines",
	Description: `Interact with Sourcegraph's continuous integration pipelines on Buildkite.

Note that Sourcegraph's CI pipelines are under our enterprise license: https://github.com/sourcegraph/sourcegraph/blob/main/LICENSE.enterprise`,
	Category: CategoryDev,
	Subcommands: []*cli.Command{{
		Name:    "preview",
		Aliases: []string{"plan"},
		Usage:   "Preview the pipeline that would be run against the currently checked out branch",
		Action: execAdapter(func(ctx context.Context, args []string) error {
			std.Out.WriteLine(output.Styled(output.StyleSuggestion,
				"If the current branch were to be pushed, the following pipeline would be run:"))

			branch, err := run.TrimResult(run.GitCmd("branch", "--show-current"))
			if err != nil {
				return err
			}
			message, err := run.TrimResult(run.GitCmd("show", "--format=%s\\n%b"))
			if err != nil {
				return err
			}
			cmd := exec.Command("go", "run", "./enterprise/dev/ci/gen-pipeline.go", "-preview")
			cmd.Env = append(os.Environ(),
				fmt.Sprintf("BUILDKITE_BRANCH=%s", branch),
				fmt.Sprintf("BUILDKITE_MESSAGE=%s", message))
			out, err := run.InRoot(cmd)
			if err != nil {
				return err
			}
			return std.Out.WriteMarkdown(out)
		}),
	}, {
		Name:    "status",
		Aliases: []string{"st"},
		Usage:   "Get the status of the CI run associated with the currently checked out branch",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "wait",
				Aliases: []string{"w"},
				Usage:   "Wait by blocking until the build is finished",
			},
			&cli.BoolFlag{
				Name:    "view",
				Aliases: []string{"v"},
				Usage:   "Open build page in browser",
			},
		},
		Action: func(cmd *cli.Context) error {
			ctx := cmd.Context
			client, err := bk.NewClient(ctx, std.Out.Output)
			if err != nil {
				return err
			}
			branch, branchFromFlag, err := getCIBranch()
			if err != nil {
				return err
			}

			// Just support main pipeline for now
			var build *buildkite.Build
			if ciBuild != "" {
				build, err = client.GetBuildByNumber(ctx, "sourcegraph", ciBuild)
			} else {
				build, err = client.GetMostRecentBuild(ctx, "sourcegraph", branch)
			}
			if err != nil {
				return errors.Newf("failed to get most recent build for branch %q: %w", branch, err)
			}
			// Print a high level overview
			printBuildOverview(build)

			if cmd.Bool("view") {
				if err := open.URL(*build.WebURL); err != nil {
					std.Out.WriteWarningf("failed to open build in browser: %s", err)
				}
			}

			if cmd.Bool("wait") && build.FinishedAt == nil {
				pending := std.Out.Pending(output.Styledf(output.StylePending, "Waiting for %d jobs...", len(build.Jobs)))
				err := statusTicker(ctx, func() (bool, error) {
					// get the next update
					build, err = client.GetMostRecentBuild(ctx, "sourcegraph", branch)
					if err != nil {
						return false, errors.Newf("failed to get most recent build for branch %q: %w", branch, err)
					}
					done := 0
					for _, job := range build.Jobs {
						if job.State != nil {
							if *job.State == "failed" && !job.SoftFailed {
								// If a job has failed, return immediately, we don't have to wait until all
								// steps are completed.
								return true, nil
							}
							if *job.State == "passed" || job.SoftFailed {
								done++
							}
						}
					}

					// once started, poll for status
					if build.StartedAt != nil {
						pending.Updatef("Waiting for %d out of %d jobs... (elapsed: %v)",
							len(build.Jobs)-done, len(build.Jobs), time.Since(build.StartedAt.Time))
					}

					if build.FinishedAt == nil {
						// No failure yet, we can keep waiting.
						return false, nil
					}
					return true, nil
				})
				pending.Destroy()
				if err != nil {
					return err
				}
			}

			// build status finalized
			failed := printBuildResults(build, cmd.Bool("wait"))

			if !branchFromFlag && ciBuild == "" {
				// If we're not on a specific branch and not asking for a specific build, warn if build commit is not your commit
				commit, err := run.GitCmd("rev-parse", "HEAD")
				if err != nil {
					return err
				}
				commit = strings.TrimSpace(commit)
				if commit != *build.Commit {
					std.Out.WriteLine(output.Linef("⚠️", output.StyleSuggestion,
						"The currently checked out commit %q does not match the commit of the build found, %q.\nHave you pushed your most recent changes yet?",
						commit, *build.Commit))
				}
			}

			if failed {
				std.Out.WriteLine(output.Linef(output.EmojiLightbulb, output.StyleSuggestion,
					"Some jobs have failed - try using 'sg ci logs' to see what went wrong, or go to the build page: %s", *build.WebURL))
			}

			return nil
		},
	}, {
		Name:      "build",
		ArgsUsage: "[runtype]",
		Usage:     "Manually request a build for the currently checked out commit and branch (e.g. to trigger builds on forks or with special run types)",
		Description: fmt.Sprintf(`Manually request a Buildkite build for the currently checked out commit and branch. Optionally provide a run type to build with.

This is useful when:

- you want to trigger a build with a particular run type, such as 'main-dry-run'
- triggering builds for PRs from forks (such as those from external contributors), which do not trigger Buildkite builds automatically for security reasons (we do not want to run insecure code on our infrastructure by default!)

Supported run types when providing an argument for 'sg ci build [runtype]':

  %s

For run types that require branch arguments, you will be prompted for an argument, or you
can provide it directly (for example, 'sg ci build [runtype] [argument]').

Learn more about pipeline run types in https://docs.sourcegraph.com/dev/background-information/ci/reference.`,
			strings.Join(getAllowedBuildTypeArgs(), "\n  ")),
		BashComplete: completeOptions(getAllowedBuildTypeArgs),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "commit",
				Aliases: []string{"c"},
				Usage:   "`commit` from the current branch to build (defaults to current commit)",
			},
		},
		Action: func(cmd *cli.Context) error {
			ctx := cmd.Context
			client, err := bk.NewClient(ctx, std.Out.Output)
			if err != nil {
				return err
			}

			branch, err := run.TrimResult(run.GitCmd("branch", "--show-current"))
			if err != nil {
				return err
			}
			commit := cmd.String("commit")
			if commit == "" {
				commit, err = run.TrimResult(run.GitCmd("rev-parse", "HEAD"))
				if err != nil {
					return err
				}
			}

			// 🚨 SECURITY: We do a simple check to see if commit is in origin, this is
			// non blocking but we ask for confirmation to double check that the user
			// is aware that potentially unknown code is going to get run on our infra.
			remoteBranches, err := run.TrimResult(run.GitCmd("branch", "-r", "--contains", commit))
			if err != nil || len(remoteBranches) == 0 || !allLinesPrefixed(strings.Split(remoteBranches, "\n"), "origin/") {
				std.Out.WriteLine(output.Linef(output.EmojiWarning, output.StyleReset,
					"Commit %q not found in in local 'origin/' branches - you might be triggering a build for a fork. Make sure all code has been reviewed before continuing.",
					commit))
				response, err := open.Prompt("Continue? (yes/no)")
				if err != nil {
					return err
				}
				if response != "yes" {
					return errors.New("Cancelling request.")
				}
			}

			var rt runtype.RunType
			if cmd.NArg() == 0 {
				rt = runtype.PullRequest
			} else {
				rt = runtype.Compute("", fmt.Sprintf("%s/%s", cmd.Args().First(), branch), nil)
				// If a special runtype is not detected then the argument was invalid
				if rt == runtype.PullRequest {
					std.Out.WriteFailuref("Unsupported runtype %q", cmd.Args().First())
					std.Out.Writef("Supported runtypes:\n\n\t%s\n\nSee 'sg ci docs' to learn more.", strings.Join(getAllowedBuildTypeArgs(), ", "))
					return NewEmptyExitErr(1)
				}
			}
			if rt != runtype.PullRequest {
				m := rt.Matcher()
				if m.BranchArgumentRequired {
					var branchArg string
					if cmd.NArg() >= 2 {
						branchArg = cmd.Args().Get(1)
					} else {
						std.Out.Write("This run type requires a branch path argument.")
						branchArg, err = open.Prompt("Enter your argument input:")
						if err != nil {
							return err
						}
					}
					branch = fmt.Sprintf("%s/%s", branchArg, branch)
				}

				branch = fmt.Sprintf("%s%s", rt.Matcher().Branch, branch)
				block := std.Out.Block(output.Line("", output.StylePending, fmt.Sprintf("Pushing %s to %s...", commit, branch)))
				gitOutput, err := run.GitCmd("push", "origin", fmt.Sprintf("%s:refs/heads/%s", commit, branch), "--force")
				if err != nil {
					return err
				}
				block.WriteLine(output.Line("", output.StyleSuggestion, strings.TrimSpace(gitOutput)))
				block.Close()
			}

			pipeline := "sourcegraph"
			var build *buildkite.Build
			if rt != runtype.PullRequest {
				pollTicker := time.NewTicker(5 * time.Second)
				std.Out.WriteLine(output.Styledf(output.StylePending, "Polling for build for branch %s at %s...", branch, commit))
				for i := 0; i < 30; i++ {
					// attempt to fetch the new build - it might take some time for the hooks so we will
					// retry up to 30 times (roughly 30 seconds)
					if build != nil && build.Commit != nil && *build.Commit == commit {
						break
					}
					<-pollTicker.C
					build, err = client.GetMostRecentBuild(ctx, pipeline, branch)
					if err != nil {
						return errors.Wrap(err, "GetMostRecentBuild")
					}
				}
			} else {
				std.Out.WriteLine(output.Styledf(output.StylePending, "Requesting build for branch %q at %q...", branch, commit))
				build, err = client.TriggerBuild(ctx, pipeline, branch, commit)
				if err != nil {
					return errors.Newf("failed to trigger build for branch %q at %q: %w", branch, commit, err)
				}
			}

			std.Out.WriteLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Created build: %s", *build.WebURL))
			return nil
		},
	}, {
		Name:  "logs",
		Usage: "Get logs from CI builds (e.g. to grep locally)",
		Description: `Get logs from CI builds, and output them in stdout or push them to Loki. By default only gets failed jobs - to change this, use the '--state' flag.

The '--job' flag can be used to narrow down the logs returned - you can provide either the ID, or part of the name of the job you want to see logs for.

To send logs to a Loki instance, you can provide '--out=http://127.0.0.1:3100' after spinning up an instance with 'sg run loki grafana'.
From there, you can start exploring logs with the Grafana explore panel.
`,
		Flags: []cli.Flag{
			&ciBuildFlag,
			&cli.StringFlag{
				Name:    "job",
				Aliases: []string{"j"},
				Usage:   "ID or name of the job to export logs for",
			},
			&cli.StringFlag{
				Name:    "state",
				Aliases: []string{"s"},
				Usage:   "Job `state` to export logs for (provide an empty value for all states)",
				Value:   "failed",
			},
			&cli.StringFlag{
				Name:    "out",
				Aliases: []string{"o"},
				Usage: fmt.Sprintf("Output `format`: one of %+v, or a URL pointing to a Loki instance, such as %q",
					[]string{ciLogsOutTerminal, ciLogsOutSimple, ciLogsOutJSON}, loki.DefaultLokiURL),
				Value: ciLogsOutTerminal,
			},
			&cli.StringFlag{
				Name:  "overwrite-state",
				Usage: "`state` to overwrite the job state metadata",
			},
		},
		Action: func(cmd *cli.Context) error {
			ctx := cmd.Context
			client, err := bk.NewClient(ctx, std.Out.Output)
			if err != nil {
				return err
			}

			branch, _, err := getCIBranch()
			if err != nil {
				return err
			}

			var build *buildkite.Build
			if ciBuild != "" {
				build, err = client.GetBuildByNumber(ctx, "sourcegraph", ciBuild)
			} else {
				build, err = client.GetMostRecentBuild(ctx, "sourcegraph", branch)
			}
			if err != nil {
				return errors.Newf("failed to get most recent build for branch %q: %w", branch, err)
			}
			std.Out.WriteLine(output.Styledf(output.StylePending, "Fetching logs for %s ...",
				*build.WebURL))

			options := bk.ExportLogsOpts{
				JobQuery: cmd.String("job"),
				State:    cmd.String("state"),
			}
			logs, err := client.ExportLogs(ctx, "sourcegraph", *build.Number, options)
			if err != nil {
				return err
			}
			if len(logs) == 0 {
				std.Out.WriteLine(output.Line("", output.StyleSuggestion,
					fmt.Sprintf("No logs found matching the given parameters (job: %q, state: %q).", options.JobQuery, options.State)))
				return nil
			}

			logsOut := cmd.String("out")
			switch logsOut {
			case ciLogsOutTerminal, ciLogsOutSimple:
				// Buildkite's timestamp thingo causes log lines to not render in terminal
				bkTimestamp := regexp.MustCompile(`\x1b_bk;t=\d{13}\x07`) // \x1b is ESC, \x07 is BEL
				for _, log := range logs {
					block := std.Out.Block(output.Linef(output.EmojiInfo, output.StyleUnderline, "%s",
						*log.JobMeta.Name))
					content := bkTimestamp.ReplaceAllString(*log.Content, "")
					if logsOut == ciLogsOutSimple {
						content = bk.CleanANSI(content)
					}
					block.Write(content)
					block.Close()
				}
				std.Out.WriteLine(output.Styledf(output.StyleSuccess, "Found and output logs for %d jobs.", len(logs)))

			case ciLogsOutJSON:
				for _, log := range logs {
					if logsOut != "" {
						failed := logsOut
						log.JobMeta.State = &failed
					}
					stream, err := loki.NewStreamFromJobLogs(log)
					if err != nil {
						return errors.Newf("build %d job %s: NewStreamFromJobLogs: %s", log.JobMeta.Build, log.JobMeta.Job, err)
					}
					b, err := json.MarshalIndent(stream, "", "\t")
					if err != nil {
						return errors.Newf("build %d job %s: Marshal: %s", log.JobMeta.Build, log.JobMeta.Job, err)
					}
					std.Out.Write(string(b))
				}

			default:
				lokiURL, err := url.Parse(logsOut)
				if err != nil {
					return errors.Newf("invalid Loki target: %w", err)
				}
				lokiClient := loki.NewLokiClient(lokiURL)
				std.Out.WriteLine(output.Styledf(output.StylePending, "Pushing to Loki instance at %q", lokiURL.Host))

				var (
					pushedEntries int
					pushedStreams int
					pushErrs      []string
					pending       = std.Out.Pending(output.Styled(output.StylePending, "Processing logs..."))
				)
				for i, log := range logs {
					job := log.JobMeta.Job
					if log.JobMeta.Label != nil {
						job = fmt.Sprintf("%q (%s)", *log.JobMeta.Label, log.JobMeta.Job)
					}
					overwriteState := cmd.String("overwrite-state")
					if overwriteState != "" {
						failed := overwriteState
						log.JobMeta.State = &failed
					}

					pending.Updatef("Processing build %d job %s (%d/%d)...",
						log.JobMeta.Build, job, i, len(logs))
					stream, err := loki.NewStreamFromJobLogs(log)
					if err != nil {
						pushErrs = append(pushErrs, fmt.Sprintf("build %d job %s: %s",
							log.JobMeta.Build, job, err))
						continue
					}

					// Set buildkite metadata if available
					if ciBranch := os.Getenv("BUILDKITE_BRANCH"); ciBranch != "" {
						stream.Stream.Branch = ciBranch
					}
					if ciQueue := os.Getenv("BUILDKITE_AGENT_META_DATA_QUEUE"); ciQueue != "" {
						stream.Stream.Queue = ciQueue
					}

					err = lokiClient.PushStreams(ctx, []*loki.Stream{stream})
					if err != nil {
						pushErrs = append(pushErrs, fmt.Sprintf("build %d job %q: %s",
							log.JobMeta.Build, job, err))
						continue
					}

					pushedEntries += len(stream.Values)
					pushedStreams += 1
				}

				if pushedEntries > 0 {
					pending.Complete(output.Linef(output.EmojiSuccess, output.StyleSuccess,
						"Pushed %d entries from %d streams to Loki", pushedEntries, pushedStreams))
				} else {
					pending.Destroy()
				}

				if pushErrs != nil {
					failedStreams := len(logs) - pushedStreams
					std.Out.WriteLine(output.Linef(output.EmojiFailure, output.StyleWarning,
						"Failed to push %d streams: \n - %s", failedStreams, strings.Join(pushErrs, "\n - ")))
					if failedStreams == len(logs) {
						return errors.New("failed to push all logs")
					}
				}
			}

			return nil
		},
	}, {
		Name:        "docs",
		Usage:       "Render reference documentation for build pipeline types",
		Description: "Render reference documentation for build pipeline types. An online version of this is also available in https://docs.sourcegraph.com/dev/background-information/ci/reference.",
		Action: execAdapter(func(ctx context.Context, args []string) error {
			cmd := exec.Command("go", "run", "./enterprise/dev/ci/gen-pipeline.go", "-docs")
			out, err := run.InRoot(cmd)
			if err != nil {
				return err
			}
			return std.Out.WriteMarkdown(out)
		}),
	}, {
		Name:        "open",
		ArgsUsage:   "[pipeline]",
		Usage:       "Open Sourcegraph's Buildkite page in browser",
		Description: "Open Sourcegraph's Buildkite page in browser. Optionally specify the pipeline you want to open.",
		Action: execAdapter(func(ctx context.Context, args []string) error {
			buildkiteURL := fmt.Sprintf("https://buildkite.com/%s", bk.BuildkiteOrg)
			if len(args) > 0 && args[0] != "" {
				pipeline := args[0]
				buildkiteURL += fmt.Sprintf("/%s", pipeline)
			}
			return open.URL(buildkiteURL)
		}),
	}},
}

func getAllowedBuildTypeArgs() []string {
	var results []string
	for _, rt := range runtype.RunTypes() {
		if rt.Matcher().IsBranchPrefixMatcher() {
			results = append(results, strings.TrimSuffix(rt.Matcher().Branch, "/"))
		}
	}
	return results
}

func allLinesPrefixed(lines []string, match string) bool {
	for _, l := range lines {
		if !strings.HasPrefix(strings.TrimSpace(l), match) {
			return false
		}
	}
	return true
}

func printBuildOverview(build *buildkite.Build) {
	std.Out.WriteLine(output.Styledf(output.StyleBold, "Most recent build: %s", *build.WebURL))
	std.Out.Writef("Commit:\t\t%s\nMessage:\t%s\nAuthor:\t\t%s <%s>",
		*build.Commit, *build.Message, build.Author.Name, build.Author.Email)
	if build.PullRequest != nil {
		std.Out.Writef("PR:\t\thttps://github.com/sourcegraph/sourcegraph/pull/%s", *build.PullRequest.ID)
	}
}

func printBuildResults(build *buildkite.Build, notify bool) (failed bool) {
	std.Out.Writef("Started:\t%s", build.StartedAt)
	if build.FinishedAt != nil {
		std.Out.Writef("Finished:\t%s (elapsed: %s)", build.FinishedAt, build.FinishedAt.Sub(build.StartedAt.Time))
	}

	// Check build state
	// Valid states: running, scheduled, passed, failed, blocked, canceled, canceling, skipped, not_run, waiting
	// https://buildkite.com/docs/apis/rest-api/builds
	var style output.Style
	var emoji string
	switch *build.State {
	case "passed":
		style = output.StyleSuccess
		emoji = output.EmojiSuccess
	case "waiting", "blocked", "scheduled":
		style = output.StyleSuggestion
	case "skipped", "not_run":
		style = output.StyleReset
	case "running":
		style = output.StylePending
		emoji = output.EmojiInfo
	case "failed":
		failed = true
		emoji = output.EmojiFailure
		fallthrough
	default:
		style = output.StyleWarning
	}
	block := std.Out.Block(output.Styledf(style, "Status:\t\t%s %s", emoji, *build.State))

	// Inspect jobs individually.
	failedSummary := []string{"Failed jobs:"}
	for _, job := range build.Jobs {
		var elapsed time.Duration
		if job.State == nil || job.Name == nil {
			continue
		}
		// Check job state.
		switch *job.State {
		case "passed":
			style = output.StyleSuccess
			elapsed = job.FinishedAt.Sub(job.StartedAt.Time)
		case "waiting", "blocked", "scheduled", "assigned":
			style = output.StyleSuggestion
		case "broken":
			// State 'broken' happens when a conditional is not met, namely the 'if' block
			// on a job. Why is it 'broken' and not 'skipped'? We don't think it be like
			// this, but it do. Anyway, we pretend it was skipped and treat it as such.
			// https://buildkite.com/docs/pipelines/conditionals#conditionals-and-the-broken-state
			*job.State = "skipped"
			fallthrough
		case "skipped", "not_run":
			style = output.StyleReset
		case "running":
			elapsed = time.Since(job.StartedAt.Time)
			style = output.StylePending
		case "failed":
			elapsed = job.FinishedAt.Sub(job.StartedAt.Time)
			if job.SoftFailed {
				*job.State = "soft failed"
				style = output.StyleReset
				break
			}
			failedSummary = append(failedSummary, fmt.Sprintf("- %s", *job.Name))
			failed = true
			fallthrough
		default:
			style = output.StyleWarning
		}
		if elapsed > 0 {
			block.WriteLine(output.Styledf(style, "- [%s] %s (%s)", *job.State, *job.Name, elapsed))
		} else {
			block.WriteLine(output.Styledf(style, "- [%s] %s", *job.State, *job.Name))
		}
	}

	block.Close()

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
	ticker := time.NewTicker(30 * time.Second)
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
