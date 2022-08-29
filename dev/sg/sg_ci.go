package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/gen2brain/beeep"
	"github.com/grafana/regexp"
	"github.com/urfave/cli/v2"

	sgrun "github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/ci/runtype"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/bk"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/loki"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/open"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/usershell"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

const (
	ciLogsOutTerminal = "terminal"
	ciLogsOutSimple   = "simple"
	ciLogsOutJSON     = "json"
)

var (
	ciBranchFlag = cli.StringFlag{
		Name:    "branch",
		Aliases: []string{"b"},
		Usage:   "Branch `name` of build to target (defaults to current branch)",
	}
	ciBuildFlag = cli.StringFlag{
		Name:    "build",
		Aliases: []string{"n"}, // 'n' for number, because 'b' is taken
		Usage:   "Override branch detection with a specific build `number`",
	}
	ciCommitFlag = cli.StringFlag{
		Name:    "commit",
		Aliases: []string{"c"},
		Usage:   "Override branch detection with the latest build for `commit`",
	}
	ciPipelineFlag = cli.StringFlag{
		Name:    "pipeline",
		Aliases: []string{"p"},
		EnvVars: []string{"SG_CI_PIPELINE"},
		Usage:   "Select a custom Buildkite `pipeline` in the Sourcegraph org",
		Value:   "sourcegraph",
	}
)

// Register the following flags on all commands that can target different builds!
var ciTargetFlags = []cli.Flag{
	&ciBranchFlag,
	&ciBuildFlag,
	&ciCommitFlag,
	&ciPipelineFlag,
}

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

