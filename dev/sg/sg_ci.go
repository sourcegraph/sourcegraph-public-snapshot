package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/bk"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	ciFlagSet = flag.NewFlagSet("sg ci", flag.ExitOnError)
	ciCommand = &ffcli.Command{
		Name:       "ci",
		ShortUsage: "sg ci [preview|status]",
		ShortHelp:  "Interact with Sourcegraph's continuous integration pipelines",
		LongHelp: `Interact with Sourcegraph's continuous integration pipelines on Buildkite.

Note that Sourcegraph's CI pipelines are under our enterprise license: https://github.com/sourcegraph/sourcegraph/blob/main/LICENSE.enterprise`,
		FlagSet: ciFlagSet,
		Exec:    ciExec,
	}
)

func ciExec(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return flag.ErrHelp
	}
	switch args[0] {
	case "preview":
		stdout.Out.WriteLine(output.Linef("", output.StyleSuggestion,
			"If the current branch were to be pushed, the following pipeline would be run:"))

		branch, err := run.GitCmd("branch", "--show-current")
		if err != nil {
			return err
		}
		message, err := run.GitCmd("show", "--format=%s\\n%b")
		if err != nil {
			return err
		}
		cmd := exec.Command("go", "run", "./enterprise/dev/ci/gen-pipeline.go", "-preview")
		cmd.Env = append(os.Environ(),
			fmt.Sprintf("BUILDKITE_BRANCH=%s", strings.TrimSpace(branch)),
			fmt.Sprintf("BUILDKITE_MESSAGE=%s", strings.TrimSpace(message)))
		out, err := run.InRoot(cmd)
		if err != nil {
			return err
		}
		stdout.Out.Write(out)
		return nil

	case "status":
		client, err := bk.NewClient(ctx, out)
		if err != nil {
			return err
		}

		branch, err := run.GitCmd("branch", "--show-current")
		if err != nil {
			return err
		}
		branch = strings.TrimSpace(branch)

		// Just support main pipeline for now
		build, err := client.GetMostRecentBuild(ctx, "sourcegraph", branch)
		if err != nil {
			return fmt.Errorf("failed to get most recent build for branch %q: %w", branch, err)
		}

		// Print a high level overview
		out.WriteLine(output.Linef("", output.StyleBold, "Most recent build: %s", *build.WebURL))
		out.Writef(`Commit: %s
Started: %s`, *build.Commit, build.StartedAt)
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

	default:
		return flag.ErrHelp
	}
}
