package main

import (
	"fmt"

	cl "github.com/sourcegraph/sourcegraph/dev/sg/client"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/internal/gqltestutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/urfave/cli/v2"
)

var clientCommand = &cli.Command{
	Name:  "client",
	Usage: "Perform common tasks on a Sourcegraph instance",
	Subcommands: []*cli.Command{
		{
			Name:      "codehost",
			ArgsUsage: " ",
			Category:  CategoryUtil,
			Usage:     "Interact with codehosts on a given Sourcegraph instance",
			Subcommands: []*cli.Command{
				{
					Name:      "add-github",
					ArgsUsage: " org1 org2",
					Usage:     "Add a GitHub codehost on a given Sourcegraph instance",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:        "baseurl",
							Usage:       "Base URL to reach the Sourcegraph instance",
							DefaultText: "https://sourcegraph.test:3443",
						},
						&cli.StringFlag{
							Name:        "email",
							Usage:       "Email or username used to sign-in",
							DefaultText: "sourcegraph@sourcegraph.com",
						},
						&cli.StringFlag{
							Name:        "password",
							Usage:       "Password used to sign-in",
							DefaultText: "sourcegraphsourcegraph",
						},
						&cli.StringFlag{
							Name:  "token",
							Usage: "GitHub token to associate with the new code host",
						},
						&cli.StringFlag{
							Name:     "display-name",
							Usage:    "Display name to give to the new code host",
							Required: true,
						},
					},
					Action: func(cmd *cli.Context) error {
						client, err := gqltestutil.SignIn(cmd.String("baseurl"), cmd.String("email"), cmd.String("password"))
						if err != nil {
							return errors.Wrap(err, "Failed to sign-in on the Sourcegraph instance")
						}
						extSvcs, err := client.ExternalServices()
						if err != nil {
							return errors.Wrap(err, "Could not list code hosts")
						}
						displayName := cmd.String("display-name")
						for _, s := range extSvcs {
							if s.DisplayName == displayName {
								std.Out.WriteFailuref("A code host named '%s' already exists with the following definition", displayName)
								std.Out.WriteMarkdown(fmt.Sprintf("`%s`", s.Config))
								return errors.New("code host already exists")
							}
						}
						std.Out.WriteNoticef("Creating new code host named '%s'", displayName)
						err = cl.AddCodeHostsByGitHubOrgs(client, cmd.String("token"), displayName, cmd.Args().Slice())
						if err != nil {
							return errors.Wrap(err, "Could not add code host")
						}
						extSvcs, err = client.ExternalServices()
						if err != nil {
							return errors.Wrap(err, "Could not list code hosts")
						}
						for _, s := range extSvcs {
							if s.DisplayName == displayName {
								std.Out.WriteSuccessf("Added CodeHost '%s'", displayName)
								std.Out.WriteMarkdown(fmt.Sprintf("`%s`", s.Config))
								return nil
							}
						}
						std.Out.WriteFailuref("A code host named '%s' already exists with the following definition", displayName)
						return errors.Wrap(err, "Could find newly added code host")
					},
				},
			},
		},
	},
}
