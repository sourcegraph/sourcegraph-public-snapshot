// Package sams exports 'sg sams' commands for the Sourcegraph Accounts Management System.
package sams

import (
	"encoding/json"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/dev/sg/sams/samsflags"
)

func clientCredentialsFlags() []cli.Flag {
	return append(samsflags.ClientCredentials(),
		&cli.StringSliceFlag{
			Name:    "scopes",
			Aliases: []string{"s"},
			Value:   cli.NewStringSlice("openid", "profile", "email"),
			Usage:   "OAuth scopes ('$SERVICE::$PERM::$ACTION') to request from the Sourcegraph Accounts Management System (SAMS) server",
		},
	)
}

// Command is the 'sg sams' toolchain for the Sourcegraph Accounts Management System (SAMS).
var Command = &cli.Command{
	Name:     "sourcegraph-accounts",
	Aliases:  []string{"sams"},
	Category: category.Company,
	Usage:    "Development utilities for integrations against the Sourcegraph Accounts Management System (SAMS)",
	Description: `Learn more in https://sourcegraph.notion.site/Sourcegraph-Accounts-Management-System-SAMS-e86f1bc3dc3b4d979818e468bba189fd.

Please reach out to #discuss-core-services for assistance if you have any questions!`,
	Subcommands: []*cli.Command{{
		Name:  "token",
		Usage: "Request and interact with SAMS tokens",
		Subcommands: []*cli.Command{{
			Name:  "introspect",
			Usage: "Generate a short-lived OAuth access token for the configured SAMS client and introspect its attributes",
			Flags: append(clientCredentialsFlags(),
				&cli.BoolFlag{
					Name:    "print-token",
					Aliases: []string{"p"},
					Usage:   "Print the requested token, in addition to the introspection response",
				}),
			Action: func(c *cli.Context) error {
				samsScopes := scopes.ToScopes(c.StringSlice("scopes"))

				client, err := samsflags.NewClientFromFlags(c, samsScopes)
				if err != nil {
					return errors.Wrap(err, "create client")
				}

				cfg, err := samsflags.NewClientCredentialsFromFlags(c, samsScopes)
				if err != nil {
					return errors.Wrap(err, "create client credentials")
				}
				token, err := cfg.TokenSource(c.Context).Token()
				if err != nil {
					return errors.Wrap(err, "generate token")
				}
				resp, err := client.Tokens().IntrospectToken(c.Context, token.AccessToken)
				if err != nil {
					return errors.Wrap(err, "introspect token")
				}

				output := map[string]any{
					"introspect": resp,
				}
				if c.Bool("print-token") {
					output["token"] = token
				}

				data, err := json.MarshalIndent(output, "", "  ")
				if err != nil {
					return err
				}
				return std.Out.WriteCode("json", string(data))
			},
		}},
	}},
}
