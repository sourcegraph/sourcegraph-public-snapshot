package hubspotutil

import (
	"log"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/hubspot"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// HubSpotAccessToken is used by some requests to access their respective API endpoints. This access
// token must have the following scopes:
//
// - crm.objects.contacts.write
// - timeline
// - forms
// - crm.objects.contacts.read
var HubSpotAccessToken = env.Get("HUBSPOT_ACCESS_TOKEN", "", "HubSpot access token for accessing certain HubSpot endpoints.")

// SurveyFormID is the ID for a satisfaction (NPS) survey.
var SurveyFormID = "ee042306-491a-4b06-bd9c-1181774dfda0"

// HappinessFeedbackFormID is the ID for a Happiness survey.
var HappinessFeedbackFormID = "417ec50b-39b4-41fa-a267-75da6f56a7cf"

// SignupEventID is the HubSpot ID for signup events.
// HubSpot Events and IDs are all defined in HubSpot "Events" web console:
// https://app.hubspot.com/reports/2762526/events
var SignupEventID = "000001776813"

// SelfHostedSiteInitEventID is the Hubstpot Event ID for when a new site is created in /site-admin/sites
var SelfHostedSiteInitEventID = "000010399089"

// CodyClientInstalledEventID is the HubSpot Event ID for when a user reports installing a Cody client.
var CodyClientInstalledEventID = "000018021981"

var client *hubspot.Client

// HasAPIKey returns true if a HubspotAPI key is present. A subset of requests require a HubSpot API key.
func HasAPIKey() bool {
	return HubSpotAccessToken != ""
}

func init() {
	// The HubSpot access token will only be available in the production sourcegraph.com environment.
	// Not having this access token only restricts certain requests (e.g. GET requests to the Contacts API),
	// while others (e.g. POST requests to the Forms API) will still go through.
	client = hubspot.New("2762526", HubSpotAccessToken)
}

// Client returns a hubspot client
func Client() *hubspot.Client {
	return client
}

// SyncUser handles creating or syncing a user profile in HubSpot, and if provided,
// logs a user event.
func SyncUser(email, eventID string, contactParams *hubspot.ContactProperties) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic in tracking.SyncUser: %s", err)
		}
	}()
	// If the user no API token present or on-prem environment, don't do any tracking
	if !HasAPIKey() || !envvar.SourcegraphDotComMode() {
		return
	}

	// Update or create user contact information in HubSpot
	err := syncHubSpotContact(email, eventID, contactParams)
	if err != nil {
		log15.Warn("syncHubSpotContact: failed to create or update HubSpot contact", "source", "HubSpot", "error", err)
	}
}

func syncHubSpotContact(email, eventID string, contactParams *hubspot.ContactProperties) error {
	if email == "" {
		return errors.New("user must have a valid email address")
	}

	// Generate a single set of user parameters for HubSpot
	if contactParams == nil {
		contactParams = &hubspot.ContactProperties{}
	}
	contactParams.UserID = email

	c := Client()

	// Create or update the contact
	_, err := c.CreateOrUpdateContact(email, contactParams)
	if err != nil {
		return err
	}

	// Log the user event
	if eventID != "" {
		err = c.LogEvent(email, eventID, map[string]string{})
		if err != nil {
			return errors.Wrap(err, "LogEvent")
		}
	}

	return nil
}
