// Package sams exports 'sg sams' commands for the Sourcegraph Accounts Management System.
package sams

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
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
	}, {
		Name:  "client",
		Usage: "Manage IdP clients registered in SAMS",
		Subcommands: []*cli.Command{{
			Name:      "create",
			ArgsUsage: "<display name>",
			Usage:     "Create an IdP client in the target SAMS instance",
			Flags: []cli.Flag{
				&cli.StringSliceFlag{
					Name:     "redirect-uris",
					Usage:    "Redirect URIs to associate with the client, e.g. 'https://sourcegraph.test:3443/.auth/callback'",
					Required: true,
				},
				// Flags are special subset of clientCredentialsFlags(), as we
				// use a static secret - in the future we should change this to
				// require client credentials instead.
				&cli.StringFlag{
					Name:    "sams-server",
					Aliases: []string{"sams"},
					EnvVars: []string{"SG_SAMS_SERVER_URL"},
					Value:   samsflags.SAMSDevURL,
					Usage: fmt.Sprintf("URL of the Sourcegraph Accounts Management System (SAMS) server - one of %q or %q",
						samsflags.SAMSProdURL, samsflags.SAMSDevURL),
				},
				&cli.StringSliceFlag{
					Name:    "scopes",
					Aliases: []string{"s"},
					Value:   cli.NewStringSlice("openid", "profile", "email"),
					Usage:   "OAuth scopes ('$SERVICE::$PERM::$ACTION') to request from the Sourcegraph Accounts Management System (SAMS) server",
				},
			},
			Action: func(c *cli.Context) error {
				samsServer := c.String("sams-server")
				displayName := c.Args().First()
				if displayName == "" {
					return errors.New("argument display name required")
				}

				ss, err := secrets.FromContext(c.Context)
				if err != nil {
					return err
				}
				// Environments specified in
				// https://github.com/sourcegraph/managed-services/blob/main/services/sourcegraph-accounts/service.yaml
				samsManagementSecretSource := map[string]*secrets.ExternalSecret{
					samsflags.SAMSDevURL: {
						Project: "sourcegraph-dev",
						Name:    "SAMS_MANAGEMENT_SECRET",
					},
					samsflags.SAMSProdURL: {
						Project: "sourcegraph-accounts-prod-csvc",
						Name:    "MANAGEMENT_SECRET",
					},
				}
				var managementSecret string
				if externalSecret := samsManagementSecretSource[samsServer]; externalSecret != nil {
					std.Out.WriteSuggestionf("Targeting SAMS instance at %q", samsServer)
					managementSecret, err = ss.GetExternal(c.Context, *externalSecret)
					if err != nil {
						std.Out.WriteAlertf("Failed to get secret - do you have Entitle access to the %q project? See https://sourcegraph.notion.site/Sourcegraph-Accounts-infrastructure-operations-b90a571da30443a8b1e7c31ade3594fb",
							externalSecret.Project)
						return err
					}
				} else {
					managementSecret, err = std.Out.PromptPasswordf(os.Stdin,
						"Enter the SAMS management secret for your target SAMS server %q: ",
						c.String("sams-server"))
					if err != nil {
						return err
					}
				}

				body := map[string]any{
					"name":          displayName,
					"scopes":        c.StringSlice("scopes"),
					"redirect_uris": c.StringSlice("redirect-uris"),
				}
				data, err := json.Marshal(body)
				if err != nil {
					return err
				}

				req, err := http.NewRequest(
					http.MethodPost,
					samsServer+"/api/management/v1/identity-provider/clients",
					bytes.NewReader(data),
				)
				if err != nil {
					return err
				}
				req.Header.Set("Authorization", "Bearer "+managementSecret)

				resp, err := httpcli.UncachedExternalDoer.Do(req.WithContext(c.Context))
				if err != nil {
					return errors.Wrap(err, "do request")
				}
				defer resp.Body.Close()
				responseData, err := io.ReadAll(resp.Body)
				if err != nil {
					return errors.Wrap(err, "read response")
				}
				if resp.StatusCode < 200 || resp.StatusCode >= 300 {
					std.Out.Write("Got responnse:")
					std.Out.Write(string(responseData))
					return errors.Newf("got unexpected response status %d", resp.StatusCode)
				}

				var v any
				if err := json.Unmarshal(responseData, &v); err != nil {
					return err
				}
				prettyResponseData, err := json.MarshalIndent(v, "", "  ")
				if err != nil {
					return err
				}
				if err := std.Out.WriteCode("json", string(prettyResponseData)); err != nil {
					return err
				}

				if u, err := url.Parse(samsServer); err == nil &&
					(u.Hostname() == "127.0.0.1" || u.Hostname() == "localhost") { // CI:LOCALHOST_OK
					std.Out.WriteSuggestionf("These client credentials can NOT be shown again - if you lose them you will need to create another one.")
				} else {
					std.Out.WriteWarningf("These client credentials are highly sensitive and can NOT be shown again. Please store them securely in Google Secret Manager or 1Password.")
				}
				return nil
			},
		}},
	}},
}
