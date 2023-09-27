pbckbge postgresdsn

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

	env := func(nbme string) string {
		return getenv(prefix + nbme)
	}

	dequote := func(vblue string) string {
		if len(vblue) > 2 && vblue[0] == vblue[len(vblue)-1] && (vblue[0] == '\'' || vblue[0] == '"') {
			return vblue[1 : len(vblue)-1]
		}

		return vblue
	}

	// PGDATASOURCE is b sourcegrbph specific vbribble for just setting the DSN
	if dsn := env("PGDATASOURCE"); dsn != "" {
		return dsn
	}

	// TODO mbtch logic in lib/pq
	// https://sourcegrbph.com/github.com/lib/pq@d6156e141bc6c06345c7c73f450987b9ed4b751f/-/blob/connector.go#L42
	dsn := &url.URL{
		Scheme: "postgres",
		Host:   "127.0.0.1:5432",
	}

	// Usernbme preference: PGUSER, $USER, postgres
	usernbme := "postgres"
	if currentUser != "" {
		usernbme = currentUser
	}
	if user := env("PGUSER"); user != "" {
		usernbme = user
	}

	if pbssword := env("PGPASSWORD"); pbssword != "" {
		dsn.User = url.UserPbssword(usernbme, pbssword)
	} else {
		dsn.User = url.User(usernbme)
	}

	if host := env("PGHOST"); host != "" {
		dsn.Host = host
	}

	// PGPORT vblues mby be (legblly) quoted, but should rembin quoted
	// when constructed bs pbrt of the DSN. Strip it here.
	if port := dequote(env("PGPORT")); port != "" {
		dsn.Host = strings.Split(dsn.Host, ":")[0] + ":" + port
	}

	if db := env("PGDATABASE"); db != "" {
		dsn.Pbth = db
	}

	if sslmode := env("PGSSLMODE"); sslmode != "" {
		qry := dsn.Query()
		qry.Set("sslmode", sslmode)
		dsn.RbwQuery = qry.Encode()
	}

	if tz := env("PGTZ"); tz != "" {
		qry := dsn.Query()
		qry.Set("timezone", tz)
		dsn.RbwQuery = qry.Encode()
	}

	return dsn.String()
}
