package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/rfc"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var rfcCommand = &cli.Command{
	Name:  "rfc",
	Usage: `List, search, and open Sourcegraph RFCs`,
	Description: fmt.Sprintf("Sourcegraph RFCs live in the following drives - see flags to configure which drive to query:\n\n%s", func() (out string) {
		for _, d := range []rfc.DriveSpec{rfc.PublicDrive, rfc.PrivateDrive} {
			out += fmt.Sprintf("* %s: https://drive.google.com/drive/folders/%s\n", d.DisplayName, d.FolderID)
		}
		return out
	}()),
	UsageText: `
# List all Public RFCs
sg rfc list

# List all Private RFCs
sg rfc --private list

# Search for a Public RFC
sg rfc search "search terms"

# Search for a Private RFC
sg rfc --private search "search terms"

# Open a specific Public RFC
sg rfc open 420

# Open a specific private RFC
sg rfc --private open 420

# Create a new public RFC
sg rfc create "title"

# Create a new private RFC. Possible types: [solution]
sg rfc --private create --type <type> "title"
`,
	Category: category.Company,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:     "private",
			Usage:    "perform the RFC action on the private RFC drive",
			Required: false,
			Value:    false,
		},
	},
	Commands: []*cli.Command{
		{
			Name:      "list",
			ArgsUsage: " ",
			Usage:     "List Sourcegraph RFCs",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				driveSpec := rfc.PublicDrive
				if cmd.Bool("private") {
					driveSpec = rfc.PrivateDrive
				}
				return rfc.List(ctx, driveSpec, std.Out)
			},
		},
		{
			Name:      "search",
			ArgsUsage: "[query]",
			Usage:     "Search Sourcegraph RFCs",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				driveSpec := rfc.PublicDrive
				if cmd.Bool("private") {
					driveSpec = rfc.PrivateDrive
				}
				if cmd.Args().Len() == 0 {
					return errors.New("no search query given")
				}
				return rfc.Search(ctx, strings.Join(cmd.Args().Slice(), " "), driveSpec, std.Out)
			},
		},
		{
			Name:      "open",
			ArgsUsage: "[number]",
			Usage:     "Open a Sourcegraph RFC - find and list RFC numbers with 'sg rfc list' or 'sg rfc search'",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				driveSpec := rfc.PublicDrive
				if cmd.Bool("private") {
					driveSpec = rfc.PrivateDrive
				}
				if cmd.Args().Len() == 0 {
					return errors.New("no RFC given")
				}
				return rfc.Open(ctx, cmd.Args().First(), driveSpec, std.Out)
			},
		},
		{
			Name:      "create",
			ArgsUsage: "--type <type> [title...]",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "type",
					Usage: "the type of the RFC to create (valid: solution)",
					Value: rfc.ProblemSolutionDriveTemplate.Name,
				},
			},
			Usage: "Create Sourcegraph RFCs",
			Action: func(ctx context.Context, cmd *cli.Command) error {
				driveSpec := rfc.PublicDrive
				if cmd.Bool("private") {
					driveSpec = rfc.PrivateDrive
				}

				rfcType := cmd.String("type")

				var template rfc.Template
				// Search for the rfcType and assign it to template
				for _, tpl := range rfc.AllTemplates {
					if tpl.Name == rfcType {
						template = tpl
						break
					}
				}
				if template.Name == "" {
					return errors.New(fmt.Sprintf("Unknown RFC type: %s", rfcType))
				}

				if cmd.Args().Len() == 0 {
					return errors.New("no title given")
				}
				return rfc.Create(ctx, template, strings.Join(cmd.Args().Slice(), " "),
					driveSpec, std.Out)
			},
		},
	},
}
