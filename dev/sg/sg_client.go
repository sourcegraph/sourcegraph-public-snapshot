package main

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/scaletesting"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
	"github.com/sourcegraph/sourcegraph/lib/output"
	"github.com/urfave/cli/v2"
)

var clientCommand = &cli.Command{
	Name:  "client",
	Usage: "Perform common tasks on a Sourcegraph instance",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name: "baseurl",
		},
		&cli.StringFlag{
			Name: "email",
		},
		&cli.StringFlag{
			Name: "password",
		},
	},
	Subcommands: []*cli.Command{
		{
			Name:      "codehost",
			ArgsUsage: " ",
			Category:  CategoryUtil,
			Usage:     "TODO",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name: "token",
				},
			},
			Subcommands: []*cli.Command{
				{
					Name:      "add",
					ArgsUsage: "",
					Usage:     "TODO",
					Action: func(cmd *cli.Context) error {
						client, err := gqltestutil.SignIn(cmd.String("baseurl"), cmd.String("email"), cmd.String("password"))
						if err != nil {
							return err
						}
						err = scaletesting.AddCodeHostsByGitHubOrgs(client, cmd.String("token"), "test", cmd.Args().Slice())
						if err != nil {
							return err
						}
						resp, err := client.ExternalService("test")
						if err != nil {
							return err
						}
						std.Out.WriteSuccessf(
							"Added CodeHost %s%s%s",
							output.StyleOrange,
							"test",
							output.StyleReset,
						)
						std.Out.WriteMarkdown(fmt.Sprintf("`%s`", resp["config"]))
						return nil
					},
				},
			},
		},
	},
}
