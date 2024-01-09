package hubspotutil

import (
	"context"
	"log" //nolint:logging // TODO move all logging to sourcegraph/log

	"github.com/inconshreveable/log15" //nolint:logging // TODO move all logging to sourcegraph/log

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
// - analytics.behavioral_events.send
var HubSpotAccessToken = env.Get("HUBSPOT_ACCESS_TOKEN", "", "HubSpot access token for accessing certain HubSpot endpoints.")

// SurveyFormID is the ID for a satisfaction (NPS) survey.
var SurveyFormID = "ee042306-491a-4b06-bd9c-1181774dfda0"

// CodySurveyFormID is the ID for a Cody usage survey on dotcom users.
var CodySurveyFormID = "fadc00c7-8cf4-48dd-8502-c386b0311f5d"

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

// CodyClientInstalledV3EventID is the HubSpot ID for the new event which support custom properties.
var CodyClientInstalledV3EventID = "pe2762526_codyinstall"

// AppDownloadButtonClickedEventID is the HubSpot Event ID for when a user clicks on a button to download Cody App.
var AppDownloadButtonClickedEventID = "000019179879"

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

	// Update or create user contact information in HubSpot, and we want to sync the
	// contact independent of the request lifecycle.
	err := syncHubSpotContact(context.Background(), email, eventID, contactParams)
	if err != nil {
		log15.Warn("syncHubSpotContact: failed to create or update HubSpot contact", "source", "HubSpot", "error", err)
	}
}

// SyncUserWithV3Event handles creating or syncing a user profile in HubSpot, and if provided,
// logs a V3 custom event along with the event params.
func SyncUserWithV3Event(email, eventName string, contactParams *hubspot.ContactProperties, eventProperties any) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic in tracking.SyncUserWithV3Event: %s", err)
		}
	}()

	// If the user no API token present or on-prem environment, don't do any tracking
	if !HasAPIKey() || !envvar.SourcegraphDotComMode() {
		return
	}

	// Update or create user contact information in HubSpot, and we want to sync the
	// contact independent of the request lifecycle.
	err := syncHubSpotContact(context.Background(), email, "", contactParams)
	if err != nil {
		log15.Warn("syncHubSpotContact: failed to create or update HubSpot contact", "source", "HubSpot", "error", err)
	}

	// Log the V3 event
	if eventName != "" {
		c := Client()
		err = c.LogV3Event(email, eventName, eventProperties)
		if err != nil {
			log.Printf("LOGV3Event: failed to event %s", err)

		}
	}
}

func syncHubSpotContact(ctx context.Context, email, eventID string, contactParams *hubspot.ContactProperties) error {
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
		err = c.LogEvent(ctx, email, eventID, map[string]string{})
		if err != nil {
			return errors.Wrap(err, "LogEvent")
		}
	}

	return nil
}
