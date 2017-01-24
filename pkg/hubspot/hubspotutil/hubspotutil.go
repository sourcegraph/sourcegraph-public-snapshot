package hubspotutil

import (
	"errors"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/hubspot"
)

// AfterPrivateCodeSignupFormID is a hardcoded identifier for the
// "After Private Code Signup" HubSpot form. To make changes to
// variables passed to this form, edit it in the HubSpot console.
const AfterPrivateCodeSignupFormID = "e9443a20-ef14-4858-971e-4925a3c6c45c"

var client *hubspot.Client

func init() {
	// TODO(dan): replace this with an env variable (e.g. see mailchimputil.go)
	portalID := "2762526"
	if portalID != "" {
		client = hubspot.New(portalID)
	}
}

// Client returns a hubspot client, or an error if HUBSPOT_KEY is not set.
func Client() (*hubspot.Client, error) {
	if client == nil {
		return nil, errors.New("hubspot: authorization key only available on production environments")
	}
	return client, nil
}
