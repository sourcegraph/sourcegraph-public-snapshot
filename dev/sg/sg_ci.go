package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/cockroachdb/errors"
	"github.com/gen2brain/beeep"
	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/bk"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/loki"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/open"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

const (
	ciLogsOutStdout = "stdout"
)

var (
	ciFlagSet = flag.NewFlagSet("sg ci", flag.ExitOnError)

	ciLogsFlagSet               = flag.NewFlagSet("sg ci logs", flag.ExitOnError)
	ciLogsBranchFlag            = ciLogsFlagSet.String("branch", "", "Branch name of build to find logs for (defaults to current branch)")
	ciLogsJobStateFlag          = ciLogsFlagSet.String("state", "failed", "Job states to export logs for.")
	ciLogsJobOverwriteStateFlag = ciLogsFlagSet.String("overwrite-state", "", "State to overwrite the job state metadata")
	ciLogsJobQueryFlag          = ciLogsFlagSet.String("job", "", "ID or name of the job to export logs for.")
	ciLogsBuildFlag             = ciLogsFlagSet.String("build", "", "Override branch detection with a specific build number")
	ciLogsOutFlag               = ciLogsFlagSet.String("out", ciLogsOutStdout,
		fmt.Sprintf("Output format, either 'stdout' or a URL pointing to a Loki instance, such as %q", loki.DefaultLokiURL))

	ciStatusFlagSet    = flag.NewFlagSet("sg ci status", flag.ExitOnError)
	ciStatusBranchFlag = ciStatusFlagSet.String("branch", "", "Branch name of build to check build status for (defaults to current branch)")
	ciStatusWaitFlag   = ciStatusFlagSet.Bool("wait", false, "Wait by blocking until the build is finished.")
	ciStatusBuildFlag  = ciStatusFlagSet.String("build", "", "Override branch detection with a specific build number")

	ciBuildFlagSet    = flag.NewFlagSet("sg ci build", flag.ExitOnError)
	ciBuildCommitFlag = ciBuildFlagSet.String("commit", "", "commit from the current branch to build (defaults to current commit)")
)

// get branch from flag or git
func getCIBranch() (branch string, fromFlag bool, err error) {
	fromFlag = true
	switch {
	case *ciLogsBranchFlag != "":
		branch = *ciLogsBranchFlag
	case *ciStatusBranchFlag != "":
		branch = *ciStatusBranchFlag
	default:
		branch, err = run.TrimResult(run.GitCmd("branch", "--show-current"))
		fromFlag = false
	}
	return
}

