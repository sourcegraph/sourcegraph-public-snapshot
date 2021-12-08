package dbtest

import (
	"net/url"
	"os"
)

func getDSN(dsn string) (*url.URL, error) {
	if dsn == "" {
		var ok bool
		if dsn, ok = os.LookupEnv("PGDATASOURCE"); !ok {
			dsn = `postgres://sourcegraph:sourcegraph@127.0.0.1:5432/sourcegraph?sslmode=disable&timezone=UTC`
		}
	}

	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	updateDSNFromEnv(u)

	return u, nil
}

// updateDSNFromEnv updates dsn based on PGXXX environment variables set on
// the frontend.
func updateDSNFromEnv(dsn *url.URL) {
	if host := os.Getenv("PGHOST"); host != "" {
		dsn.Host = host
	}

	if port := os.Getenv("PGPORT"); port != "" {
		dsn.Host += ":" + port
	}

	if user := os.Getenv("PGUSER"); user != "" {
		if password := os.Getenv("PGPASSWORD"); password != "" {
			dsn.User = url.UserPassword(user, password)
		} else {
			dsn.User = url.User(user)
		}
	}

	if db := os.Getenv("PGDATABASE"); db != "" {
		dsn.Path = db
	}

	if sslmode := os.Getenv("PGSSLMODE"); sslmode != "" {
		qry := dsn.Query()
		qry.Set("sslmode", sslmode)
		dsn.RawQuery = qry.Encode()
	}
}
