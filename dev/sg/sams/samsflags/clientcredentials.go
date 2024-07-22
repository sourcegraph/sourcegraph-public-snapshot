package samsflags

import (
	"fmt"

	"github.com/urfave/cli/v2"
	"golang.org/x/oauth2/clientcredentials"

	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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
			Usage: fmt.Sprintf("URL of the Sourcegraph Accounts Management System (SAMS) server - one of %q or %q,"+
				// TODO: 9091 currently conflicts with embeddings, we may want
				// to change the default in the future
				" or http://127.0.0.1:9091 for a locally running server.",
				SAMSProdURL, SAMSDevURL),
		},
		&cli.StringFlag{
			Name:    "client-id",
			EnvVars: []string{"SG_SAMS_CLIENT_ID"},
			Usage: fmt.Sprintf("Client ID of the Sourcegraph Accounts Management System (SAMS) client - defaults to shared local dev credentials for %q",
				SAMSDevURL),
		},
		&cli.StringFlag{
			Name:    "client-secret",
			EnvVars: []string{"SG_SAMS_CLIENT_SECRET"},
			Usage: fmt.Sprintf("Client Secret of the Sourcegraph Accounts Management System (SAMS) client - defaults to shared local dev credentials for %q",
				SAMSDevURL),
		},
	}
}

// NewClientCredentialsFromFlags returns a new client credentials config from
// clientCredentialsFlags.
func NewClientCredentialsFromFlags(c *cli.Context, ss scopes.Scopes) (*clientcredentials.Config, error) {
	cfg := &clientcredentials.Config{
		ClientID:     c.String("client-id"),
		ClientSecret: c.String("client-secret"),
		TokenURL:     c.String("sams-server") + "/oauth/token",
		Scopes:       scopes.ToStrings(ss),
	}

	var err error
	if cfg.ClientID == "" {
		cfg.ClientID, err = defaultLocalDevClientID(c.Context)
		if err != nil {
			return nil, errors.Wrapf(err, "get local dev client ID for %s", SAMSDevURL)
		}
	}
	if cfg.ClientSecret == "" {
		cfg.ClientSecret, err = defaultLocalDevClientSecret(c.Context)
		if err != nil {
			return nil, errors.Wrapf(err, "get local dev client secret for %s", SAMSDevURL)
		}
	}

	return cfg, nil
}

func NewClientFromFlags(c *cli.Context, ss scopes.Scopes) (*sams.ClientV1, error) {
	cfg, err := NewClientCredentialsFromFlags(c, ss)
	if err != nil {
		return nil, err
	}
	return sams.NewClientV1(sams.ClientV1Config{
		ConnConfig: sams.ConnConfig{
			ExternalURL: c.String("sams-server"),
		},
		TokenSource: cfg.TokenSource(c.Context),
	})
}
