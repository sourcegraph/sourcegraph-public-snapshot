package tracking

import (
	"log"

	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/hubspot"
	"github.com/sourcegraph/sourcegraph/pkg/hubspot/hubspotutil"
)

// SyncUser handles creating or syncing a user profile in HubSpot, and if provided,
// logs a user event.
func SyncUser(email string, eventID string, contactParams *hubspot.ContactProperties) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic in tracking.SyncUser: %s", err)
		}
	}()

	// If the user is in a dev or on-prem environment, don't do any tracking
	if env.Version == "dev" || !envvar.SourcegraphDotComMode() {
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

	c := hubspotutil.Client()

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
