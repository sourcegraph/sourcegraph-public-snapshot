package samsflags

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2/clientcredentials"

	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
)

const (
	SAMSDevURL  = "https://accounts.sgdev.org"
	SAMSProdURL = "https://accounts.sourcegraph.com"
)

func ClientCredentials() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "sams-server",
			Aliases: []string{"sams"},
			EnvVars: []string{"SG_SAMS_SERVER_URL"},
			Value:   SAMSDevURL,
			Usage: fmt.Sprintf("URL of the Sourcegraph Accounts Management System (SAMS) server - one of %q or %q",
				SAMSProdURL, SAMSDevURL),
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
	}
}

// NewClientCredentialsFromFlags returns a new client credentials config from
// clientCredentialsFlags.
func NewClientCredentialsFromFlags(c *cli.Context, ss scopes.Scopes) *clientcredentials.Config {
	return &clientcredentials.Config{
		ClientID:     c.String("client-id"),
		ClientSecret: c.String("client-secret"),
		TokenURL:     c.String("sams-server") + "/oauth/token",
		Scopes:       scopes.ToStrings(ss),
	}
}

func NewClientFromFlags(c *cli.Context, ss scopes.Scopes) (*sams.ClientV1, error) {
	return sams.NewClientV1(sams.ClientV1Config{
		ConnConfig: sams.ConnConfig{
			ExternalURL: c.String("sams-server"),
		},
		TokenSource: NewClientCredentialsFromFlags(c, ss).TokenSource(c.Context),
	})
}