var ciCommand = &cli.Command{
	Name:        "ci",
	Usage:       "Interact with Sourcegraph's Buildkite continuous integration pipelines",
	Description: `Note that Sourcegraph's CI pipelines are under our enterprise license: https://github.com/sourcegraph/sourcegraph/blob/main/LICENSE.enterprise`,
	UsageText: `
# Preview what a CI run for your current changes will look like
sg ci preview

# Check on the status of your changes on the current branch in the Buildkite pipeline
sg ci status
# Check on the status of a specific branch instead
sg ci status --branch my-branch
# Block until the build has completed (it will send a system notification)
sg ci status --wait
# Get status for a specific build number
sg ci status --build 123456

# Pull logs of failed jobs to stdout
sg ci logs
# Push logs of most recent main failure to local Loki for analysis
# You can spin up a Loki instance with 'sg run loki grafana'
sg ci logs --branch main --out http://127.0.0.1:3100
# Get the logs for a specific build number, useful when debugging
sg ci logs --build 123456

# Manually trigger a build on the CI with the current branch
sg ci build
# Manually trigger a build on the CI on the current branch, but with a specific commit
sg ci build --commit my-commit
# Manually trigger a main-dry-run build of the HEAD commit on the current branch
sg ci build main-dry-run
sg ci build --force main-dry-run
# Manually trigger a main-dry-run build of a specified commit on the current ranch
sg ci build --force --commit my-commit main-dry-run
# View the available special build types
sg ci build --help
`,
	Category: CategoryDev,
	Subcommands: []*cli.Command{{
		Name:    "preview",
		Aliases: []string{"plan"},
		Usage:   "Preview the pipeline that would be run against the currently checked out branch",
		Flags: []cli.Flag{
			&ciBranchFlag,
		},
		Action: func(cmd *cli.Context) error {
			std.Out.WriteLine(output.Styled(output.StyleSuggestion,
				"If the current branch were to be pushed, the following pipeline would be run:"))

			target, err := getBuildTarget(cmd)
			if err != nil {
				return err
			}
			if target.targetType != buildTargetTypeBranch {
				// Should never happen because we only register the branch flag
				return errors.New("target is not a branch")
			}

			message, err := run.TrimResult(run.GitCmd("show", "--format=%s\\n%b"))
			if err != nil {
				return err
			}

			previewCmd := usershell.Command(cmd.Context, "go run ./enterprise/dev/ci/gen-pipeline.go -preview").
				Env(map[string]string{
					"BUILDKITE_BRANCH":  target.target, // this must be a branch
					"BUILDKITE_MESSAGE": message,
				})
			out, err := root.Run(previewCmd).String()
			if err != nil {
				return err
			}
			return std.Out.WriteMarkdown(out)
		},
	}, {
		Name:    "status",
		Aliases: []string{"st"},
		Usage:   "Get the status of the CI run associated with the currently checked out branch",
		Flags: append(ciTargetFlags,
			&cli.BoolFlag{
				Name:    "wait",
				Aliases: []string{"w"},
				Usage:   "Wait by blocking until the build is finished",
			},
			&cli.BoolFlag{
				Name:    "view",
				Aliases: []string{"v"},
				Usage:   "Open build page in browser",
			}),
		Action: func(cmd *cli.Context) error {
			client, err := bk.NewClient(cmd.Context, std.Out)
			if err != nil {
				return err
			}
			target, err := getBuildTarget(cmd)
			if err != nil {
				return err
			}

			// Just support main pipeline for now
			build, err := target.GetBuild(cmd.Context, client)
			if err != nil {
				return err
			}

			// Print a high level overview, and jump into a browser
			printBuildOverview(build)
			if cmd.Bool("view") {
				if err := open.URL(*build.WebURL); err != nil {
					std.Out.WriteWarningf("failed to open build in browser: %s", err)
				}
			}

			// If we are waiting and unfinished, poll for a build
			if cmd.Bool("wait") && build.FinishedAt == nil {
				if build.Branch == nil {
					return errors.Newf("build %d not associated with a branch", *build.Number)
				}

				pending := std.Out.Pending(output.Styledf(output.StylePending, "Waiting for %d jobs...", len(build.Jobs)))
				err := statusTicker(cmd.Context, func() (bool, error) {
					// get the next update for this specific build
					build, err = client.GetBuildByNumber(cmd.Context, target.pipeline, strconv.Itoa(*build.Number))
					if err != nil {
						return false, errors.Newf("failed to get most recent build for branch %q: %w", *build.Branch, err)
					}

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
				})
				pending.Destroy()
				if err != nil {
					return err
				}
			}

			// lets get annotations (if any) for the build
			var annotations bk.JobAnnotations
			annotations, err = client.GetJobAnnotationsByBuildNumber(cmd.Context, "sourcegraph", strconv.Itoa(*build.Number))
			if err != nil {
				return errors.Newf("failed to get annotations for build %d: %w", *build.Number, err)
			}

			// render resutls
			failed := printBuildResults(build, annotations, cmd.Bool("wait"))

			// If we're not on a specific branch and not asking for a specific build,
			// warn if build commit is not your local copy - we are building an
			// unknown revision.
			if !target.fromFlag && target.targetType == buildTargetTypeBranch {
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
		Description: fmt.Sprintf(`Optionally provide a run type to build with.

This command is useful when:

- you want to trigger a build with a particular run type, such as 'main-dry-run'
- triggering builds for PRs from forks (such as those from external contributors), which do not trigger Buildkite builds automatically for security reasons (we do not want to run insecure code on our infrastructure by default!)

Supported run types when providing an argument for 'sg ci build [runtype]':

* %s

For run types that require branch arguments, you will be prompted for an argument, or you
can provide it directly (for example, 'sg ci build [runtype] [argument]').

Learn more about pipeline run types in https://docs.sourcegraph.com/dev/background-information/ci/reference.`,
			strings.Join(getAllowedBuildTypeArgs(), "\n* ")),
		BashComplete: completeOptions(getAllowedBuildTypeArgs),
		Flags: []cli.Flag{
			&ciPipelineFlag,
			&cli.StringFlag{
				Name:    "commit",
				Aliases: []string{"c"},
				Usage:   "`commit` from the current branch to build (defaults to current commit)",
			},
		},
		Action: func(cmd *cli.Context) error {
			ctx := cmd.Context
			client, err := bk.NewClient(ctx, std.Out)
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
			if !repo.HasCommit(ctx, commit) {
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

			var (
				pipeline = ciPipelineFlag.Get(cmd)
				build    *buildkite.Build
			)
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

To send logs to a Loki instance, you can provide --out=http://127.0.0.1:3100 after spinning up an instance with 'sg run loki grafana'.
From there, you can start exploring logs with the Grafana explore panel.
`,
		Flags: append(ciTargetFlags,
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
				Usage: fmt.Sprintf("Output `format`: one of [%s], or a URL pointing to a Loki instance, such as %s",
					strings.Join([]string{ciLogsOutTerminal, ciLogsOutSimple, ciLogsOutJSON}, "|"), loki.DefaultLokiURL),
				Value: ciLogsOutTerminal,
			},
			&cli.StringFlag{
				Name:  "overwrite-state",
				Usage: "`state` to overwrite the job state metadata",
			},
		),
		Action: func(cmd *cli.Context) error {
			ctx := cmd.Context
			client, err := bk.NewClient(ctx, std.Out)
			if err != nil {
				return err
			}

			target, err := getBuildTarget(cmd)
			if err != nil {
				return err
			}

			build, err := target.GetBuild(ctx, client)
			if err != nil {
				return err
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
		Description: "An online version of the rendered documentation is also available in https://docs.sourcegraph.com/dev/background-information/ci/reference.",
		Action: func(ctx *cli.Context) error {
			cmd := exec.Command("go", "run", "./enterprise/dev/ci/gen-pipeline.go", "-docs")
			out, err := run.InRoot(cmd)
			if err != nil {
				return err
			}
			return std.Out.WriteMarkdown(out)
		},
	}, {
		Name:      "open",
		ArgsUsage: "[pipeline]",
		Usage:     "Open Sourcegraph's Buildkite page in browser",
		Action: func(ctx *cli.Context) error {
			buildkiteURL := fmt.Sprintf("https://buildkite.com/%s", bk.BuildkiteOrg)
			args := ctx.Args().Slice()
			if len(args) > 0 && args[0] != "" {
				pipeline := args[0]
				buildkiteURL += fmt.Sprintf("/%s", pipeline)
			}
			return open.URL(buildkiteURL)
		},
	}, {
		Name:      "search-failures",
		ArgsUsage: "[text to search for]",
		Usage:     "Open Sourcegraph's CI failures Grafana logs page in browser",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "step",
				Usage: "Filter by step name (--step STEP_NAME will translate to '.*STEP_NAME.*')",
			},
		},
		Action: func(ctx *cli.Context) error {
			text := "TODO"
			stepName := ctx.String("step")

			if ctx.Args().Len() > 0 {
				text = ctx.Args().Slice()[0]
			}
			grafanaURL := buildGrafanaURL(text, stepName)
			return open.URL(grafanaURL)
		},
	}},
}

func buildGrafanaURL(text string, stepName string) string {
	var base string
	if stepName == "" {
		base = "https://sourcegraph.grafana.net/explore?orgId=1&left=%7B%22datasource%22:%22grafanacloud-sourcegraph-logs%22,%22queries%22:%5B%7B%22refId%22:%22A%22,%22editorMode%22:%22code%22,%22expr%22:%22%7Bapp%3D%5C%22buildkite%5C%22%7D%20%7C%3D%20%60_TEXT_%60%22,%22queryType%22:%22range%22%7D%5D,%22range%22:%7B%22from%22:%22now-10d%22,%22to%22:%22now%22%7D%7D"
	} else {
		base = "https://sourcegraph.grafana.net/explore?orgId=1&left=%7B%22datasource%22:%22grafanacloud-sourcegraph-logs%22,%22queries%22:%5B%7B%22refId%22:%22A%22,%22editorMode%22:%22code%22,%22expr%22:%22%7Bapp%3D%5C%22buildkite%5C%22,%20step_key%3D~%5C%22_STEP_%5C%22%7D%20%7C%3D%20%60_TEXT_%60%22,%22queryType%22:%22range%22%7D%5D,%22range%22:%7B%22from%22:%22now-10d%22,%22to%22:%22now%22%7D%7D"
	}
	url := strings.ReplaceAll(base, "_TEXT_", text)
	return strings.ReplaceAll(url, "_STEP_", fmt.Sprintf(".*%s.*", stepName))
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

func printBuildResults(build *buildkite.Build, annotations bk.JobAnnotations, notify bool) (failed bool) {
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
		style = output.StyleFailure
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
			style = output.StyleFailure
			failed = true
		default:
			style = output.StyleWarning
		}
		if elapsed > 0 {
			block.WriteLine(output.Styledf(style, "- [%s] %s (%s)", *job.State, *job.Name, elapsed))
		} else {
			block.WriteLine(output.Styledf(style, "- [%s] %s", *job.State, *job.Name))
		}

		if annotation, exist := annotations[*job.ID]; exist {
			block.WriteMarkdown(annotation.Content, output.MarkdownNoMargin, output.MarkdownIndent(2))
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
