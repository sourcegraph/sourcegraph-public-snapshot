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

var clientCredentialsFlags = append(samsflags.ClientCredentials(),
	&cli.StringSliceFlag{
		Name:    "scopes",
		Aliases: []string{"s"},
		Value:   cli.NewStringSlice("openid", "profile", "email"),
		Usage:   "OAuth scopes ('$SERVICE::$PERM::$ACTION') to request from the Sourcegraph Accounts Management System (SAMS) server",
	},
)

// Command is the 'sg sams' toolchain for the Sourcegraph Accounts Management System (SAMS).
var Command = &cli.Command{
	Name:     "sourcegraph-accounts",
	Aliases:  []string{"sams"},
	Category: category.Company,
	Usage:    "Development utilities for integrations against the Sourcegraph Accounts Management System (SAMS)",
	Description: `Learn more in https://sourcegraph.notion.site/Sourcegraph-Accounts-Management-System-SAMS-e86f1bc3dc3b4d979818e468bba189fd.

Please reach out to #discuss-core-services for assistance if you have any questions!`,
	Subcommands: []*cli.Command{{
		Name:  "introspect-token",
		Usage: "Generate a short-lived OAuth access token and introspect it from the Sourcegraph Accounts Management System (SAMS)",
		Flags: clientCredentialsFlags,
		Action: func(c *cli.Context) error {
			samsScopes := scopes.ToScopes(c.StringSlice("scopes"))

			client, err := samsflags.NewClientFromFlags(c, samsScopes)
			if err != nil {
				return errors.Wrap(err, "create client")
			}

			token, err := samsflags.NewClientCredentialsFromFlags(c, samsScopes).
				TokenSource(c.Context).
				Token()
			if err != nil {
				return errors.Wrap(err, "generate token")
			}
			resp, err := client.Tokens().IntrospectToken(c.Context, token.AccessToken)
			if err != nil {
				return errors.Wrap(err, "introspect token")
			}

			data, err := json.MarshalIndent(resp, "", "  ")
			if err != nil {
				return err
			}
			std.Out.Write(string(data))
			return nil
		},
	}, {
		Name:  "create-client-token",
		Usage: "Generate a short-lived OAuth access token for use as a bearer token to SAMS clients",
		Flags: clientCredentialsFlags,
		Action: func(c *cli.Context) error {
			samsScopes := scopes.ToScopes(c.StringSlice("scopes"))
			tokenSource := samsflags.NewClientCredentialsFromFlags(c, samsScopes).
				TokenSource(c.Context)
			token, err := tokenSource.Token()
			if err != nil {
				return errors.Wrap(err, "generate token")
			}

			data, err := json.MarshalIndent(token, "", "  ")
			if err != nil {
				return err
			}
			std.Out.Write(string(data))
			return nil
		},
	}},
}
