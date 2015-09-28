package conf

import (
	"log"
	"net/url"
	"os"
)

var (
	// PrivateRavenDSN is the private DSN for Raven (Sentry) error
	// reporting. It contains a secret and should not be revealed to
	// clients (e.g., JavaScript on client browsers).
	PrivateRavenDSN = os.Getenv("SG_RAVEN_DSN")

	// PublicRavenDSN is the public DSN for Raven (Sentry) error
	// reporting. It may be revealed to clients. The URL origins on
	// which it may be used are limited in the Sentry project's
	// settings (e.g., to "*.sourcegraph.com").
	PublicRavenDSN = stripURLPassword(PrivateRavenDSN)
)

func stripURLPassword(urlWithPassword string) string {
	if urlWithPassword == "" {
		return ""
	}
	u, err := url.Parse(urlWithPassword)
	if err != nil {
		log.Fatalf("Error parsing URL %q in stripURLPassword: %s.", urlWithPassword, err)
	}
	if u.User != nil {
		u.User = url.User(u.User.Username())
	}
	return u.String()
}
