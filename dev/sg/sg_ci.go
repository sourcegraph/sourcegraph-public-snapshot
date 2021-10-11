package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/bk"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/open"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	ciFlagSet = flag.NewFlagSet("sg ci", flag.ExitOnError)
	ciCommand = &ffcli.Command{
		Name:       "ci",
		ShortUsage: "sg ci [preview|status|build]",
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
			Exec: func(ctx context.Context, args []string) error {
				client, err := bk.NewClient(ctx, out)
				if err != nil {
					return err
				}

				branch, err := run.TrimResult(run.GitCmd("branch", "--show-current"))
				if err != nil {
					return err
				}

				// Just support main pipeline for now
				build, err := client.GetMostRecentBuild(ctx, "sourcegraph", branch)
				if err != nil {
					return fmt.Errorf("failed to get most recent build for branch %q: %w", branch, err)
				}

				// Print a high level overview
				out.WriteLine(output.Linef("", output.StyleBold, "Most recent build: %s", *build.WebURL))
				out.Writef("Commit: %s\nStarted: %s", *build.Commit, build.StartedAt)
				if build.FinishedAt != nil {
					out.Writef("Finished: %s (elapsed: %s)", build.FinishedAt, build.FinishedAt.Sub(build.StartedAt.Time))
				}

				// Valid states: running, scheduled, passed, failed, blocked, canceled, canceling, skipped, not_run
				// https://buildkite.com/docs/apis/rest-api/builds
				var style output.Style
				var emoji string
				switch *build.State {
				case "passed":
					style = output.StyleSuccess
					emoji = output.EmojiSuccess
				case "running", "scheduled":
					style = output.StylePending
					emoji = output.EmojiInfo
				case "failed":
					emoji = output.EmojiFailure
					fallthrough
				default:
					style = output.StyleWarning
				}
				out.WriteLine(output.Linef(emoji, style, "Status: %s", *build.State))

				// Warn if build commit is not your commit
				commit, err := run.GitCmd("rev-parse", "HEAD")
				if err != nil {
					return err
				}
				commit = strings.TrimSpace(commit)
				if commit != *build.Commit {
					out.WriteLine(output.Linef(output.EmojiWarning, output.StyleWarning,
						"The currently checked out commit %q does not match the commit of the build found, %q.\nHave you pushed your most recent changes yet?",
						commit, *build.Commit))
				}
				return nil
			},
		}, {
			Name:      "build",
			ShortHelp: "Manually request a build for the currently checked out commit and branch (e.g. to trigger builds on forks)",
			LongHelp:  "Manually request a Buildkite build for the currently checked out commit and branch. This is most useful when triggering builds for PRs from forks (such as those from external contributors), which do not trigger Buildkite builds automatically for security reasons (we do not want to run insecure code on our infrastructure by default!)",
			Exec: func(ctx context.Context, args []string) error {
				client, err := bk.NewClient(ctx, out)
				if err != nil {
					return err
				}

				branch, err := run.TrimResult(run.GitCmd("branch", "--show-current"))
				if err != nil {
					return err
				}
				commit, err := run.TrimResult(run.GitCmd("rev-parse", "HEAD"))
				if err != nil {
					return err
				}
				out.WriteLine(output.Linef("", output.StylePending, "Requesting build for branch %q at %q...", branch, commit))

				// simple check to see if commit is in origin, this is non blocking but
				// we ask for confirmation to double check.
				remoteBranches, err := run.TrimResult(run.GitCmd("branch", "-r", "--contains", commit))
				if err != nil || len(remoteBranches) == 0 || !allLinesPrefixed(strings.Split(remoteBranches, "\n"), "origin/") {
					out.WriteLine(output.Linef(output.EmojiWarning, output.StyleReset,
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
				out.WriteLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Created build: %s", *build.WebURL))
				return nil
			},
		}},
	}
)

func allLinesPrefixed(lines []string, match string) bool {
	for _, l := range lines {
		if !strings.HasPrefix(l, match) {
			return false
		}
	}
	return true
}
