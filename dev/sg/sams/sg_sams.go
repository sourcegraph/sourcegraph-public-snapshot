// Package msp exports 'sg sams' commands for the Sourcegraph Accounts Management System.
package sams

import (
	"encoding/json"

	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2/clientcredentials"

	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var clientCredentialsFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "sams-server",
		Aliases: []string{"sams"},
		EnvVars: []string{"SG_SAMS_SERVER_URL"},
		Value:   "https://accounts.sgdev.org",
		Usage:   "URL of the Sourcegraph Accounts Management System (SAMS) server",
	},
	&cli.StringFlag{
		Name:     "client-id",
		EnvVars:  []string{"SG_SAMS_CLIENT_ID"},
		Usage:    "Client ID of the Sourcegraph Accounts Management System (SAMS) client",
		Required: true,
	},
	&cli.StringFlag{
		Name:     "client-secret",
		EnvVars:  []string{"SG_SAMS_CLIENT_SECRET"},
		Usage:    "Client secret for the Sourcegraph Accounts Management System (SAMS) client",
		Required: true,
	},
	&cli.StringSliceFlag{
		Name:    "scopes",
		Aliases: []string{"s"},
		Value:   cli.NewStringSlice("openid", "profile", "email"),
		Usage:   "OAuth scopes ('$SERVICE::$PERM::$ACTION') to request from the Sourcegraph Accounts Management System (SAMS) server",
	},
}

// newClientCredentialsFromFlags returns a new client credentials config from
// clientCredentialsFlags.
func newClientCredentialsFromFlags(c *cli.Context) *clientcredentials.Config {
	return &clientcredentials.Config{
		ClientID:     c.String("client-id"),
		ClientSecret: c.String("client-secret"),
		TokenURL:     c.String("sams-server") + "/oauth/token",
		Scopes:       c.StringSlice("scopes"),
	}
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
		Name:  "introspect-token",
		Usage: "Generate a short-lived OAuth access token and introspect it from the Sourcegraph Accounts Management System (SAMS)",
		Flags: clientCredentialsFlags,
		Action: func(c *cli.Context) error {
			tokenSource := newClientCredentialsFromFlags(c).
				TokenSource(c.Context)

			client, err := sams.NewClientV1(
				sams.ClientV1Config{
					ConnConfig: sams.ConnConfig{
						ExternalURL: c.String("sams-server"),
					},
					TokenSource: tokenSource,
				},
			)
			if err != nil {
				return errors.Wrap(err, "create client")
			}

			token, err := tokenSource.Token()
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
			tokenSource := newClientCredentialsFromFlags(c).
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
