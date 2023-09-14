package postgresdsn

import (
	"fmt"
	"net/url"
	"strings"
)

func New(prefix, currentUser string, getenv func(string) string) string {
	if prefix == "frontend" {
		prefix = ""
	}
	if prefix != "" {
		prefix = fmt.Sprintf("%s_", strings.ToUpper(prefix))
	}

	env := func(name string) string {
		return getenv(prefix + name)
	}

	dequote := func(value string) string {
		if len(value) > 2 && value[0] == value[len(value)-1] && (value[0] == '\'' || value[0] == '"') {
			return value[1 : len(value)-1]
		}

		return value
	}

	// PGDATASOURCE is a sourcegraph specific variable for just setting the DSN
	if dsn := env("PGDATASOURCE"); dsn != "" {
		return dsn
	}

	// TODO match logic in lib/pq
	// https://sourcegraph.com/github.com/lib/pq@d6156e141ac6c06345c7c73f450987a9ed4b751f/-/blob/connector.go#L42
	dsn := &url.URL{
		Scheme: "postgres",
		Host:   "127.0.0.1:5432",
	}

	// Username preference: PGUSER, $USER, postgres
	username := "postgres"
	if currentUser != "" {
		username = currentUser
	}
	if user := env("PGUSER"); user != "" {
		username = user
	}

	if password := env("PGPASSWORD"); password != "" {
		dsn.User = url.UserPassword(username, password)
	} else {
		dsn.User = url.User(username)
	}

	if host := env("PGHOST"); host != "" {
		dsn.Host = host
	}

	// PGPORT values may be (legally) quoted, but should remain quoted
	// when constructed as part of the DSN. Strip it here.
	if port := dequote(env("PGPORT")); port != "" {
		dsn.Host = strings.Split(dsn.Host, ":")[0] + ":" + port
	}

	if db := env("PGDATABASE"); db != "" {
		dsn.Path = db
	}

	if sslmode := env("PGSSLMODE"); sslmode != "" {
		qry := dsn.Query()
		qry.Set("sslmode", sslmode)
		dsn.RawQuery = qry.Encode()
	}

	if tz := env("PGTZ"); tz != "" {
		qry := dsn.Query()
		qry.Set("timezone", tz)
		dsn.RawQuery = qry.Encode()
	}

	return dsn.String()
}
