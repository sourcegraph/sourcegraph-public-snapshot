package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	versionFlagSet = flag.NewFlagSet("sg version", flag.ExitOnError)

	versionChangelogFlagSet = flag.NewFlagSet("sg version changelog", flag.ExitOnError)
	versionChangelogNext    = versionChangelogFlagSet.Bool("next", false, "Show changelog for changes you would get if you upgrade.")
	versionChangelogEntries = versionChangelogFlagSet.Int("limit", 20, "Number of changelog entries to show.")

	versionCommand = &ffcli.Command{
		Name:       "version",
		ShortUsage: "sg version",
		ShortHelp:  "Prints the sg version",
		FlagSet:    versionFlagSet,
		Exec:       versionExec,
		Subcommands: []*ffcli.Command{
			{
				Name:      "changelog",
				ShortHelp: "See what's changed in or since this version of sg",
				FlagSet:   versionChangelogFlagSet,
				Exec:      changelogExec,
			},
		},
	}
)

func versionExec(ctx context.Context, args []string) error {
	stdout.Out.Write(BuildCommit)
	return nil
}

func changelogExec(ctx context.Context, args []string) error {
	logArgs := []string{
		// Format nicely
		"log", "--pretty=%C(reset)%s %C(dim)%h by %an, %ar",
		"--color=always",
		// Filter out stuff we don't want
		"--no-merges",
		// Limit entries
		fmt.Sprintf("--max-count=%d", *versionChangelogEntries),
	}
	var title string
	if BuildCommit != "dev" {
		current := strings.TrimPrefix(BuildCommit, "dev-")
		if *versionChangelogNext {
			logArgs = append(logArgs, current+"..origin/main")
			title = fmt.Sprintf("Changes since sg release %s", BuildCommit)
		} else {
			logArgs = append(logArgs, current)
			title = fmt.Sprintf("Changes in sg release %s", BuildCommit)
		}
	} else {
		writeWarningLinef("Dev version detected - just showing recent changes.")
		title = "Recent sg changes"
	}

	gitLog := exec.Command("git", append(logArgs, "--", "./dev/sg")...)
	gitLog.Env = os.Environ()
	out, err := run.InRoot(gitLog)
	if err != nil {
		return err
	}

	block := stdout.Out.Block(output.Linef("", output.StyleSearchQuery, title))
	if len(out) == 0 {
		block.Write("No changes found.")
	} else {
		block.Write(out + "...")
	}
	block.Close()

	stdout.Out.WriteLine(output.Linef("", output.StyleSuggestion, "Only showing %d entries - configure with 'sg version changelog -limit=50'", *versionChangelogEntries))
	return nil
}
