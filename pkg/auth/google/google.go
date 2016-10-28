// Package google contains configuration for Google API client used
// by Sourcegraph code.
package google

import (
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Default is the default OAuth2 configuration for Google.
// Client ID/secret values are taken from the environment.
var Default = oauth2.Config{
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	Endpoint:     google.Endpoint,
}
