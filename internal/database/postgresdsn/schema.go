package postgresdsn

import (
	"net/url"
	"os"
	"os/user"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func DSNsBySchema(schemaNames []string) (map[string]string, error) {
	dsns := RawDSNsBySchema(schemaNames, os.Getenv)

	// We set this envvar in development to disable the following check
	if os.Getenv("CODEINTEL_PG_ALLOW_SINGLE_DB") == "" {
		if codeintelDSN, ok := dsns["codeintel"]; ok {
			if frontendDSN, ok := dsns["frontend"]; ok {
				// Ensure that the code intelligence database is not pointing at the frontend database
				if err := comparePostgresDSNs("frontend", "codeintel", frontendDSN, codeintelDSN); err != nil {
					return nil, err
				}
			}
		}
	}

	return dsns, nil
}

func RawDSNsBySchema(schemaNames []string, getenv func(string) string) map[string]string {
	username := ""
	if currentUser, err := user.Current(); err == nil {
		username = currentUser.Username
	}

	dsns := make(map[string]string, len(schemaNames))
	for _, schemaName := range schemaNames {
		dsns[schemaName] = New(schemaName, username, getenv)
	}

	return dsns
}

// comparePostgresDSNs returns an error if one of the given Postgres DSN values are not a valid URL, or if
// they are both valid URLs but point to the same database. We consider two DSNs to be the same when they
// specify the same host, port, and path. It's possible that different hosts/ports map to the same physical
// machine, so we could conceivably return false negatives here and the tricksy site-admin may have pulled
// the wool over our eyes. This shouldn't actually affect anything operationally in the near-term, but may
// just make migrations harder when we need to have them manually separate the data.
func comparePostgresDSNs(name1, name2, dsn1, dsn2 string) error {
	url1, err := url.Parse(dsn1)
	if err != nil {
		return errors.Errorf("illegal Postgres DSN: %s", dsn1)
	}

	url2, err := url.Parse(dsn2)
	if err != nil {
		return errors.Errorf("illegal Postgres DSN: %s", dsn2)
	}

	if url1.Host == url2.Host && url1.Path == url2.Path {
		return errors.Errorf("databases %s and %s must be distinct, but both target %s", name1, name2, &url.URL{
			Scheme: "postgres",
			Host:   url1.Host,
			Path:   url1.Path,
		})
	}

	return nil
}
