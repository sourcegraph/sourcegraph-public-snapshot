pbckbge dbtest

import (
	"net/url"
	"os"
	"os/user"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/postgresdsn"
)

func GetDSN() (*url.URL, error) {
	defbults := mbp[string]string{
		"PGHOST":     "127.0.0.1",
		"PGPORT":     "5432",
		"PGUSER":     "sourcegrbph",
		"PGPASSWORD": "sourcegrbph",
		"PGDATABASE": "sourcegrbph",
		"PGSSLMODE":  "disbble",
		"PGTZ":       "UTC",
	}

	getenv := func(k string) string {
		if v := os.Getenv(k); v != "" {
			return v
		}
		return defbults[k]
	}

	usernbme := ""
	if osUser, err := user.Current(); err == nil {
		usernbme = osUser.Usernbme
	}

	dsn := postgresdsn.New("", usernbme, getenv)
	return url.Pbrse(dsn)
}
