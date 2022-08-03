package main

import (
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	versionCommand = &cli.Command{
		Name:     "version",
		Usage:    "View details for this installation of sg",
		Action:   versionExec,
		Category: CategoryUtil,
		Subcommands: []*cli.Command{
			{
				Name:    "changelog",
				Aliases: []string{"c"},
				Usage:   "See what's changed in or since this version of sg",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "next",
						Usage: "Show changelog for changes you would get if you upgrade.",
					},
					&cli.IntFlag{
						Name:  "count",
						Usage: "Number of changelog entries to show.",
						Value: 10,
					},
				},
				Action: func(cmd *cli.Context) error {
					if _, err := run.GitCmd("fetch", "origin", "main"); err != nil {
						return errors.Newf("failed to update main: %s", err)
					}

					count := cmd.Int("count")
					title, changes, err := repo.RecentSGChanges(BuildCommit, repo.RecentChangesOpts{
						Count: count,
						Next:  cmd.Bool("next"),
					})
					if err != nil {
						return err
					}

					block := std.Out.Block(output.Styled(output.StyleSearchQuery, title))
					if len(changes) == 0 {
						block.Write("No changes found.")
					} else {
						block.Write(strings.Join(changes, "\n") + "\n...")
					}
					block.Close()

					std.Out.WriteLine(output.Styledf(output.StyleSuggestion,
						"Only showing %d entries - configure with 'sg version changelog -limit=50'", count))
					return nil
				},
			},
		},
	}
)

func versionExec(ctx *cli.Context) error {
	std.Out.Write(BuildCommit)
	return nil
}
