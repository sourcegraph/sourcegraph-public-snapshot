package hubspotutil

import (
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/hubspot"
)

// HubSpotHAPIKey is used by some requests to access their respective API endpoints
var HubSpotHAPIKey = env.Get("HUBSPOT_HAPI_KEY", "", "HubSpot HAPIkey for accessing certain HubSpot endpoints.")

var SurveyFormID = "a86bbac5-576d-4ca0-86c1-0c60837c3eab"

// SignupEventID is the HubSpot ID for signup events.
// HubSpot Events and IDs are all defined in HubSpot "Events" web console:
// https://app.hubspot.com/reports/2762526/events
var SignupEventID = "000001776813"

var client *hubspot.Client

func init() {
	// The HubSpot API key will only be available in the production sourcegraph.com environment.
	// Not having this key only restricts certain requests (e.g. GET requests to the Contacts API),
	// while others (e.g. POST requests to the Forms API) will still go through.
	client = hubspot.New("2762526", HubSpotHAPIKey)
}

// Client returns a hubspot client
func Client() *hubspot.Client {
	return client
}