var (
	ciCommand = &ffcli.Command{
		Name:       "ci",
		ShortUsage: "sg ci [preview|status|build|logs]",
		ShortHelp:  "Interact with Sourcegraph's continuous integration pipelines",
		LongHelp: `Interact with Sourcegraph's continuous integration pipelines on Buildkite.

Note that Sourcegraph's CI pipelines are under our enterprise license: https://github.com/sourcegraph/sourcegraph/blob/main/LICENSE.enterprise`,
		FlagSet: ciFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
		Subcommands: []*ffcli.Command{{
			Name:      "preview",
			ShortHelp: "Preview the pipeline that would be run against the currently checked out branch",
			Exec: func(ctx context.Context, args []string) error {
				stdout.Out.WriteLine(output.Linef("", output.StyleSuggestion,
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
				stdout.Out.Write(out)
				return nil
			},
		}, {
			Name:      "status",
			ShortHelp: "Get the status of the CI run associated with the currently checked out branch",
			FlagSet:   ciStatusFlagSet,
			Exec: func(ctx context.Context, args []string) error {
				client, err := bk.NewClient(ctx, stdout.Out)
				if err != nil {
					return err
				}
				branch, branchFromFlag, err := getCIBranch()
				if err != nil {
					return err
				}

				// Just support main pipeline for now
				var build *buildkite.Build
				if *ciStatusBuildFlag != "" {
					build, err = client.GetBuildByNumber(ctx, "sourcegraph", *ciStatusBuildFlag)
				} else {
					build, err = client.GetMostRecentBuild(ctx, "sourcegraph", branch)
				}
				if err != nil {
					return fmt.Errorf("failed to get most recent build for branch %q: %w", branch, err)
				}
				// Print a high level overview
				printBuildOverview(build)

				if *ciStatusWaitFlag && build.FinishedAt == nil {
					pending := stdout.Out.Pending(output.Linef("", output.StylePending, "Waiting for %d jobs...", len(build.Jobs)))
					err := statusTicker(ctx, func() (bool, error) {
						// get the next update
						build, err = client.GetMostRecentBuild(ctx, "sourcegraph", branch)
						if err != nil {
							return false, fmt.Errorf("failed to get most recent build for branch %q: %w", branch, err)
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
				failed := printBuildResults(build, *ciStatusWaitFlag)

				if !branchFromFlag && *ciStatusBuildFlag == "" {
					// If we're not on a specific branch and not asking for a specific build, warn if build commit is not your commit
					commit, err := run.GitCmd("rev-parse", "HEAD")
					if err != nil {
						return err
					}
					commit = strings.TrimSpace(commit)
					if commit != *build.Commit {
						stdout.Out.WriteLine(output.Linef("⚠️", output.StyleSuggestion,
							"The currently checked out commit %q does not match the commit of the build found, %q.\nHave you pushed your most recent changes yet?",
							commit, *build.Commit))
					}
				}

				if failed {
					stdout.Out.WriteLine(output.Linef(output.EmojiLightbulb, output.StyleSuggestion,
						"Some jobs have failed - try using 'sg ci logs' to see what went wrong, or go to the build page: %s", *build.WebURL))
				}

				return nil
			},
		}, {
			Name:      "build",
			FlagSet:   ciBuildFlagSet,
			ShortHelp: "Manually request a build for the currently checked out commit and branch (e.g. to trigger builds on forks)",
			LongHelp:  "Manually request a Buildkite build for the currently checked out commit and branch. This is most useful when triggering builds for PRs from forks (such as those from external contributors), which do not trigger Buildkite builds automatically for security reasons (we do not want to run insecure code on our infrastructure by default!)",
			Exec: func(ctx context.Context, args []string) error {
				client, err := bk.NewClient(ctx, stdout.Out)
				if err != nil {
					return err
				}

				branch, err := run.TrimResult(run.GitCmd("branch", "--show-current"))
				if err != nil {
					return err
				}

				commit := *ciBuildCommitFlag
				if commit == "" {
					commit, err = run.TrimResult(run.GitCmd("rev-parse", "HEAD"))
					if err != nil {
						return err
					}
				}

				stdout.Out.WriteLine(output.Linef("", output.StylePending, "Requesting build for branch %q at %q...", branch, commit))

				// simple check to see if commit is in origin, this is non blocking but
				// we ask for confirmation to double check.
				remoteBranches, err := run.TrimResult(run.GitCmd("branch", "-r", "--contains", commit))
				if err != nil || len(remoteBranches) == 0 || !allLinesPrefixed(strings.Split(remoteBranches, "\n"), "origin/") {
					stdout.Out.WriteLine(output.Linef(output.EmojiWarning, output.StyleReset,
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

				build, err := client.TriggerBuild(ctx, "sourcegraph", branch, commit)
				if err != nil {
					return fmt.Errorf("failed to trigger build for branch %q at %q: %w", branch, commit, err)
				}
				stdout.Out.WriteLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Created build: %s", *build.WebURL))
				return nil
			},
		}, {
			Name:      "logs",
			ShortHelp: "Get logs from CI builds.",
			LongHelp: `Get logs from CI builds, and output them in stdout or push them to Loki. By default only gets failed jobs - to change this, use the '--state' flag.

The '--job' flag can be used to narrow down the logs returned - you can provide either the ID, or part of the name of the job you want to see logs for.

To send logs to a Loki instance, you can provide '--out=http://127.0.0.1:3100' after spinning up an instance with 'sg run loki grafana'.
From there, you can start exploring logs with the Grafana explore panel.
`,
			FlagSet: ciLogsFlagSet,
			Exec: func(ctx context.Context, args []string) error {
				client, err := bk.NewClient(ctx, stdout.Out)
				if err != nil {
					return err
				}

				branch, _, err := getCIBranch()
				if err != nil {
					return err
				}

				var build *buildkite.Build
				if *ciLogsBuildFlag != "" {
					build, err = client.GetBuildByNumber(ctx, "sourcegraph", *ciLogsBuildFlag)
				} else {
					build, err = client.GetMostRecentBuild(ctx, "sourcegraph", branch)
				}
				if err != nil {
					return fmt.Errorf("failed to get most recent build for branch %q: %w", branch, err)
				}
				stdout.Out.WriteLine(output.Linef("", output.StylePending, "Fetching logs for %s ...",
					*build.WebURL))

				options := bk.ExportLogsOpts{
					JobQuery: *ciLogsJobQueryFlag,
					State:    *ciLogsJobStateFlag,
				}
				logs, err := client.ExportLogs(ctx, "sourcegraph", *build.Number, options)
				if err != nil {
					return err
				}
				if len(logs) == 0 {
					stdout.Out.WriteLine(output.Line("", output.StyleSuggestion,
						fmt.Sprintf("No logs found matching the given parameters (job: %q, state: %q).", options.JobQuery, options.State)))
					return nil
				}

				switch *ciLogsOutFlag {
				case ciLogsOutStdout:
					// Buildkite's timestamp thingo causes log lines to not render in terminal
					bkTimestamp := regexp.MustCompile(`\x1b_bk;t=\d{13}\x07`) // \x1b is ESC, \x07 is BEL
					for _, log := range logs {
						block := stdout.Out.Block(output.Linef(output.EmojiInfo, output.StyleUnderline, "%s",
							*log.JobMeta.Name))
						block.Write(bkTimestamp.ReplaceAllString(*log.Content, ""))
						block.Close()
					}
					stdout.Out.WriteLine(output.Linef("", output.StyleSuccess, "Found and output logs for %d jobs.", len(logs)))

				default:
					lokiURL, err := url.Parse(*ciLogsOutFlag)
					if err != nil {
						return fmt.Errorf("invalid Loki target: %w", err)
					}
					lokiClient := loki.NewLokiClient(lokiURL)
					stdout.Out.WriteLine(output.Linef("", output.StylePending, "Pushing to Loki instance at %q", lokiURL.Host))

					var (
						pushedEntries int
						pushedStreams int
						pushErrs      []string
						pending       = stdout.Out.Pending(output.Linef("", output.StylePending, "Processing logs..."))
					)
					for i, log := range logs {
						job := log.JobMeta.Job
						if log.JobMeta.Label != nil {
							job = fmt.Sprintf("%q (%s)", *log.JobMeta.Label, log.JobMeta.Job)
						}
						if *ciLogsJobOverwriteStateFlag != "" {
							failed := *ciLogsJobOverwriteStateFlag
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

						// Set buildkite branch if available
						if ciBranch := os.Getenv("BUILDKITE_BRANCH"); ciBranch != "" {
							stream.Stream.Branch = ciBranch
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
						stdout.Out.WriteLine(output.Linef(output.EmojiFailure, output.StyleWarning,
							"Failed to push %d streams: \n - %s", failedStreams, strings.Join(pushErrs, "\n - ")))
						if failedStreams == len(logs) {
							return errors.New("failed to push all logs")
						}
					}
				}

				return nil
			},
		}},
	}
)

func allLinesPrefixed(lines []string, match string) bool {
	for _, l := range lines {
		if !strings.HasPrefix(strings.TrimSpace(l), match) {
			return false
		}
	}
	return true
}

func printBuildOverview(build *buildkite.Build) {
	stdout.Out.WriteLine(output.Linef("", output.StyleBold, "Most recent build: %s", *build.WebURL))
	stdout.Out.Writef("Commit:\t\t%s\nMessage:\t%s\nAuthor:\t\t%s <%s>",
		*build.Commit, *build.Message, build.Author.Name, build.Author.Email)
}

func printBuildResults(build *buildkite.Build, notify bool) (failed bool) {
	stdout.Out.Writef("Started:\t%s", build.StartedAt)
	if build.FinishedAt != nil {
		stdout.Out.Writef("Finished:\t%s (elapsed: %s)", build.FinishedAt, build.FinishedAt.Sub(build.StartedAt.Time))
	}

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
	block := stdout.Out.Block(output.Linef("", style, "Status:\t\t%s %s", emoji, *build.State))

	// Inspect jobs individually.
	failedSummary := []string{"Failed jobs:"}
	for _, job := range build.Jobs {
		var elapsed time.Duration
		if job.State == nil || job.Name == nil {
			continue
		}
		switch *job.State {
		case "passed":
			style = output.StyleSuccess
			elapsed = job.FinishedAt.Sub(job.StartedAt.Time)
		case "waiting", "blocked", "scheduled":
			style = output.StyleSuggestion
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
			block.WriteLine(output.Linef("", style, "- [%s] %s (%s)", *job.State, *job.Name, elapsed))
		} else {
			block.WriteLine(output.Linef("", style, "- [%s] %s", *job.State, *job.Name))
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
			return fmt.Errorf("polling timeout reached")
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
