// Package google contains configuration for Google API client used
// by Sourcegraph code.
package google

import (
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

var clientID = env.Get("GOOGLE_CLIENT_ID", "", "OAuth client ID for Google")
var clientSecret = env.Get("GOOGLE_CLIENT_SECRET", "", "OAuth client secret for Google")

// Default is the default OAuth2 configuration for Google.
// Client ID/secret values are taken from the environment.
var Default = oauth2.Config{
	ClientID:     clientID,
	ClientSecret: clientSecret,
	Endpoint:     google.Endpoint,
}
